package repository

import (
	"context"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies all embedded SQL migrations. Each file runs in its own
// transaction and is idempotent (CREATE TABLE IF NOT EXISTS).
//
// Tidak dipanggil otomatis dari main.go: cmd/server di luar path yang
// diizinkan contract BE-301. Handler BE-302/BE-303 memanggil Migrate saat
// startup sebelum memakai query backup.
func (d *DB) Migrate(ctx context.Context) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := "migrations/" + entry.Name()

		sql, err := migrationsFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := d.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}
