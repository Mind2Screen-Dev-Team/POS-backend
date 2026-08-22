package repository

import (
	"context"
	"strings"
)

// ProductRepository handles product data operations
type ProductRepository struct {
	db *DB
}

// NewProductRepository creates a new ProductRepository with DB connection
func NewProductRepository(db *DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Product represents a product entity
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	HargaBeli   float64 `json:"harga_beli"`
	HargaJual   float64 `json:"harga_jual"`
	Stok        int     `json:"stok"`
	KategoriID  string  `json:"kategori_id"`
	FotoURL     string  `json:"foto_url"`
}

// SearchProducts searches products with LIKE query (case-insensitive)
// ORDER BY name A-Z
func (r *ProductRepository) SearchProducts(ctx context.Context, query string, categoryID string) ([]map[string]interface{}, error) {
	query = strings.TrimSpace(query)

	queryParam := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, harga_beli, harga_jual, stok, kategori_id, foto_url
		FROM products
		WHERE name ILIKE $1
		AND ($2 = '' OR kategori_id = $2)
		ORDER BY name ASC`,
		queryParam, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.HargaBeli, &p.HargaJual, &p.Stok, &p.KategoriID, &p.FotoURL); err != nil {
			return nil, err
		}
		products = append(products, map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"description":  p.Description,
			"harga_beli":   p.HargaBeli,
			"harga_jual":   p.HargaJual,
			"stok":         p.Stok,
			"kategori_id":  p.KategoriID,
			"foto_url":     p.FotoURL,
		})
	}
	return products, nil
}

// FilterProductsByCategory returns products filtered by category
// ORDER BY name A-Z
func (r *ProductRepository) FilterProductsByCategory(ctx context.Context, categoryID string) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, harga_beli, harga_jual, stok, kategori_id, foto_url
		FROM products
		WHERE kategori_id = $1
		ORDER BY name ASC`,
		categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.HargaBeli, &p.HargaJual, &p.Stok, &p.KategoriID, &p.FotoURL); err != nil {
			return nil, err
		}
		products = append(products, map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"description":  p.Description,
			"harga_beli":   p.HargaBeli,
			"harga_jual":   p.HargaJual,
			"stok":         p.Stok,
			"kategori_id":  p.KategoriID,
			"foto_url":     p.FotoURL,
		})
	}
	return products, nil
}