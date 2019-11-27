package mail

import (
	"crypto/tls"
	"errors"
	"github.com/go-gomail/gomail"
)

var Debug = false

type Sender struct {
	//Bcc      string
	//BccName  string
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
	Subject  string `yaml:"subject"`
	Body     string `yaml:"body"`
}

type Message struct {
	To     string
	ToName string
	Sender
}

func (mm *Message) Sent() error {
	if Debug {
		return errors.New("can not send email")
	}
	m := gomail.NewMessage()
	m.SetAddressHeader("From", mm.Email, "TaoTie")
	m.SetAddressHeader("To", mm.To, mm.ToName)
	m.SetHeader("Subject", mm.Subject)

	//m.SetHeader("Bcc",
	//	m.FormatAddress(mm.Bcc, mm.BccName))

	m.SetBody("text/html", mm.Body)

	d := gomail.NewDialer(mm.Host, mm.Port, mm.Email, mm.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
