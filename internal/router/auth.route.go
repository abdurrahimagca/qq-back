package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/handler/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoute(mux *http.ServeMux, db *pgxpool.Pool) {
	mux.HandleFunc("/api/v1/auth/signin-or-up-with-otp", auth.NewHandler(db).SignInOrUpWithOtp)
}