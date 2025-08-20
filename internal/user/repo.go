package user

import (
	"context"
	"errors"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	GetUserByID(ctx context.Context, userID uuid.UUID) (User, error)
	CreateUser(ctx context.Context, username string) (User, error)
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

func (r *pgxRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (User, error) {
	dbUser, err := r.q.GetUserByID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return mapDBUserToDomainUser(dbUser), nil
}

func (r *pgxRepository) CreateUser(ctx context.Context, username string) (User, error) {
	dbUser, err := r.q.InsertUser(ctx, db.InsertUserParams{
		Username: username,
	})
	if err != nil {
		return User{}, err
	}
	return mapDBUserToDomainUser(dbUser), nil
}

func mapDBUserToDomainUser(dbUser db.User) User {
	var displayName *string
	if dbUser.DisplayName.Valid {
		displayName = &dbUser.DisplayName.String
	}

	var avatarKey *string
	if dbUser.AvatarKey.Valid {
		avatarKey = &dbUser.AvatarKey.String
	}

	return User{
		ID:           dbUser.ID.Bytes,
		Username:     dbUser.Username,
		DisplayName:  displayName,
		PrivacyLevel: string(dbUser.PrivacyLevel),
		AvatarKey:    avatarKey,
	}
}