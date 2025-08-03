package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/health"
	middlewareChain "github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HealthRoute(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	handler := health.NewHandler(db, config)
	
	// Public health check (no auth required)
	publicHealthMiddlewares := middlewareChain.Chain(
		// Add any health-specific middlewares here if needed
	)
	
	// Health check endpoint
	mux.HandleFunc("/health", 
		middlewareChain.ChainFunc(handler.Health, publicHealthMiddlewares))
}