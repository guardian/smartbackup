package netapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
)

func GetJobData(response *http.Response) (*NetappJob, error) {
	var jobData NetappJob
	responseBytes, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Printf("Could not read bytes: %s", readErr)
		return nil, readErr
	}
	unmarshalErr := json.Unmarshal(responseBytes, &jobData)
	if unmarshalErr != nil {
		log.Printf("Server sent something we can't understand: %s", string(responseBytes))
		return nil, unmarshalErr
	}
	return &jobData, nil
}

func GetJob(config *NetappConfig, jobId string) (*NetappJob, error) {
	httpClient := &http.Client{}

	uri := url2.URL{
		Scheme: "https",
		Host: config.Host,
		Path: fmt.Sprintf("/api/cluster/jobs/%s", jobId),
	}

	req, _ := http.NewRequest("GET",uri.String(),nil)
	req.SetBasicAuth(config.User, config.Passwd)

	response, httpErr := httpClient.Do(req)
	if httpErr != nil {
		log.Printf("Could not make http request: %s", httpErr)
		return nil, httpErr
	}

	if response.StatusCode==200 {
		return GetJobData(response)
	} else {
		log.Printf("Server returned %d getting job", response.StatusCode)
		content, err := GetErrorData(response)
		if err != nil {
			log.Printf("Could not interpret server response: %s", err)
			return nil, err
		}
		log.Printf("Server said %s", content.Error.Message)
		return nil, errors.New(content.Error.Message)
	}
}