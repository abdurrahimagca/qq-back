package server

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/api"
)

func (s *Server) SendOtp(ctx context.Context, request api.SendOtpRequestObject) (api.SendOtpResponseObject, error) {
	if request.Body == nil {
		return api.SendOtp400ApplicationProblemPlusJSONResponse{
			
			Message:   stringPtr("Request body is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC	()),
		}, nil
	}

	email := request.Body.Email
	if email == "" {
		return api.SendOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Email is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	err := s.registrationUC.RegisterOrLoginOTP(ctx, email)
	if err != nil {
		return api.SendOtp500ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Failed to send OTP: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	isNewUser := true
	return api.SendOtp200JSONResponse{
		Data: &struct {
			IsNewUser *bool `json:"isNewUser,omitempty"`
		}{
			IsNewUser: &isNewUser,
		},
		Message:   stringPtr("OTP sent successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(time.Now().UTC()),
	}, nil
}

func (s *Server) VerifyOtp(ctx context.Context, request api.VerifyOtpRequestObject) (api.VerifyOtpResponseObject, error) {
	if request.Body == nil {
		return api.VerifyOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Request body is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	email := request.Body.Email
	otpCode := request.Body.OtpCode

	if email == "" || otpCode == "" {
		return api.VerifyOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Email and OTP code are required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	tokens, err := s.registrationUC.VerifyOTPAndLogin(ctx, email, otpCode)
	if err != nil {
		return api.VerifyOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Invalid OTP code: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	return api.VerifyOtp200JSONResponse{
		Data: &struct {
			AccessToken  *string `json:"accessToken,omitempty"`
			RefreshToken *string `json:"refreshToken,omitempty"`
		}{
			AccessToken:  &tokens.AccessToken,
			RefreshToken: &tokens.RefreshToken,
		},
		Message:   stringPtr("Login successful"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(time.Now().UTC()),
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	if request.Body == nil {
		return api.RefreshToken400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Request body is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	refreshToken := request.Body.RefreshToken
	if refreshToken == "" {
		return api.RefreshToken400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Refresh token is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	newTokens, err := s.registrationUC.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return api.RefreshToken401ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Invalid or expired refresh token: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	return api.RefreshToken200JSONResponse{
		Data: &map[string]interface{}{
			"accessToken":  newTokens.AccessToken,
			"refreshToken": newTokens.RefreshToken,
		},
		Message:   stringPtr("Tokens refreshed successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(time.Now().UTC()),
	}, nil
}