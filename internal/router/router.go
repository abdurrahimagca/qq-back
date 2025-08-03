package router

import (
	"encoding/json"
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)


func Router(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	AuthRoute(mux, db)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		middleware.ApiAuth(config, w, r, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Hello, This is working World!"})
		})
	})
}