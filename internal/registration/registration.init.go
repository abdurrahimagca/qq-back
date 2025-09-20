package registration

import (
	"github.com/abdurrahimagca/qq-back/internal/auth"
	mailport "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationModule struct {
	usecase RegistrationUsecase
	server  RegistrationServer
}

func NewRegistrationModule(
	mailer mailport.Service,
	authService auth.Service,
	userService user.Service,
	tokenService tokenport.Service,
	pool *pgxpool.Pool,
) *RegistrationModule {
	usecase := NewRegistrationUsecase(mailer, authService, userService, pool, tokenService)
	server := NewRegistrationServer(usecase)

	return &RegistrationModule{
		usecase: usecase,
		server:  server,
	}
}

func (rm *RegistrationModule) RegisterEndpoints(api huma.API) {
	rm.server.RegisterRegistrationEndpoints(api)
}
