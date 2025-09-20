package user

import (
	"errors"

	
)

var (
	ErrNotFound  = errors.New("user not found")
	ErrInvalidID = errors.New("invalid user id")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidUsername = errors.New("invalid username")
)

