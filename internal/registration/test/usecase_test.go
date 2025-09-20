package registration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/registration"
	"github.com/abdurrahimagca/qq-back/internal/user"
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRegistrationUsecaseForTest(
	h *registrationTestHarness,
	mailSvc *fakeMailer,
	tokenSvc *fakeTokenService,
) registration.Usecase {
	authService := auth.NewService(h.authRepo)
	userService := user.NewService(h.userRepo)
	return registration.NewUsecase(mailSvc, authService, userService, h.pool, tokenSvc)
}

func TestRegisterOrLoginOTP_ExistingUser(t *testing.T) {
	h := newRegistrationTestHarness(t)
	ctx := context.Background()

	email := fmt.Sprintf("existing-%d@example.com", time.Now().UnixNano())
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	authID, userRecord := createAuthAndUser(t, h, email, username)

	// Seed stale OTP to ensure cleanup occurs.
	sql := "INSERT INTO auth_otp_codes (auth_id, code) VALUES ($1, $2)"
	_, err := h.pool.Exec(ctx, sql, authID, hashOTP("OLDOTP"))
	require.NoError(t, err)

	mailerFake := &fakeMailer{}
	mailerFake.setTemplate("Your OTP is {{.OTP}}")
	tokenFake := &fakeTokenService{}

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	useDeterministicRand(t, []byte{0x0a, 0x0b, 0x0c})

	isNewUserPtr, err := usecase.RegisterOrLoginOTP(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, isNewUserPtr)
	assert.False(t, *isNewUserPtr)

	require.Equal(t, 1, mailerFake.emailCount())
	emailParams, err := mailerFake.lastEmail()
	require.NoError(t, err)
	assert.Equal(t, email, emailParams.To)
	assert.Equal(t, "OTP Verification", emailParams.Subject)

	otpCode := "0A0B0C"
	assert.Contains(t, emailParams.Body, otpCode)

	// Only one OTP should exist for the auth id and it should be the new hash.
	assert.Equal(t, 1, countOTPs(t, h.pool, authID))
	var storedHash string
	require.NoError(
		t, h.pool.QueryRow(ctx, "SELECT code FROM auth_otp_codes WHERE auth_id = $1", authID).Scan(&storedHash))
	assert.Equal(t, hashOTP(otpCode), storedHash)

	retrieved := fetchUserByEmail(t, user.NewService(h.userRepo), ctx, email)
	assert.Equal(t, userRecord.ID, retrieved.ID)
}

func TestRegisterOrLoginOTP_NewUser(t *testing.T) {
	h := newRegistrationTestHarness(t)
	ctx := context.Background()

	email := fmt.Sprintf("new-%d@example.com", time.Now().UnixNano())

	mailerFake := &fakeMailer{}
	mailerFake.setTemplate("Code: {{.OTP}}")
	tokenFake := &fakeTokenService{}

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	useDeterministicRand(t, []byte{0x1a, 0x2b, 0x3c})

	isNewUserPtr, err := usecase.RegisterOrLoginOTP(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, isNewUserPtr)
	assert.True(t, *isNewUserPtr)

	emailParams, err := mailerFake.lastEmail()
	require.NoError(t, err)
	assert.Equal(t, email, emailParams.To)
	assert.Contains(t, emailParams.Body, "1A2B3C")

	userSvc := user.NewService(h.userRepo)
	retrieved := fetchUserByEmail(t, userSvc, ctx, email)
	require.NotNil(t, retrieved)
	assert.True(t, retrieved.ID.Valid)
	assert.True(t, retrieved.AuthID.Valid)
	assert.NotEmpty(t, retrieved.Username)
}

func TestRegisterOrLoginOTP_MailerFailureAfterCommit(t *testing.T) {
	h := newRegistrationTestHarness(t)
	ctx := context.Background()

	mailerFake := &fakeMailer{}
	mailerFake.setTemplate("OTP {{.OTP}}")
	mailerFake.setSendErr(fmt.Errorf("smtp down"))

	tokenFake := &fakeTokenService{}

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	email := fmt.Sprintf("mailer-%d@example.com", time.Now().UnixNano())
	useDeterministicRand(t, []byte{0x09, 0x09, 0x09})

	_, err := usecase.RegisterOrLoginOTP(ctx, email)
	require.Error(t, err)
	assert.EqualError(t, err, "smtp down")

	var count int
	require.NoError(t, h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM auth_otp_codes").Scan(&count))
	assert.Equal(t, 1, count)
}

func TestVerifyOTPAndLogin_Success(t *testing.T) {
	h := newRegistrationTestHarness(t)
	ctx := context.Background()

	email := fmt.Sprintf("verify-%d@example.com", time.Now().UnixNano())

	mailerFake := &fakeMailer{}
	mailerFake.setTemplate("OTP {{.OTP}}")

	tokenFake := &fakeTokenService{}
	tokenFake.setGenerateResult(token.GenerateTokenResult{AccessToken: "access", RefreshToken: "refresh"})

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	useDeterministicRand(t, []byte{0xaa, 0xbb, 0xcc})
	_, err := usecase.RegisterOrLoginOTP(ctx, email)
	require.NoError(t, err)
	emailParams, err := mailerFake.lastEmail()
	require.NoError(t, err)
	otpCode := strings.TrimPrefix(emailParams.Body, "OTP ")
	otpCode = strings.TrimSpace(otpCode)

	result, err := usecase.VerifyOTPAndLogin(ctx, email, otpCode)
	require.NoError(t, err)
	assert.Equal(t, "access", result.AccessToken)
	assert.Equal(t, "refresh", result.RefreshToken)

	require.Equal(t, 1, tokenFake.generateCallCount())
	call, err := tokenFake.lastGenerateCall()
	require.NoError(t, err)
	assert.NotEmpty(t, call.UserID)
}

func TestVerifyOTPAndLogin_InvalidOTP(t *testing.T) {
	h := newRegistrationTestHarness(t)
	ctx := context.Background()

	email := fmt.Sprintf("invalid-%d@example.com", time.Now().UnixNano())

	mailerFake := &fakeMailer{}
	mailerFake.setTemplate("OTP {{.OTP}}")
	tokenFake := &fakeTokenService{}

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	useDeterministicRand(t, []byte{0x01, 0x02, 0x03})
	_, err := usecase.RegisterOrLoginOTP(ctx, email)
	require.NoError(t, err)

	_, err = usecase.VerifyOTPAndLogin(ctx, email, "WRONGOTP")
	require.Error(t, err)
}

func TestRefreshTokens_Success(t *testing.T) {
	h := newRegistrationTestHarness(t)
	ctx := context.Background()

	email := fmt.Sprintf("refresh-%d@example.com", time.Now().UnixNano())
	authID, userRecord := createAuthAndUser(t, h, email, fmt.Sprintf("refresh_user_%d", time.Now().UnixNano()))

	// Ensure user service can fetch by ID
	require.NotEqual(t, pgtype.UUID{}, userRecord.ID)

	mailerFake := &fakeMailer{}
	tokenFake := &fakeTokenService{}
	tokenFake.setValidateResult(token.ValidateTokenResult{Claims: &token.Claims{UserID: userRecord.ID.String()}})
	tokenFake.setExpectedValidateToken("valid-refresh")
	tokenFake.setGenerateResult(token.GenerateTokenResult{AccessToken: "new-access", RefreshToken: "new-refresh"})

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	result, err := usecase.RefreshTokens(ctx, "valid-refresh")
	require.NoError(t, err)
	assert.Equal(t, "new-access", result.AccessToken)
	assert.Equal(t, "new-refresh", result.RefreshToken)

	require.Equal(t, 1, tokenFake.generateCallCount())
	call, err := tokenFake.lastGenerateCall()
	require.NoError(t, err)
	assert.Equal(t, userRecord.ID.String(), call.UserID)

	// Ensure stored auth remains linked
	assert.Equal(t, 0, countOTPs(t, h.pool, authID))
}

func TestRefreshTokens_EmptyUserIDClaims(t *testing.T) {
	h := newRegistrationTestHarness(t)

	mailerFake := &fakeMailer{}
	tokenFake := &fakeTokenService{}
	tokenFake.setValidateResult(token.ValidateTokenResult{Claims: &token.Claims{UserID: ""}})

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	_, err := usecase.RefreshTokens(context.Background(), "any")
	require.ErrorIs(t, err, qqerrors.ErrValidationError)
}

func TestRefreshTokens_InvalidUUID(t *testing.T) {
	h := newRegistrationTestHarness(t)

	mailerFake := &fakeMailer{}
	tokenFake := &fakeTokenService{}
	tokenFake.setValidateResult(token.ValidateTokenResult{Claims: &token.Claims{UserID: "not-a-uuid"}})

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	_, err := usecase.RefreshTokens(context.Background(), "token")
	require.Error(t, err)
}

func TestRefreshTokens_ValidateError(t *testing.T) {
	h := newRegistrationTestHarness(t)

	mailerFake := &fakeMailer{}
	tokenFake := &fakeTokenService{}
	tokenFake.setValidateErr(fmt.Errorf("invalid token"))

	usecase := newRegistrationUsecaseForTest(h, mailerFake, tokenFake)

	_, err := usecase.RefreshTokens(context.Background(), "bad-token")
	require.Error(t, err)
	assert.EqualError(t, err, "invalid token")
}
