package resend_mail

import (
	"context"
	"fmt"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/resend/resend-go/v2"
)

type SendOTPMailParams struct {
	To     string
	Code   string
	Config *environment.Environment
}

func SendOTPMail(ctx context.Context, params SendOTPMailParams) error {
	client := resend.NewClient(params.Config.Resend.Key)

	emailParams := &resend.SendEmailRequest{
		From:    "QQ App <qq@homelab-kaleici.space>",
		To:      []string{params.To},
		Html:    fmt.Sprintf("<div style='font-family: Arial, sans-serif;'>Thanks for using QQ App!</div><div style='font-family: Arial, sans-serif;'>Your OTP code is: <strong>%s</strong></div><div style='font-family: Arial, sans-serif;'>This code will expire in 3 minutes.</div>", params.Code),
		Subject: "Your OTP Code For QQ App",
	}

	sent, err := client.Emails.Send(emailParams)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Log the sent email ID for debugging (optional)
	_ = sent // sent.Id contains the email ID if needed for logging

	return nil
}
