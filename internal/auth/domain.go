package auth

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidOtpCode = errors.New("invalid otp code")
	ErrInvalidEmail   = errors.New("invalid email")
	ErrNotFound       = errors.New("not found")
)

type GetUserIdAndEmailByOtpCodeResult struct {
	ID    uuid.UUID
	Email string
}

