package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/platform/mailer"
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

	mailer       mailer.Service
	authService  auth.Service
	userService  user.Service
	tokenService tokenport.Service
	logger       *slog.Logger
}

func New(env *environment.Environment) *Bootstrap {
	b := &Bootstrap{
		env:    env,
		logger: slog.Default(),
	}

	b.initInfrastructure()
	b.initDependencies()

	return b
}

func (b *Bootstrap) initInfrastructure() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, b.env.DatabaseURL)
	if err != nil {
		b.logger.Error("Error creating pool", "error", err)
	}
	b.pool = pool

	b.mux = http.NewServeMux()
	humaConfig := huma.DefaultConfig(b.env.API.Title, b.env.API.Version)
	humaConfig.DocsPath = ""
	b.api = humago.New(b.mux, humaConfig)

	b.setupDocsEndpoint()
}

func (b *Bootstrap) setupDocsEndpoint() {
	b.mux.HandleFunc("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html>
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
	b.mailer = mailer.NewResendMailer(b.env)
	b.tokenService = tokenport.NewJWTTokenService(b.env)
}

func (b *Bootstrap) registrationModule() {
	rm := registration.NewModule(
		b.mailer,
		b.authService,
		b.userService,
		b.pool,
		b.tokenService,
	)
	rm.RegisterEndpoints(b.api)
}

func (b *Bootstrap) Bootstrap() {
	b.registrationModule()
}
func (b *Bootstrap) StartServer() {
	readTimeout := 15
	readHeaderTimeout := 5
	writeTimeout := 15
	idleTimeout := 60

	srv := &http.Server{
		Addr:              ":" + b.env.API.Port,
		Handler:           b.mux,
		ReadTimeout:       time.Duration(readTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(readHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(writeTimeout) * time.Second,
		IdleTimeout:       time.Duration(idleTimeout) * time.Second,
	}

	b.logger.Info("Server starting", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		b.logger.Error("Error starting server", "error", err)
	}
}

func (b *Bootstrap) Close() {
	if b.pool != nil {
		b.pool.Close()
	}
}
