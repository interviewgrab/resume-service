package email

import (
	"crypto/tls"
	"errors"
	"os"

	"github.com/Shopify/gomail"
)

type EmailClient struct {
	from       string
	sendCloser *gomail.SendCloser
}

const (
	senderEmailKey = "SENDER_EMAIL"
	senderPassKey  = "SENDER_PASS"
)

func NewClient() (*EmailClient, error) {
	email, password := os.Getenv(senderEmailKey), os.Getenv(senderPassKey)
	d := gomail.NewDialer("smtp.gmail.com", 587, email, password)
	sendCloser, err := d.Dial()
	if err != nil {
		return nil, errors.New("Unable to setup email-client, check your email / password")
	}
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return &EmailClient{from: email, sendCloser: &sendCloser}, nil
}

func (c *EmailClient) SendMail(to string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Welcome to resume-service!")
	m.SetBody("text/html", "<h1>Welcome to resume service</h1><br/><p>Your OTP is 123456</p>")

	return gomail.Send(*c.sendCloser, m)
}
