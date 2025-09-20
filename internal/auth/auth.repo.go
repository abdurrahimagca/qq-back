package auth

import (
	"context"
	"errors"

	"github.com/abdurrahimagca/qq-back/internal/db"
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	CreateAuthForOTPLogin(ctx context.Context, email string) (*pgtype.UUID, error)
	CreateOTP(ctx context.Context, userID pgtype.UUID, otpHash string) error
	GetUserIdAndEmailByOtpCode(ctx context.Context, otpHash string) (db.GetUserIdAndEmailByOtpCodeRow, error)
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

func (r *pgxRepository) CreateAuthForOTPLogin(ctx context.Context, email string) (*pgtype.UUID, error) {
	id, err := r.q.InsertAuth(ctx, db.InsertAuthParams{
		Email:    email,
		Provider: "email_otp",
	})
	if err != nil {
		return nil, qqerrors.GetDbErrAsQQError(err)
	}
	return &id, nil
}

func (r *pgxRepository) CreateOTP(ctx context.Context, userID pgtype.UUID, otpHash string) error {
	_, err := r.q.InsertAuthOtpCode(ctx, db.InsertAuthOtpCodeParams{
		AuthID: userID,
		Code:   otpHash,
	})
	if err != nil {
		return qqerrors.GetDbErrAsQQError(err)
	}
	return nil
}

func (r *pgxRepository) GetUserIdAndEmailByOtpCode(
	ctx context.Context,
	otpHash string,
) (db.GetUserIdAndEmailByOtpCodeRow, error) {
	row, err := r.q.GetUserIdAndEmailByOtpCode(ctx, otpHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetUserIdAndEmailByOtpCodeRow{}, ErrNotFound
		}
		return db.GetUserIdAndEmailByOtpCodeRow{}, qqerrors.GetDbErrAsQQError(err)
	}
	return row, nil
}
func (r *pgxRepository) KillOrphanedOTPs(ctx context.Context, email string) error {
	err := r.q.DeleteOtpCodesByEmail(ctx, email)
	if err != nil {
		return qqerrors.GetDbErrAsQQError(err)
	}
	return nil
}

func (r *pgxRepository) KillOrphanedOTPsByUserID(ctx context.Context, userID pgtype.UUID) error {
	err := r.q.DeleteOtpCodesByUserID(ctx, userID)
	if err != nil {
		return qqerrors.GetDbErrAsQQError(err)
	}
	return nil
}
