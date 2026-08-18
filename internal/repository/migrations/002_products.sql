-- 002_products.sql — tabel produk untuk katalog barang.
-- id: UUID primary key.
-- nama: nama produk (wajib).
-- deskripsi: deskripsi produk (opsional).
-- harga_beli: harga beli ke supplier (wajib, >= 0).
-- harga_jual: harga jual ke customer (wajib, >= 0).
-- stok: jumlah stok tersedia (wajib, >= 0).
-- kategori_id: UUID kategori (wajib).
-- foto_url: URL foto produk (opsional).

CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY,
    nama        TEXT NOT NULL,
    deskripsi   TEXT,
    harga_beli  BIGINT NOT NULL CHECK (harga_beli >= 0),
    harga_jual  BIGINT NOT NULL CHECK (harga_jual >= 0),
    stok        BIGINT NOT NULL CHECK (stok >= 0),
    kategori_id UUID NOT NULL,
    foto_url    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_kategori_id ON products (kategori_id);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);