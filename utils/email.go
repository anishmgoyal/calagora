package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"

	"github.com/anishmgoyal/calagora/constants"
)

// ses-smtp-user.20161103-222837
const (
	emailThreadCount = 5
	emailChanSize    = 100
)

// Email contains fields necessary to define an email
type Email struct {
	From          string   `json:"from"`
	To            []string `json:"to"`
	Subject       string   `json:"subject"`
	PlainText     string   `json:"plain_text"`
	FormattedText string   `json:"formatted_text"`
}

// StartEmailService spawns a few threads to handle emails, and returns
// the channel through which emails can be queued up for delivery
func StartEmailService() chan *Email {
	ch := make(chan *Email, emailChanSize)
	for i := 0; i < emailThreadCount; i++ {
		go SendMail(ch)
	}
	return ch
}

// SendMail sends an email
func SendMail(ch chan *Email) {
	for {
		email := <-ch

		var buff bytes.Buffer
		var boundaryStr = boundary()

		buff.WriteString("From: ")
		buff.WriteString(email.From)
		buff.WriteString("\r\n")

		buff.WriteString("To: ")
		buff.WriteString(strings.Join(email.To, ","))
		buff.WriteString("\r\n")

		buff.WriteString("Subject: ")
		buff.WriteString(email.Subject)
		buff.WriteString("\r\n")

		buff.WriteString("Content-Type: multipart/alternative; boundary=")
		buff.WriteString(boundaryStr)
		buff.WriteString("\r\n\r\n")

		buff.WriteString("--" + boundaryStr)

		buff.WriteString("\r\n")
		buff.WriteString("Content-Type: text/plain; charset=UTF-8")
		buff.WriteString("\r\n\r\n")

		buff.WriteString(email.PlainText)
		buff.WriteString("\r\n\r\n")

		buff.WriteString("--" + boundaryStr)

		buff.WriteString("\r\n")
		buff.WriteString("Content-Type: text/html; charset=UTF-8")
		buff.WriteString("\r\n\r\n")

		buff.WriteString(email.FormattedText)
		buff.WriteString("\r\n\r\n")

		buff.WriteString("--" + boundaryStr + "--")

		if constants.DoSendEmails {
			auth := smtp.PlainAuth(
				"",
				constants.SMTPAuthUser,
				constants.SMTPAuthPassword,
				constants.SMTPHostname,
			)

			rfcCompliantFrom := email.From
			minIndex := strings.Index(rfcCompliantFrom, "<")
			maxIndex := strings.Index(rfcCompliantFrom, ">")
			if maxIndex > minIndex {
				rfcCompliantFrom = rfcCompliantFrom[minIndex+1 : maxIndex]
			}

			err := smtp.SendMail(
				constants.SMTPHostname+":"+constants.SMTPPort,
				auth,
				rfcCompliantFrom,
				email.To,
				buff.Bytes(),
			)
			if err != nil {
				fmt.Println("Failed to send an email:", err.Error())
				fmt.Print("Email was to: ")
				fmt.Println(email.To)
			}
		} else {
			fmt.Println("To: " + strings.Join(email.To, "; "))
			fmt.Println("From: " + email.From)
			fmt.Println("Subject: " + email.Subject)
			fmt.Println(email.PlainText)
		}
	}
}

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

func boundary() string {
	length := 20 + rand.Intn(10)
	buff := make([]byte, length)
	for i := 0; i < length; i++ {
		buff[i] = characters[rand.Intn(len(characters))]
	}
	return string(buff)
}
