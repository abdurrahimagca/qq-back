package user

import (
	"context"
	"github.com/google/uuid"
)

type Service interface {
    // Transaction detayı (tx) artık burada yok!
	CreateDefaultUser(ctx context.Context) (User, error)
	GetUserByID(ctx context.Context, userID string) (User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateDefaultUser(ctx context.Context) (User, error) {
	// Service'in görevi iş mantığını uygulamaktır: "default username" oluşturmak.
	prefix := "user_"
	username := prefix + uuid.New().String()[:12] // Biraz daha uzun ve eşsiz olması için

    // Bu metot artık bir tx almaz. Transaction yönetimi UseCase'in işidir.
    // UseCase, bu servise transaction'a bağlı bir repository (`repo.WithTx(tx)`) verecektir.
	return s.repo.CreateUser(ctx, username)
}

func (s *service) GetUserByID(ctx context.Context, userID string) (User, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, ErrInvalidID
	}

	return s.repo.GetUserByID(ctx, parsedID)
}