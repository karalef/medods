package mockmail

import (
	"fmt"
)

// New creates new mock mail sender.
func New() Mock {
	return Mock{}
}

// Mock is a mock mail sender.
type Mock struct{}

// Send sends the message.
func (s Mock) Send(to string, text string) error {
	fmt.Println("Message:", text)
	fmt.Println("Sent to:", to)
	return nil
}
