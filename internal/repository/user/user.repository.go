package user

import (
	"context"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetUserByID(ctx context.Context, tx pgx.Tx, userID pgtype.UUID) (*db.User, error) {
	queries := db.New(tx)

	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUserProfile(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, displayName *string, avatarURL *string) error {
	queries := db.New(tx)
	newData := db.UpdateUserParams{}
	if displayName != nil {
		newData.DisplayName = pgtype.Text{String: *displayName, Valid: true}
	}
	if avatarURL != nil {
		newData.AvatarUrl = pgtype.Text{String: *avatarURL, Valid: true}
	}
	if displayName == nil && avatarURL == nil {
		return nil
	}

	_, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          userID,
		DisplayName: newData.DisplayName,
		AvatarUrl:   newData.AvatarUrl,
	})

	if err != nil {
		return err
	}

	return nil
}
