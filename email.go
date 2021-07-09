package sendmail

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"
)

type EmailPart struct {
	Header textproto.MIMEHeader
	body *bytes.Buffer
}

func (p *EmailPart) Write(data []byte) (int, error) {
	return p.body.Write(data)
}

type EmailMessage struct {
	Header textproto.MIMEHeader
	parts []*EmailPart
}

func NewEmailMessage(sender EmailAddress) *EmailMessage {
	e := &EmailMessage{
		Header: textproto.MIMEHeader{},
		parts: []*EmailPart{},
	}
	e.Header.Set("From", fmt.Sprintf("%s <%s>", sender.Name(), sender.Address()))
	e.Header.Set("Date", time.Now().Format("Mon, 2 Jan 2006 15:04:05 -0700"))
	e.Header.Set("MIME-Version", "1.0")
	e.NewPart(`text/plain; charset="UTF-8"`)
	return e
}

func (e *EmailMessage) Recipients() []EmailAddress {
	xaddrs := []string{}
	vals, ok := e.Header["To"]
	if ok {
		xaddrs = append(xaddrs, vals...)
	}
	vals, ok = e.Header["Cc"]
	if ok {
		xaddrs = append(xaddrs, vals...)
	}
	vals, ok = e.Header["Bcc"]
	if ok {
		xaddrs = append(xaddrs, vals...)
	}
	addrs := []EmailAddress{}
	for _, xaddr := range xaddrs {
		addr, err := ParseEmailAddress(xaddr)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

func (e *EmailMessage) AddTo(addr EmailAddress) {
	e.Header.Add("To", EmailAddressString(addr))
}

func (e *EmailMessage) AddCc(addr EmailAddress) {
	e.Header.Add("Cc", EmailAddressString(addr))
}

func (e *EmailMessage) AddBcc(addr EmailAddress) {
	e.Header.Add("Bcc", EmailAddressString(addr))
}

func (e *EmailMessage) SetSubject(sub string) {
	e.Header.Set("Subject", sub)
}

func (e *EmailMessage) NewPart(contentType string) *EmailPart {
	part := &EmailPart{
		Header: textproto.MIMEHeader{},
		body: &bytes.Buffer{},
	}
	part.Header.Set("Content-Type", contentType)
	e.parts = append(e.parts, part)
	return part
}

func (e *EmailMessage) TextPart() *EmailPart {
	return e.parts[0]
}

func (e *EmailMessage) HTMLPart() *EmailPart {
	if len(e.parts) == 1 {
		return e.NewPart(`text/html; charset="UTF-8"`)
	}
	for _, part := range e.parts {
		ct := strings.Split(part.Header.Get("Content-Type"), ";")
		if ct[0] == "text/html" {
			return part
		}
	}
	return e.NewPart(`text/html; charset="UTF-8"`)
}

func (e *EmailMessage) Write(data []byte) (int, error) {
	part := e.parts[len(e.parts) - 1]
	return part.Write(data)
}

func (e *EmailMessage) WriteText(str string) (int, error) {
	return e.TextPart().Write([]byte(str))
}

func (e *EmailMessage) WriteHTML(str string) (int, error) {
	return e.HTMLPart().Write([]byte(str))
}

func (e *EmailMessage) writeHeaderLine(buf io.Writer, key string) string {
	ckey := textproto.CanonicalMIMEHeaderKey(key)
	vals, ok := e.Header[ckey]
	if ok {
		buf.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, strings.Join(vals, ", "))))
	}
	return ckey
}

func (e *EmailMessage) Bytes() []byte {
	buf := &bytes.Buffer{}

	mp := multipart.NewWriter(buf)
	if len(e.parts) == 1 {
		e.Header.Set("Content-Type", e.parts[0].Header.Get("Content-Type"))
	} else {
		if len(e.parts) == 2 && strings.Split(e.parts[1].Header.Get("Content-Type"), ";")[0] == "text/html" {
			e.Header.Set("Content-Type", fmt.Sprintf(`multipart/alternative; boundary="%s"`, mp.Boundary()))
		} else {
			e.Header.Set("Content-Type", fmt.Sprintf(`multipart/mixed; boundary="%s"`, mp.Boundary()))
		}
	}
	seen := map[string]bool{}
	keys := []string{"From", "To", "Cc", "Subject", "MIME-Version", "Date", "Content-Type"}
	for _, k := range keys {
		ck := e.writeHeaderLine(buf, k)
		seen[ck] = true
	}
	for k := range e.Header {
		if seen[k] {
			continue
		}
		if k == "Bcc" {
			continue
		}
		e.writeHeaderLine(buf, k)
	}
	buf.Write([]byte("\r\n"))
	if len(e.parts) == 1 {
		buf.Write(e.parts[0].body.Bytes())
	} else {
		for _, part := range e.parts {
			w, _ := mp.CreatePart(part.Header)
			w.Write(part.body.Bytes())
		}
		mp.Close()
	}
	return buf.Bytes()
}
