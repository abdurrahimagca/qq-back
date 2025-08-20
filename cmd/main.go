package main

import (
	"context"
	"log"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/server"
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

	// Initialize the unified server with all dependencies
	handler, err := server.NewUnifiedServer(pool, config)
	if err != nil {
		log.Fatal("Error creating unified server", err)
	}

	log.Println("Server is running on port 3003")
	log.Println("API Documentation available at: http://localhost:3003/docs")
	log.Fatal(http.ListenAndServe(":3003", handler))
}
