package mailer_test

import (
	"context"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestEnv(apiKey string) *environment.Environment {
	return &environment.Environment{
		Ctx: context.Background(),
		Resend: environment.ResendEnvironment{
			Key: apiKey,
		},
	}
}

func TestNewResendMailer(t *testing.T) {
	env := buildTestEnv("test-api-key")

	service := mailer.NewResendMailer(env)

	assert.NotNil(t, service, "Service should not be nil")
}

func TestResendMailer_GetTemplate_ExistingTemplate(t *testing.T) {
	env := buildTestEnv("test-api-key")
	service := mailer.NewResendMailer(env)
	ctx := context.Background()

	template, err := service.GetTemplate(ctx, "otp")

	require.NoError(t, err)
	assert.NotEmpty(t, template, "Template content should not be empty")
	assert.Contains(t, template, "<!DOCTYPE html>", "Should contain HTML doctype")
	assert.Contains(t, template, "OTP Verification", "Should contain OTP verification title")
	assert.Contains(t, template, "{{.OTP}}", "Should contain OTP placeholder")
}

func TestResendMailer_GetTemplate_NonExistingTemplate(t *testing.T) {
	env := buildTestEnv("test-api-key")
	service := mailer.NewResendMailer(env)
	ctx := context.Background()

	template, err := service.GetTemplate(ctx, "nonexistent")

	require.Error(t, err, "Should return error for non-existing template")
	assert.Empty(t, template, "Template content should be empty on error")
	assert.Contains(t, err.Error(), "template not found", "Error should mention template not found")
}

func TestResendMailer_GetTemplate_PathTraversal(t *testing.T) {
	env := buildTestEnv("test-api-key")
	service := mailer.NewResendMailer(env)
	ctx := context.Background()

	tests := []struct {
		name         string
		templateName string
	}{
		{
			name:         "Double dot traversal",
			templateName: "../secrets",
		},
		{
			name:         "Complex path traversal",
			templateName: "../../etc/passwd",
		},
		{
			name:         "Hidden double dots",
			templateName: "template..name",
		},
		{
			name:         "URL encoded traversal",
			templateName: "..%2F..%2Fetc%2Fpasswd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := service.GetTemplate(ctx, tt.templateName)

			require.Error(t, err, "Should return error for path traversal attempt")
			assert.Empty(t, template, "Template content should be empty on error")
			assert.Contains(t, err.Error(), "invalid template name", "Error should mention invalid template name")
		})
	}
}

// Note: The following tests would require modifying the mailer to accept a mock client
// Since the current implementation creates the client internally, these tests demonstrate
// the testing approach but won't actually run without refactoring the code structure

func TestResendMailer_SendEmail_HappyPath(t *testing.T) {
	// This test demonstrates how we would test SendEmail with a mock client
	// In a real implementation, we'd need to refactor the mailer to accept
	// the client as a dependency injection

	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient) // hypothetical constructor
	// ctx := context.Background()

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Test Subject",
	// 	Body:    "<h1>Test Body</h1>",
	// }

	// err := service.SendEmail(ctx, params)

	// require.NoError(t, err)
	// assert.Equal(t, 1, mockClient.GetSendCallCount())

	// lastCall := mockClient.GetLastSendCall()
	// assert.Equal(t, params.From, lastCall.From)
	// assert.Equal(t, []string{params.To}, lastCall.To)
	// assert.Equal(t, params.Subject, lastCall.Subject)
	// assert.Equal(t, params.Body, lastCall.Html)
}

func TestResendMailer_SendEmail_ClientError(t *testing.T) {
	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// mockClient.SetSendError(ErrNetworkFailure)

	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient)
	// ctx := context.Background()

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Test Subject",
	// 	Body:    "<h1>Test Body</h1>",
	// }

	// err := service.SendEmail(ctx, params)

	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to send email")
	// assert.Contains(t, err.Error(), ErrNetworkFailure.Error())
}

func TestResendMailer_SendEmail_EmptyBody(t *testing.T) {
	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient)
	// ctx := context.Background()

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Test Subject",
	// 	Body:    "",
	// }

	// err := service.SendEmail(ctx, params)

	// require.NoError(t, err, "Should handle empty body gracefully")
	// assert.Equal(t, 1, mockClient.GetSendCallCount())

	// lastCall := mockClient.GetLastSendCall()
	// assert.Equal(t, "", lastCall.Html, "Should send empty HTML body")
}

func TestResendMailer_SendEmail_WhitespaceBody(t *testing.T) {
	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient)
	// ctx := context.Background()

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Test Subject",
	// 	Body:    "   \n\t   ",
	// }

	// err := service.SendEmail(ctx, params)

	// require.NoError(t, err, "Should handle whitespace body gracefully")
	// assert.Equal(t, 1, mockClient.GetSendCallCount())

	// lastCall := mockClient.GetLastSendCall()
	// assert.Equal(t, params.Body, lastCall.Html, "Should preserve whitespace in body")
}

func TestResendMailer_SendEmail_UnicodeContent(t *testing.T) {
	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient)
	// ctx := context.Background()

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Test Subject with 🚀 emoji and ñoño characters",
	// 	Body:    "<h1>Hello 世界! 🌍</h1><p>Testing unicode: αβγδε</p>",
	// }

	// err := service.SendEmail(ctx, params)

	// require.NoError(t, err, "Should handle unicode content")
	// assert.Equal(t, 1, mockClient.GetSendCallCount())

	// lastCall := mockClient.GetLastSendCall()
	// assert.Equal(t, params.Subject, lastCall.Subject)
	// assert.Equal(t, params.Body, lastCall.Html)
}

func TestResendMailer_SendEmail_LargeBody(t *testing.T) {
	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient)
	// ctx := context.Background()

	// // Create a large HTML body (1MB)
	// largeContent := strings.Repeat("<p>This is a test paragraph. </p>", 50000)

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Large Email Test",
	// 	Body:    largeContent,
	// }

	// err := service.SendEmail(ctx, params)

	// require.NoError(t, err, "Should handle large email body")
	// assert.Equal(t, 1, mockClient.GetSendCallCount())

	// lastCall := mockClient.GetLastSendCall()
	// assert.Equal(t, largeContent, lastCall.Html)
}

func TestResendMailer_SendEmail_SpecialCharacters(t *testing.T) {
	t.Skip("Requires refactoring mailer to accept injectable client")

	// mockClient := NewMockEmailsAPI()
	// env := buildTestEnv("test-api-key")
	// service := mailer.NewResendMailerWithClient(env, mockClient)
	// ctx := context.Background()

	// params := mailer.SendParams{
	// 	From:    "test@example.com",
	// 	To:      "recipient@example.com",
	// 	Subject: "Special chars: <>&\"'",
	// 	Body:    "<h1>HTML with &lt;special&gt; chars &amp; &quot;quotes&quot;</h1>",
	// }

	// err := service.SendEmail(ctx, params)

	// require.NoError(t, err, "Should handle special characters")
	// assert.Equal(t, 1, mockClient.GetSendCallCount())

	// lastCall := mockClient.GetLastSendCall()
	// assert.Equal(t, params.Subject, lastCall.Subject)
	// assert.Equal(t, params.Body, lastCall.Html)
}

// Integration-style test that can actually run with the current implementation
func TestResendMailer_Integration_InvalidAPIKey(t *testing.T) {
	// This test can run because it doesn't require mocking - it uses the real client
	// but with an invalid API key to test error handling

	env := buildTestEnv("invalid-api-key")
	service := mailer.NewResendMailer(env)
	ctx := context.Background()

	params := mailer.SendParams{
		From:    "test@example.com",
		To:      "recipient@example.com",
		Subject: "Test Subject",
		Body:    "<h1>Test Body</h1>",
	}

	err := service.SendEmail(ctx, params)

	require.Error(t, err, "Should return error for invalid API key")
	assert.Contains(t, err.Error(), "failed to send email", "Error should be wrapped appropriately")
}
