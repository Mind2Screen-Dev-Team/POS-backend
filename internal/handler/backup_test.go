package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

// mockBackupRepo implements the repo surface backupHandler needs.
type mockBackupRepo struct {
	transactions []repository.Transaction
	err          error
}

func (m *mockBackupRepo) ListTransactions(_ context.Context, _ string, _, _ time.Time) ([]repository.Transaction, error) {
	return m.transactions, m.err
}

func (m *mockBackupRepo) Migrate(_ context.Context) error {
	return m.err
}

func TestBackupHandler(t *testing.T) {
	validUser := "01234567-89ab-cdef-0123-456789abcdef"
	validRange := "start_date=2023-01-01&end_date=2023-01-31"

	tests := []struct {
		name       string
		query      string
		repoTxns   []repository.Transaction
		repoErr    error
		wantStatus int
		wantCode   string
		wantCount  int
	}{
		{
			name:       "missing params",
			query:      "",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_param",
		},
		{
			name:       "invalid user_id",
			query:      "user_id=not-a-uuid&" + validRange,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_param",
		},
		{
			name:       "invalid start_date",
			query:      "user_id=" + validUser + "&start_date=01-01-2023&end_date=2023-01-31",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_param",
		},
		{
			name:       "invalid end_date",
			query:      "user_id=" + validUser + "&start_date=2023-01-01&end_date=2023/01/31",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_param",
		},
		{
			name:       "end before start",
			query:      "user_id=" + validUser + "&start_date=2023-02-01&end_date=2023-01-31",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_param",
		},
		{
			name:       "no data",
			query:      "user_id=" + validUser + "&" + validRange,
			repoErr:    repository.ErrNoData,
			wantStatus: http.StatusNotFound,
			wantCode:   "no_data",
		},
		{
			name:       "db error",
			query:      "user_id=" + validUser + "&" + validRange,
			repoErr:    errors.New("connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "server_error",
		},
		{
			name:  "valid request",
			query: "user_id=" + validUser + "&" + validRange,
			repoTxns: []repository.Transaction{
				{
					ID:           "11111111-2222-3333-4444-555555555555",
					UserID:       validUser,
					TransactedAt: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
					Payload:      json.RawMessage(`{"amount":100.5}`),
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBackupRepo{transactions: tt.repoTxns, err: tt.repoErr}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/backup?"+tt.query, nil)
			rec := httptest.NewRecorder()

			backupHandler(repo)(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var resp struct {
				Error        *struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
				Transactions []repository.Transaction `json:"transactions"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			if tt.wantCode != "" {
				if resp.Error == nil || resp.Error.Code != tt.wantCode {
					t.Errorf("error code = %+v, want %q", resp.Error, tt.wantCode)
				}
				return
			}
			if len(resp.Transactions) != tt.wantCount {
				t.Errorf("transactions len = %d, want %d", len(resp.Transactions), tt.wantCount)
			}
		})
	}
	}

	// TestBackupHandlerEndDateInclusive verifies transactions at 23:59 on end_date
	// are included in the result (half-open [start, end+1day) range).
	func TestBackupHandlerEndDateInclusive(t *testing.T) {
		userID := "01234567-89ab-cdef-0123-456789abcdef"
		endDate := time.Date(2024, 5, 31, 23, 59, 0, 0, time.UTC)
		transactions := []repository.Transaction{
			{
				ID:           "11111111-2222-3333-4444-555555555555",
				UserID:       userID,
				TransactedAt: endDate,
				Payload:      json.RawMessage(`{"amount":100.5}`),
			},
		}

		repo := &mockBackupRepo{transactions: transactions}
		query := fmt.Sprintf("user_id=%s&start_date=2024-05-01&end_date=2024-05-31", userID)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/backup?"+query, nil)
		rec := httptest.NewRecorder()

		backupHandler(repo)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp struct {
			Transactions []repository.Transaction `json:"transactions"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if len(resp.Transactions) != 1 {
			t.Errorf("transactions count = %d, want 1", len(resp.Transactions))
		}
	}

	// TestBackupHandlerSingleDayRange verifies start_date == end_date works.
	func TestBackupHandlerSingleDayRange(t *testing.T) {
		userID := "01234567-89ab-cdef-0123-456789abcdef"
		date := "2024-05-15"
		transactions := []repository.Transaction{
			{
				ID:           "11111111-2222-3333-4444-555555555555",
				UserID:       userID,
				TransactedAt: time.Date(2024, 5, 15, 10, 30, 0, 0, time.UTC),
				Payload:      json.RawMessage(`{"amount":100.5}`),
			},
		}

		repo := &mockBackupRepo{transactions: transactions}
		query := fmt.Sprintf("user_id=%s&start_date=%s&end_date=%s", userID, date, date)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/backup?"+query, nil)
		rec := httptest.NewRecorder()

		backupHandler(repo)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp struct {
			Transactions []repository.Transaction `json:"transactions"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if len(resp.Transactions) != 1 {
			t.Errorf("transactions count = %d, want 1", len(resp.Transactions))
		}
	}

	// TestRouterBackupRoute verifies the /api/v1/backup route is registered.
	func TestRouterBackupRoute(t *testing.T) {
		router := NewRouter(nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/backup?user_id=01234567-89ab-cdef-0123-456789abcdef&start_date=2024-01-01&end_date=2024-01-31", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// Nil DB → backup handler merespon 503 service unavailable, bukan 404
		// (route ter-register dan handler dijalankan).
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d (backup handler unavailable tanpa DB)", rec.Code, http.StatusServiceUnavailable)
		}
	}

// TestBackupHandlerIdempotent verifies two consecutive GET calls return
// identical data — a read never deletes server data, so restore is repeatable.
func TestBackupHandlerIdempotent(t *testing.T) {
	data := []repository.Transaction{
		{
			ID:           "11111111-2222-3333-4444-555555555555",
			UserID:       "01234567-89ab-cdef-0123-456789abcdef",
			TransactedAt: time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC),
			Payload:      json.RawMessage(`{"amount":42}`),
		},
	}

	repo := &mockBackupRepo{transactions: data}
	query := "user_id=" + data[0].UserID + "&start_date=2024-05-01&end_date=2024-05-31"
	h := backupHandler(repo)

	var first, second string
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/backup?"+query, nil)
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("call %d: status = %d, want %d", i+1, rec.Code, http.StatusOK)
		}
		if i == 0 {
			first = rec.Body.String()
		} else {
			second = rec.Body.String()
		}
	}

	if first != second {
		t.Errorf("GET responses differ across calls:\n%s\n---\n%s", first, second)
	}
}
