-- 001_init.sql — skema awal backup transaksi.
-- users: anonymous device UUID sebagai identitas user server.
-- backups: rekaman backup per rentang tanggal (ISO 8601).
-- transactions: payload JSONB lengkap + FK ke users dan backups.
-- Semua FK ON DELETE CASCADE; index (user_id, transacted_at) untuk
-- query backup/restore per user per rentang.

CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS backups (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS transactions (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    transacted_at TIMESTAMPTZ NOT NULL,
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    backup_id     UUID REFERENCES backups (id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Query backup/restore per rentang: WHERE user_id = $1 AND transacted_at BETWEEN $2 AND $3.
CREATE INDEX IF NOT EXISTS idx_transactions_user_transacted
    ON transactions (user_id, transacted_at);
