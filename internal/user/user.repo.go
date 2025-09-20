package user

import (
	"context"

	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"

	"github.com/abdurrahimagca/qq-back/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	GetUserByID(ctx context.Context, userID pgtype.UUID) (*db.User, error)
	CreateUserWithAuthID(ctx context.Context, authID pgtype.UUID, username string) (*db.User, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	UpdateUser(ctx context.Context, user db.UpdateUserParams) (*db.User, error)
	UserNameExists(ctx context.Context, username string) (bool, error)
}

type pgxRepository struct {
	q *db.Queries
}

func NewPgxRepository(pool *pgxpool.Pool) Repository {
	return &pgxRepository{
		q: db.New(pool),
	}
}

func (r *pgxRepository) WithTx(tx pgx.Tx) Repository {
	return &pgxRepository{
		q: r.q.WithTx(tx),
	}
}

func (r *pgxRepository) GetUserByID(ctx context.Context, userID pgtype.UUID) (*db.User, error) {
	dbUser, err := r.q.GetUserByID(ctx, userID)
	if err != nil {
		return nil, qqerrors.GetDBErrAsQQError(err)
	}
	return &dbUser, nil
}

func (r *pgxRepository) CreateUserWithAuthID(
	ctx context.Context, authID pgtype.UUID, username string) (*db.User, error) {
	dbUser, err := r.q.InsertUser(ctx, db.InsertUserParams{
		AuthID:   authID,
		Username: username,
	})
	if err != nil {
		return nil, qqerrors.GetDBErrAsQQError(err)
	}
	return &dbUser, nil
}
func (r *pgxRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	dbUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, qqerrors.GetDBErrAsQQError(err)
	}
	return &dbUser, nil
}
func (r *pgxRepository) UpdateUser(ctx context.Context, user db.UpdateUserParams) (*db.User, error) {
	dbUser, err := r.q.UpdateUser(ctx, user)
	if err != nil {
		return nil, qqerrors.GetDBErrAsQQError(err)
	}
	return &dbUser, nil
}

func (r *pgxRepository) UserNameExists(ctx context.Context, username string) (bool, error) {
	exists, err := r.q.UserNameExists(ctx, username)
	if err != nil {
		return false, qqerrors.GetDBErrAsQQError(err)
	}
	return exists > 0, nil
}
