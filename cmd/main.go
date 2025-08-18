package main

import (
	"context"
	"log"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
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

	log.Println("Server is running on port 3003")
	log.Fatal(http.ListenAndServe(":3003", mux))
}
