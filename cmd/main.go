package main

import (
	"context"
	"log"

	"github.com/abdurrahimagca/qq-back/internal/environment"

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
}
