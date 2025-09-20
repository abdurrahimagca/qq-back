package user

import (
	"context"
	"encoding/hex"

	"github.com/abdurrahimagca/qq-back/internal/db"
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateDefaultUserWithAuthID(ctx context.Context, authID pgtype.UUID) (*db.User, error)
	GetUserByID(ctx context.Context, userID pgtype.UUID) (*db.User, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	UpdateUser(ctx context.Context, user db.UpdateUserParams) (*db.User, error)
	UserNameAvailable(ctx context.Context, username string) (bool, error)
	WithTx(tx pgx.Tx) Service
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) WithTx(tx pgx.Tx) Service {
	return &service{repo: s.repo.WithTx(tx)}
}

func (s *service) CreateDefaultUserWithAuthID(ctx context.Context, authID pgtype.UUID) (*db.User, error) {
	prefix := "user_"
	// Generate a simple username using authID bytes
	username := prefix + hex.EncodeToString(authID.Bytes[:6])

	return s.repo.CreateUserWithAuthID(ctx, authID, username)
}

func (s *service) GetUserByID(ctx context.Context, userID pgtype.UUID) (*db.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (s *service) UpdateUser(ctx context.Context, user db.UpdateUserParams) (*db.User, error) {
	if user.Username.Valid {
		available, err := s.UserNameAvailable(ctx, user.Username.String)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, qqerrors.ErrUniqueViolation
		}
	}

	return s.repo.UpdateUser(ctx, user)
}

func (s *service) UserNameAvailable(ctx context.Context, username string) (bool, error) {
	exists, err := s.repo.UserNameExists(ctx, username)
	if err != nil {
		return false, err
	}
	return !exists, nil
}
