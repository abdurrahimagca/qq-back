package newauth

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateNewUserOtp(ctx context.Context, tx pgx.Tx, params CreateNewUserOtpParams) (CreateNewUserResult, error)
	GetUserByEmail(ctx context.Context, params GetUserByEmailParams) (GetUserByEmailResult, error)
	InsertNewOtpCodeForUser(ctx context.Context, params InsertNewOtpCodeForUserParams) (InsertNewOtpCodeForUserResult, error)
	DeleteOtpCodesByEmail(ctx context.Context, params DeleteOtpCodesByEmailParams) (DeleteOtpCodesByEmailResult, error)
	GetUserIdAndEmailByOtpCodeAndDelete(tx pgx.Tx, ctx context.Context, params GetUserIdAndEmailByOtpCodeParams) (GetUserIdAndEmailByOtpCodeResult, error)
	GetUserByID(ctx context.Context, params GetUserByIDParams) (GetUserByIDResult, error)
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
		Email:    params.Email,
		Provider: db.AuthProviderEmailOtp,
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
func (r *pgxRepository) GetUserByEmail(ctx context.Context, params GetUserByEmailParams) (GetUserByEmailResult, error) {

	user, err := r.q.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return GetUserByEmailResult{}, err
	}

	return GetUserByEmailResult{
		User:  user,
		Email: params.Email,
	}, nil
}
func (r *pgxRepository) GetUserIdAndEmailByOtpCodeAndDelete(tx pgx.Tx, ctx context.Context, params GetUserIdAndEmailByOtpCodeParams) (GetUserIdAndEmailByOtpCodeResult, error) {
	queries := db.New(tx)
	row, err := queries.GetUserIdAndEmailByOtpCode(ctx, params.OtpCode)

	if err != nil {
		return GetUserIdAndEmailByOtpCodeResult{}, err
	}
	_, err = queries.DeleteOtpCodesByEmail(ctx, row.Email)
	if err != nil {
		return GetUserIdAndEmailByOtpCodeResult{}, err
	}

	return GetUserIdAndEmailByOtpCodeResult{
		UserID: &row.ID,
		Email:  row.Email,
	}, nil

}

func (r *pgxRepository) GetUserByID(ctx context.Context, params GetUserByIDParams) (GetUserByIDResult, error) {
	user, err := r.q.GetUserByID(ctx, params.ID)
	if err != nil {
		return GetUserByIDResult{}, err
	}
	return GetUserByIDResult{
		User: user,
	}, nil
}

