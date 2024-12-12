package mail

// Mailer represents the mail sender.
type Mailer interface {
	Send(to, text string) error
}
