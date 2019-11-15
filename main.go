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

func SyncPerformSnapshot(config *netapp.NetappConfig, targetVolume *netapp.NetappEntity, snapshotName string) error {
	response, err := netapp.CreateSnapshot(config, targetVolume, snapshotName)
	if err != nil {
		log.Fatalf("Could not create snapshot: %s", err)
	}

	log.Printf("Created job with id %s", response.Job.UUID)

	time.Sleep(1 * time.Second)
	for {
		jobData, readErr := netapp.GetJob(config, response.Job.UUID)
		if readErr != nil {
			errMsg := fmt.Sprintf("Could not get job data: %s", readErr)
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
	allowInvalid := flag.Bool("continue", true, "Don't terminate if any config is invalid but continue to work with the ones that are")
	flag.Parse()

	config, configErr := GetConfig(*configFilePtr)
	if configErr != nil {
		log.Fatalf("Could not load config: %s", configErr)
	}

	resolvedTargets, unresolvedTargets := config.ResolveBackupTargets()

	if len(unresolvedTargets) > 0 {
		log.Printf("WARNING: The following database definitions are not valid, please check that they refer to correct database and netapp entries")
		for _, entry := range unresolvedTargets {
			log.Printf("\t%s", entry)
		}
		if *allowInvalid == false {
			log.Fatalf("Exiting as --continue is set to false")
		}
	}

	if len(resolvedTargets) == 0 {
		log.Fatalf("ERROR: There are no valid configurations to back up!")
	}

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

		targetVolumeEntity := &netapp.NetappEntity{UUID: target.VolumeId}
		snapshotErr := SyncPerformSnapshot(target.Netapp, targetVolumeEntity, backupName)

		if snapshotErr != nil {
			log.Printf("ERROR: Could not perform snapshot! %s", snapshotErr)
		}

		endpoint, err := postgres.StopBackup(target.Database)
		if err != nil {
			log.Printf("ERROR: Database %s did not unquiesce! %s", target.Database.Name, err)
			break
		}
		log.Printf("Databse unquiesced, completed state ID is %s", endpoint)
	}

}
