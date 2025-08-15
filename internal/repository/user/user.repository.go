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

func UpdateUserProfile(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, displayName *string, avatarKeySmall *string, avatarKeyMedium *string, avatarKeyLarge *string) error {
	queries := db.New(tx)
	newData := db.UpdateUserParams{}
	if displayName != nil {
		newData.DisplayName = pgtype.Text{String: *displayName, Valid: true}
	}
	if avatarKeySmall != nil {
		newData.AvatarKeySmall = pgtype.Text{String: *avatarKeySmall, Valid: true}
	}
	if avatarKeyMedium != nil {
		newData.AvatarKeyMedium = pgtype.Text{String: *avatarKeyMedium, Valid: true}
	}
	if avatarKeyLarge != nil {
		newData.AvatarKeyLarge = pgtype.Text{String: *avatarKeyLarge, Valid: true}
	}
	if displayName == nil && avatarKeySmall == nil && avatarKeyMedium == nil && avatarKeyLarge == nil {
		return nil
	}

	_, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          userID,
		DisplayName: newData.DisplayName,
		AvatarKeySmall: newData.AvatarKeySmall,
		AvatarKeyMedium: newData.AvatarKeyMedium,
		AvatarKeyLarge: newData.AvatarKeyLarge,
	})

	if err != nil {
		return err
	}

	return nil
}
