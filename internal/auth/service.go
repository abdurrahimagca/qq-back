package auth

import (
	"context"
	"crypto/sha256"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service arayüzü, auth domain'inin temel yeteneklerini tanımlar.
type Service interface {
	WithTx(tx pgx.Tx) Service
	GenerateAndSaveOTP(ctx context.Context, userID uuid.UUID) (string, error) // Deprecated - use GenerateAndSaveOTPForAuth
	GenerateAndSaveOTPForAuth(ctx context.Context, authID uuid.UUID) (string, error)
	VerifyOTP(ctx context.Context, email string, otpCode string) error
	KillOrphanedOTPsByUserID(ctx context.Context, userID uuid.UUID) error
	KillOrphanedOTPs(ctx context.Context, email string) error
	CreateNewAuthForOTPLogin(ctx context.Context, email string) (*uuid.UUID, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}
func (s *service) WithTx(tx pgx.Tx) Service {
	return &service{repo: s.repo.WithTx(tx)}
}

func (s *service) CreateNewAuthForOTPLogin(ctx context.Context, email string) (*uuid.UUID, error) {
	id, err := s.repo.CreateAuthForOTPLogin(ctx, email)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (s *service) GenerateAndSaveOTP(ctx context.Context, userID uuid.UUID) (string, error) {
	// Deprecated: This method incorrectly treats userID as authID
	// Use GenerateAndSaveOTPForAuth instead
	otpCode := strings.ToUpper("QQ" + uuid.NewString()[:6])
	otpHash := sha256.Sum256([]byte(otpCode))

	if err := s.repo.CreateOTP(ctx, userID, string(otpHash[:])); err != nil {
		return "", err
	}

	return otpCode, nil
}

func (s *service) GenerateAndSaveOTPForAuth(ctx context.Context, authID uuid.UUID) (string, error) {
	otpCode := strings.ToUpper("QQ" + uuid.NewString()[:6])
	otpHash := sha256.Sum256([]byte(otpCode))

	if err := s.repo.CreateOTP(ctx, authID, string(otpHash[:])); err != nil {
		return "", err
	}

	return otpCode, nil
}

func (s *service) KillOrphanedOTPsByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.repo.KillOrphanedOTPsByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
}

func (s *service) KillOrphanedOTPs(ctx context.Context, email string) error {
	return s.repo.KillOrphanedOTPs(ctx, email)
}

func (s *service) VerifyOTP(ctx context.Context, email string, otpCode string) error {
	otpHash := sha256.Sum256([]byte(otpCode))
	usr, err := s.repo.GetUserIdAndEmailByOtpCode(ctx, string(otpHash[:]))
	if err != nil {
		return err
	}
	if usr.Email != email {
		return ErrInvalidOtpCode
	}
	return nil

}
