package qqerrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestQQError_Error(t *testing.T) {
	serr := &QQError{
		Message:    "test error",
		StatusCode: http.StatusBadRequest,
		Original:   errors.New("original error"),
	}

	if serr.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", serr.Error())
	}
}

func TestQQError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	serr := &QQError{
		Message:    "test error",
		StatusCode: http.StatusBadRequest,
		Original:   originalErr,
	}

	if serr.Unwrap() != originalErr {
		t.Errorf("Expected original error, got %v", serr.Unwrap())
	}
}

func TestGetDbErrAsQQError_NilError(t *testing.T) {
	result := GetDbErrAsQQError(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestGetDbErrAsQQError_NoRows(t *testing.T) {
	result := GetDbErrAsQQError(pgx.ErrNoRows)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	if result.Message != ErrNotFound.Error() {
		t.Errorf("Expected '%s', got '%s'", ErrNotFound.Error(), result.Message)
	}

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected %d, got %d", http.StatusNotFound, result.StatusCode)
	}

	if result.Original != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", result.Original)
	}
}

func TestGetDbErrAsQQError_UniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    SQL_UNIQUE_VIOLATION,
		Message: "duplicate key value violates unique constraint",
	}

	result := GetDbErrAsQQError(pgErr)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	if result.Message != ErrUniqueViolation.Error() {
		t.Errorf("Expected '%s', got '%s'", ErrUniqueViolation.Error(), result.Message)
	}

	if result.StatusCode != http.StatusConflict {
		t.Errorf("Expected %d, got %d", http.StatusConflict, result.StatusCode)
	}
}

func TestGetDbErrAsQQError_ForeignKeyViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    SQL_FOREIGN_KEY_VIOLATION,
		Message: "foreign key constraint violation",
	}

	result := GetDbErrAsQQError(pgErr)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	if result.Message != ErrConstraintViolation.Error() {
		t.Errorf("Expected '%s', got '%s'", ErrConstraintViolation.Error(), result.Message)
	}

	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, result.StatusCode)
	}
}

func TestGetDbErrAsQQError_CheckViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    SQL_CHECK_VIOLATION,
		Message: "check constraint violation",
	}

	result := GetDbErrAsQQError(pgErr)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	if result.Message != ErrValidationError.Error() {
		t.Errorf("Expected '%s', got '%s'", ErrValidationError.Error(), result.Message)
	}

	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, result.StatusCode)
	}
}

func TestGetDbErrAsQQError_NotNullViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    SQL_NOT_NULL_VIOLATION,
		Message: "not null constraint violation",
	}

	result := GetDbErrAsQQError(pgErr)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	if result.Message != ErrValidationError.Error() {
		t.Errorf("Expected '%s', got '%s'", ErrValidationError.Error(), result.Message)
	}

	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, result.StatusCode)
	}
}

func TestGetDbErrAsQQError_UnknownPgError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    "99999",
		Message: "unknown database error",
	}

	result := GetDbErrAsQQError(pgErr)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	expectedMessage := "database error: : unknown database error (SQLSTATE 99999)"
	if result.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestGetDbErrAsQQError_GenericError(t *testing.T) {
	genericErr := errors.New("some generic error")

	result := GetDbErrAsQQError(genericErr)

	if result == nil {
		t.Fatal("Expected SError, got nil")
	}

	expectedMessage := "database error: some generic error"
	if result.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestGetHumaErrorFromError_ErrNotFound(t *testing.T) {
	result := GetHumaErrorFromError(ErrNotFound)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, result.GetStatus())
	}

	if result.Error() != "Not found" {
		t.Errorf("Expected message 'Not found', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_ErrValidationError(t *testing.T) {
	result := GetHumaErrorFromError(ErrValidationError)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusUnprocessableEntity {
		t.Errorf("Expected status %d, got %d", http.StatusUnprocessableEntity, result.GetStatus())
	}

	if result.Error() != "Validation error" {
		t.Errorf("Expected message 'Validation error', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_ErrInternalServer(t *testing.T) {
	result := GetHumaErrorFromError(ErrInternalServer)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, result.GetStatus())
	}

	if result.Error() != "Internal server error" {
		t.Errorf("Expected message 'Internal server error', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_ErrUniqueViolation(t *testing.T) {
	result := GetHumaErrorFromError(ErrUniqueViolation)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, result.GetStatus())
	}

	if result.Error() != "Unique violation" {
		t.Errorf("Expected message 'Unique violation', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_ErrConstraintViolation(t *testing.T) {
	result := GetHumaErrorFromError(ErrConstraintViolation)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, result.GetStatus())
	}

	if result.Error() != "Constraint violation" {
		t.Errorf("Expected message 'Constraint violation', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_ErrDuplicateRow(t *testing.T) {
	result := GetHumaErrorFromError(ErrDuplicateRow)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, result.GetStatus())
	}

	if result.Error() != "Duplicate row" {
		t.Errorf("Expected message 'Duplicate row', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_UnknownError(t *testing.T) {
	unknownErr := errors.New("unknown error")
	result := GetHumaErrorFromError(unknownErr)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, result.GetStatus())
	}

	if result.Error() != "Internal server error" {
		t.Errorf("Expected message 'Internal server error', got '%s'", result.Error())
	}
}

func TestGetHumaErrorFromError_NilError(t *testing.T) {
	result := GetHumaErrorFromError(nil)

	if result == nil {
		t.Fatal("Expected huma.StatusError, got nil")
	}

	if result.GetStatus() != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, result.GetStatus())
	}

	if result.Error() != "Internal server error" {
		t.Errorf("Expected message 'Internal server error', got '%s'", result.Error())
	}
}
