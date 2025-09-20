package mailer

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/resend/resend-go/v2"
)

var (
	//go:embed templates/*.html
	templatesFS embed.FS
)

type resendMailer struct {
	environment *environment.Environment
	client      *resend.Client
}

func NewResendMailer(conf *environment.Environment) Service {
	client := resend.NewClient(conf.Resend.Key)

	return &resendMailer{
		environment: conf,
		client:      client,
	}
}

func (m *resendMailer) SendEmail(ctx context.Context, params SendParams) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

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

func (m *resendMailer) GetTemplate(ctx context.Context, templateName string) (string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}
	if strings.Contains(templateName, "..") {
		return "", errors.New("invalid template name")
	}

	filePath := fmt.Sprintf("templates/%s.html", templateName)
	content, err := templatesFS.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}
	return string(content), nil
}
