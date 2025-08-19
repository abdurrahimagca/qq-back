package main

import (
	"context"
	"log"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	authHandler "github.com/abdurrahimagca/qq-back/internal/handler/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Starting server...")

	// Load configuration
	config, err := environment.Load()
	if err != nil {
		log.Fatal("Error loading environment", err)
	}

	// Initialize database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatal("Error creating pool", err)
	}
	defer pool.Close()

	// Initialize auth handler - this satisfies the StrictServerInterface
	authH := authHandler.NewAuthHandler(pool, config)

	// Create a strict handler that ENFORCES type safety!
	strictHandler := api.NewStrictHandler(authH, nil)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register our strict handler using the generated HandlerFromMuxWithBaseURL
	// This automatically handles the routing based on the OpenAPI spec
	// Base URL matches the server URL in openapi.yml: /api/v1.1
	// This function registers routes directly on our mux
	api.HandlerFromMuxWithBaseURL(strictHandler, mux, "/api/v1.1")

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./cmd/_docs.html")
	})

	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, "./openapi.json")
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	log.Println("Server is running on port 3003")
	log.Println("API Documentation available at: http://localhost:3003/docs")
	log.Println("API endpoints:")
	log.Println("  POST /api/v1.1/auth/otp")
	log.Println("  POST /api/v1.1/auth/otp-verify")
	log.Fatal(http.ListenAndServe(":3003", mux))
}
