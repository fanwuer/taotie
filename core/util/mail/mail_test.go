package mail

import (
	"fmt"
	"testing"
)

func TestMail_Sent(t *testing.T) {
	s := Sender{}
	s.Host = "smtp-mail.outlook.com"
	s.Port = 587
	s.Email = "gdccmcm14@live.com"
	s.Password = "ddd"

	m := new(Message)
	m.Sender = s
	m.To = "gao.yu@husky.neu.edu"
	m.ToName = "user"
	m.Subject = "register"
	m.Body = "ddddddddd"

	err := m.Sent()
	if err != nil {
		fmt.Println(err.Error())
	}
}
