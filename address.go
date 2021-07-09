package sendmail

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email address")

type EmailAddress interface {
	Address() string
	Name() string
}

type address struct {
	username string
	domain string
	name string
}

func (a address) Address() string {
	return fmt.Sprintf("%s@%s", a.username, a.domain)
}

func (a address) Name() string {
	return a.name
}

func (a address) String() string {
	return EmailAddressString(a)
}

func EmailAddressString(addr EmailAddress) string {
	name := addr.Name()
	if name != "" {
		return fmt.Sprintf("%s <%s>", name, addr.Address())
	}
	return addr.Address()
}

func ParseEmailAddress(addr string) (EmailAddress, error) {
	left := strings.Index(addr, "<")
	right := strings.Index(addr, ">")
	var name, username, domain string
	if left >= 0 && right > left {
		name = strings.Trim(strings.TrimSpace(addr[:left]), "\"")
		addr = addr[left+1:right]
	}
	parts := strings.Split(addr, "@")
	if len(parts) == 1 {
		return nil, ErrInvalidEmail
	}
	domain = strings.ToLower(strings.TrimSpace(parts[len(parts) - 1]))
	parts = strings.Fields(parts[len(parts) - 2])
	username = parts[len(parts) - 1]
	return &address{
		username: username,
		domain: domain,
		name: name,
	}, nil
}
