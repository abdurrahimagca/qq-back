package user

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("user not found")
	ErrInvalidID = errors.New("invalid user id")
)

// Bu bizim saf domain modelimiz. JSON veya DB tag'leri içermez.
type User struct {
	ID           uuid.UUID
	AuthID       uuid.UUID // Foreign key to auth table
	Username     string
	DisplayName  *string
	PrivacyLevel string
	AvatarKey    *string
}
