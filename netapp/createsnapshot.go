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
	"time"
)

/**
Extracts a Job Data payload from the given http response
Parameters: pointer to an http.Response object that contains the response
Returns:
- a pointer to a CreateSnapshotResponse object that contains the returned job data if it parses properly
- an error if we can't obtain or parse the information properly
*/
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

/**
tells the NetApp REST api to create a snapshot for the given volume on the given appliance with a given name
Parameters:
- pointer to a NetappConfig object that describes the appliance to target
- pointer to a NetappEntity object that describes the volume to target
- string that contains the name for the given snapshot
Returns:
- a pointer to a CreateSnapshotResponse object if the operation succeeds, nil if it fails
- an error object if the operation fails, nil if it succeeds
*/
func CreateSnapshot(config *NetappConfig, volume *NetappEntity, snapshotName string, expiryTime time.Time) (*CreateSnapshotResponse, error) {
	httpClient := &http.Client{}

	log.Printf("Starting create snapshot operation on %s (%s)", volume.Name, volume.UUID)

	expiryTimeString := expiryTime.Format(time.RFC3339)
	requestContent := map[string]interface{}{
		"name":        snapshotName,
		"expiry_time": expiryTimeString,
	}

	requestString, marshalErr := json.Marshal(requestContent)
	if marshalErr != nil {
		log.Printf("Could not format request for server: %s", marshalErr)
		return nil, marshalErr
	}

	log.Printf("DEBUG: request is %s", string(requestString))
	url := url2.URL{
		Scheme: "https",
		Host:   config.Host,
		Path:   fmt.Sprintf("/api/storage/volumes/%s/snapshots", volume.UUID),
	}

	urlString := url.String()
	log.Printf("DEBUG: making request to %s", urlString)
	req, _ := http.NewRequest("POST", urlString, bytes.NewReader(requestString))
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(config.User, config.Passwd)

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
	if problem != nil {
		log.Printf("Server returned success but we could not understand the response: %s", problem)
		return nil, problem
	}
	return snapshotResponse, nil
}
