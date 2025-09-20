package registration_test

import (
	"context"
	"errors"
	"sync"

	"github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	token "github.com/abdurrahimagca/qq-back/internal/platform/token"
)

type fakeMailer struct {
	mu          sync.Mutex
	template    string
	templateErr error
	sendErr     error
	sentEmails  []mailer.SendParams
}

func (f *fakeMailer) GetTemplate(ctx context.Context, templateName string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.templateErr != nil {
		return "", f.templateErr
	}
	return f.template, nil
}

func (f *fakeMailer) SendEmail(ctx context.Context, params mailer.SendParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sendErr != nil {
		return f.sendErr
	}
	f.sentEmails = append(f.sentEmails, params)
	return nil
}

func (f *fakeMailer) setTemplate(template string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.template = template
}

func (f *fakeMailer) setSendErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sendErr = err
}

func (f *fakeMailer) lastEmail() (mailer.SendParams, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sentEmails) == 0 {
		return mailer.SendParams{}, errors.New("no emails sent")
	}
	return f.sentEmails[len(f.sentEmails)-1], nil
}

func (f *fakeMailer) emailCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sentEmails)
}

type fakeTokenService struct {
	mu                    sync.Mutex
	expectedValidateToken string
	validateResult        token.ValidateTokenResult
	validateErr           error
	generateResult        token.GenerateTokenResult
	generateErr           error
	generateCalls         []token.GenerateTokenParams
}

func (f *fakeTokenService) GenerateTokens(
	ctx context.Context,
	params token.GenerateTokenParams,
) (token.GenerateTokenResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.generateErr != nil {
		return token.GenerateTokenResult{}, f.generateErr
	}
	f.generateCalls = append(f.generateCalls, params)
	return f.generateResult, nil
}

func (f *fakeTokenService) ValidateToken(
	ctx context.Context,
	params token.ValidateTokenParams,
) (token.ValidateTokenResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.validateErr != nil {
		return token.ValidateTokenResult{}, f.validateErr
	}
	if f.expectedValidateToken != "" && params.Token != f.expectedValidateToken {
		return token.ValidateTokenResult{}, errors.New("unexpected token value")
	}
	return f.validateResult, nil
}

func (f *fakeTokenService) setGenerateResult(result token.GenerateTokenResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.generateResult = result
}

func (f *fakeTokenService) setValidateResult(
	result token.ValidateTokenResult,
) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validateResult = result
}

func (f *fakeTokenService) setValidateErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validateErr = err
}

func (f *fakeTokenService) setExpectedValidateToken(token string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expectedValidateToken = token
}

func (f *fakeTokenService) lastGenerateCall() (token.GenerateTokenParams, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.generateCalls) == 0 {
		return token.GenerateTokenParams{}, errors.New("no generate calls")
	}
	return f.generateCalls[len(f.generateCalls)-1], nil
}

func (f *fakeTokenService) generateCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.generateCalls)
}
