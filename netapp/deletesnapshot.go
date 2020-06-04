package netapp

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	url2 "net/url"
)

/**
asks the netapp API to delete the given snapshot from the given volume.
this requires write access to volumes, e.g.
	```security login rest-role create -vserver my-svm -role SmartBackupRest -api /api/storage/volumes -access all```
*/
func DeleteSnapshot(config *NetappConfig, volume *NetappEntity, snapshotId string) error {
	httpClient := &http.Client{}

	log.Printf("INFO Deleting snapshot %s from volume %s", snapshotId, volume.Name)

	url := url2.URL{
		Scheme: "https",
		Host:   config.Host,
		Path:   fmt.Sprintf("/api/storage/volumes/%s/snapshots/%s", volume.UUID, snapshotId),
	}

	req, _ := http.NewRequest("DELETE", url.String(), nil)
	req.SetBasicAuth(config.User, config.Passwd)

	response, httpErr := httpClient.Do(req)

	if httpErr != nil {
		log.Printf("Could not make HTTP request: %s", httpErr)
		return httpErr
	}

	defer response.Body.Close()
	if response.StatusCode != 202 {
		log.Printf("ERROR Could not delete snapshot, server returned %d", response.StatusCode)
		return errors.New("could not delete snapshot")
	}
	return nil
}
