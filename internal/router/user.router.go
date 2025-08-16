package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateUserRouter(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	mux.HandleFunc("POST /me/update", func(w http.ResponseWriter, r *http.Request) {
		user.UpdateUserHandler(w, r, db, *config)
	})
}
