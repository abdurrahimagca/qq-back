package newauth

import (
	"context"
	"errors"
)

type Service interface {
	OtpLoginSignup(ctx context.Context, email string) error
	VerifyOtpCode() error
	RefreshTokens() error

}

type service struct {
	repo Repository
	
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}



func (s *service) OtpLoginSignup(ctx context.Context, email string) error {
	return errors.New("not implemented")
}

func (s *service) VerifyOtpCode() error {
	return errors.New("not implemented")
}

func (s *service) RefreshTokens() error {
	return errors.New("not implemented")
}
