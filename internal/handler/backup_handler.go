package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
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
		var req backupRequest
		query := r.URL.Query()
		id := query.Get("id")
		startDate := query.Get("start")
		endDate := query.Get("end")
		batchStr := query.Get("batch")

		if id == "" || startDate == "" || endDate == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_parameters"})
			return
		}

		if !uuidRegex.MatchString(id) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_user_id"})
			return
		}

		if !dateRegex.MatchString(startDate) || !dateRegex.MatchString(endDate) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_date_format"})
			return
		}

		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_date_format"})
			return
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_date_format"})
			return
		}

		if end.Before(start) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "end_before_start"})
			return
		}

		batch := 500
		if batchStr != "" {
			batch, err = strconv.Atoi(batchStr)
			if err != nil || batch > 500 || batch < 1 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_batch"})
				return
			}
		}

		req = backupRequest{
			UserID:    id,
			StartDate: startDate,
			EndDate:   endDate,
			Transactions: make([]transaction, batch),
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
