package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/ports"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationUsecase interface {
	RegisterOrLoginOTP(ctx context.Context, email string) error
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

func (uc *registrationUsecase) RegisterOrLoginOTP(ctx context.Context, email string) error {
	tx, err := uc.dbpool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		// Only rollback if transaction is still active (not committed)
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	txAuthService := uc.authService.WithTx(tx)
	txUserService := uc.userService.WithTx(tx)

	// Try to get existing user first
	foundUser, err := txUserService.GetUserByEmail(ctx, email)
	if err != nil && err != user.ErrNotFound {
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	var authID uuid.UUID
	// If user doesn't exist, create auth record first, then user
	if foundUser.ID == uuid.Nil {
		// Create auth record with email
		authIDPtr, err := txAuthService.CreateNewAuthForOTPLogin(ctx, email)
		if err != nil {
			return fmt.Errorf("failed to create auth record: %w", err)
		}
		authID = *authIDPtr

		// Create user with the auth_id
		foundUser, err = txUserService.CreateDefaultUserWithAuthID(ctx, authID)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists, get their auth_id
		authID = foundUser.AuthID
	}

	err = uc.authService.KillOrphanedOTPsByUserID(ctx, foundUser.ID)
	if err != nil {
		return fmt.Errorf("failed to kill orphaned otps: %w", err)
	}

	// Generate and save OTP using auth_id (not user_id)
	otp, err := txAuthService.GenerateAndSaveOTPForAuth(ctx, authID)
	if err != nil {
		return fmt.Errorf("failed to generate and save otp: %w", err)
	}

	// Get email template (this doesn't need transaction)
	template, err := uc.mailer.GetEmailTemplate(ctx, "otp")
	if err != nil {
		return fmt.Errorf("failed to get email template: %w", err)
	}

	// Commit transaction before sending email
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	tx = nil // Mark transaction as committed so defer won't try to rollback

	// Send email (this should be outside transaction)
	body := strings.Replace(template, "{{.OTP}}", otp, 1)
	err = uc.mailer.SendEmail(ctx, ports.SendEmailParams{
		From:    "noreply@qq.com",
		To:      email,
		Subject: "OTP Verification",
		Body:    body,
	})
	if err != nil {
		// Note: Transaction is already committed, so OTP is saved
		// This is intentional - we don't want to rollback user creation
		// just because email failed. Log the error and continue.
		return fmt.Errorf("OTP created successfully but failed to send email: %w", err)
	}

	return nil
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

	// Verify OTP code
	err = txAuthService.VerifyOTP(ctx, email, otp)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid OTP: %w", err)
	}

	// Get user by email (should exist since OTP was valid)
	user, err := txUserService.GetUserByEmail(ctx, email)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Clean up used OTP codes for this user
	err = txAuthService.KillOrphanedOTPsByUserID(ctx, user.ID)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to clean up OTP codes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	tx = nil

	// Generate JWT tokens (outside transaction)
	tokens, err := uc.tokenService.GenerateTokens(ctx, ports.GenerateTokenParams{
		UserID: user.ID.String(),
	})
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}
func (uc *registrationUsecase) RefreshTokens(ctx context.Context, refreshToken string) (ports.GenerateTokenResult, error) {
	// Validate the refresh token
	tokenResult, err := uc.tokenService.ValidateToken(ctx, ports.ValidateTokenParams{
		Token: refreshToken,
	})
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Extract user ID from claims
	userID := tokenResult.Claims.UserID
	if userID == "" {
		return ports.GenerateTokenResult{}, fmt.Errorf("invalid token: missing user ID")
	}

	// Verify user still exists
	user, err := uc.userService.GetUserByID(ctx, userID)
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("user not found: %w", err)
	}

	// Generate new token pair
	newTokens, err := uc.tokenService.GenerateTokens(ctx, ports.GenerateTokenParams{
		UserID: user.ID.String(),
	})
	if err != nil {
		return ports.GenerateTokenResult{}, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return newTokens, nil
}
