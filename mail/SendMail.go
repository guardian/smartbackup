package mail

import (
	"fmt"
	"io"
	"log"
	"net/smtp"
	"os"
	"strings"
)

func HeaderBodies(headers *map[string]string) io.Reader {
	var headersString string

	for k, v := range *headers {
		headersString = headersString + fmt.Sprintf("%s: %s\n", k, v)
	}
	headersString += "\n"

	return strings.NewReader(headersString)
}

func concatenateSenders(senders *[]string) string {
	var senderString string

	for _, entry := range *senders {
		senderString += entry + ","
	}
	return senderString
}

func SendMail(mailConfig *MailConfig, subject string, body *io.Reader) error {
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

	_, contentErr := io.Copy(writer, *body)
	if contentErr != nil {
		log.Printf("Could not write body: %s", contentErr)
		return contentErr
	}

	return nil //we are done now, let the deferred writer.Close and smtpClient.Close calls clean up
}
