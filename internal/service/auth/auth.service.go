package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/external/mail"
	"github.com/abdurrahimagca/qq-back/internal/repository/auth"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func createFirstTimeUserWithOtp(email string, tx pgx.Tx, config *environment.Config) error {
	userName := strings.Split(email, "@")[0] + "_" + uuid.New().String()[:8]

	otpCode := uuid.New().String()[:6]

	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])
	provider := strings.Split(email, "@")[1]

	_, err := auth.CreateFirstTimeUserWithOtp(context.Background(), tx, auth.CreateFirstTimeUserParams{
		Email:      email,
		Provider:   provider,
		Username:   userName,
		AuthType:   db.AuthProviderEmailOtp,
		OtpCode:    hashedOtpCode,
		ProviderID: "",
	})

	if err != nil {
		return err
	}
	err = mail.SendOTPMail(context.Background(), mail.SendOTPMailParams{
		To:     email,
		Code:   otpCode,
		Config: config,
	})

	if err != nil {
		return err
	}

	return nil
}
func handleAlreadyExistsUser(email string, userID pgtype.UUID, tx pgx.Tx, config *environment.Config) error {
	otpCode := uuid.New().String()[:6]

	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])

	err := auth.InsertNewOtpCodeForUser(context.Background(), tx, userID, hashedOtpCode)

	if err != nil {
		return err
	}

	err = mail.SendOTPMail(context.Background(), mail.SendOTPMailParams{
		To:     email,
		Code:   otpCode,
		Config: config,
	})

	if err != nil {
		return err
	}

	return nil
}

func CreateUserIfNotExistWithOtpService(email string, tx pgx.Tx, config *environment.Config) error {
	user, _ := auth.GetUserByEmail(context.Background(), tx, email)
	if user != nil {
		return handleAlreadyExistsUser(email, user.ID, tx, config)

	}

	err := createFirstTimeUserWithOtp(email, tx, config)

	if err != nil {
		return err
	}

	return nil
}
