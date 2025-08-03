package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db     *pgxpool.Pool
	config *environment.Config
}

func NewHandler(db *pgxpool.Pool, config *environment.Config) *Handler {
	return &Handler{db: db, config: config}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]Check)
	overallStatus := "healthy"

	// Database connection check
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Config check
	configCheck := h.checkConfig()
	checks["config"] = configCheck
	if configCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0", // You can make this dynamic
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	
	if overallStatus == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) checkDatabase() Check {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test database connection
	if err := h.db.Ping(ctx); err != nil {
		return Check{
			Status:  "unhealthy",
			Message: "Database connection failed: " + err.Error(),
			Latency: time.Since(start).String(),
		}
	}
	
	// Test simple query
	var result int
	err := h.db.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return Check{
			Status:  "unhealthy",
			Message: "Database query failed: " + err.Error(),
			Latency: time.Since(start).String(),
		}
	}
	
	return Check{
		Status:  "healthy",
		Message: "Database connection successful",
		Latency: time.Since(start).String(),
	}
}

func (h *Handler) checkConfig() Check {
	start := time.Now()
	
	// Check if essential config values are present
	if h.config.DatabaseURL == "" {
		return Check{
			Status:  "unhealthy",
			Message: "DATABASE_URL not configured",
			Latency: time.Since(start).String(),
		}
	}
	
	// Add more config checks as needed
	// if h.config.JWTSecret == "" {
	//     return Check{
	//         Status:  "unhealthy",
	//         Message: "JWT_SECRET not configured",
	//         Latency: time.Since(start).String(),
	//     }
	// }
	
	return Check{
		Status:  "healthy",
		Message: "Configuration is valid",
		Latency: time.Since(start).String(),
	}
}