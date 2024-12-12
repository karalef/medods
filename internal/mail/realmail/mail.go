package realmail

import (
	"net"
	"net/smtp"
)

// New creates new mail sender.
func New(smtpAddr, from, pass string) (*Sender, error) {
	host, _, err := net.SplitHostPort(smtpAddr)
	if err != nil {
		return nil, err
	}
	return &Sender{
		addr: smtpAddr,
		from: from,
		auth: smtp.PlainAuth("", from, pass, host),
	}, nil
}

// Sender is a mail sender.
type Sender struct {
	addr string
	from string
	auth smtp.Auth
}

// Send sends the message.
func (s *Sender) Send(to string, text string) error {
	return smtp.SendMail(s.addr, s.auth, s.from, []string{to}, []byte(text))
}
