package user

import (
	"context"
	"errors"
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



func UpdateUserProfile(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, params db.UpdateUserParams) (*db.User, error) {
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
	if !params.DisplayName.Valid && !params.AvatarKey.Valid {
		return nil, errors.New("no data to update")
	}

	user, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          userID,
		DisplayName: newData.DisplayName,
		AvatarKey:   newData.AvatarKey,
		Username:    newData.Username,
		PrivacyLevel: newData.PrivacyLevel,
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}