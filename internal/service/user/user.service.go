package user

import (
	"context"
	"io"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/external/bucket"
	userRepository "github.com/abdurrahimagca/qq-back/internal/repository/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	db       *pgxpool.Pool
	config   *environment.Config
	userRepo *userRepository.UserRepository
}

func NewUserService(db *pgxpool.Pool, config *environment.Config) *UserService {
	return &UserService{
		db:       db,
		config:   config,
		userRepo: userRepository.NewUserRepository(db),
	}
}

type UpdateUserParams struct {
	DisplayName  pgtype.Text
	Username     pgtype.Text
	PrivacyLevel db.NullPrivacyLevel
	File         io.Reader
}

func (s *UserService) GetUserProfile(ctx context.Context, userID pgtype.UUID) (*db.User, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user, err := s.userRepo.GetUserByID(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID pgtype.UUID, params UpdateUserParams) (*db.User, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	dbParams := db.UpdateUserParams{
		ID:           userID,
		DisplayName:  params.DisplayName,
		Username:     params.Username,
		PrivacyLevel: params.PrivacyLevel,
	}

	if params.File != nil {
		bucketService, err := bucket.NewService(s.config.R2)
		if err != nil {
			return nil, err
		}
		uploadImageResult, err := bucketService.UploadImage(params.File, false)
		if err != nil {
			return nil, err
		}
		dbParams.AvatarKey = pgtype.Text{String: *uploadImageResult.Key, Valid: true}
	}

	user, err := s.userRepo.UpdateUserProfile(ctx, tx, dbParams)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}