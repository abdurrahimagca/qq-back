package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	CreateAuthForOTPLogin(ctx context.Context, email string) (uuid.UUID, error)
	CreateOTP(ctx context.Context, userID uuid.UUID, otpHash string) error
	GetUserIdAndEmailByOtpCode(ctx context.Context, otpHash string) (GetUserIdAndEmailByOtpCodeResult, error)
	KillOrphanedOTPs(ctx context.Context, email string) error
	KillOrphanedOTPsByUserID(ctx context.Context, userID pgtype.UUID) error
}

type pgxRepository struct {
	q *db.Queries
}

func NewPgxRepository(pool *pgxpool.Pool) Repository {
	return &pgxRepository{
		q: db.New(pool),
	}
}

func (r *pgxRepository) WithTx(tx pgx.Tx) Repository {
	return &pgxRepository{
		q: r.q.WithTx(tx),
	}
}

func (r *pgxRepository) CreateAuthForOTPLogin(ctx context.Context, email string) (uuid.UUID, error) {
	id, err := r.q.InsertAuth(ctx, db.InsertAuthParams{
		Email:    email,
		Provider: "email_otp",
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("CreateOtpAuth failed: %w", err)
	}
	return uuid.UUID(id.Bytes), nil
}

func (r *pgxRepository) CreateOTP(ctx context.Context, userID uuid.UUID, otpHash string) error {
	_, err := r.q.InsertAuthOtpCode(ctx, db.InsertAuthOtpCodeParams{
		AuthID: pgtype.UUID{Bytes: userID, Valid: true},
		Code:   otpHash,
	})
	if err != nil {
		return fmt.Errorf("CreateOTP failed: %w", err)
	}
	return nil
}

func (r *pgxRepository) GetUserIdAndEmailByOtpCode(ctx context.Context, otpHash string) (GetUserIdAndEmailByOtpCodeResult, error) {
	row, err := r.q.GetUserIdAndEmailByOtpCode(ctx, otpHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetUserIdAndEmailByOtpCodeResult{}, ErrNotFound
		}
		return GetUserIdAndEmailByOtpCodeResult{}, fmt.Errorf("GetUserIdAndEmailByOtpCode failed: %w", err)
	}
	return GetUserIdAndEmailByOtpCodeResult{
		ID:    uuid.UUID(row.ID.Bytes),
		Email: row.Email,
	}, nil
}

func (r *pgxRepository) KillOrphanedOTPs(ctx context.Context, email string) error {
	 err := r.q.DeleteOtpCodesByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("KillOrphanedOTPs failed: %w", err)
	}
	return nil
}

func (r *pgxRepository) KillOrphanedOTPsByUserID(ctx context.Context, userID pgtype.UUID) error {
	err := r.q.DeleteOtpCodesByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("DeleteOtpCodesByUserID failed: %w", err)
	}
	return nil
}