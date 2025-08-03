package auth

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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

func CreateFirstTimeUserWithOtp(ctx context.Context, tx pgx.Tx, params CreateFirstTimeUserParams) (CreateFirstTimeUserWithOtpResult, error) {
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

func GetUserIdAndEmailByOtpCode(ctx context.Context, tx pgx.Tx, code string) (*db.GetUserIdAndEmailByOtpCodeRow, error) {
	queries := db.New(tx)

	user, err := queries.GetUserIdAndEmailByOtpCode(ctx, code)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(ctx context.Context, tx pgx.Tx, email string) (*db.User, error) {
	queries := db.New(tx)

	user, err := queries.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func InsertNewOtpCodeForUser(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, code string) error {
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

func DeleteOtpCodeById(ctx context.Context, tx pgx.Tx, id pgtype.UUID) error {
	queries := db.New(tx)

	err := queries.DeleteOtpCodeById(ctx, id)

	if err != nil {
		return err
	}

	return nil
}
