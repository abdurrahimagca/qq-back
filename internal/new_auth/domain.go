package newauth

import (
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