package mailer

import (
	"context"
	"fmt"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/ports"
	"github.com/resend/resend-go/v2"
)

type resendMailer struct {
	environment *environment.Environment
	client      *resend.Client
}

func NewResendMailer(conf *environment.Environment) ports.MailerPort {
	client := resend.NewClient(conf.Resend.Key)

	return &resendMailer{
		environment: conf,
		client:      client,
	}
}

func (m *resendMailer) SendEmail(ctx context.Context, params ports.SendEmailParams) error {

	emailParams := &resend.SendEmailRequest{
		From:    params.From,
		To:      []string{params.To},
		Html:    params.Body,
		Subject: params.Subject,
	}

	sent, err := m.client.Emails.Send(emailParams)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	_ = sent

	return nil
}

func (m *resendMailer) GetEmailTemplate(ctx context.Context, templateName string) (string, error) {

	if templateName == "otp" {
		return `	
		<html>
		<body>
		<h1>OTP Verification</h1>
		<p>Your OTP is {{.OTP}}</p>
		</body>
		</html>
		`, nil
	}
	return "", fmt.Errorf("template not found")
}
