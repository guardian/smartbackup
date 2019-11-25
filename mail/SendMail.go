package mail

import (
	"fmt"
	"io"
	"log"
	"net/smtp"
	"os"
	"strings"
)

/**
converts the provided string-string map of headers into a single block of text for the email
body, and returns an io.Reader for it to be streamed into the email content
Parameters: pointer to a string-string map of headers
Returns: an io.Reader that renders out the text of the headers
 */
func HeaderBodies(headers *map[string]string) io.Reader {
	var headersString string

	for k, v := range *headers {
		headersString = headersString + fmt.Sprintf("%s: %s\n", k, v)
	}
	headersString += "\n"

	return strings.NewReader(headersString)
}

/**
helper function to take the array of sender addresses and convert it to a single comma-delimited field
Parameters: pointer to an array of strings
Returns: a string containing the content of all of the array separated by commas
 */
func concatenateSenders(senders *[]string) string {
	var senderString string

	for _, entry := range *senders {
		senderString += entry + ","
	}
	return senderString
}

/**
main function to send an email with the given subject and content.
Use the messenger object to handle templating values into subject and body.
returns an error if there is a problem or nil if it works
 */
func SendMail(mailConfig *MailConfig, subject string, body io.Reader) error {
	smtpClient, err := smtp.Dial(mailConfig.SMTPServer)
	if err != nil {
		log.Printf("Could not connect to SMTP server '%s': %s", mailConfig.SMTPServer, err)
		return err
	}

	defer smtpClient.Close()

	if mailConfig.SMTPUser != "" && mailConfig.SMTPPasswd != "" {
		auth := smtp.CRAMMD5Auth(mailConfig.SMTPUser, mailConfig.SMTPPasswd)
		authErr := smtpClient.Auth(auth)
		log.Printf("Could not authenticate to SMTP server '%s' with provided credentials: %s", mailConfig.SMTPServer, authErr)
		return authErr
	}

	hostName, hstErr := os.Hostname()
	if hstErr != nil {
		log.Printf("WARNING - could not determine hostname: %s. Falling back to 'localhost'", hstErr)
		hostName = "localhost"
	}

	heloErr := smtpClient.Hello(hostName)
	if heloErr != nil {
		log.Printf("Could not send HELLO to server: %s", heloErr)
		return heloErr
	}

	mailErr := smtpClient.Mail(mailConfig.FromAddress)
	if mailErr != nil {
		log.Printf("Could not send MAIL to server: %s", mailErr)
		return mailErr
	}

	for _, sendTo := range mailConfig.SendTo {
		rcptErr := smtpClient.Rcpt(sendTo)
		if rcptErr != nil {
			log.Printf("WARNING - could not add recipient %s: %s", sendTo, rcptErr)
			return rcptErr
		}
	}

	headers := map[string]string{
		"From":    mailConfig.FromAddress,
		"To":      concatenateSenders(&mailConfig.SendTo),
		"Subject": subject,
	}

	headerReader := HeaderBodies(&headers)

	writer, dataErr := smtpClient.Data()
	if dataErr != nil {
		log.Printf("Could not initiate")
	}
	defer writer.Close()

	_, headerErr := io.Copy(writer, headerReader)
	if headerErr != nil {
		log.Printf("Could not write headers: %s", headerErr)
		return headerErr
	}

	_, contentErr := io.Copy(writer, body)
	if contentErr != nil {
		log.Printf("Could not write body: %s", contentErr)
		return contentErr
	}

	return nil //we are done now, let the deferred writer.Close and smtpClient.Close calls clean up
}
