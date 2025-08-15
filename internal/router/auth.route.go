package router

import (
	"context"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/auth"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoute(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	handler := auth.NewHandler(db, config)

	// Public auth routes (no additional auth needed)
	publicAuthMiddlewares := middleware.Chain(
	// Add public route middlewares here (rate limiting, etc.)
	)

	// Protected auth routes (require API auth) - uncomment when needed
	/*
		protectedAuthMiddlewares := middlewareChain.Chain(
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					middleware.ApiAuth(config, w, r, func(w http.ResponseWriter, r *http.Request) {
						next.ServeHTTP(w, r)
					})
				})
			},
		)
	*/

	// Public routes
	mux.HandleFunc("/auth/signin-or-up-with-otp",
		middleware.ChainFunc(handler.SignInOrUpWithOtp, publicAuthMiddlewares))

	mux.HandleFunc("/auth/login-otp",
		middleware.ChainFunc(handler.SignInWithOtpCode, publicAuthMiddlewares))

	// Protected routes requiring JWT authentication
	userAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer tx.Rollback(context.Background())

			middleware.UserAuth(config, tx)(next).ServeHTTP(w, r)
		})
	}

	protectedAuthMiddlewares := middleware.Chain(userAuthMiddleware)

	// User profile route
	mux.HandleFunc("/auth/profile",
		middleware.ChainFunc(handler.GetUserProfile, protectedAuthMiddlewares))
}
