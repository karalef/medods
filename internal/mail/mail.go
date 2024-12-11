package mail

import (
	"net"
	"net/smtp"
)

func NewSender(smtpAddr, from, pass string) (*Sender, error) {
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

type Sender struct {
	addr string
	from string
	auth smtp.Auth
}

func (s *Sender) Send(to string, text string) error {
	return smtp.SendMail(s.addr, s.auth, s.from, []string{to}, []byte(text))
}
