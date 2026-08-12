package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// connectTestDB returns a connected DB from TEST_DATABASE_URL, or nil when the
// env var is unset. The migration test is skipped in that case so the default
// `go test ./...` (no external service) stays green.
func connectTestDB(t *testing.T) *DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL unset — skipping repository integration test")
	}

	ctx := context.Background()
	db, err := New(ctx, dsn, 5)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(db.Close)

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestInsertAndListTransactionsByRange(t *testing.T) {
	db := connectTestDB(t)
	if db == nil {
		return
	}
	ctx := context.Background()

	userID := newTestUUID(t, "u")
	otherUser := newTestUUID(t, "o")

	if err := db.EnsureUser(ctx, userID); err != nil {
		t.Fatalf("ensure user: %v", err)
	}
	if err := db.EnsureUser(ctx, otherUser); err != nil {
		t.Fatalf("ensure other user: %v", err)
	}

	base := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	txns := []Transaction{
		{ID: newTestUUID(t, "t1"), UserID: userID, TransactedAt: base.Add(48 * time.Hour)},
		{ID: newTestUUID(t, "t2"), UserID: userID, TransactedAt: base.Add(72 * time.Hour), Payload: []byte(`{"total":25000}`)},
		{ID: newTestUUID(t, "t3"), UserID: otherUser, TransactedAt: base.Add(72 * time.Hour)},
	}

	n, err := db.InsertTransactions(ctx, userID, txns[:2])
	if err != nil {
		t.Fatalf("insert txns: %v", err)
	}
	if n != 2 {
		t.Fatalf("inserted %d, want 2", n)
	}
	if _, err := db.InsertTransactions(ctx, otherUser, txns[2:]); err != nil {
		t.Fatalf("insert other user txn: %v", err)
	}

	start := base.Add(24 * time.Hour)
	end := base.Add(96 * time.Hour)

	got, err := db.ListTransactions(ctx, userID, start, end)
	if err != nil {
		t.Fatalf("list txns: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d transactions, want 2", len(got))
	}

	// Urut ascending oleh transacted_at, tanpa bocor transaksi user lain.
	if got[0].ID != txns[0].ID || got[1].ID != txns[1].ID {
		t.Errorf("order or selection wrong: got %+v, want %s, %s", got, txns[0].ID, txns[1].ID)
	}
	if got[1].UserID != userID {
		t.Errorf("expected only transactions of %s, got user %s", userID, got[1].UserID)
	}
	if got[1].Payload == nil || !strings.Contains(string(got[1].Payload), "25000") {
		t.Errorf("payload not preserved: %s", got[1].Payload)
	}

	// Transaksi t2 berada tepat pada transacted_at = base+72h. End = base+72h
	// mengeluarkan t2 karena rentang half-open [start, end): t2 == end tidak
	// termasuk, t1 (base+48h) masih termasuk.
	got, err = db.ListTransactions(ctx, userID, start, base.Add(72*time.Hour))
	if err != nil {
		t.Fatalf("list txns after narrowing end: %v", err)
	}
	if len(got) != 1 || got[0].ID != txns[0].ID {
		t.Errorf("half-open end returned %d (%+v), want only %s", len(got), got, txns[0].ID)
	}
}

func TestListTransactionsNoData(t *testing.T) {
	db := connectTestDB(t)
	if db == nil {
		return
	}
	ctx := context.Background()

	userID := newTestUUID(t, "n")
	if err := db.EnsureUser(ctx, userID); err != nil {
		t.Fatalf("ensure user: %v", err)
	}

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
	if _, err := db.ListTransactions(ctx, userID, start, end); err != ErrNoData {
		t.Fatalf("expected ErrNoData, got %v", err)
	}
}

func TestBackupRecord(t *testing.T) {
	db := connectTestDB(t)
	if db == nil {
		return
	}
	ctx := context.Background()

	userID := newTestUUID(t, "b")
	if err := db.EnsureUser(ctx, userID); err != nil {
		t.Fatalf("ensure user: %v", err)
	}

	b, err := db.InsertBackup(ctx, newTestUUID(t, "bk"), userID, "2026-08-01", "2026-08-10")
	if err != nil {
		t.Fatalf("insert backup: %v", err)
	}
	if b.StartDate != "2026-08-01" || b.EndDate != "2026-08-10" {
		t.Fatalf("backup range wrong: %+v", b)
	}
	if b.CreatedAt.IsZero() {
		t.Fatal("backup created_at not set")
	}
}

// newTestUUID derives a deterministic UUIDv4-shaped string from the test name
// so isolated test databases never collide across runs.
func newTestUUID(t *testing.T, tag string) string {
	t.Helper()
	hexName := fmt.Sprintf("%x", tag+"-"+t.Name())
	if len(hexName) > 12 {
		hexName = hexName[:12]
	}
	return "00000000-0000-4000-8000-" + padTo12(hexName)
}

func padTo12(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s + strings.Repeat("0", 12-len(s))
}
