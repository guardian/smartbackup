package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/fredex42/smartbackup/mail"
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
		log.Printf("Could not create snapshot: %s", err)
		return err
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

/**
generate a message for the given ResolvedBackupTarget based on templates for subject and bodytext, and send it
Returns an error if the operation fails or nil if it succeeds
 */
func GenerateAndSend(messenger *Messenger, config *ConfigData, target *ResolvedBackupTarget, subjectTemplate string, bodytextTemplate string, error string) error {
	subject, body, msgErr := messenger.GenerateMessage(target, subjectTemplate, bodytextTemplate, error)
	if msgErr != nil {
		log.Printf("ERROR: Could not generate error message: %s", msgErr)
		//Hmm, should this be fatal??
		return msgErr
	}
	sendErr := mail.SendMail(&config.SMTP, subject, body)
	if sendErr != nil {
		log.Printf("ERROR: Could not send error email: %s", sendErr)
		return sendErr
	}
	return nil
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	configFilePtr := flag.String("config", "smartbackup.yaml", "YAML config file")
	allowInvalid := flag.Bool("continue", true, "Don't terminate if any config is invalid but continue to work with the ones that are")
	testSmtp := flag.Bool("test-smtp", false, "Send a test message as if a backup had failed")
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

	messenger, msgErr := NewMessenger()
	if msgErr != nil {
		log.Fatalf("Could not initialise messaging: %s", msgErr)
	}

	if *testSmtp {
		log.Printf("Sending test message to %s", spew.Sdump(config.SMTP.SendTo))
		fakeBackupTarget := &ResolvedBackupTarget{
			Netapp:   &netapp.NetappConfig{},
			Database: &postgres.DatabaseConfig{},
			VolumeId: "",
		}
		sendErr := GenerateAndSend(messenger, config, fakeBackupTarget, "Test message from smartbackup at {time}", "This is a test message from smartbackup, if you can read it then SMTP is working correctly", "")
		if sendErr != nil {
			log.Fatal("Could not send test message: ", sendErr)
		}
		log.Fatal("Successfully sent message")
	}

	for _, target := range resolvedTargets {
		dateString := time.Now().Format(time.RFC3339)
		backupName := fmt.Sprintf("%s_%s", target.Database.Name, dateString)
		log.Printf("Database backup name is %s", backupName)
		checkpoint, err := postgres.StartBackup(target.Database, backupName)
		if err != nil {
			log.Printf("ERROR: Database %s did not quiesce! %s", target.Database.Name, err)
			sendErr := GenerateAndSend(messenger, config, target, FailureSubjectTemplate, FailureMessage, err.Error())
			if sendErr != nil {
				log.Printf("ERROR: Could not send error message: %s", sendErr)
			}
			break
		}
		log.Printf("Database quiesced, consistent state ID is %s", checkpoint)

		targetVolumeEntity := &netapp.NetappEntity{UUID: target.VolumeId}
		snapshotErr := SyncPerformSnapshot(target.Netapp, targetVolumeEntity, backupName)

		if snapshotErr != nil {
			log.Printf("ERROR: Could not perform snapshot! %s", snapshotErr)
			sendErr := GenerateAndSend(messenger, config, target, FailureSubjectTemplate, FailureMessage, snapshotErr.Error())
			if sendErr != nil {
				log.Printf("ERROR: Could not send error message: %s", sendErr)
			}
		}

		endpoint, err := postgres.StopBackup(target.Database)
		if err != nil {
			log.Printf("ERROR: Database %s did not unquiesce! %s", target.Database.Name, err)
			sendErr := GenerateAndSend(messenger, config, target, FailureSubjectTemplate, FailureMessage, err.Error())
			if sendErr != nil {
				log.Printf("ERROR: Could not send error message: %s", sendErr)
			}
			break
		}

		log.Printf("Database unquiesced, completed state ID is %s", endpoint)
		if config.SMTP.AlwaysSend {
			sendErr := GenerateAndSend(messenger, config, target, SuccessSubjectTemplate, SuccessMessage, "")
			if sendErr != nil {
				log.Printf("ERROR: Could not send success message: %s", sendErr)
			}
		}
	}

}
