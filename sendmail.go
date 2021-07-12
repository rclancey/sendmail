package sendmail

import (
	"fmt"
	"net/smtp"
)

type SMTPClient struct {
	username *string
	password *string
	host string
	port int
}

func NewSMTPClient(host string, port int) *SMTPClient {
	return &SMTPClient{
		host: host,
		port: port,
	}
}

func (c *SMTPClient) SetAuth(username, password string) {
	c.username = &username
	c.password = &password
}

func (c *SMTPClient) hostAndPort() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}

func (c *SMTPClient) auth() smtp.Auth {
	if c.username == nil || c.password == nil {
		return nil
	}
	return smtp.PlainAuth("", *c.username, *c.password, c.host)
}

func (c *SMTPClient) Sendmail(msg *EmailMessage) error {
	auth := c.auth()
	from, err := ParseEmailAddress(msg.Header.Get("From"))
	if err != nil {
		return err
	}
	addrs := msg.Recipients()
	recips := make([]string, len(addrs))
	for i, addr := range addrs {
		recips[i] = addr.Address()
	}
	content := msg.Bytes()
	return smtp.SendMail(c.hostAndPort(), auth, from.Address(), recips, content)
}

func (c *SMTPClient) Send(from, to, subject, textContent string, htmlContent *string) error {
	fromAddr, err := ParseEmailAddress(from)
	if err != nil {
		return err
	}
	toAddr, err := ParseEmailAddress(to)
	if err != nil {
		return err
	}
	msg := NewEmailMessage(fomAddr)
	msg.AddTo(toAddr)
	msg.SetSubject(subject)
	_, err = msg.WriteText(textContent)
	if err != nil {
		return err
	}
	if htmlContent != nil {
		_, err = msg.WriteHTML(*htmlContent)
		if err != nil {
			return err
		}
	}
	return c.Sendmail(msg)
}
