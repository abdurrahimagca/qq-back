package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/db"
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GenerateAndSaveOTPForAuth_Success(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	svc := auth.NewService(fakeRepo)

	authID := newPGUUID()
	email := "user@example.com"
	fakeRepo.setAuthEmail(authID, email)

	useRandReader(t, &deterministicReader{data: []byte{0x1a, 0x2b, 0x3c}})

	code, err := svc.GenerateAndSaveOTPForAuth(ctx, authID)
	require.NoError(t, err)
	assert.Equal(t, "1A2B3C", code)

	hash := hashOTP(code)
	stored, ok := fakeRepo.getOTP(hash)
	require.True(t, ok, "expected OTP hash to be stored")
	assert.Equal(t, uuidToString(authID), uuidToString(stored.AuthID))
	assert.Equal(t, email, stored.Email)
}

func TestService_GenerateAndSaveOTPForAuth_CreateOTPError(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	fakeRepo.setCreateOTPErr(errors.New("db unavailable"))
	svc := auth.NewService(fakeRepo)

	useRandReader(t, &deterministicReader{data: []byte{0xaa, 0xbb, 0xcc}})

	_, err := svc.GenerateAndSaveOTPForAuth(ctx, newPGUUID())
	require.Error(t, err)
}

func TestService_VerifyOTP_Success(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	svc := auth.NewService(fakeRepo)

	otpCode := "ABC123"
	authID := newPGUUID()
	userID := newPGUUID()
	email := "user@example.com"

	fakeRepo.setAuthEmail(authID, email)
	fakeRepo.setUserForAuth(authID, userID)
	fakeRepo.setOTP(hashOTP(otpCode), db.GetUserIdAndEmailByOtpCodeRow{
		AuthID: authID,
		Email:  email,
		ID:     userID,
	})

	err := svc.VerifyOTP(ctx, email, otpCode)
	require.NoError(t, err)
}

func TestService_VerifyOTP_EmailMismatch(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	svc := auth.NewService(fakeRepo)

	fakeRepo.setOTP(hashOTP("ABC123"), db.GetUserIdAndEmailByOtpCodeRow{Email: "other@example.com"})

	err := svc.VerifyOTP(ctx, "user@example.com", "ABC123")
	require.ErrorIs(t, err, auth.ErrInvalidOtpCode)
}

func TestService_VerifyOTP_NotFound(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	fakeRepo.setGetOtpErr(auth.ErrNotFound)
	svc := auth.NewService(fakeRepo)

	err := svc.VerifyOTP(ctx, "user@example.com", "ABC123")
	require.ErrorIs(t, err, auth.ErrNotFound)
}

func TestService_VerifyOTP_RepoFailure(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	repoErr := errors.New("query failed")
	fakeRepo.setGetOtpErr(repoErr)
	svc := auth.NewService(fakeRepo)

	err := svc.VerifyOTP(ctx, "user@example.com", "ABC123")
	require.ErrorIs(t, err, repoErr)
}

func TestService_CreateNewAuthForOTPLogin_Success(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	expectedID := newPGUUID()
	fakeRepo.setNextAuthID(expectedID)
	svc := auth.NewService(fakeRepo)

	id, err := svc.CreateNewAuthForOTPLogin(ctx, "user@example.com")
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, uuidToString(expectedID), uuidToString(*id))
}

func TestService_CreateNewAuthForOTPLogin_Error(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	fakeRepo.setCreateAuthErr(qqerrors.ErrUniqueViolation)
	svc := auth.NewService(fakeRepo)

	_, err := svc.CreateNewAuthForOTPLogin(ctx, "user@example.com")
	require.ErrorIs(t, err, qqerrors.ErrUniqueViolation)
}

func TestService_KillOrphanedOTPs_Success(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	svc := auth.NewService(fakeRepo)

	err := svc.KillOrphanedOTPs(ctx, "user@example.com")
	require.NoError(t, err)

	emails := fakeRepo.killOrphanedEmails()
	require.Len(t, emails, 1)
	assert.Equal(t, "user@example.com", emails[0])
}

func TestService_KillOrphanedOTPs_Error(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	fakeRepo.setKillOrphanedErr(errors.New("db error"))
	svc := auth.NewService(fakeRepo)

	err := svc.KillOrphanedOTPs(ctx, "user@example.com")
	require.Error(t, err)
}

func TestService_KillOrphanedOTPsByUserID_Success(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	svc := auth.NewService(fakeRepo)
	userID := newPGUUID()

	err := svc.KillOrphanedOTPsByUserID(ctx, userID)
	require.NoError(t, err)

	ids := fakeRepo.killOrphanedUserIDs()
	require.Len(t, ids, 1)
	assert.Equal(t, uuidToString(userID), uuidToString(ids[0]))
}

func TestService_KillOrphanedOTPsByUserID_Error(t *testing.T) {
	ctx := context.Background()
	fakeRepo := newFakeRepository()
	fakeRepo.setKillOrphanedByUserIDErr(errors.New("db error"))
	svc := auth.NewService(fakeRepo)

	err := svc.KillOrphanedOTPsByUserID(ctx, newPGUUID())
	require.Error(t, err)
}

func TestService_WithTx_DelegatesToRepository(t *testing.T) {
	fakeRepo := newFakeRepository()
	svc := auth.NewService(fakeRepo)

	require.Equal(t, 0, fakeRepo.withTxCount())

	svc.WithTx(nil)

	require.Equal(t, 1, fakeRepo.withTxCount())
}
