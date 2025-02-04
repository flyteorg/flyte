package mocks

import (
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
)

type AwsSendEmailFunc func(input *ses.SendEmailInput) (*ses.SendEmailOutput, error)

type SESClient struct {
	sesiface.SESAPI
	sendEmail AwsSendEmailFunc
}

func (m *SESClient) SetSendEmailFunc(emailFunc AwsSendEmailFunc) {
	m.sendEmail = emailFunc
}

func (m *SESClient) SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	if m.sendEmail != nil {
		return m.sendEmail(input)
	}
	return &ses.SendEmailOutput{}, nil
}
