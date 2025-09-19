package server

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/api"
)

func (s *Server) SendOtp(ctx context.Context, request api.SendOtpRequestObject) (api.SendOtpResponseObject, error) {
	
	if err := s.validator.Struct(request.Body); err != nil {
		now := time.Now().UTC()
		return api.SendOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr(validationErrorMessage(err)),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	email := string(request.Body.Email)

	isNewUser, err := s.registrationUC.RegisterOrLoginOTP(ctx, email)
	if err != nil {
		now := time.Now().UTC()
		return api.SendOtp500ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Failed to send OTP: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	now := time.Now().UTC()
	return api.SendOtp200JSONResponse{
		Data: &struct {
			IsNewUser *bool `json:"isNewUser,omitempty"`
		}{
			IsNewUser: isNewUser,
		},
		Message:   stringPtr("OTP sent successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(now),
	}, nil
}

func (s *Server) VerifyOtp(ctx context.Context, request api.VerifyOtpRequestObject) (api.VerifyOtpResponseObject, error) {
	if request.Body == nil {
		now := time.Now().UTC()
		return api.VerifyOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Request body is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	if err := s.validator.Struct(request.Body); err != nil {
		now := time.Now().UTC()
		return api.VerifyOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr(validationErrorMessage(err)),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	email := string(request.Body.Email)
	otpCode := request.Body.OtpCode

	tokens, err := s.registrationUC.VerifyOTPAndLogin(ctx, email, otpCode)
	if err != nil {
		now := time.Now().UTC()
		return api.VerifyOtp400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Invalid OTP code: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	now := time.Now().UTC()
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
		Timestamp: timeStrPtr(now),
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	if err := s.validator.Struct(request.Body); err != nil {
		now := time.Now().UTC()
		return api.RefreshToken400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr(validationErrorMessage(err)),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	refreshToken := request.Body.RefreshToken
	newTokens, err := s.registrationUC.RefreshTokens(ctx, refreshToken)
	if err != nil {
		now := time.Now().UTC()
		return api.RefreshToken401ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Invalid or expired refresh token: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	now := time.Now().UTC()
	return api.RefreshToken200JSONResponse{
		Data: &map[string]interface{}{
			"accessToken":  newTokens.AccessToken,
			"refreshToken": newTokens.RefreshToken,
		},
		Message:   stringPtr("Tokens refreshed successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(now),
	}, nil
}
