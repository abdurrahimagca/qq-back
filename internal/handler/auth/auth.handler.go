package auth

import (
	"context"
	"regexp"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	authService "github.com/abdurrahimagca/qq-back/internal/service/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	authService *authService.AuthService
}

var _ api.StrictServerInterface = (*AuthHandler)(nil)

func NewAuthHandler(db *pgxpool.Pool, config *environment.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService.NewAuthService(db, config),
	}
}

// PostAuthOtp implements StrictServerInterface - this FORCES you to return the correct type!
func (h *AuthHandler) PostAuthOtp(ctx context.Context, request api.PostAuthOtpRequestObject) (api.PostAuthOtpResponseObject, error) {
	const email_regex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Validate email
	if request.Body.Email == "" {
		message := "Email is required"
		success := false
		return api.PostAuthOtp400JSONResponse{
			Message: message,
			Success: success,
		}, nil
	}

	if !regexp.MustCompile(email_regex).MatchString(request.Body.Email) {
		//message := "Invalid email address"
		success := false
		return api.PostAuthOtp400JSONResponse{
			//Message: message,
			Success: success,
		}, nil
	}

	// Call service to create user if not exists and send OTP
	result, err := h.authService.CreateUserIfNotExistWithOtpService(ctx, request.Body.Email)
	if err != nil {
		message := "Internal server error"
		success := false
		return api.PostAuthOtp500JSONResponse{
			Message: message,
			Success: success,
		}, nil
	}

	if result.Error != nil {
		message := "Failed to process OTP request"
		success := false
		return api.PostAuthOtp500JSONResponse{
			Message: message,
			Success: success,
		}, nil
	}

	success := true
	return api.PostAuthOtp200JSONResponse{
		Data: &struct {
			IsNewUser *bool `json:"isNewUser,omitempty"`
		}{
			IsNewUser: &result.IsNewUser,
		},
		Success: success,
	}, nil
}

// PostAuthOtpVerify implements StrictServerInterface - this FORCES you to return the correct type!
func (h *AuthHandler) PostAuthOtpVerify(ctx context.Context, request api.PostAuthOtpVerifyRequestObject) (api.PostAuthOtpVerifyResponseObject, error) {
	// Validate input
	if request.Body.Email == "" || request.Body.OtpCode == "" {
		message := "Email and OTP code are required"
		success := false
		return api.PostAuthOtpVerify400JSONResponse{
			Message: message,
			Success: success,
		}, nil
	}

	// Verify OTP
	userID, _, err := h.authService.VerifyOtpCodeService(ctx, request.Body.Email, request.Body.OtpCode)
	if err != nil {
		message := "Invalid OTP code"
		success := false
		return api.PostAuthOtpVerify400JSONResponse{
			Message: message,
			Success: success,
		}, nil
	}

	// Generate tokens
	accessToken, refreshToken, err := h.authService.GenerateTokens(*userID)
	if err != nil {
		message := "Failed to generate tokens"
		success := false
		return api.PostAuthOtpVerify500JSONResponse{
			Message: message,
			Success: success,
		}, nil
	}

	// SUCCESS: You MUST return the correct type or it won't compile!
	success := true
	return api.PostAuthOtpVerify200JSONResponse{
		Data: &struct {
			AccessToken  *string `json:"accessToken,omitempty"`
			RefreshToken *string `json:"refreshToken,omitempty"`
		}{
			AccessToken:  &accessToken,
			RefreshToken: &refreshToken,
		},
		Success: success,
	}, nil
}
