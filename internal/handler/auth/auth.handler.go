package auth

import (
	"context"
	"regexp"

	"time"

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

func (h *AuthHandler) PostAuthOtp(ctx context.Context, request api.PostAuthOtpRequestObject) (api.PostAuthOtpResponseObject, error) {
	const email_regex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	timestamp := time.Now().Format(time.RFC3339)

	if request.Body.Email == "" {
		message := "Email is required, please check your request"
		success := false
		return api.PostAuthOtp400JSONResponse{
			Message:   message,
			Success:   success,
			ErrorCode: "OTP_MAIL_REQUIRED_1",
			Timestamp: timestamp,
		}, nil
	}

	if !regexp.MustCompile(email_regex).MatchString(request.Body.Email) {
		return api.PostAuthOtp400JSONResponse{
			Message:   "Invalid email address, please check your request the regex for an email should be `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`",
			Success:   false,
			ErrorCode: "OTP_MAIL_INVALID_1",
			Timestamp: timestamp,
		}, nil
	}

	result, err := h.authService.CreateUserIfNotExistWithOtpService(ctx, request.Body.Email)
	if err != nil {
		return api.PostAuthOtp500JSONResponse{
			Message:   "Internal server error orijinal error: " + err.Error(),
			Success:   false,
			ErrorCode: "OTP_SERVICE_FAILED_1",
			Timestamp: timestamp,
		}, nil
	}

	if result.Error != nil {
		return api.PostAuthOtp500JSONResponse{
			Message:   "Failed to process OTP request, please try again later orijinal error: " + result.Error.Error(),
			Success:   false,
			ErrorCode: "OTP_SERVICE_FAILED_RESULT_1",
			Timestamp: timestamp,
		}, nil
	}

	return api.PostAuthOtp200JSONResponse{
		Data: &struct {
			IsNewUser *bool `json:"isNewUser,omitempty"`
		}{
			IsNewUser: &result.IsNewUser,
		},
		Success:   true,
		Timestamp: timestamp,
		Message:   "OTP sent successfully",
	}, nil
}

func (h *AuthHandler) PostAuthOtpVerify(ctx context.Context, request api.PostAuthOtpVerifyRequestObject) (api.PostAuthOtpVerifyResponseObject, error) {
	timestamp := time.Now().Format(time.RFC3339)

	if request.Body.Email == "" || request.Body.OtpCode == "" {
		return api.PostAuthOtpVerify400JSONResponse{
			Message:   "Email and OTP code are required, please check your request",
			Success:   false,
			ErrorCode: "OTP_VERIFY_CODE_AND_EMAIL_REQUIRED_1",
			Timestamp: timestamp,
		}, nil
	}

	userID, _, err := h.authService.VerifyOtpCodeService(ctx, request.Body.Email, request.Body.OtpCode)
	if err != nil {
		return api.PostAuthOtpVerify400JSONResponse{
			Message:   "Invalid OTP code, please check your request orijinal error: " + err.Error(),
			Success:   false,
			ErrorCode: "OTP_VERIFY_CODE_INVALID_1",
			Timestamp: timestamp,
		}, nil
	}

	accessToken, refreshToken, err := h.authService.GenerateTokens(*userID)
	if err != nil {
		return api.PostAuthOtpVerify500JSONResponse{
			Message:   "Failed to generate tokens",
			Success:   false,
			ErrorCode: "OTP_GENERATE_TOKENS_FAILED_1",
			Timestamp: timestamp,
		}, nil
	}

	return api.PostAuthOtpVerify200JSONResponse{
		Data: &struct {
			AccessToken  *string `json:"accessToken,omitempty"`
			RefreshToken *string `json:"refreshToken,omitempty"`
		}{
			AccessToken:  &accessToken,
			RefreshToken: &refreshToken,
		},
		Success:   true,
		Timestamp: timestamp,
		Message:   "OTP verified successfully",
	}, nil
}
