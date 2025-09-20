package qqerrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	SQL_UNIQUE_VIOLATION      = "23505"
	SQL_FOREIGN_KEY_VIOLATION = "23503"
	SQL_CHECK_VIOLATION       = "23514"
	SQL_NOT_NULL_VIOLATION    = "23502"
)

type QQError struct {
	Message    string
	StatusCode int
	Original   error
}

func (e *QQError) Error() string {
	return e.Message
}

func (e *QQError) Unwrap() error {
	return e.Original
}

var errMap = map[string]*QQError{
	SQL_UNIQUE_VIOLATION:      {Message: ErrUniqueViolation.Error(), StatusCode: http.StatusConflict, Original: ErrUniqueViolation},
	SQL_FOREIGN_KEY_VIOLATION: {Message: ErrConstraintViolation.Error(), StatusCode: http.StatusBadRequest, Original: ErrConstraintViolation},
	SQL_CHECK_VIOLATION:       {Message: ErrValidationError.Error(), StatusCode: http.StatusBadRequest, Original: ErrValidationError},
	SQL_NOT_NULL_VIOLATION:    {Message: ErrValidationError.Error(), StatusCode: http.StatusBadRequest, Original: ErrValidationError},
}

func GetDbErrAsQQError(err error) *QQError {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return &QQError{
			Message:    ErrNotFound.Error(),
			StatusCode: http.StatusNotFound,
			Original:   ErrNotFound,
		}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if stashlyErr, exists := errMap[pgErr.Code]; exists {
			return &QQError{
				Message:    stashlyErr.Message,
				StatusCode: stashlyErr.StatusCode,
				Original:   stashlyErr.Original,
			}
		}
	}

	return &QQError{
		Message:    fmt.Sprintf("database error: %v", err),
		StatusCode: http.StatusInternalServerError,
		Original:   err,
	}
}

func GetHumaErrorFromError(err error) huma.StatusError {
	// Check if error is already a properly typed SError, and if so, use its status code directly
	var qqErr *QQError
	if errors.As(err, &qqErr) {
		switch qqErr.StatusCode {
		case http.StatusNotFound:
			return huma.Error404NotFound("Not found", err)
		case http.StatusConflict:
			return huma.Error409Conflict("Unique violation", err)
		case http.StatusUnprocessableEntity:
			return huma.Error422UnprocessableEntity("Validation error", err)
		case http.StatusBadRequest:
			return huma.Error400BadRequest("Constraint violation", err)
		default:
			return huma.Error500InternalServerError("Internal server error", err)
		}
	}

	// Handle base errors
	switch err {
	case ErrNotFound:
		return huma.Error404NotFound("Not found", err)
	case ErrValidationError:
		return huma.Error422UnprocessableEntity("Validation error", err)
	case ErrInternalServer:
		return huma.Error500InternalServerError("Internal server error", err)
	case ErrUniqueViolation:
		return huma.Error409Conflict("Unique violation", err)
	case ErrConstraintViolation:
		return huma.Error400BadRequest("Constraint violation", err)
	case ErrDuplicateRow:
		return huma.Error409Conflict("Duplicate row", err)
	default:
		return huma.Error500InternalServerError("Internal server error", err)
	}
}
