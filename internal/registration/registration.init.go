package registration

import (
	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/environment"
	mail "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	mailport "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationDependencies struct {
	Mailer mailport.Service
	AuthService auth.Service
	UserService user.Service
	TokenService tokenport.Service
	Pool *pgxpool.Pool
}

type registrationInit struct {
	api huma.API
	deps RegistrationDependencies
}


type RegistrationInit interface {
	initAndRegisterRegistrationEndpoints(api huma.API, deps RegistrationDependencies)
	initDependencies(pool *pgxpool.Pool, environment *environment.Environment) RegistrationDependencies
	Factory(pool *pgxpool.Pool, environment *environment.Environment, api huma.API)

}

func NewRegistrationInit() RegistrationInit {
	return &registrationInit{}
}

func (ri *registrationInit) initAndRegisterRegistrationEndpoints(api huma.API, deps RegistrationDependencies) {	
  registrationUsecase := NewRegistrationUsecase(deps.Mailer, deps.AuthService, deps.UserService, deps.Pool, deps.TokenService)
  registrationServer := NewRegistrationServer(registrationUsecase)
  registrationServer.RegisterRegistrationEndpoints(api)
}

func (ri *registrationInit) initDependencies(pool *pgxpool.Pool, environment *environment.Environment) RegistrationDependencies {
	authRepo := auth.NewPgxRepository(pool)
	userRepo := user.NewPgxRepository(pool)
	authService := auth.NewService(authRepo)
	userService := user.NewService(userRepo)
	mailerService := mail.NewResendMailer(environment)
	tokenService := tokenport.NewJWTTokenService(environment)
	return RegistrationDependencies{
		Mailer: mailerService,
		AuthService: authService,
		UserService: userService,
		TokenService: tokenService,
		Pool: pool,
	}
}

func (ri *registrationInit) Factory(pool *pgxpool.Pool, environment *environment.Environment, api huma.API) {
     deps := ri.initDependencies(pool, environment)
     ri.initAndRegisterRegistrationEndpoints(api, deps)
}