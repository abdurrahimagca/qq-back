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
	CreateUserWithAuthID(ctx context.Context, authID uuid.UUID, username string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, user PartialUser) (User, error)
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
	// This method is deprecated - use CreateUserWithAuthID instead
	// Keeping for backward compatibility but should not be used
	dbUser, err := r.q.InsertUser(ctx, db.InsertUserParams{
		Username: username,
		// AuthID is required but not provided - this will fail
	})
	if err != nil {
		return User{}, err
	}
	return mapDBUserToDomainUser(dbUser), nil
}

func (r *pgxRepository) CreateUserWithAuthID(ctx context.Context, authID uuid.UUID, username string) (User, error) {
	dbUser, err := r.q.InsertUser(ctx, db.InsertUserParams{
		AuthID:   pgtype.UUID{Bytes: authID, Valid: true},
		Username: username,
	})
	if err != nil {
		return User{}, err
	}
	return mapDBUserToDomainUser(dbUser), nil
}
func (r *pgxRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	dbUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	return mapDBUserToDomainUser(dbUser), nil
}
func (r *pgxRepository) UpdateUser(ctx context.Context, user PartialUser) (User, error) {
	dbUser, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:           pgtype.UUID{Bytes: user.ID, Valid: true},
		DisplayName:  pgtype.Text{String: *user.DisplayName, Valid: user.DisplayName != nil},
		PrivacyLevel: db.NullPrivacyLevel{PrivacyLevel: db.PrivacyLevel(*user.PrivacyLevel), Valid: user.PrivacyLevel != nil},
	})
	if err != nil {
		return User{}, err
	}
	return mapDBUserToDomainUser(dbUser), nil
}

func (r *pgxRepository) UserNameExists(ctx context.Context, username string) (bool, error) {
	exists, err := r.q.UserNameExists(ctx, username)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
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
		AuthID:       dbUser.AuthID.Bytes,
		Username:     dbUser.Username,
		DisplayName:  displayName,
		PrivacyLevel: PrivacyLevel(dbUser.PrivacyLevel),
		AvatarKey:    avatarKey,
	}
}
