package auth

import (
	"encoding/json"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/api"
	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	authService "github.com/abdurrahimagca/qq-back/internal/service/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"regexp"
)

type AuthHandler struct {
	authService *authService.AuthService
	api.StrictHandlerFunc
}

func NewAuthHandler(db *pgxpool.Pool, config *environment.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService.NewAuthService(db, config),
	}
}

func (h *AuthHandler) PostAuthOtp(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body api.PostAuthOtpJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	const email_regex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Validate email
	if body.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if !regexp.MustCompile(email_regex).MatchString(body.Email) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Call service to create user if not exists and send OTP
	result, err := h.authService.CreateUserIfNotExistWithOtpService(r.Context(), body.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.Error != nil {
		http.Error(w, "Failed to process OTP request", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"isNewUser": result.IsNewUser,
			"message":   "OTP sent successfully",
		},
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PostAuthOtpVerify handles OTP verification
func (h *AuthHandler) PostAuthOtpVerify(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body api.PostAuthOtpVerifyJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if body.Email == "" || body.OtpCode == "" {
		http.Error(w, "Email and OTP code are required", http.StatusBadRequest)
		return
	}

	// Verify OTP
	userID, email, err := h.authService.VerifyOtpCodeService(r.Context(), body.Email, body.OtpCode)
	if err != nil {
		http.Error(w, "Invalid OTP code", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.authService.GenerateTokens(*userID)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"user": map[string]interface{}{
				"id":    userID,
				"email": email,
			},
		},
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
