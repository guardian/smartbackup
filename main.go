package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/postgres"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func GetConfig(filePath string) (*ConfigData, error) {
	file, openErr := os.Open(filePath)
	if openErr != nil {
		return nil, openErr
	}

	bytes, readErr := ioutil.ReadAll(file)
	if readErr != nil {
		return nil, readErr
	}

	var config ConfigData
	marshalErr := yaml.Unmarshal(bytes, &config)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return &config, nil
}

func SyncPerformSnapshot(config *netapp.NetappConfig, targetVolume *netapp.NetappEntity, targetSVM *netapp.NetappEntity, snapshotName string) error {
	response, err := netapp.CreateSnapshot(config, targetVolume, targetSVM, snapshotName)
	if err != nil {
		log.Fatalf("Could not create snapshot: %s", err)
	}

	log.Printf("Created job with id %s", response.Job.UUID)

	time.Sleep(1 * time.Second)
	for {
		jobData, readErr := netapp.GetJob(config, response.Job.UUID)
		if readErr != nil {
			errMsg := fmt.Sprintf("Could not get job data: ", readErr)
			return errors.New(errMsg)
		}
		if jobData.State == "success" {
			log.Printf("Job succeeded at %s", jobData.EndTime)
			break
		}
		if jobData.State == "failure" {
			log.Printf("Job failed at %s: %d %s", jobData.EndTime, jobData.Code, jobData.Message)
			break
		}
		log.Printf("Waiting for snapshot job to complete, current status is %s", jobData.State)
		time.Sleep(5 * time.Second)
	}
	return nil
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	configFilePtr := flag.String("config", "smartbackup.yaml", "YAML config file")
	flag.Parse()

	config, configErr := GetConfig(*configFilePtr)
	if configErr != nil {
		log.Fatalf("Could not load config: %s", configErr)
	}

	resolvedTargets := config.ResolveBackupTargets()

	for _, target := range resolvedTargets {
		dateString := time.Now().Format(time.RFC3339)
		backupName := fmt.Sprintf("%s_%s", target.Database.Name, dateString)
		log.Printf("Database backup name is %s", backupName)
		checkpoint, err := postgres.StartBackup(target.Database, backupName)
		if err != nil {
			log.Printf("ERROR: Database %s did not quiesce! %s", target.Database.Name, err)
			break
		}
		log.Printf("Database quiesced, consistent state ID is %s", checkpoint)
		time.Sleep(10 * time.Second)

		endpoint, err := postgres.StopBackup(target.Database)
		if err != nil {
			log.Printf("ERROR: Database %s did not unquiesce! %s", target.Database.Name, err)
			break
		}
		log.Printf("Databse unquiesced, completed state ID is %s", endpoint)
	}

}
