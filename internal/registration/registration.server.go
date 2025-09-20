package registration

import (
	"context"

	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/danielgtaylor/huma/v2"
)

type registrationServer struct {
	uc RegistrationUsecase
}
type RegistrationServer interface {
	SendOtpHandler(ctx context.Context, input *SendOtpInput) (*SendOtpOutput, error)
	VerifyOtpHandler(ctx context.Context, input *VerifyOtpInput) (*VerifyOtpOutput, error)
	RefreshTokensHandler(ctx context.Context, input *RefreshTokensInput) (*RefreshTokensOutput, error)
	RegisterRegistrationEndpoints(api huma.API)
}

func NewRegistrationServer(uc RegistrationUsecase) RegistrationServer {
	return &registrationServer{uc: uc}
}

func (s *registrationServer) SendOtpHandler(ctx context.Context, input *SendOtpInput) (*SendOtpOutput, error) {
	isNewUser, err := s.uc.RegisterOrLoginOTP(ctx, input.Body.Email)
	if err != nil {
		return nil, qqerrors.GetHumaErrorFromError(err)
	}

	return &SendOtpOutput{
		Body: struct {
			Data SendOtpData
		}{
			Data: SendOtpData{
				IsNewUser: *isNewUser,
			},
		},
	}, nil
}

func (s *registrationServer) VerifyOtpHandler(ctx context.Context, input *VerifyOtpInput) (*VerifyOtpOutput, error) {
	tokenPair, err := s.uc.VerifyOTPAndLogin(ctx, input.Body.Email, input.Body.OtpCode)
	if err != nil {
		return nil, qqerrors.GetHumaErrorFromError(err)
	}
	return &VerifyOtpOutput{
		Body: struct {
			Data TokenData
		}{
			Data: TokenData{
				AccessToken:  tokenPair.AccessToken,
				RefreshToken: tokenPair.RefreshToken,
			},
		},
	}, nil
}

func (s *registrationServer) RefreshTokensHandler(
	ctx context.Context, input *RefreshTokensInput) (*RefreshTokensOutput, error) {
	tokenPair, err := s.uc.RefreshTokens(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, qqerrors.GetHumaErrorFromError(err)
	}
	return &RefreshTokensOutput{
		Body: struct {
			Data TokenData
		}{
			Data: TokenData{
				AccessToken:  tokenPair.AccessToken,
				RefreshToken: tokenPair.RefreshToken,
			},
		},
	}, nil
}

func (s *registrationServer) RegisterRegistrationEndpoints(api huma.API) {
	huma.Register(api, operations[SendOtp], s.SendOtpHandler)
	huma.Register(api, operations[VerifyOtp], s.VerifyOtpHandler)
	huma.Register(api, operations[RefreshTokens], s.RefreshTokensHandler)
}
