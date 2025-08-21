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
		return api.GetMeProfile500JSONResponse{
			ErrorCode: "INTERNAL_ERROR",
			Message:   "Failed to get profile: user not found",
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	return api.GetMeProfile200JSONResponse{
		Data: &map[string]interface{}{
			"username":     user.Username,
			"displayName":  user.DisplayName,
			"privacyLevel": user.PrivacyLevel,
		},
		Success:   true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
func (s *Server) GetMeAvatar(ctx context.Context, request api.GetMeAvatarRequestObject) (api.GetMeAvatarResponseObject, error) {
	// Get user from context (added by middleware)
	user := middleware.MustGetUserFromContext(ctx)

	if user == nil {
		return api.GetMeAvatar401JSONResponse{}, nil
	}

	if user.AvatarKey == nil {
		return api.GetMeAvatar404JSONResponse{
			Message:   stringPtr("Avatar not found"),
			Timestamp: stringPtr(time.Now().UTC().Format(time.RFC3339)),
		}, nil
	}

	avatarSignedUrl, err := s.fileUC.GetSignedUrlForKey(ctx, *user.AvatarKey, 30*24*time.Hour)
	if err != nil {
		return api.GetMeAvatar500JSONResponse{
			Message:   "Failed to get avatar signed url: " + err.Error(),
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	expiresInSeconds := int(30 * 24 * time.Hour / time.Second)
	return api.GetMeAvatar200JSONResponse{

		Message:   "Avatar signed url fetched successfully",
		SignedUrl: avatarSignedUrl,
		ExpiresIn: &expiresInSeconds,
		Success:   true,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
func (s *Server) PostMeUpdateProfile(ctx context.Context, request api.PostMeUpdateProfileRequestObject) (api.PostMeUpdateProfileResponseObject, error) {
	ctxUser := middleware.MustGetUserFromContext(ctx)

	if ctxUser == nil {
		return api.PostMeUpdateProfile401JSONResponse{}, nil
	}

	var privacyLevel *user.PrivacyLevel
	if request.Body != nil && request.Body.PrivacyLevel != nil {
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
		return api.PostMeUpdateProfile500JSONResponse{
			Message:   "Failed to update profile: " + err.Error(),
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	return api.PostMeUpdateProfile200JSONResponse{
		Message:      "Profile updated successfully",
		Success:      true,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Username:     stringPtr(updatedUser.Username),
		DisplayName:  updatedUser.DisplayName,
		PrivacyLevel: stringPtr(string(updatedUser.PrivacyLevel)),
	}, nil
}
func (s *Server) PostUserUsernameAvailable(ctx context.Context, request api.PostUserUsernameAvailableRequestObject) (api.PostUserUsernameAvailableResponseObject, error) {
	user := middleware.MustGetUserFromContext(ctx)

	if user == nil {
		return api.PostUserUsernameAvailable401JSONResponse{}, nil
	}

	available, err := s.userService.UserNameAvailable(ctx, *request.Body.Username)
	if err != nil {
		return api.PostUserUsernameAvailable500JSONResponse{
			Message:   "Failed to check username availability: " + err.Error(),
			Success:   false,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}
	if available {
		return api.PostUserUsernameAvailable200JSONResponse{
			Message:   "Username available",
			Success:   true,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}
	return api.PostUserUsernameAvailable422JSONResponse{
		Message:   "Username already exists",
		Success:   false,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
