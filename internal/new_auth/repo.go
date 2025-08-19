package newauth

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateNewUserOtp(ctx context.Context, tx pgx.Tx, params CreateNewUserOtpParams) (CreateNewUserResult, error)
	InsertNewOtpCodeForUser(ctx context.Context, params InsertNewOtpCodeForUserParams) (InsertNewOtpCodeForUserResult, error)
	DeleteOtpCodesByEmail(ctx context.Context, params DeleteOtpCodesByEmailParams) (DeleteOtpCodesByEmailResult, error)
}
type pgxRepository struct {
	q *db.Queries
}

func NewPgxRepository(pool *pgxpool.Pool) Repository {
	return &pgxRepository{
		q: db.New(pool),
	}
}

func (r *pgxRepository) CreateNewUserOtp(ctx context.Context, tx pgx.Tx, params CreateNewUserOtpParams) (CreateNewUserResult, error) {
	queries := db.New(tx)

	authId, err := queries.InsertAuth(ctx, db.InsertAuthParams{
		Email:      params.Email,
		Provider:   db.AuthProviderEmailOtp,
	})
	if err != nil {
		return CreateNewUserResult{}, err
	}

	userID, err := queries.InsertUser(ctx, db.InsertUserParams{
		AuthID:   authId,
		Username: params.Username,
	})
	if err != nil {
		return CreateNewUserResult{}, err
	}

	otpCode, err := queries.InsertAuthOtpCode(ctx, db.InsertAuthOtpCodeParams{
		AuthID: authId,
		Code:   params.OtpCode,
	})
	if err != nil {
		return CreateNewUserResult{}, err
	}

	return CreateNewUserResult{
		UserID:    &userID,
		OtpCodeID: &otpCode,
	}, nil
}

func (r *pgxRepository) InsertNewOtpCodeForUser(ctx context.Context, params InsertNewOtpCodeForUserParams) (InsertNewOtpCodeForUserResult, error) {
	otpCode, err := r.q.InsertAuthOtpCode(ctx, db.InsertAuthOtpCodeParams{
		AuthID: params.UserID,
		Code:   params.Code,
	})
	if err != nil {
		return InsertNewOtpCodeForUserResult{}, err
	}

	return InsertNewOtpCodeForUserResult{
		OtpCodeID: &otpCode,
	}, nil
}
func (r *pgxRepository) DeleteOtpCodesByEmail(ctx context.Context, params DeleteOtpCodesByEmailParams) (DeleteOtpCodesByEmailResult, error) {
	deletedCount, err := r.q.DeleteOtpCodesByEmail(ctx, params.Email)
	if err != nil {
		return DeleteOtpCodesByEmailResult{}, err
	}

	return DeleteOtpCodesByEmailResult{
		DeletedCount: deletedCount,
	}, nil
}
