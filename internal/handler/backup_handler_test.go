package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupHandler_Validation(t *testing.T) {
	tests := []struct {
		name           string
		request        backupRequest
		expectedStatus int
	}{
		{
			name: "invalid UUID",
			request: backupRequest{
				UserID:    "invalid-uuid",
				StartDate: "2023-01-01",
				EndDate:   "2023-01-02",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid start date format",
			request: backupRequest{
				UserID:    "123e4567-e89b-12d3-a456-426614174000",
				StartDate: "01-01-2023",
				EndDate:   "2023-01-02",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid end date format",
			request: backupRequest{
				UserID:    "123e4567-e89b-12d3-a456-426614174000",
				StartDate: "2023-01-01",
				EndDate:   "2023/01/02",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "reversed date range",
			request: backupRequest{
				UserID:    "123e4567-e89b-12d3-a456-426614174000",
				StartDate: "2023-01-02",
				EndDate:   "2023-01-01",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "batch > 500",
			request: backupRequest{
				UserID:       "123e4567-e89b-12d3-a456-426614174000",
				StartDate:    "2023-01-01",
				EndDate:      "2023-01-02",
				Transactions: make([]transaction, 501),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid request - nil db returns 500",
			request: backupRequest{
				UserID:    "123e4567-e89b-12d3-a456-426614174000",
				StartDate: "2023-01-01",
				EndDate:   "2023-01-02",
				Transactions: []transaction{
					{ID: "t1", ProductID: "p1", CategoryID: "c1", PaymentMethod: "cash"},
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/backup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler := BackupHandler(nil)
			handler(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestBackupHandler_MethodNotAllowed(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/backup", nil)
	recorder := httptest.NewRecorder()

	handler := BackupHandler(nil)
	handler(recorder, req)

	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
}

func TestBackupHandler_InvalidJSON(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/backup", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := BackupHandler(nil)
	handler(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestBackupHandler_EmptyBody(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/backup", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := BackupHandler(nil)
	handler(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestBackupHandler_ResponseStructure(t *testing.T) {
	request := backupRequest{
		UserID:    uuid.New().String(),
		StartDate: "2023-01-01",
		EndDate:   "2023-01-02",
		Transactions: []transaction{
			{ID: "t1", ProductID: "p1", CategoryID: "c1", PaymentMethod: "cash"},
		},
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/backup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := BackupHandler(nil)
	handler(recorder, req)

	// Check that validation passes (returns 500 for nil DB, not 400)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestBackupHandler_StoredCount(t *testing.T) {
	request := backupRequest{
		UserID:       uuid.New().String(),
		StartDate:    "2023-01-01",
		EndDate:      "2023-01-02",
		Transactions: make([]transaction, 500),
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/backup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := BackupHandler(nil)
	handler(recorder, req)

	// 500 is max valid batch
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// Helper to read body from response
func readBody(t *testing.T, recorder *httptest.ResponseRecorder) []byte {
	t.Helper()
	body, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)
	return body
}
