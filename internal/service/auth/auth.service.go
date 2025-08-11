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
)

func createFirstTimeUserWithOtp(email string, tx pgx.Tx, config *environment.Config) error {
	userName := strings.Split(email, "@")[0] + "_" + uuid.New().String()[:8]

	otpCode := uuid.New().String()[:6]

	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])
	provider := strings.Split(email, "@")[1]

	_, err := auth.CreateFirstTimeUserWithOtp(context.Background(), tx, auth.CreateFirstTimeUserParams{
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
	err = mail.SendOTPMail(context.Background(), mail.SendOTPMailParams{
		To:     email,
		Code:   otpCode,
		Config: config,
	})

	if err != nil {
		return err
	}

	return nil
}
func handleAlreadyExistsUser(email string, userID pgtype.UUID, tx pgx.Tx, config *environment.Config) error {
	otpCode := uuid.New().String()[:6]

	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])

	err := auth.InsertNewOtpCodeForUser(context.Background(), tx, userID, hashedOtpCode)

	if err != nil {
		return err
	}

	err = mail.SendOTPMail(context.Background(), mail.SendOTPMailParams{
		To:     email,
		Code:   otpCode,
		Config: config,
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

func CreateUserIfNotExistWithOtpService(email string, tx pgx.Tx, config *environment.Config) (CreateUserIfNotExistWithOtpServiceResult, error) {
	user, _ := auth.GetUserByEmail(context.Background(), tx, email)
	if user != nil {
		return CreateUserIfNotExistWithOtpServiceResult{
			IsNewUser: false,
			Error:     handleAlreadyExistsUser(email, user.ID, tx, config),
		}, nil

	}

	err := createFirstTimeUserWithOtp(email, tx, config)

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
func VerifyOtpCodeService(email string, otpCode string, tx pgx.Tx, config *environment.Config) (*pgtype.UUID, *string, error) {
	hash := sha256.Sum256([]byte(otpCode))
	hashedOtpCode := hex.EncodeToString(hash[:])
	
	user, err := auth.GetUserIdAndEmailByOtpCode(context.Background(), tx, hashedOtpCode)

	if err != nil {
		return nil, nil, err
	}

	if user.Email != email {
		return nil, nil, errors.New("otp code is incorrect")
	}

	err = auth.DeleteOtpCodeEntryByAuthID(context.Background(), tx, user.AuthID)

	if err != nil {
		return nil, nil, err
	}

	return &user.ID, &user.Email, nil
}

func GenerateTokens(config *environment.Config, userID pgtype.UUID, userEmail string) (string, string, error) {
	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":   uuid.UUID(userID.Bytes).String(),
			"email": userEmail,
			"exp":   time.Now().Add(time.Duration(config.Token.AccessTokenExpireTime) * time.Minute).Unix(),
			"iat":   time.Now().Unix(),
			"iss":   config.Token.Issuer,
			"aud":   config.Token.Audience,
		},
	).SignedString([]byte(config.Token.Secret))

	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":   uuid.UUID(userID.Bytes).String(),
			"email": userEmail,
			"exp":   time.Now().Add(time.Duration(config.Token.RefreshTokenExpireTime) * time.Minute).Unix(),
			"iat":   time.Now().Unix(),
			"iss":   config.Token.Issuer,
			"aud":   config.Token.Audience,
		},
	).SignedString([]byte(config.Token.Secret))

	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
func ValidateAndGetUserFromAccessToken(tokenString string, config *environment.Config, tx pgx.Tx) (user *db.User, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.Token.Secret), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	
	// Verify issuer and audience
	if claims["iss"] != config.Token.Issuer {
		return nil, errors.New("invalid issuer")
	}
	if claims["aud"] != config.Token.Audience {
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
	
	user, userErr := auth.GetUserByID(context.Background(), tx, pgtype.UUID{Bytes: userId, Valid: true})
	if userErr != nil {
		return nil, userErr
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
