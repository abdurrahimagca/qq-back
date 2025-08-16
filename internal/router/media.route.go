package router

import (
	"net/http"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/handler/media"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MediaRoute(mux *http.ServeMux, db *pgxpool.Pool, config *environment.Config) {
	handler := media.NewHandler(config)

	mux.HandleFunc("POST /media/upload", handler.UploadImage)
	mux.HandleFunc("GET /media/presigned-url", handler.GetPresignedURL)
}

