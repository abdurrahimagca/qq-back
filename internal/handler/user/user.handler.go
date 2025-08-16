package user

import (
	"encoding/json"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/service/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpdateUserHandler now accepts the connection pool and config, and manages the transaction.
func UpdateUserHandler(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, config environment.Config) {
	// Get User ID from context (set by auth middleware). Adjust the key if necessary.
	userID, ok := r.Context().Value("userID").(pgtype.UUID)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("picture")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Could not get file from form", http.StatusBadRequest)
		return
	}
	if file != nil {
		defer file.Close()
	}

	jsonData := r.FormValue("data")
	var request user.UpdateUserParams // Use the service-level params struct
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &request); err != nil {
			http.Error(w, "Could not decode JSON data", http.StatusBadRequest)
			return
		}
	}

	// Begin a transaction
	tx, err := db.Begin(r.Context())
	if err != nil {
		http.Error(w, "Could not begin transaction", http.StatusInternalServerError)
		return
	}
	// Defer a rollback in case of panic or error
	defer tx.Rollback(r.Context())

	// Call the service with the transaction
	updatedUser, err := user.UpdateUserProfile(r.Context(), tx, userID, request, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}
