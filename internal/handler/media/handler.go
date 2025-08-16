package media

import (
	"encoding/json"
	"log"
	"net/http"

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

	image, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Invalid image", http.StatusBadRequest)
		return
	}
	defer image.Close()

	target := r.URL.Query().Get("target")
	bucketService, err := bucket.NewService(h.config.R2)
	if err != nil {
		log.Printf("Failed to create bucket service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	imageTypeTarget := bucket.ImageTypeTarget(target)

	if imageTypeTarget == "" {
		http.Error(w, "Invalid target", http.StatusBadRequest)
		return
	}

	uploadResult, err := bucketService.UploadImage(image, imageTypeTarget, false)

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
