package implementations

type ExternalEmailer = string

const (
	Sendgrid ExternalEmailer = "sendgrid"
	Smtp     ExternalEmailer = "smtp"
)
