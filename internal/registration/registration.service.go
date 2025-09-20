package registration

import (
	"context"

	"strings"

	"errors"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	mail "github.com/abdurrahimagca/qq-back/internal/platform/mailer"
	tokenport "github.com/abdurrahimagca/qq-back/internal/platform/token"
	"github.com/abdurrahimagca/qq-back/internal/user"
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Usecase interface {
	RegisterOrLoginOTP(ctx context.Context, email string) (*bool, error)
	VerifyOTPAndLogin(ctx context.Context, email string, otp string) (tokenport.GenerateTokenResult, error)
	RefreshTokens(ctx context.Context, refreshToken string) (tokenport.GenerateTokenResult, error)
}

type registrationUsecase struct {
	mailer       mail.Service
	authService  auth.Service
	userService  user.Service
	dbpool       *pgxpool.Pool
	tokenService tokenport.Service
}

func NewUsecase(
	mailer mail.Service,
	authService auth.Service,
	userService user.Service,
	pool *pgxpool.Pool,
	tokenService tokenport.Service,
) Usecase {
	return &registrationUsecase{
		mailer:       mailer,
		authService:  authService,
		userService:  userService,
		dbpool:       pool,
		tokenService: tokenService,
	}
}

func (uc *registrationUsecase) RegisterOrLoginOTP(ctx context.Context, emailAddr string) (*bool, error) {
	tx, err := uc.dbpool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	txAuthService := uc.authService.WithTx(tx)
	txUserService := uc.userService.WithTx(tx)
	var isNewUser bool

	foundUser, err := txUserService.GetUserByEmail(ctx, emailAddr)
	if err != nil && !errors.Is(err, qqerrors.ErrNotFound) {
		return nil, err
	}

	var authID pgtype.UUID
	if foundUser != nil && foundUser.ID.Valid {
		isNewUser = false
		authID = foundUser.AuthID
	} else {
		isNewUser = true
		authIDPtr, createAuthErr := txAuthService.CreateNewAuthForOTPLogin(ctx, emailAddr)
		if createAuthErr != nil {
			return nil, createAuthErr
		}
		authID = *authIDPtr

		foundUser, err = txUserService.CreateDefaultUserWithAuthID(ctx, authID)
		if err != nil {
			return nil, err
		}
	}

	err = txAuthService.KillOrphanedOTPsByUserID(ctx, foundUser.ID)
	if err != nil {
		return nil, err
	}

	otp, err := txAuthService.GenerateAndSaveOTPForAuth(ctx, authID)
	if err != nil {
		return nil, err
	}

	template, err := uc.mailer.GetTemplate(ctx, "otp")
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, commitErr
	}
	tx = nil

	body := strings.Replace(template, "{{.OTP}}", otp, 1)
	err = uc.mailer.SendEmail(ctx, mail.SendParams{
		From:    "qq@homelab-kaleici.space",
		To:      emailAddr,
		Subject: "OTP Verification",
		Body:    body,
	})
	if err != nil {
		return nil, err
	}

	return &isNewUser, nil
}
func (uc *registrationUsecase) VerifyOTPAndLogin(
	ctx context.Context, emailAddr string, otp string,
) (tokenport.GenerateTokenResult, error) {
	tx, err := uc.dbpool.Begin(ctx)
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	txAuthService := uc.authService.WithTx(tx)
	txUserService := uc.userService.WithTx(tx)

	err = txAuthService.VerifyOTP(ctx, emailAddr, otp)
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	user, err := txUserService.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	err = txAuthService.KillOrphanedOTPsByUserID(ctx, user.ID)
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return tokenport.GenerateTokenResult{}, commitErr
	}
	tx = nil

	tokenPair, err := uc.tokenService.GenerateTokens(ctx, tokenport.GenerateTokenParams{
		UserID: user.ID.String(),
	})
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	return tokenPair, nil
}
func (uc *registrationUsecase) RefreshTokens(
	ctx context.Context, refreshToken string,
) (tokenport.GenerateTokenResult, error) {
	tokenResult, err := uc.tokenService.ValidateToken(ctx, tokenport.ValidateTokenParams{
		Token: refreshToken,
	})
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	userID := tokenResult.Claims.UserID
	if userID == "" {
		return tokenport.GenerateTokenResult{}, qqerrors.ErrValidationError
	}

	userUUID := pgtype.UUID{}
	if scanErr := userUUID.Scan(userID); scanErr != nil {
		return tokenport.GenerateTokenResult{}, scanErr
	}
	user, err := uc.userService.GetUserByID(ctx, userUUID)
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	newTokens, err := uc.tokenService.GenerateTokens(ctx, tokenport.GenerateTokenParams{
		UserID: user.ID.String(),
	})
	if err != nil {
		return tokenport.GenerateTokenResult{}, err
	}

	return newTokens, nil
}
