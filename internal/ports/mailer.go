
package ports

import (
	"context"
)

type SendEmailParams struct {
	To      string
	From    string
	Subject string
	Body    string
}

type MailerPort interface {
	SendEmail(ctx context.Context, params SendEmailParams) error
	GetEmailTemplate(ctx context.Context, templateName string) (string, error)
}
