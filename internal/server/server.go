package server

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/app"
	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	registrationUC app.RegistrationUsecase
}

// Compile-time interface compliance check
var _ api.StrictServerInterface = (*Server)(nil)

func NewServer(registrationUC app.RegistrationUsecase) api.StrictServerInterface {
	return &Server{
		registrationUC: registrationUC,
	}
}


func stringPtr(s string) *string {
	return &s
}

func NewUnifiedServer(pool *pgxpool.Pool, config *environment.Environment) (http.Handler, error) {
	// Initialize repositories
	authRepo := auth.NewPgxRepository(pool)
	userRepo := user.NewPgxRepository(pool)

	// Initialize services
	authService := auth.NewService(authRepo)
	userService := user.NewService(userRepo)

	// Initialize platform services
	mailerService := mailer.NewResendMailer(config)
	tokenService := token.NewJWTTokenService(config)

	// Initialize use cases
	registrationUC := app.NewRegistrationUsecase(mailerService, authService, userService, pool, tokenService)

	// Create the server that implements StrictServerInterface
	server := NewServer(registrationUC)

	// Create strict handler
	strictHandler := api.NewStrictHandler(server, nil)

	// Create ServeMux and register routes
	mux := http.NewServeMux()

	// Register the OpenAPI routes
	api.HandlerFromMuxWithBaseURL(strictHandler, mux, "/api/v1.1")

	// Add documentation and utility routes
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

	// Return the unified handler
	return mux, nil
}