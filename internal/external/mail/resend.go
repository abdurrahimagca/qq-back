package mail

import (
	"context"
	"fmt"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/resend/resend-go/v2"
)

type SendOTPMailParams struct {
	To     string
	Code   string
	Config *environment.Config
}

func SendOTPMail(ctx context.Context, params SendOTPMailParams) error {
	client := resend.NewClient(params.Config.Resend.Key)

	emailParams := &resend.SendEmailRequest{
		From:    "QQ Quote <qq@homelab-kaleici.space>",
		To:      []string{params.To},
		Html:    fmt.Sprintf("<p>Thanks for using QQ Quote!</p><p>Your OTP code is: <strong>%s</strong></p><p>This code will expire in 3 minutes.</p>", params.Code),
		Subject: "Your OTP Code For QQ Quote",
	}

	sent, err := client.Emails.Send(emailParams)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Log the sent email ID for debugging (optional)
	_ = sent // sent.Id contains the email ID if needed for logging

	return nil
}
