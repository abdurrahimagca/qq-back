package user

import (
	"context"
	"io"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/external/bucket"
	"github.com/abdurrahimagca/qq-back/internal/repository/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpdateUserParams defines the set of parameters that can be updated by a user
// through the service layer. It intentionally omits fields like AvatarKey
// which should not be set directly by the client.
type UpdateUserParams struct {
	DisplayName  pgtype.Text
	Username     pgtype.Text
	PrivacyLevel db.NullPrivacyLevel
	File         io.Reader
}

func updateUserProfileAndHandleInsertNewProfilePictureService(
	ctx context.Context, tx pgx.Tx, file io.Reader, userID pgtype.UUID, params UpdateUserParams, env environment.Config) (*db.User, error) {
	bucketService, errr := bucket.NewService(env.R2)
	if errr != nil {
		return nil, errr
	}
	uploadImageResult, err := bucketService.UploadImage(file, false)
	if err != nil {
		return nil, err
	}

	// Map from service-level params to db-level params, adding the new avatar key.
	dbParams := db.UpdateUserParams{
		AvatarKey:    pgtype.Text{String: *uploadImageResult.Key, Valid: true},
		DisplayName:  params.DisplayName,
		Username:     params.Username,
		PrivacyLevel: params.PrivacyLevel,
	}

	user, err := user.UpdateUserProfile(ctx, tx, userID, dbParams)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserProfile handles the business logic for updating a user's profile.
func UpdateUserProfile(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, params UpdateUserParams, env environment.Config) (*db.User, error) {
	if params.File == nil {
		// If no file, map service params to DB params and update.
		dbParams := db.UpdateUserParams{
			DisplayName:  params.DisplayName,
			Username:     params.Username,
			PrivacyLevel: params.PrivacyLevel,
		}
		return user.UpdateUserProfile(ctx, tx, userID, dbParams)
	}

	// If a file is provided, use the helper to upload and then update the profile.
	return updateUserProfileAndHandleInsertNewProfilePictureService(ctx, tx, params.File, userID, params, env)
}
