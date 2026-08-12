package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	// ErrNoData is returned when a range query matches no transactions.
	ErrNoData = errors.New("no transactions in range")
)

// Transaction is one POS transaction as stored and returned by the backup API.
type Transaction struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	TransactedAt time.Time       `json:"transacted_at"`
	Payload      json.RawMessage `json:"payload"`
}

// User holds a backup user record.
type User struct {
	ID        string
	CreatedAt time.Time
}

// Backup holds one backup run over a date range.
type Backup struct {
	ID        string
	UserID    string
	StartDate string
	EndDate   string
	CreatedAt time.Time
}

const ensureUserSQL = `
	INSERT INTO users (id)
	VALUES ($1)
	ON CONFLICT (id) DO NOTHING`

// EnsureUser inserts the user row if absent (anonymous device UUID). No-op
// when the user already exists.
func (d *DB) EnsureUser(ctx context.Context, userID string) error {
	_, err := d.Pool.Exec(ctx, ensureUserSQL, userID)
	if err != nil {
		return fmt.Errorf("ensure user: %w", err)
	}
	return nil
}

const insertBackupSQL = `
	INSERT INTO backups (id, user_id, start_date, end_date)
	VALUES ($1, $2, $3, $4)
	RETURNING id, user_id, start_date::text, end_date::text, created_at`

// InsertBackup stores a backup record for a date range and returns it.
func (d *DB) InsertBackup(ctx context.Context, id, userID, startDate, endDate string) (*Backup, error) {
	var b Backup
	err := d.Pool.QueryRow(ctx, insertBackupSQL, id, userID, startDate, endDate).
		Scan(&b.ID, &b.UserID, &b.StartDate, &b.EndDate, &b.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert backup: %w", err)
	}
	return &b, nil
}

const insertTransactionSQL = `
	INSERT INTO transactions (id, user_id, transacted_at, payload, backup_id)
	VALUES ($1, $2, $3, $4, $5)`

// InsertTransactions stores a batch of transactions and returns the number
// inserted. payload may be nil, which is stored as the empty JSONB object.
func (d *DB) InsertTransactions(ctx context.Context, userID string, txns []Transaction) (int64, error) {
	payloads := make([][]byte, len(txns))
	for i, txn := range txns {
		if len(txn.Payload) == 0 {
			payloads[i] = []byte("{}")
		} else {
			payloads[i] = txn.Payload
		}
	}

	batch := &pgx.Batch{}
	for i, txn := range txns {
		batch.Queue(insertTransactionSQL, txn.ID, userID, txn.TransactedAt, payloads[i], nil)
	}

	results := d.Pool.SendBatch(ctx, batch)
	defer results.Close()

	var inserted int64
	for range txns {
		if _, err := results.Exec(); err != nil {
			return inserted, fmt.Errorf("insert transaction: %w", err)
		}
		inserted++
	}
	return inserted, nil
}

const listTransactionsSQL = `
	SELECT id, user_id, transacted_at, payload
	FROM transactions
	WHERE user_id = $1 AND transacted_at >= $2 AND transacted_at < $3
	ORDER BY transacted_at ASC`

// ListTransactions returns all transactions for a user within the
// half-open range [start, end). Returns ErrNoData when the range is empty.
func (d *DB) ListTransactions(ctx context.Context, userID string, start, end time.Time) ([]Transaction, error) {
	rows, err := d.Pool.Query(ctx, listTransactionsSQL, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txns []Transaction
	for rows.Next() {
		var txn Transaction
		if err := rows.Scan(&txn.ID, &txn.UserID, &txn.TransactedAt, &txn.Payload); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, txn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}

	if len(txns) == 0 {
		return nil, ErrNoData
	}
	return txns, nil
}
