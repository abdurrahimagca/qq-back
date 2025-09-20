package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service arayüzü, auth domain'inin temel yeteneklerini tanımlar.
type Service interface {
	WithTx(tx pgx.Tx) Service
	GenerateAndSaveOTPForAuth(ctx context.Context, authID pgtype.UUID) (string, error)
	VerifyOTP(ctx context.Context, email string, otpCode string) error
	KillOrphanedOTPsByUserID(ctx context.Context, userID pgtype.UUID) error
	KillOrphanedOTPs(ctx context.Context, email string) error
	CreateNewAuthForOTPLogin(ctx context.Context, email string) (*pgtype.UUID, error)
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

func (s *service) CreateNewAuthForOTPLogin(ctx context.Context, email string) (*pgtype.UUID, error) {
	id, err := s.repo.CreateAuthForOTPLogin(ctx, email)
	if err != nil {
		return nil, err
	}
	return id, nil
}



func (s *service) GenerateAndSaveOTPForAuth(ctx context.Context, authID pgtype.UUID) (string, error) {
	// Generate a random string for OTP code
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	// Convert to alphanumeric characters
	for i, b := range randomBytes {
		randomBytes[i] = 'A' + (b % 26)
	}
	otpCode := strings.ToUpper("QQ" + string(randomBytes))
	otpHash := sha256.Sum256([]byte(otpCode))

	if err := s.repo.CreateOTP(ctx, authID, hex.EncodeToString(otpHash[:])); err != nil {
		return "", err
	}

	return otpCode, nil
}

func (s *service) KillOrphanedOTPsByUserID(ctx context.Context, userID pgtype.UUID) error {
	return s.repo.KillOrphanedOTPsByUserID(ctx, userID)
}

func (s *service) KillOrphanedOTPs(ctx context.Context, email string) error {
	return s.repo.KillOrphanedOTPs(ctx, email)
}

func (s *service) VerifyOTP(ctx context.Context, email string, otpCode string) error {
	otpHash := sha256.Sum256([]byte(otpCode))
	usr, err := s.repo.GetUserIdAndEmailByOtpCode(ctx, hex.EncodeToString(otpHash[:]))
	if err != nil {
		return err
	}
	if usr.Email != email {
		return ErrInvalidOtpCode
	}
	return nil

}
