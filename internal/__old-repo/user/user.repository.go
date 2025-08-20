package user

import (
	"context"
	"errors"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, tx pgx.Tx, userID pgtype.UUID) (*db.User, error) {
	queries := db.New(tx)

	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateUserProfile(ctx context.Context, tx pgx.Tx, params db.UpdateUserParams) (*db.User, error) {
	queries := db.New(tx)
	newData := db.UpdateUserParams{}
	if params.DisplayName.Valid {
		newData.DisplayName = pgtype.Text{String: params.DisplayName.String, Valid: true}
	}
	if params.AvatarKey.Valid {
		newData.AvatarKey = pgtype.Text{String: params.AvatarKey.String, Valid: true}
	}
	if params.Username.Valid {
		newData.Username = pgtype.Text{String: params.Username.String, Valid: true}
	}
	if params.PrivacyLevel.Valid {
		newData.PrivacyLevel = db.NullPrivacyLevel{PrivacyLevel: params.PrivacyLevel.PrivacyLevel, Valid: true}
	}
	if !params.DisplayName.Valid && !params.AvatarKey.Valid && !params.Username.Valid && !params.PrivacyLevel.Valid {
		return nil, errors.New("no data to update")
	}

	user, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:           params.ID,
		DisplayName:  newData.DisplayName,
		AvatarKey:    newData.AvatarKey,
		Username:     newData.Username,
		PrivacyLevel: newData.PrivacyLevel,
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}
