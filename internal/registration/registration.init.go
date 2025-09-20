package registration

import (
	"github.com/abdurrahimagca/qq-back/internal/auth"
	mailport "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	usecase Usecase
	server  Server
}

func NewModule(
	mailer mailport.Service,
	authService auth.Service,
	userService user.Service,
	pool *pgxpool.Pool,
	tokenService tokenport.Service,
) *Module {
	usecase := NewUsecase(mailer, authService, userService, pool, tokenService)
	server := NewServer(usecase)

	return &Module{
		usecase: usecase,
		server:  server,
	}
}

func (rm *Module) RegisterEndpoints(api huma.API) {
	rm.server.RegisterRegistrationEndpoints(api)
}
