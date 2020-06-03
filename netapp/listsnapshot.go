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

/**
calls out to the NetApp API to obtain details about the given snapshot.
this is intended to be called internally by ListSnapshots and it therefore expects a pointer
to an existing http.Client in order to re-use connections.
the `hrefPath` is the path-only specifier that is usually obtained from the "_links/self" keys in the json
ListSnapshots response.

returns either a SnapshotEntry struct containing the full information as provided by the REST API or an error
*/
func SnapshotDetail(config *NetappConfig, httpClient *http.Client, hrefPath string) (SnapshotEntry, error) {
	hrefUrl := url2.URL{
		Scheme: "https",
		Host:   config.Host,
		Path:   hrefPath,
	}
	req, _ := http.NewRequest("GET", hrefUrl.String(), nil)
	req.SetBasicAuth(config.User, config.Passwd)

	response, httpErr := httpClient.Do(req)

	if httpErr != nil {
		log.Printf("ERROR Could not make http request: %s", httpErr)
		return SnapshotEntry{}, httpErr
	}

	defer response.Body.Close()
	bodyContent, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Print("ERROR Could not read response from server: ", readErr)
		return SnapshotEntry{}, readErr
	}

	var entry SnapshotEntry
	unmarshalErr := json.Unmarshal(bodyContent, &entry)
	if unmarshalErr != nil {
		log.Print("ERROR Could not parse returned json: ", unmarshalErr)
		return SnapshotEntry{}, unmarshalErr
	}

	return entry, nil
}

/**
calls out to the NetApp API to list all of the snapshots on the given volume.
the provided credentials must have at least "readonly" access to "/api/volumes":
   ```create -vserver my-svm -role SmartBackupRest -api /api/storage/volume -access readonly```

*/
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
		var output ListSnapshotsResponse
		marshalErr := json.Unmarshal(bodyContent, &output)
		if marshalErr != nil {
			log.Printf("ERROR Offending content was '%s'", string(bodyContent))
			log.Print("ERROR Could not parse response from server: ", marshalErr)
			return nil, marshalErr
		}

		final := ListSnapshotsResponse{
			Records:      make([]SnapshotEntry, output.RecordsCount),
			RecordsCount: output.RecordsCount,
		}
		for i, snap := range output.Records {
			if link, hasLink := snap.Links["self"]; hasLink {
				var err error
				//the list response does not have ctime, expiry time etc. so we must do an explicit lookup
				//on each item
				final.Records[i], err = SnapshotDetail(config, httpClient, link.HRef)
				if err != nil {
					//if the lookup fails then keep the previous record.
					final.Records[i] = output.Records[i]
				}
			} else {
				//if we have no "self" link we can't look up, keep the previous record
				final.Records[i] = output.Records[i]
			}
		}

		return &final, nil
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
