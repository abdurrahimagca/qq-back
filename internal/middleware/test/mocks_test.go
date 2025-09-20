package middleware_test

import (
	"context"
	"errors"

	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockTokenService implements token.Service for testing.
type MockTokenService struct {
	ValidateTokenFunc func(ctx context.Context, params token.ValidateTokenParams) (token.ValidateTokenResult, error)
	ValidateCalls     []token.ValidateTokenParams
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{
		ValidateCalls: make([]token.ValidateTokenParams, 0),
	}
}

func (m *MockTokenService) GenerateTokens(
	ctx context.Context, params token.GenerateTokenParams,
) (token.GenerateTokenResult, error) {
	// Not used in middleware tests.
	return token.GenerateTokenResult{}, errors.New("not implemented in mock")
}

func (m *MockTokenService) ValidateToken(
	ctx context.Context, params token.ValidateTokenParams,
) (token.ValidateTokenResult, error) {
	m.ValidateCalls = append(m.ValidateCalls, params)

	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, params)
	}

	// Default success behavior.
	return token.ValidateTokenResult{
		Claims: &token.Claims{
			UserID: "550e8400-e29b-41d4-a716-446655440000",
		},
	}, nil
}

// SetValidateTokenResult configures the mock to return specific result.
func (m *MockTokenService) SetValidateTokenResult(userID string, err error) {
	m.ValidateTokenFunc = func(
		ctx context.Context, params token.ValidateTokenParams,
	) (token.ValidateTokenResult, error) {
		if err != nil {
			return token.ValidateTokenResult{}, err
		}
		return token.ValidateTokenResult{
			Claims: &token.Claims{
				UserID: userID,
			},
		}, nil
	}
}

// SetValidateTokenError configures the mock to return an error.
func (m *MockTokenService) SetValidateTokenError(err error) {
	m.ValidateTokenFunc = func(
		ctx context.Context, params token.ValidateTokenParams,
	) (token.ValidateTokenResult, error) {
		return token.ValidateTokenResult{}, err
	}
}

// GetValidateTokenCallCount returns the number of times ValidateToken was called.
func (m *MockTokenService) GetValidateTokenCallCount() int {
	return len(m.ValidateCalls)
}

// GetLastValidateTokenCall returns the last call to ValidateToken.
func (m *MockTokenService) GetLastValidateTokenCall() *token.ValidateTokenParams {
	if len(m.ValidateCalls) == 0 {
		return nil
	}
	return &m.ValidateCalls[len(m.ValidateCalls)-1]
}

// Reset clears all recorded calls and resets behavior.
func (m *MockTokenService) Reset() {
	m.ValidateCalls = make([]token.ValidateTokenParams, 0)
	m.ValidateTokenFunc = nil
}

// MockUserService implements user.Service for testing.
type MockUserService struct {
	GetUserByIDFunc  func(ctx context.Context, userID pgtype.UUID) (*db.User, error)
	GetUserByIDCalls []pgtype.UUID
	Users            map[string]*db.User // keyed by UUID string
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		GetUserByIDCalls: make([]pgtype.UUID, 0),
		Users:            make(map[string]*db.User),
	}
}

func (m *MockUserService) CreateDefaultUserWithAuthID(ctx context.Context, authID pgtype.UUID) (*db.User, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID pgtype.UUID) (*db.User, error) {
	m.GetUserByIDCalls = append(m.GetUserByIDCalls, userID)

	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}

	// Check if user exists in mock storage.
	if user, exists := m.Users[userID.String()]; exists {
		return user, nil
	}

	return nil, errors.New("user not found")
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockUserService) UpdateUser(ctx context.Context, user db.UpdateUserParams) (*db.User, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockUserService) UserNameAvailable(ctx context.Context, username string) (bool, error) {
	return false, errors.New("not implemented in mock")
}

func (m *MockUserService) WithTx(tx pgx.Tx) user.Service {
	_ = tx
	return m
}

// AddUser adds a user to the mock storage.
func (m *MockUserService) AddUser(userID string, user *db.User) {
	m.Users[userID] = user
}

// SetGetUserByIDResult configures the mock to return specific result.
func (m *MockUserService) SetGetUserByIDResult(user *db.User, err error) {
	m.GetUserByIDFunc = func(
		ctx context.Context, userID pgtype.UUID,
	) (*db.User, error) {
		return user, err
	}
}

// SetGetUserByIDError configures the mock to return an error.
func (m *MockUserService) SetGetUserByIDError(err error) {
	m.GetUserByIDFunc = func(
		ctx context.Context, userID pgtype.UUID,
	) (*db.User, error) {
		return nil, err
	}
}

// GetGetUserByIDCallCount returns the number of times GetUserByID was called.
func (m *MockUserService) GetGetUserByIDCallCount() int {
	return len(m.GetUserByIDCalls)
}

// GetLastGetUserByIDCall returns the last call to GetUserByID.
func (m *MockUserService) GetLastGetUserByIDCall() *pgtype.UUID {
	if len(m.GetUserByIDCalls) == 0 {
		return nil
	}
	return &m.GetUserByIDCalls[len(m.GetUserByIDCalls)-1]
}

// Reset clears all recorded calls and resets behavior.
func (m *MockUserService) Reset() {
	m.GetUserByIDCalls = make([]pgtype.UUID, 0)
	m.GetUserByIDFunc = nil
	m.Users = make(map[string]*db.User)
}

// Common errors for testing.
var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrUserNotFound  = errors.New("user not found")
	ErrDatabaseError = errors.New("database connection failed")
	ErrInvalidUserID = errors.New("invalid user ID format")
)
