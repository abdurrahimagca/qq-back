package auth

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

type CreateFirstTimeUserParams struct {
	Email      string
	Provider   string
	ProviderID string
	Username   string
	AuthType   db.AuthProvider
	OtpCode    string
}

type CreateFirstTimeUserWithOtpResult struct {
	UserID    *pgtype.UUID
	OtpCodeID *pgtype.UUID
}

func (r *AuthRepository) CreateFirstTimeUserWithOtp(ctx context.Context, tx pgx.Tx, params CreateFirstTimeUserParams) (CreateFirstTimeUserWithOtpResult, error) {
	queries := db.New(tx)

	authId, err := queries.InsertAuth(ctx, db.InsertAuthParams{
		Email:      params.Email,
		Provider:   params.AuthType,
		ProviderID: pgtype.Text{String: params.ProviderID, Valid: params.ProviderID != ""},
	})
	if err != nil {
		return CreateFirstTimeUserWithOtpResult{}, err
	}

	userID, err := queries.InsertUser(ctx, db.InsertUserParams{
		AuthID:   authId,
		Username: params.Username,
	})
	if err != nil {
		return CreateFirstTimeUserWithOtpResult{}, err
	}

	otpCode, err := queries.InsertAuthOtpCode(ctx, db.InsertAuthOtpCodeParams{
		AuthID: authId,
		Code:   params.OtpCode,
	})
	if err != nil {
		return CreateFirstTimeUserWithOtpResult{}, err
	}

	return CreateFirstTimeUserWithOtpResult{
		UserID:    &userID,
		OtpCodeID: &otpCode,
	}, nil
}

func (r *AuthRepository) GetUserIdAndEmailByOtpCode(ctx context.Context, tx pgx.Tx, code string) (*db.GetUserIdAndEmailByOtpCodeRow, error) {
	queries := db.New(tx)

	user, err := queries.GetUserIdAndEmailByOtpCode(ctx, code)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, tx pgx.Tx, email string) (*db.User, error) {
	queries := db.New(tx)

	user, err := queries.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) InsertNewOtpCodeForUser(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, code string) error {
	queries := db.New(tx)

	// Get the user to find their auth_id
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	_, err = queries.InsertAuthOtpCode(ctx, db.InsertAuthOtpCodeParams{
		AuthID: user.AuthID,
		Code:   code,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) DeleteOtpCodeById(ctx context.Context, tx pgx.Tx, id pgtype.UUID) error {
	queries := db.New(tx)

	err := queries.DeleteOtpCodeById(ctx, id)

	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) DeleteOtpCodeEntryByAuthID(ctx context.Context, tx pgx.Tx, authID pgtype.UUID) error {
	queries := db.New(tx)

	err := queries.DeleteOtpCodeEntryByAuthID(ctx, authID)

	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, dbtx db.DBTX, userID pgtype.UUID) (*db.User, error) {
	queries := db.New(dbtx)

	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
