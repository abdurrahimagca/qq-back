package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/auth"
	middlewareChain "github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoute(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	handler := auth.NewHandler(db, config)
	
	// Public auth routes (no additional auth needed)
	publicAuthMiddlewares := middlewareChain.Chain(
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
		middlewareChain.ChainFunc(handler.SignInOrUpWithOtp, publicAuthMiddlewares))
	
	// Protected routes (examples for future use)
	// mux.HandleFunc("/auth/logout", 
	//     middlewareChain.ChainFunc(handler.Logout, protectedAuthMiddlewares))
	// mux.HandleFunc("/auth/refresh", 
	//     middlewareChain.ChainFunc(handler.Refresh, protectedAuthMiddlewares))
}