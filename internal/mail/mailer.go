package mail

type Mailer interface {
	Send(to, text string) error
}
