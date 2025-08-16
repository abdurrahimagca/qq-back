package media

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/abdurrahimagca/qq-back/internal/external/bucket"
)

type Handler struct {
	config *environment.Config
}

func NewHandler(config *environment.Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// TODO: Add authentication back later
	// _, ok := r.Context().Value(middleware.UserContextKey).(*db.User)
	// if !ok {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	// Read image directly from request body
	image := r.Body
	defer image.Close()

	bucketService, err := bucket.NewService(h.config.R2)
	if err != nil {
		log.Printf("Failed to create bucket service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	uploadResult, err := bucketService.UploadImage(image, false)

	if err != nil {
		log.Printf("Failed to upload image: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(uploadResult)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)

}

func (h *Handler) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	bucketService, err := bucket.NewService(h.config.R2)
	if err != nil {
		log.Printf("Failed to create bucket service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	presignedURL, err := bucketService.GetPresignedURL(key, 40*time.Second)
	if err != nil {
		log.Printf("Failed to get presigned URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(presignedURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)

}
