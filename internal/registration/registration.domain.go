package registration

import (
	"github.com/danielgtaylor/huma/v2"
)


var moduleErrors = []int{400, 401, 403, 404, 500}
var moduleTags = []string{"Registration"}
const (
	SendOtp = "sendOtp"
	VerifyOtp = "verifyOtp"
	RefreshTokens = "refreshTokens"
)
var operations = map[string]huma.Operation{
	SendOtp: {
		Summary: "Send OTP code to email for existing users or create new user account",
		Description: "Send OTP code to email for existing users or create new user account",
		OperationID: SendOtp,
		Tags: moduleTags,
	},
	VerifyOtp: {
		Summary: "Verify OTP code",
		Description: "Verify OTP code",
		OperationID: VerifyOtp,
		Tags: moduleTags,
	},
	RefreshTokens: {
		Summary: "Refresh tokens",
		Description: "Refresh tokens",
		OperationID: RefreshTokens,
		Tags: moduleTags,
	},
}

type SendOtpInput struct {
	Body struct {
		Email string `json:"email required format:email"`
	}
}

type SendOtpOutput struct {
	Data struct {
		IsNewUser bool `json:"isNewUser"`
	}
}


type VerifyOtpInput struct {
	Body struct {
	Email string `json:"email required format:email"`
	OtpCode string `json:"otpCode required minLength:6 maxLength:6"`
}
}

type VerifyOtpOutput struct {
	Data struct {
		AccessToken string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
}

type RefreshTokensInput struct {
	Body struct {
		RefreshToken string `json:"refreshToken required"`
	}
}

type RefreshTokensOutput struct {
	Data struct {
		AccessToken string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
}
