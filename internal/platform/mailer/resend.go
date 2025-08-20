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
		return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="margin: 0; padding: 20px; font-family: Arial, sans-serif; background-color: #f5f5f5;">
    <table width="100%" cellpadding="0" cellspacing="0" style="max-width: 600px; margin: 0 auto; background-color: white; border-radius: 8px;">
        <tr>
            <td style="background-color: #4f46e5; padding: 30px; text-align: center; border-radius: 8px 8px 0 0;">
                <h1 style="color: white; margin: 0; font-size: 24px;">OTP Verification</h1>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px; text-align: center;">
                <p style="font-size: 16px; color: #333; margin-bottom: 30px;">Your verification code is:</p>
                <div style="background-color: #4f46e5; color: white; font-size: 36px; font-weight: bold; padding: 20px; border-radius: 8px; letter-spacing: 4px; margin: 20px 0; display: inline-block; font-family: monospace;">{{.OTP}}</div>
                <p style="color: red; font-size: 20x; margin-top: 30px;">This code expires in 3 minutes. Don't share it with anyone.</p>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8f9fa; padding: 20px; text-align: center; border-radius: 0 0 8px 8px; border-top: 1px solid #e9ecef;">
                <p style="color: #666; font-size: 12px; margin: 0;">QQ Application - Automated Message</p>
            </td>
        </tr>
    </table>
</body>
</html>`, nil
	}
	return "", fmt.Errorf("template not found")
}
