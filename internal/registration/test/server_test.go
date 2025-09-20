package registration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/registration"
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRegistrationUsecase struct {
	registerResult    *bool
	registerErr       error
	verifyResult      token.GenerateTokenResult
	verifyErr         error
	refreshResult     token.GenerateTokenResult
	refreshErr        error
	lastRegisterEmail string
	lastVerifyEmail   string
	lastVerifyOTP     string
	lastRefreshToken  string
}

func (f *fakeRegistrationUsecase) RegisterOrLoginOTP(
	ctx context.Context, email string,
) (*bool, error) {
	f.lastRegisterEmail = email
	return f.registerResult, f.registerErr
}

func (f *fakeRegistrationUsecase) VerifyOTPAndLogin(
	ctx context.Context, email string, otp string,
) (token.GenerateTokenResult, error) {
	f.lastVerifyEmail = email
	f.lastVerifyOTP = otp
	return f.verifyResult, f.verifyErr
}

func (f *fakeRegistrationUsecase) RefreshTokens(
	ctx context.Context, refreshToken string,
) (token.GenerateTokenResult, error) {
	f.lastRefreshToken = refreshToken
	return f.refreshResult, f.refreshErr
}

func TestRegistrationServer_SendOtpHandler_Success(t *testing.T) {
	isNew := true
	uc := &fakeRegistrationUsecase{registerResult: &isNew}
	server := registration.NewRegistrationServer(uc)

	input := &registration.SendOtpInput{}
	input.Body.Email = "user@example.com"

	resp, err := server.SendOtpHandler(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Body.Data.IsNewUser)
	assert.Equal(t, "user@example.com", uc.lastRegisterEmail)
}

func TestRegistrationServer_SendOtpHandler_Error(t *testing.T) {
	uc := &fakeRegistrationUsecase{registerErr: qqerrors.ErrValidationError}
	server := registration.NewRegistrationServer(uc)

	input := &registration.SendOtpInput{}
	input.Body.Email = "invalid"

	resp, err := server.SendOtpHandler(context.Background(), input)
	require.Nil(t, resp)
	require.Error(t, err)
}

func TestRegistrationServer_VerifyOtpHandler_Success(t *testing.T) {
	uc := &fakeRegistrationUsecase{verifyResult: token.GenerateTokenResult{AccessToken: "acc", RefreshToken: "ref"}}
	server := registration.NewRegistrationServer(uc)

	input := &registration.VerifyOtpInput{}
	input.Body.Email = "user@example.com"
	input.Body.OtpCode = "123456"

	resp, err := server.VerifyOtpHandler(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "acc", resp.Body.Data.AccessToken)
	assert.Equal(t, "ref", resp.Body.Data.RefreshToken)
	assert.Equal(t, "user@example.com", uc.lastVerifyEmail)
	assert.Equal(t, "123456", uc.lastVerifyOTP)
}

func TestRegistrationServer_VerifyOtpHandler_Error(t *testing.T) {
	uc := &fakeRegistrationUsecase{verifyErr: errors.New("invalid otp")}
	server := registration.NewRegistrationServer(uc)

	input := &registration.VerifyOtpInput{}
	input.Body.Email = "user@example.com"
	input.Body.OtpCode = "000000"

	resp, err := server.VerifyOtpHandler(context.Background(), input)
	require.Nil(t, resp)
	require.Error(t, err)
}

func TestRegistrationServer_RefreshTokensHandler_Success(t *testing.T) {
	uc := &fakeRegistrationUsecase{
		refreshResult: token.GenerateTokenResult{AccessToken: "new-acc", RefreshToken: "new-ref"}}
	server := registration.NewRegistrationServer(uc)

	input := &registration.RefreshTokensInput{}
	input.Body.RefreshToken = "token"

	resp, err := server.RefreshTokensHandler(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "new-acc", resp.Body.Data.AccessToken)
	assert.Equal(t, "new-ref", resp.Body.Data.RefreshToken)
	assert.Equal(t, "token", uc.lastRefreshToken)
}

func TestRegistrationServer_RefreshTokensHandler_Error(t *testing.T) {
	uc := &fakeRegistrationUsecase{refreshErr: qqerrors.ErrUnauthorized}
	server := registration.NewRegistrationServer(uc)

	input := &registration.RefreshTokensInput{}
	input.Body.RefreshToken = "bad"

	resp, err := server.RefreshTokensHandler(context.Background(), input)
	require.Nil(t, resp)
	require.Error(t, err)
}
