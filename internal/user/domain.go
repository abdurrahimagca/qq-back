package user

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("user not found")
	ErrInvalidID = errors.New("invalid user id")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidUsername = errors.New("invalid username")
)

type PrivacyLevel string

const (
	PrivacyLevelPublic      PrivacyLevel = "public"
	PrivacyLevelPrivate     PrivacyLevel = "private"
	PrivacyLevelFullPrivate PrivacyLevel = "full_private"
)

type User struct {
	ID           uuid.UUID
	AuthID       uuid.UUID // Foreign key to auth table
	Username     string
	DisplayName  *string
	PrivacyLevel PrivacyLevel
	AvatarKey    *string
}
type ReadUser struct {
	Username        string
	DisplayName     *string
	PrivacyLevel    PrivacyLevel
	AvatarSignedUrl *string
}

type PartialUser struct {
	ID           uuid.UUID
	DisplayName  *string
	PrivacyLevel *PrivacyLevel
	Username     *string
}
const userNameRegex = `^[a-zA-Z0-9_-]+$`