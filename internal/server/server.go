package server

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/app"
	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/abdurrahimagca/qq-back/internal/platform/file-upload"
	"github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	registrationUC app.RegistrationUsecase
	fileUC         app.FileUsecase
	userService    user.Service
}

// Compile-time interface compliance check
var _ api.StrictServerInterface = (*Server)(nil)

func NewServer(registrationUC app.RegistrationUsecase, fileUC app.FileUsecase, userService user.Service) api.StrictServerInterface {
	return &Server{
		registrationUC: registrationUC,
		fileUC:         fileUC,
		userService:    userService,
	}
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
	r2Service := fileupload.NewR2Service(config.R2)

	// Initialize use cases
	registrationUC := app.NewRegistrationUsecase(mailerService, authService, userService, pool, tokenService)
	fileUC := app.NewFileUsecase(r2Service, *config)

	// Create the server that implements StrictServerInterface
	server := NewServer(registrationUC, fileUC, userService)

	// Initialize strict middleware
	strictAuthMiddleware := middleware.NewStrictAuthMiddleware(tokenService, userService, []string{"/docs", "/openapi.json"})
	
	// Create strict handler with middleware
	strictHandler := api.NewStrictHandler(server, []api.StrictMiddlewareFunc{strictAuthMiddleware.Middleware})

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

	return mux, nil
}