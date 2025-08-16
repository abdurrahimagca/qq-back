package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/user"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateUserRouter(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	protectedMiddlewares := []middleware.Middleware{
		middleware.TransactionMiddleware(db),
		middleware.UserAuth(config),
	}

	mux.HandleFunc("POST /me/update",
		middleware.ChainFunc(
			func(w http.ResponseWriter, r *http.Request) {
				user.UpdateUserHandler(w, r, db, *config)
			},
			protectedMiddlewares...,
		),
	)
}
func UpdateUserProfilePictureRouter(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	protectedMiddlewares := []middleware.Middleware{
		middleware.TransactionMiddleware(db),
		middleware.UserAuth(config),
	}

	mux.HandleFunc("PUT /me/avatar",
		middleware.ChainFunc(
			func(w http.ResponseWriter, r *http.Request) {
				user.UpdateUserProfilePicture(w, r, db, *config)
			},
			protectedMiddlewares...,
		),
	)
}
