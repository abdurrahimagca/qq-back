package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func resend(ctx context.Context, url string, key string, body map[string]interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: %d", resp.StatusCode)
	}

	return nil
}

type SendOTPMailParams struct {
	To     string
	Code   string
	Config *environment.Config
}

func SendOTPMail(ctx context.Context, params SendOTPMailParams) error {
	body := map[string]interface{}{
		"from":    "qq@homelab-kaleici.space",
		"to":      params.To,
		"subject": "Your OTP Code For QQ",
		"html":    fmt.Sprintf("Thanks for using QQ. Your OTP Code is %s", params.Code),
	}

	return resend(ctx, params.Config.Resend.Url, params.Config.Resend.Key, body)
}
