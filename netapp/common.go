package netapp

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

/**
parse the http response's content into an ErrorResponse document
 */
func GetErrorData(response *http.Response) (*ErrorResponse, error) {
	var errorData ErrorResponse
	responseBytes, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Printf("Could not read bytes: %s", readErr)
		return nil, readErr
	}
	unmarshalErr := json.Unmarshal(responseBytes, &errorData)
	if unmarshalErr != nil {
		log.Printf("Server sent something we can't understand: %s", string(responseBytes))
		return nil, unmarshalErr
	}
	return &errorData, nil
}