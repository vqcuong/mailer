package sender

import (
	"net/smtp"
	"strconv"
)

func SendEmail(username, password, server string, port int, to, subject, body string) error {
	auth := smtp.PlainAuth("", username, password, server)

	msg := "From: " + username + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body

	err := smtp.SendMail(server+":"+strconv.Itoa(port), auth, username, []string{to}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}
