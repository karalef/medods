package mail

import (
	"fmt"
)

func NewMock() (Mock, error) {
	return Mock{}, nil
}

type Mock struct{}

func (s Mock) Send(to string, text string) error {
	fmt.Println("Message:", text)
	fmt.Println("Sent to:", to)
	return nil
}
