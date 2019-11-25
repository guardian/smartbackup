package pagerduty

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const pagerDutyUrl = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

/**
creates a new CreateIncidentRequest object
*/
func NewIncident(config *PagerDutyConfig, incidentKey string, details map[string]string, description string) *CreateIncidentRequest {
	i := &CreateIncidentRequest{
		ServiceKey:  config.ServiceKey,
		EventType:   "trigger",
		IncidentKey: incidentKey,
		Description: description,
		Client:      "SmartBackup",
		Details:     details,
	}

	return i
}

/**
send an alert to PagerDuty
Parameters: - pointer to a CreateIncidentRequest with the incident to request
Returns: - an error if the operation failed or nil if it succeeded
*/
func SendAlert(req *CreateIncidentRequest) error {
	jsonToSend, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		log.Printf("Could not create indicent JSON: %s", marshalErr)
		return marshalErr
	}

	log.Print("DEBUG: json to send is ", string(jsonToSend))
	response, sendErr := http.Post(pagerDutyUrl, "application/json", bytes.NewReader(jsonToSend))
	if sendErr != nil {
		log.Printf("Could not send pager duty alert: %s", sendErr)
		return sendErr
	}

	responseBodyBytes, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Printf("Could not read server response: %s", readErr)
	}
	responseBodyText := string(responseBodyBytes)

	log.Printf("Server said: %s", responseBodyText)
	switch response.StatusCode {
	case 200:
		log.Printf("Sent content successfully")
		return nil
	case 400:
		log.Printf("Server claimed an invalid event format, this is a code bug, please report at https://github.com/fredex42/smartbackup along with full logging output")
		return errors.New("invalid event format")
	case 403:
		log.Printf("Server claimed permission denied, could be due to rate-limiting or an incorrect key")
		return errors.New("permission denied")
	}
	return errors.New(fmt.Sprintf("Unexpected server response: %d", response.StatusCode))
}
