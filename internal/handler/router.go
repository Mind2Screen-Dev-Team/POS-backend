package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

// healthResponse is the JSON payload for the /health endpoint.
type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Time     string `json:"time"`
}

// NewRouter creates the HTTP handler with all routes registered.
// db may be nil if running without a database connection.
func NewRouter(db *repository.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler(db))

	return mux
}

// healthHandler returns a handler that pings the database and reports status.
func healthHandler(db *repository.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			Status: "ok",
			Time:   time.Now().UTC().Format(time.RFC3339),
		}

		if db != nil {
			if err := db.Ping(r.Context()); err != nil {
				resp.Status = "degraded"
				resp.Database = "unhealthy: " + err.Error()
				writeJSON(w, http.StatusServiceUnavailable, resp)
				return
			}
			resp.Database = "healthy"
		} else {
			resp.Database = "not configured"
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck // writing to ResponseWriter
}
