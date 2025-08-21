package user

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	// Transaction detayı (tx) artık burada yok!
	CreateDefaultUser(ctx context.Context) (User, error)
	CreateDefaultUserWithAuthID(ctx context.Context, authID uuid.UUID) (User, error)
	GetUserByID(ctx context.Context, userID string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, user PartialUser) (User, error)
	UserNameAvailable(ctx context.Context, username string) (bool, error)
	// Transaction-aware methods
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

func (s *service) CreateDefaultUser(ctx context.Context) (User, error) {
	// This method is deprecated - use CreateDefaultUserWithAuthID instead
	// Service'in görevi iş mantığını uygulamaktır: "default username" oluşturmak.
	prefix := "user_"
	username := prefix + uuid.New().String()[:12] // Biraz daha uzun ve eşsiz olması için

	// Bu metot artık bir tx almaz. Transaction yönetimi UseCase'in işidir.
	// UseCase, bu servise transaction'a bağlı bir repository (`repo.WithTx(tx)`) verecektir.
	return s.repo.CreateUser(ctx, username)
}

func (s *service) CreateDefaultUserWithAuthID(ctx context.Context, authID uuid.UUID) (User, error) {
	// Service'in görevi iş mantığını uygulamaktır: "default username" oluşturmak.
	prefix := "user_"
	username := prefix + uuid.New().String()[:12] // Biraz daha uzun ve eşsiz olması için

	return s.repo.CreateUserWithAuthID(ctx, authID, username)
}

func (s *service) GetUserByID(ctx context.Context, userID string) (User, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, ErrInvalidID
	}

	return s.repo.GetUserByID(ctx, parsedID)
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, ErrNotFound
	}
	return user, nil
}
func (s *service) UpdateUser(ctx context.Context, user PartialUser) (User, error) {

	if user.Username != nil {
		available, err := s.UserNameAvailable(ctx, *user.Username)
		if err != nil {
			fmt.Println("Error checking username availability:", err)
			return User{}, err
		}
		if !available {
			fmt.Println("Username already exists:", *user.Username)
			return User{}, ErrUsernameAlreadyExists
		}
		matched, err := regexp.MatchString(userNameRegex, *user.Username)
		if err != nil {
			fmt.Println("Error checking username regex:", err)
			return User{}, err
		}
		if !matched {
			return User{}, ErrInvalidUsername
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
