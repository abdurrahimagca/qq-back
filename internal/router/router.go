package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Router(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	// API v1 routes
	api := http.NewServeMux()

	// Route groups (no middleware yet - added per route as needed)
	AuthRoute(api, db, config)
	HealthRoute(api, db, config)
	MediaRoute(api, db, config)

	// Base API middleware chain for all /api/v1/* routes
	baseMiddlewares := middleware.Chain(
	// Add common middlewares here (CORS, logging, etc.)
	// middleware.CORS,
	// middleware.Logging,
	)

	// Apply base middleware to API v1
	mux.Handle("/api/v1/", baseMiddlewares(http.StripPrefix("/api/v1", api)))

	// Root route for health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		middleware.ApiAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			time := time.Now().Format(time.RFC3339)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Hello, This is working World!", "time": time})
		}))
	})
}
