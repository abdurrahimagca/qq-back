package bootstrap

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/environment"
	mail "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	mailport "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/registration"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Bootstrap struct {
	pool *pgxpool.Pool
	env  *environment.Environment
	api  huma.API
	mux  *http.ServeMux

	mailer       mailport.Service
	authService  auth.Service
	userService  user.Service
	tokenService tokenport.Service
}

func New(env *environment.Environment) *Bootstrap {
	b := &Bootstrap{
		env: env,
	}

	b.initInfrastructure()
	b.initDependencies()

	return b
}

func (b *Bootstrap) initInfrastructure() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, b.env.DatabaseURL)
	if err != nil {
		slog.Error("Error creating pool", "error", err)
	}
	b.pool = pool

	b.mux = http.NewServeMux()
	humaConfig := huma.DefaultConfig(b.env.API.Title, b.env.API.Version)
	humaConfig.DocsPath = ""
	b.api = humago.New(b.mux, humaConfig)

	b.setupDocsEndpoint()
}

func (b *Bootstrap) setupDocsEndpoint() {
	b.mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
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
}

func (b *Bootstrap) initDependencies() {
	authRepo := auth.NewPgxRepository(b.pool)
	userRepo := user.NewPgxRepository(b.pool)
	b.authService = auth.NewService(authRepo)
	b.userService = user.NewService(userRepo)
	b.mailer = mail.NewResendMailer(b.env)
	b.tokenService = tokenport.NewJWTTokenService(b.env)
}

func (b *Bootstrap) registrationModule() {
	rm := registration.NewRegistrationModule(
		b.mailer,
		b.authService,
		b.userService,
		b.tokenService,
		b.pool,
	)
	rm.RegisterEndpoints(b.api)
}

func (b *Bootstrap) Bootstrap() {
	b.registrationModule()
}

func (b *Bootstrap) StartServer() {
	slog.Info("Server starting on :" + b.env.API.Port)
	slog.Error("Error starting server", "error", http.ListenAndServe(":"+b.env.API.Port, b.mux))
}

func (b *Bootstrap) Close() {
	if b.pool != nil {
		b.pool.Close()
	}
}
