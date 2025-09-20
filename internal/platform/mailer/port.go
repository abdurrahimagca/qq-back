package mailer

import "context"

// SendParams captures the information needed to send an email.
type SendParams struct {
	To      string
	From    string
	Subject string
	Body    string
}

// Service exposes email-sending capabilities to the application layer.
type Service interface {
	SendEmail(ctx context.Context, params SendParams) error
	GetTemplate(ctx context.Context, templateName string) (string, error)
}
