package server

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/api"
)

func (s *Server) PostAuthOtp(ctx context.Context, request api.PostAuthOtpRequestObject) (api.PostAuthOtpResponseObject, error) {
	if request.Body == nil {
		return api.PostAuthOtp400JSONResponse{
			ErrorCode: "INVALID_REQUEST",
			Message:   "Request body is required",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	email := request.Body.Email
	if email == "" {
		return api.PostAuthOtp400JSONResponse{
			ErrorCode: "INVALID_EMAIL",
			Message:   "Email is required",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	err := s.registrationUC.RegisterOrLoginOTP(ctx, email)
	if err != nil {
		return api.PostAuthOtp500JSONResponse{
			ErrorCode: "INTERNAL_ERROR",
			Message:   "Failed to send OTP: " + err.Error(),
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	isNewUser := true
	return api.PostAuthOtp200JSONResponse{
		Data: &struct {
			IsNewUser *bool `json:"isNewUser,omitempty"`
		}{
			IsNewUser: &isNewUser,
		},
		Message:   "OTP sent successfully",
		Success:   true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) PostAuthOtpVerify(ctx context.Context, request api.PostAuthOtpVerifyRequestObject) (api.PostAuthOtpVerifyResponseObject, error) {
	if request.Body == nil {
		return api.PostAuthOtpVerify400JSONResponse{
			ErrorCode: "INVALID_REQUEST",
			Message:   "Request body is required",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	email := request.Body.Email
	otpCode := request.Body.OtpCode

	if email == "" || otpCode == "" {
		return api.PostAuthOtpVerify400JSONResponse{
			ErrorCode: "INVALID_REQUEST",
			Message:   "Email and OTP code are required",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	tokens, err := s.registrationUC.VerifyOTPAndLogin(ctx, email, otpCode)
	if err != nil {
		return api.PostAuthOtpVerify400JSONResponse{
			ErrorCode: "INVALID_OTP",
			Message:   "Invalid OTP code: " + err.Error(),
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	return api.PostAuthOtpVerify200JSONResponse{
		Data: &struct {
			AccessToken  *string `json:"accessToken,omitempty"`
			RefreshToken *string `json:"refreshToken,omitempty"`
		}{
			AccessToken:  &tokens.AccessToken,
			RefreshToken: &tokens.RefreshToken,
		},
		Message:   "Login successful",
		Success:   true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) PostAuthRefreshToken(ctx context.Context, request api.PostAuthRefreshTokenRequestObject) (api.PostAuthRefreshTokenResponseObject, error) {
	if request.Body == nil {
		return api.PostAuthRefreshToken400JSONResponse{
			ErrorCode: "INVALID_REQUEST",
			Message:   "Request body is required",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	refreshToken := request.Body.RefreshToken
	if refreshToken == "" {
		return api.PostAuthRefreshToken400JSONResponse{
			ErrorCode: "INVALID_REQUEST",
			Message:   "Refresh token is required",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	newTokens, err := s.registrationUC.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return api.PostAuthRefreshToken401JSONResponse{
			Message:   stringPtr("Invalid or expired refresh token: " + err.Error()),
			Timestamp: stringPtr(time.Now().UTC().Format(time.RFC3339)),
		}, nil
	}

	return api.PostAuthRefreshToken200JSONResponse{
		AccessToken:  &newTokens.AccessToken,
		RefreshToken: &newTokens.RefreshToken,
		Data:         &map[string]interface{}{},
		Message:      "Tokens refreshed successfully",
		Success:      true,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}