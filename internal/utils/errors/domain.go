package qqerrors

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrValidationError     = errors.New("validation error")
	ErrUniqueViolation     = errors.New("unique violation")
	ErrConstraintViolation = errors.New("constraint violation")
	ErrDuplicateRow        = errors.New("duplicate row")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInternalServer      = errors.New("internal server error")
)

