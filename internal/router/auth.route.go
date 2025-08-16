package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/auth"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoute(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	handler := auth.NewHandler(db, config)

	// Public routes
	mux.HandleFunc("POST /auth/signin-or-up-with-otp", handler.SignInOrUpWithOtp)
	mux.HandleFunc("POST /auth/login-otp", handler.SignInWithOtpCode)

	// Protected routes requiring JWT authentication
	protectedMiddlewares := []middleware.Middleware{
		middleware.TransactionMiddleware(db),
		middleware.UserAuth(config),
	}

	mux.HandleFunc("GET /auth/profile",
		middleware.ChainFunc(handler.GetUserProfile, protectedMiddlewares...))

	mux.HandleFunc("POST /auth/refresh-token",
		middleware.ChainFunc(handler.RefreshToken, protectedMiddlewares...))
}
