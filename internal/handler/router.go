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

	// box db lewat interface bernilai nil bila db nil; mengikat (*DB)(nil)
	// langsung ke interface backupRepo menghasilkan interface non-nil yang
	// menyembunyikan nil pointer dari guard db == nil di backupHandler.
	var backup backupRepo
	if db != nil {
		backup = db
	}
	mux.HandleFunc("GET /api/v1/backup", backupHandler(backup))

	// Product CRUD endpoints
	mux.HandleFunc("POST /api/v1/products", ProductHandler(db))
	mux.HandleFunc("GET /api/v1/products", ProductHandler(db))
	mux.HandleFunc("GET /api/v1/products/{id}", ProductHandler(db))
	mux.HandleFunc("PUT /api/v1/products/{id}", ProductHandler(db))
	mux.HandleFunc("DELETE /api/v1/products/{id}", ProductHandler(db))

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

// writeError writes a CONTRACT-compatible error envelope:
// {"error": {"code": "...", "message": "..."}}.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// isUUID reports whether s is a valid UUID (8-4-4-4-12 hex with dashes).
// Implemented with the standard library to avoid extra dependencies.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !isHexDigit(c) {
			return false
		}
	}
	return true
}

// isHexDigit reports whether c is a hexadecimal digit.
func isHexDigit(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// parseDate parses a YYYY-MM-DD date string.
func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
