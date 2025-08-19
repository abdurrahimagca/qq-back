package user

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	userService "github.com/abdurrahimagca/qq-back/internal/service/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	userService *userService.UserService
}

func NewUserHandler(db *pgxpool.Pool, config *environment.Config) *UserHandler {
	return &UserHandler{
		userService: userService.NewUserService(db, config),
	}
}

func (h *UserHandler) GetUserProfile(ctx context.Context, request api.GetUserProfileRequestObject) (api.GetUserProfileResponseObject, error) {
	// Implementation for GetUserProfile
	return api.GetUserProfile200JSONResponse{}, nil
}

func (h *UserHandler) PostUserProfileUpdate(ctx context.Context, request api.PostUserProfileUpdateRequestObject) (api.PostUserProfileUpdateResponseObject, error) {
	// Implementation for PostProfileUpdate
	return api.PostUserProfileUpdate200JSONResponse{}, nil
}

func (h *UserHandler) PutUserProfileUpdateAvatar(ctx context.Context, request api.PutUserProfileUpdateAvatarRequestObject) (api.PutUserProfileUpdateAvatarResponseObject, error) {
	// Implementation for PutProfileUpdateAvatar
	return api.PutUserProfileUpdateAvatar200JSONResponse{}, nil
}
