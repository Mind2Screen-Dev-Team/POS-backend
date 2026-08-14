package handler

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

// backupRepo is the data access surface the backup handler needs.
type backupRepo interface {
	ListTransactions(ctx context.Context, userID string, start, end time.Time) ([]repository.Transaction, error)
	Migrate(ctx context.Context) error
}

var (
	migrateOnce sync.Once
	migrateErr  error
)

// backupHandler handles GET /api/v1/backup — restore data transaksi.
// Idempoten: GET tidak menghapus data server, bisa diulang.
func backupHandler(db backupRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if DB is available
		if db == nil {
			writeError(w, http.StatusServiceUnavailable, "server_error", "Database connection not available")
			return
		}

		q := r.URL.Query()
		userID := q.Get("user_id")
		startDate := q.Get("start_date")
		endDate := q.Get("end_date")

		// Semua param wajib
		if userID == "" || startDate == "" || endDate == "" {
			writeError(w, http.StatusBadRequest, "invalid_param", "Param user_id, start_date, dan end_date wajib diisi")
			return
		}

		// user_id wajib UUID valid
		if !isUUID(userID) {
			writeError(w, http.StatusBadRequest, "invalid_param", "user_id harus berupa UUID")
			return
		}

		// start_date/end_date wajib YYYY-MM-DD; end_date >= start_date
		start, ok := parseDate(startDate)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_param", "Tanggal harus format YYYY-MM-DD")
			return
		}
		end, ok := parseDate(endDate)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_param", "Tanggal harus format YYYY-MM-DD")
			return
		}
		if end.Before(start) {
			writeError(w, http.StatusBadRequest, "invalid_param", "end_date tidak boleh sebelum start_date")
			return
		}

		// Rentang half-open [start, end+1hari) agar end_date tercakup penuh
		transactions, err := db.ListTransactions(r.Context(), userID, start, end.Add(24*time.Hour))
		if err != nil {
			if errors.Is(err, repository.ErrNoData) {
				writeError(w, http.StatusNotFound, "no_data", "Tidak ada data transaksi pada rentang tersebut")
				return
			}
			writeError(w, http.StatusInternalServerError, "server_error", "Terjadi kesalahan server")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"transactions": transactions})
	}
}
