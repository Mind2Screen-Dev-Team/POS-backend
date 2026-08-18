package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

type backupRequest struct {
	UserID       string        `json:"user_id"`
	StartDate    string        `json:"start_date"`
	EndDate      string        `json:"end_date"`
	Transactions []transaction `json:"transactions"`
}

type transaction struct {
	ID            string `json:"id"`
	ProductID     string `json:"product_id"`
	CategoryID    string `json:"category_id"`
	PaymentMethod string `json:"payment_method"`
}

type backupResponse struct {
	Status      string `json:"status"`
	BackupID    string `json:"backup_id"`
	StoredCount int    `json:"stored_count"`
}

func BackupHandler(db *repository.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
			return
		}

		var req backupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}

		if !uuidRegex.MatchString(req.UserID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_user_id"})
			return
		}

		if !dateRegex.MatchString(req.StartDate) || !dateRegex.MatchString(req.EndDate) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_date_format"})
			return
		}

		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_date_format"})
			return
		}
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_date_format"})
			return
		}

		if end.Before(start) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "end_before_start"})
			return
		}

		if len(req.Transactions) > 500 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "schema_CONTRACT-301"})
			return
		}

		if db == nil || db.Pool == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_CONTRACT-301"})
			return
		}

		backupID := uuid.New().String()
		ctx := r.Context()
		tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_CONTRACT-301"})
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `
			INSERT INTO backups (id, user_id, start_date, end_date)
			VALUES ($1, $2, $3, $4)`,
			backupID, req.UserID, req.StartDate, req.EndDate)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_CONTRACT-301"})
			return
		}

		for _, t := range req.Transactions {
			data, _ := json.Marshal(t)
			_, err := tx.Exec(ctx, `
				INSERT INTO transactions (backup_id, data)
				VALUES ($1, $2)`,
				backupID, data)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_CONTRACT-301"})
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_CONTRACT-301"})
			return
		}

		writeJSON(w, http.StatusOK, backupResponse{
			Status:      "success",
			BackupID:    backupID,
			StoredCount: len(req.Transactions),
		})
	}
}
