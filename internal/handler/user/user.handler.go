package user

import (
	"encoding/json"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	dbModule "github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/middleware"
	"github.com/abdurrahimagca/qq-back/internal/service/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UpdateUserProfileRequest struct {
	DisplayName   string `json:"displayName"`
	Username      string `json:"username"`
	PrivacyLevel *dbModule.PrivacyLevel `json:"privacyLevel"`
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, config environment.Config) {
	// Get User from context (set by auth middleware).
	userFromCtx, ok := r.Context().Value(middleware.UserContextKey).(*dbModule.User)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}
	userID := userFromCtx.ID

	// Get Transaction from context (started by middleware).
	tx, ok := middleware.GetTxFromContext(r.Context())
	if !ok {
		http.Error(w, "Transaction not found in context", http.StatusInternalServerError)
		return
	}
	request := UpdateUserProfileRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Could not decode JSON data", http.StatusBadRequest)
		return
	}

	// Call the service with the transaction from the middleware.
	params := user.UpdateUserParams{}
	
	if request.DisplayName != "" {
		params.DisplayName = pgtype.Text{String: request.DisplayName, Valid: true}
	}
	
	if request.Username != "" {
		params.Username = pgtype.Text{String: request.Username, Valid: true}
	}
	
	if request.PrivacyLevel != nil {
		params.PrivacyLevel = dbModule.NullPrivacyLevel{PrivacyLevel: *request.PrivacyLevel, Valid: true}
	}
	
	updatedUser, err := user.UpdateUserProfile(r.Context(), tx, userID, params, config)
	if err != nil {
		// The middleware will rollback. We just need to set the error status.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// The middleware will commit. We just need to write the response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

func UpdateUserProfilePicture(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, config environment.Config) {
	// Get User from context (set by auth middleware).
	userFromCtx, ok := r.Context().Value(middleware.UserContextKey).(*dbModule.User)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}
	userID := userFromCtx.ID

	tx, ok := middleware.GetTxFromContext(r.Context())
	if !ok {
		http.Error(w, "Transaction not found in context", http.StatusInternalServerError)
		return
	}

	// Read file directly from request body
	file := r.Body
	defer file.Close()

	updatedUser, err := user.UpdateUserProfile(r.Context(), tx, userID, user.UpdateUserParams{
		File: file,
	}, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}
