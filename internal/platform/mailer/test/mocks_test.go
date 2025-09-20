package mailer_test

import (
	"errors"

	"github.com/resend/resend-go/v2"
)

// EmailsAPI defines the interface we need from resend client.
type EmailsAPI interface {
	Send(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error)
}

// MockEmailsAPI implements EmailsAPI for testing.
type MockEmailsAPI struct {
	SendFunc   func(*resend.SendEmailRequest) (*resend.SendEmailResponse, error)
	SendCalls  []*resend.SendEmailRequest
	SendError  error
	SendResult *resend.SendEmailResponse
}

func NewMockEmailsAPI() *MockEmailsAPI {
	return &MockEmailsAPI{
		SendCalls: make([]*resend.SendEmailRequest, 0),
	}
}

func (m *MockEmailsAPI) Send(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	m.SendCalls = append(m.SendCalls, params)

	if m.SendFunc != nil {
		return m.SendFunc(params)
	}

	if m.SendError != nil {
		return nil, m.SendError
	}

	if m.SendResult != nil {
		return m.SendResult, nil
	}

	// Default success response.
	return &resend.SendEmailResponse{
		Id: "test-email-id-123",
	}, nil
}

// GetLastSendCall returns the last call to Send, or nil if no calls were made.
func (m *MockEmailsAPI) GetLastSendCall() *resend.SendEmailRequest {
	if len(m.SendCalls) == 0 {
		return nil
	}
	return m.SendCalls[len(m.SendCalls)-1]
}

// GetSendCallCount returns the number of times Send was called.
func (m *MockEmailsAPI) GetSendCallCount() int {
	return len(m.SendCalls)
}

// Reset clears all recorded calls.
func (m *MockEmailsAPI) Reset() {
	m.SendCalls = make([]*resend.SendEmailRequest, 0)
	m.SendError = nil
	m.SendResult = nil
	m.SendFunc = nil
}

// SetSendError configures the mock to return an error on Send.
func (m *MockEmailsAPI) SetSendError(err error) {
	m.SendError = err
}

// SetSendResult configures the mock to return a specific result on Send.
func (m *MockEmailsAPI) SetSendResult(result *resend.SendEmailResponse) {
	m.SendResult = result
}

// Common error responses for testing.
var (
	ErrNetworkFailure = errors.New("network connection failed")
	ErrInvalidAPIKey  = errors.New("invalid API key")
	ErrRateLimited    = errors.New("rate limit exceeded")
)
