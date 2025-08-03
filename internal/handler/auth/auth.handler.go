package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/service/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db     *pgxpool.Pool
	config *environment.Config
}

func NewHandler(db *pgxpool.Pool, config *environment.Config) *Handler {
	return &Handler{db: db, config: config}
}

type SignupRequest struct {
	Email string `json:"email"`
}

type SignupResponse struct {
	Message string `json:"message"`
}

type SigninRequest struct {
	Email   string `json:"email"`
	OtpCode string `json:"otp_code"`
}

type SigninResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func (h *Handler) SignInOrUpWithOtp(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	tx, err := h.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	err = auth.CreateUserIfNotExistWithOtpService(req.Email, tx, h.config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := tx.Commit(context.Background()); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignupResponse{
		Message: "OTP code sent to your email",
	})
}

