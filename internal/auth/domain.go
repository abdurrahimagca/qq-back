package newauth

import (
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/tokens"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateNewUserOtpParams struct {
	Email      string
	Username   string
	OtpCode    string
}

type CreateNewUserResult struct {
	UserID    *pgtype.UUID
	OtpCodeID *pgtype.UUID
}

type InsertNewOtpCodeForUserParams struct {
	UserID pgtype.UUID
	Code   string
}

type InsertNewOtpCodeForUserResult struct {
	OtpCodeID *pgtype.UUID
}

type DeleteOtpCodesByEmailParams struct {
	Email string
}

type DeleteOtpCodesByEmailResult struct {
	DeletedCount int64
}

type LoginOtpParams struct {
	Email string
}

type LoginOtpResult struct {
	UserID *pgtype.UUID
}

type GetUserByEmailParams struct {
	Email string
}

type GetUserByEmailResult struct {
	User db.User
	Email string
}

type VerifyOtpCodeParams struct {
	Email string
	Code  string
}

type VerifyOtpCodeResult struct {
	Tokens *tokens.Tokens
}

type GetUserIdAndEmailByOtpCodeParams struct {
	OtpCode string
}

type GetUserIdAndEmailByOtpCodeResult struct {
	UserID *pgtype.UUID
	Email  string
}

type RefreshTokensParams struct {
	RefreshToken string
}

type RefreshTokensResult struct {
	Tokens *tokens.Tokens
}

type GetUserByIDParams struct {
	ID pgtype.UUID
}

type GetUserByIDResult struct {
	User db.User
}

type ValidateAccessAndGetUserParams struct {
	AccessToken string
}

type ValidateAccessAndGetUserResult struct {
	User *db.User
}