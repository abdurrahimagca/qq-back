package main

import (
	"context"

	"log"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/environment"
	mail "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/registration"
	"github.com/abdurrahimagca/qq-back/internal/user"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Starting server...")
	environment, err := environment.Load()
	if err != nil {
		log.Fatal("Error loading environment", err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, environment.DatabaseURL)
	if err != nil {
		log.Fatal("Error creating pool", err)
	}
	defer pool.Close()
	authRepo := auth.NewPgxRepository(pool)
	userRepo := user.NewPgxRepository(pool)
	authService := auth.NewService(authRepo)
	userService := user.NewService(userRepo)
	mailerService := mail.NewResendMailer(environment)
	tokenService := tokenport.NewJWTTokenService(environment)
	registrationUsecase := registration.NewRegistrationUsecase(
		mailerService,
		authService,
		userService,
		pool,
		tokenService,
	)
	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig(environment.API.Title, environment.API.Version)
	humaConfig.DocsPath = "" 
	api := humago.New(mux, humaConfig)

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!doctype html>
						<html>
						<head>
							<title>API Reference</title>
							<meta charset="utf-8" />
							<meta
							name="viewport"
							content="width=device-width, initial-scale=1" />
						</head>
						<body>
							<script
							id="api-reference"
							data-url="/openapi.json"></script>
							<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
						</body>
						</html>`))
	})
	registrationServer := registration.NewRegistrationServer(registrationUsecase)
	registrationServer.RegisterRegistrationEndpoints(api)

	log.Println("Server starting on :" + environment.API.Port)
	log.Fatal(http.ListenAndServe(":"+environment.API.Port, mux))
}
