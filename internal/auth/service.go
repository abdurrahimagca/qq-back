package newauth

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	resend_mail "github.com/abdurrahimagca/qq-back/internal/external/mail"
	"github.com/abdurrahimagca/qq-back/internal/tokens"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	OtpLoginSignup(ctx context.Context, params LoginOtpParams) (LoginOtpResult, error)
	VerifyOtpCode(ctx context.Context, params VerifyOtpCodeParams) (VerifyOtpCodeResult, error)
	RefreshTokens(ctx context.Context, params RefreshTokensParams) (RefreshTokensResult, error)
	ValidateAccessAndGetUser(ctx context.Context, params ValidateAccessAndGetUserParams) (ValidateAccessAndGetUserResult, error)
}

type service struct {
	repo   Repository
	db     *pgxpool.Pool
	config *environment.Environment
	tokens tokens.Service
}

func NewService(repo Repository, db *pgxpool.Pool, config *environment.Environment, tokens tokens.Service) Service {
	return &service{
		repo:   repo,
		db:     db,
		config: config,
		tokens: tokens,
	}
}

func (s *service) InsertNewOtpCode(ctx context.Context, params InsertNewOtpCodeForUserParams) (InsertNewOtpCodeForUserResult, error) {
	return s.repo.InsertNewOtpCodeForUser(ctx, params)
}

func (s *service) handleUserExists(ctx context.Context, params LoginOtpParams, userID pgtype.UUID) (LoginOtpResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginOtpResult{}, err
	}
	defer tx.Rollback(ctx)

	_, err = s.repo.DeleteOtpCodesByEmail(ctx, DeleteOtpCodesByEmailParams(params))
	if err != nil {
		return LoginOtpResult{}, err
	}

	codePrefix := "QQ"
	otpCode := strings.ToUpper(codePrefix + uuid.New().String()[:6])
	otpHash := sha256.New().Sum([]byte(otpCode))

	_, err = s.repo.InsertNewOtpCodeForUser(ctx, InsertNewOtpCodeForUserParams{
		UserID: userID,
		Code:   string(otpHash),
	})
	if err != nil {
		return LoginOtpResult{}, err
	}

	err = resend_mail.SendOTPMail(ctx, resend_mail.SendOTPMailParams{
		To:     params.Email,
		Code:   otpCode,
		Config: s.config,
	})
	if err != nil {
		return LoginOtpResult{}, err
	}
	tx.Commit(ctx)

	return LoginOtpResult{
		UserID: &userID,
	}, nil
}
func (s *service) handleUserNotFound(ctx context.Context, params LoginOtpParams) (LoginOtpResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginOtpResult{}, err
	}
	defer tx.Rollback(ctx)

	codePrefix := "QQ"
	otpCode := strings.ToUpper(codePrefix + uuid.New().String()[:6])
	otpHash := sha256.New().Sum([]byte(otpCode))
	username := "user_" + uuid.New().String()[:10]

	result, err := s.repo.CreateNewUserOtp(ctx, tx, CreateNewUserOtpParams{
		Email:    params.Email,
		Username: username,
		OtpCode:  string(otpHash),
	})

	if err != nil {
		return LoginOtpResult{}, err
	}

	err = resend_mail.SendOTPMail(ctx, resend_mail.SendOTPMailParams{
		To:     params.Email,
		Code:   otpCode,
		Config: s.config,
	})
	if err != nil {
		return LoginOtpResult{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return LoginOtpResult{}, err
	}

	return LoginOtpResult{
		UserID: result.UserID,
	}, nil
}

func (s *service) OtpLoginSignup(ctx context.Context, params LoginOtpParams) (LoginOtpResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, GetUserByEmailParams(params))
	if err != nil || !user.User.ID.Valid {
		return s.handleUserNotFound(ctx, params)
	}
	return s.handleUserExists(ctx, params, user.User.ID)
}

func (s *service) VerifyOtpCode(ctx context.Context, params VerifyOtpCodeParams) (VerifyOtpCodeResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return VerifyOtpCodeResult{}, err
	}
	defer tx.Rollback(ctx)

	result, err := s.repo.GetUserIdAndEmailByOtpCodeAndDelete(tx, ctx, GetUserIdAndEmailByOtpCodeParams{
		OtpCode: params.Code,
	})
	if err != nil {
		return VerifyOtpCodeResult{}, err
	}
	if result.UserID == nil {
		return VerifyOtpCodeResult{}, errors.New("invalid otp code")
	}
	tokens, err := s.tokens.GenerateTokens(tokens.GenerateTokenParams{
		UserID: result.UserID.String(),
	})
	if err != nil {
		return VerifyOtpCodeResult{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return VerifyOtpCodeResult{}, err
	}

	return VerifyOtpCodeResult{
		Tokens: tokens.Tokens,
	}, nil
}

func (s *service) RefreshTokens(ctx context.Context, params RefreshTokensParams) (RefreshTokensResult, error) {
	claims, err := s.tokens.ValidateToken(tokens.ValidateTokenParams{
		Token: params.RefreshToken,
	})
	if err != nil {
		return RefreshTokensResult{}, err
	}
	userID, err := uuid.Parse(claims.Claims.UserID)
	if err != nil {
		return RefreshTokensResult{}, err
	}
	user, err := s.repo.GetUserByID(ctx, GetUserByIDParams{
		ID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	})
	if err != nil || !user.User.ID.Valid {
		return RefreshTokensResult{}, err
	}
	tokens, err := s.tokens.GenerateTokens(tokens.GenerateTokenParams{
		UserID: user.User.ID.String(),
	})
	if err != nil {
		return RefreshTokensResult{}, err
	}
	return RefreshTokensResult{
		Tokens: tokens.Tokens,
	}, nil
}

func (s *service) ValidateAccessAndGetUser(ctx context.Context, params ValidateAccessAndGetUserParams) (ValidateAccessAndGetUserResult, error) {
	claims, err := s.tokens.ValidateToken(tokens.ValidateTokenParams{
		Token: params.AccessToken,
	})
	if err != nil {
		return ValidateAccessAndGetUserResult{}, err
	}
	user, err := s.repo.GetUserByID(ctx, GetUserByIDParams{
		ID: pgtype.UUID{
			Bytes: uuid.MustParse(claims.Claims.UserID),
			Valid: true,
		},
	})
	if err != nil || !user.User.ID.Valid {
		return ValidateAccessAndGetUserResult{}, err
	}
	return ValidateAccessAndGetUserResult{
		User: &user.User,
	}, nil
}
