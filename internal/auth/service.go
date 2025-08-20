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
	GenerateAndSaveOTP(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (string, error)
	VerifyOTP(ctx context.Context, tx pgx.Tx, email string, otpCode string) error
	CreateNewAuthForOTPLogin(ctx context.Context, tx pgx.Tx, email string) (*uuid.UUID, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateNewAuthForOTPLogin(ctx context.Context, tx pgx.Tx, email string) (*uuid.UUID, error) {
	id, err := s.repo.CreateAuthForOTPLogin(ctx, email)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (s *service) GenerateAndSaveOTP(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (string, error) {
	// Transaction'a bağlı bir repo al
	txRepo := s.repo.WithTx(tx)

	// 1. İş Kuralı: Önce kullanıcının mevcut, geçerli tüm OTP'lerini sil.
	if err := txRepo.KillOrphanedOTPsByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true}); err != nil { // repo'da böyle bir metot olmalı
		return "", err
	}

	// 2. İş Kuralı: Yeni OTP kodunu ve hash'ini oluştur.
	otpCode := strings.ToUpper("QQ" + uuid.NewString()[:6])
	otpHash := sha256.Sum256([]byte(otpCode))

	// 3. Yeni OTP'yi DB'ye kaydet.
	if err := txRepo.CreateOTP(ctx, userID, string(otpHash[:])); err != nil {
		return "", err
	}

	// Ham kodu (hash'lenmemiş) geri dön, çünkü bu e-posta ile gönderilecek.
	return otpCode, nil
}

func (s *service) VerifyOTP(ctx context.Context, tx pgx.Tx, email string, otpCode string) error {
	txRepo := s.repo.WithTx(tx)

	// İş Kuralı: Gelen kodu hash'le ve DB'deki ile karşılaştır.
	otpHash := sha256.Sum256([]byte(otpCode))
	usr, err := txRepo.GetUserIdAndEmailByOtpCode(ctx, string(otpHash[:]))
	if err != nil {
		return err
	}
	if usr.Email != email {
		return ErrInvalidOtpCode
	}
	if err := txRepo.KillOrphanedOTPs(ctx, usr.Email); err != nil {
		return err
	}
	return nil

}
