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
	mux := http.NewServeMux()
	config, err := environment.Load()
	if err != nil {
		log.Fatal("Error loading environment", err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatal("Error creating pool", err)
	}
	defer pool.Close()

	// Initialize auth handler
	authH := authHandler.NewAuthHandler(pool, config)
	
	// Setup API routes using the generated OpenAPI handler
	apiHandler := api.Handler(authH)
	mux.Handle("/api/v1.1/", http.StripPrefix("/api/v1.1", apiHandler))

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
	log.Fatal(http.ListenAndServe(":3003", mux))
}
