package server

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/abdurrahimagca/qq-back/internal/user"
)

func (s *Server) GetMeProfile(ctx context.Context, request api.GetMeProfileRequestObject) (api.GetMeProfileResponseObject, error) {
	// Get user from context (added by middleware)
	user := middleware.MustGetUserFromContext(ctx)

	if user == nil {
		now := time.Now().UTC()
		return api.GetMeProfile401JSONResponse{
			Message:   stringPtr("Unauthorized"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	now := time.Now().UTC()
	return api.GetMeProfile200JSONResponse{
		Data: &map[string]interface{}{
			"username":     user.Username,
			"displayName":  user.DisplayName,
			"privacyLevel": string(user.PrivacyLevel),
		},
		Message:   stringPtr("Profile retrieved successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(now),
	}, nil
}
func (s *Server) GetMeAvatar(ctx context.Context, request api.GetMeAvatarRequestObject) (api.GetMeAvatarResponseObject, error) {
	// Get user from context (added by middleware)
	user := middleware.MustGetUserFromContext(ctx)

	if user == nil {
		return api.GetMeAvatar401ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Unauthorized"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	if user.AvatarKey == nil {
		return api.GetMeAvatar404ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Avatar not found"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	avatarSignedUrl, err := s.fileUC.GetSignedUrlForKey(ctx, *user.AvatarKey, 30*24*time.Hour)
	if err != nil {
		return api.GetMeAvatar500ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Failed to get avatar signed url: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(time.Now().UTC()),
		}, nil
	}

	expiresInSeconds := int(30 * 24 * time.Hour / time.Second)
	return api.GetMeAvatar200JSONResponse{
		Data: &map[string]interface{}{
			"signedUrl": avatarSignedUrl,
			"expiresIn": expiresInSeconds,
		},
		Message:   stringPtr("Avatar signed url fetched successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(time.Now().UTC()),
	}, nil
}
func (s *Server) UpdateMeProfile(ctx context.Context, request api.UpdateMeProfileRequestObject) (api.UpdateMeProfileResponseObject, error) {
	ctxUser := middleware.MustGetUserFromContext(ctx)

	if ctxUser == nil {
		now := time.Now().UTC()
		return api.UpdateMeProfile401ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Unauthorized"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	if request.Body == nil {
		now := time.Now().UTC()
		return api.UpdateMeProfile400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Request body is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	if err := s.validator.Struct(request.Body); err != nil {
		now := time.Now().UTC()
		return api.UpdateMeProfile400ApplicationProblemPlusJSONResponse{
			Message:   stringPtr(validationErrorMessage(err)),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	var privacyLevel *user.PrivacyLevel
	if request.Body.PrivacyLevel != nil {
		pl := user.PrivacyLevel(*request.Body.PrivacyLevel)
		privacyLevel = &pl
	}

	updatedUser, err := s.userService.UpdateUser(ctx, user.PartialUser{
		ID:           ctxUser.ID,
		DisplayName:  request.Body.DisplayName,
		PrivacyLevel: privacyLevel,
		Username:     request.Body.Username,
	})
	if err != nil {
		now := time.Now().UTC()
		return api.UpdateMeProfile500ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Failed to update profile: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	now := time.Now().UTC()
	return api.UpdateMeProfile200JSONResponse{
		Data: &map[string]interface{}{
			"username":     updatedUser.Username,
			"displayName":  updatedUser.DisplayName,
			"privacyLevel": string(updatedUser.PrivacyLevel),
		},
		Message:   stringPtr("Profile updated successfully"),
		Success:   boolPtr(true),
		Timestamp: timeStrPtr(now),
	}, nil
}
func (s *Server) CheckUsernameAvailable(ctx context.Context, request api.CheckUsernameAvailableRequestObject) (api.CheckUsernameAvailableResponseObject, error) {
	user := middleware.MustGetUserFromContext(ctx)

	if user == nil {
		now := time.Now().UTC()
		return api.CheckUsernameAvailable401ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Unauthorized"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	if request.Body == nil {
		now := time.Now().UTC()
		return api.CheckUsernameAvailable422ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Request body is required"),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	if err := s.validator.Struct(request.Body); err != nil {
		now := time.Now().UTC()
		return api.CheckUsernameAvailable422ApplicationProblemPlusJSONResponse{
			Message:   stringPtr(validationErrorMessage(err)),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	username := *request.Body.Username
	available, err := s.userService.UserNameAvailable(ctx, username)
	if err != nil {
		now := time.Now().UTC()
		return api.CheckUsernameAvailable500ApplicationProblemPlusJSONResponse{
			Message:   stringPtr("Failed to check username availability: " + err.Error()),
			Success:   boolPtr(false),
			Timestamp: timeStrPtr(now),
		}, nil
	}

	now := time.Now().UTC()
	if available {
		return api.CheckUsernameAvailable200JSONResponse{
			Message:   stringPtr("Username available"),
			Success:   boolPtr(true),
			Timestamp: timeStrPtr(now),
		}, nil
	}
	return api.CheckUsernameAvailable422ApplicationProblemPlusJSONResponse{
		Message:   stringPtr("Username already exists"),
		Success:   boolPtr(false),
		Timestamp: timeStrPtr(now),
	}, nil
}
