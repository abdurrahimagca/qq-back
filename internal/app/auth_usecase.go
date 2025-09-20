package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/ports"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationUsecase interface {
	RegisterOrLoginOTP(ctx context.Context, email string) (*bool, error)
	VerifyOTPAndLogin(ctx context.Context, email string, otp string) (ports.GenerateTokenResult, error)
	RefreshTokens(ctx context.Context, refreshToken string) (ports.GenerateTokenResult, error)
}

type registrationUsecase struct {
	mailer       ports.MailerPort
	authService  auth.Service
	userService  user.Service
	dbpool       *pgxpool.Pool
	tokenService ports.TokenPort
}

func NewRegistrationUsecase(mailer ports.MailerPort, authService auth.Service, userService user.Service, pool *pgxpool.Pool, tokenService ports.TokenPort) RegistrationUsecase { // <-- EKLENDİ & DÖNÜŞ TİPİ DEĞİŞTİ
	return &registrationUsecase{
		mailer:       mailer,
		authService:  authService,
		userService:  userService,
		dbpool:       pool,
		tokenService: tokenService,
	}
}

func (uc *registrationUsecase) RegisterOrLoginOTP(ctx context.Context, email string) (*bool, error) {
	tx, err := uc.dbpool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	txAuthService := uc.authService.WithTx(tx)
	txUserService := uc.userService.WithTx(tx)
	var isNewUser bool

	foundUser, err := txUserService.GetUserByEmail(ctx, email)
	if err != nil && err != user.ErrNotFound {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if foundUser.ID.Valid {
		isNewUser = false
	} else {
		isNewUser = true
	}

	var authID pgtype.UUID
	if !foundUser.ID.Valid {
		authIDPtr, err := txAuthService.CreateNewAuthForOTPLogin(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("failed to create auth record: %w", err)
		}
		authID = *authIDPtr

		foundUser, err = txUserService.CreateDefaultUserWithAuthID(ctx, authID)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		authID = foundUser.AuthID
	}

	err = txAuthService.KillOrphanedOTPsByUserID(ctx, foundUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to kill orphaned otps: %w", err)
	}

	otp, err := txAuthService.GenerateAndSaveOTPForAuth(ctx, authID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate and save otp: %w", err)
	}

	template, err := uc.mailer.GetEmailTemplate(ctx, "otp")
	if err != nil {
		return nil, fmt.Errorf("failed to get email template: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	tx = nil

	body := strings.Replace(template, "{{.OTP}}", otp, 1)
	err = uc.mailer.SendEmail(ctx, ports.SendEmailParams{
		From:    "qq@homelab-kaleici.space",
		To:      email,
		Subject: "OTP Verification",
		Body:    body,
	})
	if err != nil {
		return nil, fmt.Errorf("OTP created successfully but failed to send email: %w", err)
	}

	return &isNewUser, nil
}
func (uc *registrationUsecase) VerifyOTPAndLogin(ctx context.Context, email string, otp string) (ports.GenerateTokenResult, error) {
	tx, err := uc.dbpool.Begin(ctx)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	txAuthService := uc.authService.WithTx(tx)
	txUserService := uc.userService.WithTx(tx)

	err = txAuthService.VerifyOTP(ctx, email, otp)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid OTP: %w", err)
	}

	user, err := txUserService.GetUserByEmail(ctx, email)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to get user: %w", err)
	}

	err = txAuthService.KillOrphanedOTPsByUserID(ctx, user.ID)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to clean up OTP codes: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	tx = nil

	tokens, err := uc.tokenService.GenerateTokens(ctx, ports.GenerateTokenParams{
		UserID: user.ID.String(),
	})
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}
func (uc *registrationUsecase) RefreshTokens(ctx context.Context, refreshToken string) (ports.GenerateTokenResult, error) {
	tokenResult, err := uc.tokenService.ValidateToken(ctx, ports.ValidateTokenParams{
		Token: refreshToken,
	})
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	userID := tokenResult.Claims.UserID
	if userID == "" {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid token: missing user ID")
	}

	userUUID := pgtype.UUID{}
	if err := userUUID.Scan(userID); err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid user ID format: %w", err)
	}
	user, err := uc.userService.GetUserByID(ctx, userUUID)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("user not found: %w", err)
	}

	newTokens, err := uc.tokenService.GenerateTokens(ctx, ports.GenerateTokenParams{
		UserID: user.ID.String(),
	})
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return newTokens, nil
}
