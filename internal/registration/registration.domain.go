package registration

import (
	"github.com/danielgtaylor/huma/v2"
)

var moduleErrors = []int{400, 401, 403, 404, 500}
var moduleTags = []string{"Registration"}

const (
	SendOtp       = "sendOtp"
	VerifyOtp     = "verifyOtp"
	RefreshTokens = "refreshTokens"
)

var operations = map[string]huma.Operation{
	SendOtp: {
		Method:      "POST",
		Path:        "/auth/send-otp",
		Summary:     "Send OTP code to email for existing users or create new user account",
		Description: "Send OTP code to email for existing users or create new user account",
		OperationID: SendOtp,
		Errors:      moduleErrors,
		Tags:        moduleTags,
	},
	VerifyOtp: {
		Method:      "POST",
		Path:        "/auth/verify-otp",
		Summary:     "Verify OTP code",
		Description: "Verify OTP code",
		OperationID: VerifyOtp,
		Errors:      moduleErrors,
		Tags:        moduleTags,
	},
	RefreshTokens: {
		Method:      "POST",
		Path:        "/auth/refresh-tokens",
		Summary:     "Refresh tokens",
		Description: "Refresh tokens",
		OperationID: RefreshTokens,
		Errors:      moduleErrors,
		Tags:        moduleTags,
	},
}

type SendOtpInput struct {
	Body struct {
		Email string `json:"email" doc:"Email address of the user" required:"true" format:"email"`
	}
}

type SendOtpData struct {
	IsNewUser bool `json:"isNewUser"`
}

type SendOtpOutput struct {
	Body struct {
		Data SendOtpData
	}
}

type VerifyOtpInput struct {
	Body struct {
		Email   string `json:"email" doc:"Email address of the user" required:"true" format:"email"`
		OtpCode string `json:"otpCode" doc:"OTP code received via email" format:"number" required:"true" minLength:"6" maxLength:"6"`
	}
}

type TokenData struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type VerifyOtpOutput struct {
	Body struct {
		Data TokenData
	}
}

type RefreshTokensInput struct {
	Body struct {
		RefreshToken string `json:"refreshToken" doc:"Refresh token" required:"true"`
	}
}

type RefreshTokensOutput struct {
	Body struct {
		Data TokenData
	}
}
