package implementations

type ExternalEmailer = string

const (
	Sendgrid ExternalEmailer = "sendgrid"
	SMTP     ExternalEmailer = "smtp"
)
