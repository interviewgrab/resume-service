package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/Shopify/gomail"
)

type EmailClient struct {
	d *gomail.Dialer
}

const (
	senderEmailKey = "SENDER_EMAIL"
	senderPassKey  = "SENDER_PASS"
)

func NewClient() (*EmailClient, error) {
	email, password := os.Getenv(senderEmailKey), os.Getenv(senderPassKey)
	d := gomail.NewDialer("smtp.gmail.com", 587, email, password)
	_, err := d.Dial()
	if err != nil {
		return nil, errors.Join(errors.New("Unable to setup email-client, check your email / password"), err)
	}
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return &EmailClient{d: d}, nil
}

func (c *EmailClient) SendMail(to string, otp string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.d.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Welcome to resume-service!")
	m.SetBody("text/html", fmt.Sprintf("<h1>Welcome to resume service</h1><br/><p>Your OTP is %s</p>", otp))

	return c.d.DialAndSend(m)
}
