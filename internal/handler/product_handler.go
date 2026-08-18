package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
	"github.com/google/uuid"
)

// Product represents a product entity.
type Product struct {
	ID         uuid.UUID `json:"id"`
	Nama       string    `json:"nama"`
	Deskripsi  string    `json:"deskripsi"`
	HargaBeli  int64     `json:"harga_beli"`
	HargaJual  int64     `json:"harga_jual"`
	Stok       int64     `json:"stok"`
	KategoriID uuid.UUID `json:"kategori_id"`
	FotoURL    string    `json:"foto_url"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

type createProductRequest struct {
	Nama       string  `json:"nama"`
	Deskripsi  string  `json:"deskripsi"`
	HargaBeli  float64 `json:"harga_beli"`
	HargaJual  float64 `json:"harga_jual"`
	Stok       float64 `json:"stok"`
	KategoriID string  `json:"kategori_id"`
	FotoURL    string  `json:"foto_url"`
}

type updateProductRequest struct {
	Nama       string  `json:"nama"`
	Deskripsi  string  `json:"deskripsi"`
	HargaBeli  float64 `json:"harga_beli"`
	HargaJual  float64 `json:"harga_jual"`
	Stok       float64 `json:"stok"`
	KategoriID string  `json:"kategori_id"`
	FotoURL    string  `json:"foto_url"`
}

// productRepo is the interface for product data operations.
type productRepo interface {
	CreateProduct(ctx context.Context, id, nama, deskripsi, fotoURL, kategoriID string, hargaBeli, hargaJual, stok int64) (string, string, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, nama, deskripsi, fotoURL, kategoriID string, hargaBeli, hargaJual, stok int64) (Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) (int64, error)
	ListProducts(ctx context.Context, limit, offset int) ([]Product, error)
	ProductExists(ctx context.Context, id uuid.UUID) (bool, error)
}

// CreateProductHandler returns a handler for POST /api/v1/products.
func CreateProductHandler(repo productRepo) http.HandlerFunc {
	return handleCreateProduct(repo)
}

// ListProductsHandler returns a handler for GET /api/v1/products (list).
func ListProductsHandler(repo productRepo) http.HandlerFunc {
	return handleGetProducts(repo)
}

// GetProductByIDHandler returns a handler for GET /api/v1/products/:id.
func GetProductByIDHandler(repo productRepo) http.HandlerFunc {
	return handleGetProductByID(repo)
}

// UpdateProductHandler returns a handler for PUT /api/v1/products/:id.
func UpdateProductHandler(repo productRepo) http.HandlerFunc {
	return handleUpdateProduct(repo)
}

// DeleteProductHandler returns a handler for DELETE /api/v1/products/:id.
func DeleteProductHandler(repo productRepo) http.HandlerFunc {
	return handleDeleteProduct(repo)
}

// WithProductRepo is the exported handler function for testing with mocks.
// It accepts a productRepo interface directly.
func WithProductRepo(repo productRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateProduct(repo)(w, r)
		case http.MethodGet:
			if r.PathValue("id") != "" {
				handleGetProductByID(repo)(w, r)
			} else {
				handleGetProducts(repo)(w, r)
			}
		case http.MethodPut:
			handleUpdateProduct(repo)(w, r)
		case http.MethodDelete:
			handleDeleteProduct(repo)(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		}
	}
}

// productDB implements productRepo using repository.DB.
type productDB struct {
	db *repository.DB
}

func (p *productDB) CreateProduct(ctx context.Context, id, nama, deskripsi, fotoURL, kategoriID string, hargaBeli, hargaJual, stok int64) (string, string, error) {
	pool := p.db.Pool
	var createdAt, updatedAt string
	err := pool.QueryRow(ctx,
		`INSERT INTO products (id, nama, deskripsi, harga_beli, harga_jual, stok, kategori_id, foto_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`,
		id, nama, deskripsi, hargaBeli, hargaJual, stok, kategoriID, fotoURL,
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		return "", "", err
	}
	return createdAt, updatedAt, nil
}

func (p *productDB) GetProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	pool := p.db.Pool
	var prod Product
	var createdAt, updatedAt sql.NullTime

	err := pool.QueryRow(ctx,
		`SELECT id, nama, deskripsi, harga_beli, harga_jual, stok, kategori_id, foto_url, created_at, updated_at
		FROM products WHERE id = $1`, id,
	).Scan(&prod.ID, &prod.Nama, &prod.Deskripsi, &prod.HargaBeli, &prod.HargaJual, &prod.Stok, &prod.KategoriID, &prod.FotoURL, &createdAt, &updatedAt)

	if err != nil {
		return Product{}, err
	}

	if createdAt.Valid {
		prod.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	if updatedAt.Valid {
		prod.UpdatedAt = updatedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
	}

	return prod, nil
}

func (p *productDB) UpdateProduct(ctx context.Context, id uuid.UUID, nama, deskripsi, fotoURL, kategoriID string, hargaBeli, hargaJual, stok int64) (Product, error) {
	pool := p.db.Pool
	result, err := pool.Exec(ctx,
		`UPDATE products SET nama=$2, deskripsi=$3, harga_beli=$4, harga_jual=$5, stok=$6, kategori_id=$7, foto_url=$8, updated_at=now()
		WHERE id=$1`,
		id, nama, deskripsi, hargaBeli, hargaJual, stok, kategoriID, fotoURL,
	)
	if err != nil {
		return Product{}, err
	}

	if result.RowsAffected() == 0 {
		return Product{}, sql.ErrNoRows
	}

	return Product{
		ID:         id,
		Nama:       nama,
		Deskripsi:  deskripsi,
		HargaBeli:  hargaBeli,
		HargaJual:  hargaJual,
		Stok:       stok,
		KategoriID: uuid.MustParse(kategoriID),
		FotoURL:    fotoURL,
	}, nil
}

func (p *productDB) DeleteProduct(ctx context.Context, id uuid.UUID) (int64, error) {
	pool := p.db.Pool
	result, err := pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (p *productDB) ListProducts(ctx context.Context, limit, offset int) ([]Product, error) {
	pool := p.db.Pool
	rows, err := pool.Query(ctx,
		`SELECT id, nama, deskripsi, harga_beli, harga_jual, stok, kategori_id, foto_url, created_at, updated_at
		FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var prod Product
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&prod.ID, &prod.Nama, &prod.Deskripsi, &prod.HargaBeli, &prod.HargaJual, &prod.Stok, &prod.KategoriID, &prod.FotoURL, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			prod.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		if updatedAt.Valid {
			prod.UpdatedAt = updatedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		products = append(products, prod)
	}
	return products, nil
}

func (p *productDB) ProductExists(ctx context.Context, id uuid.UUID) (bool, error) {
	pool := p.db.Pool
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

// extractProductID extracts the product ID from the URL path.
// Path format: /api/v1/products/{id} or /api/v1/products
func extractProductID(path string) string {
	// Remove /api/v1/products prefix
	const prefix = "/api/v1/products"
	if len(path) <= len(prefix) {
		return ""
	}
	id := path[len(prefix):]
	// Remove leading slash
	id = strings.TrimPrefix(id, "/")
	return id
}

func ProductHandler(db *repository.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil || db.Pool == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "service_unavailable"})
			return
		}

		repo := &productDB{db: db}

		switch r.Method {
		case http.MethodPost:
			handleCreateProduct(repo)(w, r)
		case http.MethodGet:
			if r.URL.Path != "/api/v1/products" {
				handleGetProductByID(repo)(w, r)
			} else {
				handleGetProducts(repo)(w, r)
			}
		case http.MethodPut:
			handleUpdateProduct(repo)(w, r)
		case http.MethodDelete:
			handleDeleteProduct(repo)(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		}
	}
}

func handleCreateProduct(repo productRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}

		// Validate required fields
		if req.Nama == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nama_wajib_diisi"})
			return
		}
		if req.KategoriID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kategori_id_wajib_diisi"})
			return
		}
		kategoriID, err := uuid.Parse(req.KategoriID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kategori_id_tidak_valid"})
			return
		}
		if req.HargaBeli < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_beli_tidak_valid"})
			return
		}
		if req.HargaBeli != float64(int64(req.HargaBeli)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_beli_bukan_integer"})
			return
		}
		if req.HargaJual < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_jual_tidak_valid"})
			return
		}
		if req.HargaJual != float64(int64(req.HargaJual)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_jual_bukan_integer"})
			return
		}
		if req.Stok < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stok_tidak_valid"})
			return
		}
		if req.Stok != float64(int64(req.Stok)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stok_bukan_integer"})
			return
		}

		productID := uuid.New()
		ctx := r.Context()

		createdAt, updatedAt, err := repo.CreateProduct(ctx,
			productID.String(), req.Nama, req.Deskripsi, req.FotoURL,
			kategoriID.String(), int64(req.HargaBeli), int64(req.HargaJual), int64(req.Stok))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_create_product"})
			return
		}

		writeJSON(w, http.StatusCreated, Product{
			ID:         productID,
			Nama:       req.Nama,
			Deskripsi:  req.Deskripsi,
			HargaBeli:  int64(req.HargaBeli),
			HargaJual:  int64(req.HargaJual),
			Stok:       int64(req.Stok),
			KategoriID: kategoriID,
			FotoURL:    req.FotoURL,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		})
	}
}

func handleGetProducts(repo productRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 100
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				if parsed > 1000 {
					parsed = 1000
				}
				limit = parsed
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		ctx := r.Context()
		products, err := repo.ListProducts(ctx, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_list_products"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"data":   products,
			"limit":  limit,
			"offset": offset,
		})
	}
}

func handleGetProductByID(repo productRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := extractProductID(r.URL.Path)
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id_tidak_valid"})
			return
		}

		ctx := r.Context()
		p, err := repo.GetProductByID(ctx, id)
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "product_not_found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_get_product"})
			return
		}

		writeJSON(w, http.StatusOK, p)
	}
}

func handleUpdateProduct(repo productRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := extractProductID(r.URL.Path)
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id_tidak_valid"})
			return
		}

		var req updateProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}

		// Validate required fields
		if req.Nama == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nama_wajib_diisi"})
			return
		}
		if req.KategoriID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kategori_id_wajib_diisi"})
			return
		}
		kategoriID, err := uuid.Parse(req.KategoriID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kategori_id_tidak_valid"})
			return
		}
		if req.HargaBeli < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_beli_tidak_valid"})
			return
		}
		if req.HargaBeli != float64(int64(req.HargaBeli)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_beli_bukan_integer"})
			return
		}
		if req.HargaJual < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_jual_tidak_valid"})
			return
		}
		if req.HargaJual != float64(int64(req.HargaJual)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harga_jual_bukan_integer"})
			return
		}
		if req.Stok < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stok_tidak_valid"})
			return
		}
		if req.Stok != float64(int64(req.Stok)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stok_bukan_integer"})
			return
		}

		ctx := r.Context()

		exists, err := repo.ProductExists(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_check_product"})
			return
		}
		if !exists {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product_not_found"})
			return
		}

		product, err := repo.UpdateProduct(ctx, id, req.Nama, req.Deskripsi, req.FotoURL, kategoriID.String(), int64(req.HargaBeli), int64(req.HargaJual), int64(req.Stok))
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "product_not_found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_update_product"})
			return
		}

		writeJSON(w, http.StatusOK, product)
	}
}

func handleDeleteProduct(repo productRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := extractProductID(r.URL.Path)
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id_tidak_valid"})
			return
		}

		ctx := r.Context()
		rowsAffected, err := repo.DeleteProduct(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db_error_delete_product"})
			return
		}

		if rowsAffected == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product_not_found"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
