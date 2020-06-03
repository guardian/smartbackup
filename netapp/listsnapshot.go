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

func ListSnapshots(config *NetappConfig, volume *NetappEntity) (*ListSnapshotsResponse, error) {
	httpClient := &http.Client{}

	log.Printf("INFO Listing snapshots from %s", volume.Name)

	url := url2.URL{
		Scheme:   "https",
		Host:     config.Host,
		Path:     fmt.Sprintf("/api/storage/volumes/%s/snapshots", volume.UUID),
		RawQuery: fmt.Sprintf("volume.uuid=%s", url2.QueryEscape(volume.UUID)),
	}

	log.Printf("DEBUG URL is %s", url.String())

	req, _ := http.NewRequest("GET", url.String(), nil)
	req.SetBasicAuth(config.User, config.Passwd)

	response, httpErr := httpClient.Do(req)

	if httpErr != nil {
		log.Printf("ERROR Could not make HTTP request: %s", httpErr)
		return nil, httpErr
	}

	bodyContent, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Print("ERROR Could not read in response from server: ", readErr)
		return nil, readErr
	}

	if response.StatusCode == 200 {
		var response ListSnapshotsResponse
		marshalErr := json.Unmarshal(bodyContent, &response)
		if marshalErr != nil {
			log.Printf("ERROR Offending content was '%s'", string(bodyContent))
			log.Print("ERROR Could not parse response from server: ", marshalErr)
			return nil, marshalErr
		}
		return &response, nil
	} else {
		var response ErrorResponse
		marshalErr := json.Unmarshal(bodyContent, &response)
		if marshalErr != nil {
			log.Printf("ERROR Offending content was '%s'", string(bodyContent))
			log.Print("ERROR Could not parse response from server: ", marshalErr)
			return nil, marshalErr
		}
		log.Print("ERROR Could not list snapshots: ", response.Error.Message)
		return nil, errors.New(response.Error.Message)
	}
}
