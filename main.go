package main

import (
	"crypto/tls"
	"flag"
	"github.com/fredex42/smartbackup/netapp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"gopkg.in/yaml.v2"
	"time"
)



type ConfigData struct {
	Netapp netapp.NetappConfig `yaml:"netapp"`
}

func GetConfig(filePath string) (*ConfigData, error){
	file, openErr := os.Open(filePath)
	if openErr!=nil {
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

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	volumeUuidPtr := flag.String("volumeid","","UUID of the volume to snapshot")
	svmName := flag.String("svm","","Storage Virtual Machine that contains the volume")
	configFilePtr := flag.String("config","smartbackup.yaml","YAML config file")
	snapshotName := flag.String("name","test","Name of the snapshot to create")

	flag.Parse()

	config, configErr := GetConfig(*configFilePtr)
	if configErr != nil {
		log.Fatalf("Could not load config: %s", configErr)
	}

	targetSVM := &netapp.NetappEntity{
		Name: *svmName,
	}
	targetVolume := &netapp.NetappEntity{
		UUID: *volumeUuidPtr,
	}

	response, err := netapp.CreateSnapshot(&config.Netapp,targetVolume, targetSVM, *snapshotName)
	if err != nil {
		log.Fatalf("Could not create snapshot: %s", err)
	}

	log.Printf("Created job with id %s", response.Job.UUID)

	for {
		jobData, readErr := netapp.GetJob(&config.Netapp, response.Job.UUID)
		if readErr != nil {
			log.Fatalf("Could not get job data: ", readErr)
		}
		if(jobData.State=="success"){
			log.Printf("Job succeeded at %s", jobData.EndTime)
			break
		}
		if jobData.State=="failure" {
			log.Printf("Job failed at %s: %d %s", jobData.EndTime, jobData.Code, jobData.Message)
			break
		}
		time.Sleep(5*time.Second)
	}
}
