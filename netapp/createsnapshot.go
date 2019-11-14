package netapp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
)

func GetJobResponseData(response *http.Response) (*CreateSnapshotResponse, error) {
	var jobData CreateSnapshotResponse
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

func CreateSnapshot(config *NetappConfig, volume *NetappEntity, svm *NetappEntity, snapshotName string) (*CreateSnapshotResponse, error) {
	httpClient := &http.Client{}

	log.Printf("Starting create snapshot operation on %s (%s)", volume.Name, volume.UUID)

	requestContent := map[string]interface{} {"name": snapshotName}
	requestString, marshalErr := json.Marshal(requestContent)
	if marshalErr != nil {
		log.Printf("Could not format request for server: %s", marshalErr)
		return nil, marshalErr
	}

	log.Printf("DEBUG: request is %s", string(requestString))
	url := url2.URL{
		Scheme: "https",
		Host: config.Host,
		Path: fmt.Sprintf("/api/storage/volumes/%s/snapshots", volume.UUID),
	}

	urlString := url.String()
	log.Printf("DEBUG: making request to %s", urlString)
	req, _ := http.NewRequest("POST",urlString,bytes.NewReader(requestString))
	req.Header.Add("Content-Type","application/json")
	req.SetBasicAuth(config.User,config.Passwd)

	response, httpErr := httpClient.Do(req)

	if httpErr != nil {
		log.Printf("Could not make HTTP request: %s", httpErr)
		return nil, httpErr
	}

	if response.StatusCode != 202 {
		log.Printf("Server responded with an error code: %d", response.StatusCode)
		errorResponse, problem := GetErrorData(response)
		if problem != nil {
			return nil, problem
		}
		log.Printf("Server could not create snapshot: %s %s", errorResponse.Error.Code, errorResponse.Error.Message)
		return nil, errors.New("Server error")
	}

	snapshotResponse, problem := GetJobResponseData(response)
	if problem!=nil {
		log.Printf("Server returned success but we could not understand the response: %s", problem)
		return nil, problem
	}
	return snapshotResponse, nil
}
