package interfaces

import (
	"crypto/tls"
	"io"
	"net/smtp"
)

// This interface is introduced to allow for mocking of the smtp.Client object.

//go:generate mockery -name=SMTPClient -output=../mocks -case=underscore
type SMTPClient interface {
	Hello(localName string) error
	Extension(ext string) (bool, string)
	Auth(a smtp.Auth) error
	StartTLS(config *tls.Config) error
	Noop() error
	Close() error
	Mail(from string) error
	Rcpt(to string) error
	Data() (io.WriteCloser, error)
}
