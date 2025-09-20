package auth

import (
	"errors"
)

var (
	ErrInvalidOtpCode = errors.New("invalid otp code")
	ErrInvalidEmail   = errors.New("invalid email")
	ErrNotFound       = errors.New("not found")
)
