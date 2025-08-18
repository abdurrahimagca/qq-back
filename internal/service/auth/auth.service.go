package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/external/mail"
	"github.com/abdurrahimagca/qq-back/internal/repository/auth"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService struct {
	db       *pgxpool.Pool
	config   *environment.Config
	authRepo *auth.AuthRepository
}

func NewAuthService(db *pgxpool.Pool, config *environment.Config) *AuthService {
	return &AuthService{
		db:       db,
		config:   config,
		authRepo: auth.NewAuthRepository(db),
	}
}

func (s *AuthService) createFirstTimeUserWithOtp(ctx context.Context, email string) error {
	userName := strings.Split(email, "@")[0] + "_" + uuid.New().String()[:8]

	otpCode := uuid.New().String()[:6]

	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])
	provider := strings.Split(email, "@")[1]

	// Start a transaction for atomic user creation
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Will be no-op if tx.Commit() succeeds

	_, err = s.authRepo.CreateFirstTimeUserWithOtp(ctx, tx, auth.CreateFirstTimeUserParams{
		Email:      email,
		Provider:   provider,
		Username:   userName,
		AuthType:   db.AuthProviderEmailOtp,
		OtpCode:    hashedOtpCode,
		ProviderID: "",
	})

	if err != nil {
		return err
	}

	// Commit the transaction before sending email
	if err = tx.Commit(ctx); err != nil {
		return err
	}

	// Send email outside transaction (non-critical operation)
	err = mail.SendOTPMail(ctx, mail.SendOTPMailParams{
		To:     email,
		Code:   otpCode,
		Config: s.config,
	})

	if err != nil {
		return err
	}

	return nil
}
func (s *AuthService) handleAlreadyExistsUser(ctx context.Context, email string, userID pgtype.UUID) error {
	otpCode := uuid.New().String()[:6]

	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])

	// Start a transaction for atomic OTP code insertion
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Will be no-op if tx.Commit() succeeds

	err = s.authRepo.InsertNewOtpCodeForUser(ctx, tx, userID, hashedOtpCode)

	if err != nil {
		return err
	}

	// Commit the transaction before sending email
	if err = tx.Commit(ctx); err != nil {
		return err
	}

	// Send email outside transaction (non-critical operation)
	err = mail.SendOTPMail(ctx, mail.SendOTPMailParams{
		To:     email,
		Code:   otpCode,
		Config: s.config,
	})

	if err != nil {
		return err
	}

	return nil
}

type CreateUserIfNotExistWithOtpServiceResult struct {
	IsNewUser bool
	Error     error
}

func (s *AuthService) CreateUserIfNotExistWithOtpService(ctx context.Context, email string) (CreateUserIfNotExistWithOtpServiceResult, error) {
	// Use a transaction to check if user exists and handle accordingly
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CreateUserIfNotExistWithOtpServiceResult{
			IsNewUser: false,
			Error:     err,
		}, nil
	}
	defer tx.Rollback(ctx) // Will be no-op if tx.Commit() succeeds

	user, _ := s.authRepo.GetUserByEmail(ctx, tx, email)

	// Commit the read transaction
	if err = tx.Commit(ctx); err != nil {
		return CreateUserIfNotExistWithOtpServiceResult{
			IsNewUser: false,
			Error:     err,
		}, nil
	}

	if user != nil {
		return CreateUserIfNotExistWithOtpServiceResult{
			IsNewUser: false,
			Error:     s.handleAlreadyExistsUser(ctx, email, user.ID),
		}, nil
	}

	err = s.createFirstTimeUserWithOtp(ctx, email)

	if err != nil {
		return CreateUserIfNotExistWithOtpServiceResult{
			IsNewUser: false,
			Error:     err,
		}, nil
	}

	return CreateUserIfNotExistWithOtpServiceResult{
		IsNewUser: true,
		Error:     nil,
	}, nil
}
func (s *AuthService) VerifyOtpCodeService(ctx context.Context, email string, otpCode string) (*pgtype.UUID, *string, error) {
	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])

	// Start a transaction for atomic OTP verification
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx) // Will be no-op if tx.Commit() succeeds

	user, err := s.authRepo.GetUserIdAndEmailByOtpCode(ctx, tx, hashedOtpCode)

	if err != nil {
		return nil, nil, err
	}

	if user.Email != email {
		return nil, nil, errors.New("otp code is incorrect")
	}

	err = s.authRepo.DeleteOtpCodeEntryByAuthID(ctx, tx, user.AuthID)

	if err != nil {
		return nil, nil, err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &user.ID, &user.Email, nil
}

func (s *AuthService) GenerateTokens(userID pgtype.UUID) (string, string, error) {
	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": uuid.UUID(userID.Bytes).String(),

			"exp": time.Now().Add(time.Duration(s.config.Token.AccessTokenExpireTime) * time.Minute).Unix(),
			"iat": time.Now().Unix(),
			"iss": s.config.Token.Issuer,
			"aud": s.config.Token.Audience,
		},
	).SignedString([]byte(s.config.Token.Secret))

	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": uuid.UUID(userID.Bytes).String(),
			"exp": time.Now().Add(time.Duration(s.config.Token.RefreshTokenExpireTime) * time.Minute).Unix(),
			"iat": time.Now().Unix(),
			"iss": s.config.Token.Issuer,
			"aud": s.config.Token.Audience,
		},
	).SignedString([]byte(s.config.Token.Secret))

	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
func (s *AuthService) ValidateAndGetUserFromAccessToken(tokenString string, tx pgx.Tx) (user *db.User, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Token.Secret), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)

	// Verify issuer and audience
	if claims["iss"] != s.config.Token.Issuer {
		return nil, errors.New("invalid issuer")
	}
	if claims["aud"] != s.config.Token.Audience {
		return nil, errors.New("invalid audience")
	}

	userIdStr, ok := claims["sub"].(string)
	if !ok || userIdStr == "" {
		return nil, errors.New("invalid user ID in token")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, userErr := s.authRepo.GetUserByID(context.Background(), tx, pgtype.UUID{Bytes: userId, Valid: true})
	if userErr != nil {
		return nil, userErr
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *AuthService) RefreshTokenService(refreshToken string, tx pgx.Tx) (string, string, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Token.Secret), nil
	})

	if err != nil {
		return "", "", err
	}
	if !token.Valid {
		return "", "", errors.New("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)

	userIdStr, ok := claims["sub"].(string)
	if !ok || userIdStr == "" {
		return "", "", errors.New("invalid user ID in token")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return "", "", errors.New("invalid user ID format")
	}

	user, userErr := s.authRepo.GetUserByID(context.Background(), tx, pgtype.UUID{Bytes: userId, Valid: true})
	if userErr != nil {
		return "", "", userErr
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	accessToken, refreshToken, err := s.GenerateTokens(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
