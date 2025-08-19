package handler

import (
	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	auth_handler "github.com/abdurrahimagca/qq-back/internal/handler/auth"
	user_handler "github.com/abdurrahimagca/qq-back/internal/handler/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiHandler struct {
	*auth_handler.AuthHandler
	*user_handler.UserHandler
}

var _ api.StrictServerInterface = (*ApiHandler)(nil)

func NewApiHandler(db *pgxpool.Pool, config *environment.Config) *ApiHandler {
	return &ApiHandler{
		AuthHandler: auth_handler.NewAuthHandler(db, config),
		UserHandler: user_handler.NewUserHandler(db, config),
	}
}
