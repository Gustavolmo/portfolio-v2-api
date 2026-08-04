package email

import (
	"fmt"
	"net/smtp"
	"os"
)

const HOST = "send.one.com"
const PORT = "587"
const FROM_EMAIL = "info@allverk.se"
const TO_EMAIL = "lmo.gustavo@gmail.com"

type Sender interface {
	Send(emailMsg []byte, destinataries []string) error
}

type SMTPSender struct{}
func (SMTPSender) Send(emailMsg []byte, destinataries []string) (error) {
	FROM_EMAIL_PW := os.Getenv("ALLVERK_EMAIL_PASSWORD")
	if FROM_EMAIL_PW == "" {
		return fmt.Errorf("Empty pw")
	}

	auth := smtp.PlainAuth("", FROM_EMAIL, FROM_EMAIL_PW, HOST)
	smtpAddr := HOST + ":" + PORT

	if err := smtp.SendMail(smtpAddr, auth, FROM_EMAIL, destinataries, emailMsg); err != nil {
		return fmt.Errorf("SMTP fail to send %w", err)
	}

	return nil
}
