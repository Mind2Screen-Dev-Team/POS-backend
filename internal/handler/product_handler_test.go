package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/handler"
	"github.com/google/uuid"
)

// pathValueKey is the context key for URL path values (Go 1.22+ mux pattern)
type pathValueKey struct{}

// setPathValue sets a path value in the request context for testing
func setPathValue(req *http.Request, name, value string) *http.Request {
	values := map[string]string{name: value}
	ctx := context.WithValue(req.Context(), pathValueKey{}, values)
	return req.WithContext(ctx)
}

// pathValueFromURL extracts path value from URL for testing
func pathValueFromURL(path, name string) string {
	// For /api/v1/products/:id, extract the id
	var id string
	fmt.Sscanf(path, "/api/v1/products/%s", &id)
	// Handle trailing slash
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '/' {
			id = id[:i]
			break
		}
	}
	return id
}

// mockProductRepo implements the productRepo interface for testing.
type mockProductRepo struct {
	products map[uuid.UUID]handler.Product
	err      error
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		products: make(map[uuid.UUID]handler.Product),
	}
}

func (m *mockProductRepo) CreateProduct(ctx context.Context, id, nama, deskripsi, fotoURL, kategoriID string, hargaBeli, hargaJual, stok int64) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	kID := uuid.MustParse(kategoriID)
	p := handler.Product{
		ID:         uuid.MustParse(id),
		Nama:       nama,
		Deskripsi:  deskripsi,
		HargaBeli:  hargaBeli,
		HargaJual:  hargaJual,
		Stok:       stok,
		KategoriID: kID,
		FotoURL:    fotoURL,
		CreatedAt:  "2024-01-01T00:00:00Z",
		UpdatedAt:  "2024-01-01T00:00:00Z",
	}
	m.products[uuid.MustParse(id)] = p
	return p.CreatedAt, p.UpdatedAt, nil
}

func (m *mockProductRepo) GetProductByID(ctx context.Context, id uuid.UUID) (handler.Product, error) {
	if m.err != nil {
		return handler.Product{}, m.err
	}
	p, ok := m.products[id]
	if !ok {
		return handler.Product{}, sql.ErrNoRows
	}
	return p, nil
}

func (m *mockProductRepo) UpdateProduct(ctx context.Context, id uuid.UUID, nama, deskripsi, fotoURL, kategoriID string, hargaBeli, hargaJual, stok int64) (handler.Product, error) {
	if m.err != nil {
		return handler.Product{}, m.err
	}
	kID := uuid.MustParse(kategoriID)
	p := handler.Product{
		ID:         id,
		Nama:       nama,
		Deskripsi:  deskripsi,
		HargaBeli:  hargaBeli,
		HargaJual:  hargaJual,
		Stok:       stok,
		KategoriID: kID,
		FotoURL:    fotoURL,
	}
	m.products[id] = p
	return p, nil
}

func (m *mockProductRepo) DeleteProduct(ctx context.Context, id uuid.UUID) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if _, exists := m.products[id]; !exists {
		return 0, nil
	}
	delete(m.products, id)
	return 1, nil
}

func (m *mockProductRepo) ListProducts(ctx context.Context, limit, offset int) ([]handler.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	var list []handler.Product
	for _, p := range m.products {
		list = append(list, p)
	}
	return list, nil
}

func (m *mockProductRepo) ProductExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	_, ok := m.products[id]
	return ok, nil
}

// Helper to create request with proper path value set
func newRequestWithPath(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	return req
}

// Tests with mock repo for validation logic

func TestCreateProduct_Success(t *testing.T) {
	repo := newMockProductRepo()

	kategoriID := uuid.New()
	reqBody := map[string]any{
		"nama":        "Produk Test",
		"deskripsi":   "Deskripsi produk",
		"harga_beli":  10000.0,
		"harga_jual":  15000.0,
		"stok":        50.0,
		"kategori_id": kategoriID.String(),
		"foto_url":    "https://example.com/foto.jpg",
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp handler.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Nama != "Produk Test" {
		t.Errorf("expected nama 'Produk Test', got '%s'", resp.Nama)
	}
	if resp.HargaBeli != 10000 {
		t.Errorf("expected harga_beli 10000, got %d", resp.HargaBeli)
	}
	if resp.HargaJual != 15000 {
		t.Errorf("expected harga_jual 15000, got %d", resp.HargaJual)
	}
	if resp.Stok != 50 {
		t.Errorf("expected stok 50, got %d", resp.Stok)
	}
}

func TestCreateProduct_MissingNama(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"deskripsi":   "Deskripsi produk",
		"harga_beli":  10000.0,
		"harga_jual":  15000.0,
		"stok":        50.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "nama_wajib_diisi" {
		t.Errorf("expected error 'nama_wajib_diisi', got '%s'", errResp["error"])
	}
}

func TestCreateProduct_MissingKategoriID(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama":       "Produk Test",
		"harga_beli": 10000.0,
		"harga_jual": 15000.0,
		"stok":       50.0,
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "kategori_id_wajib_diisi" {
		t.Errorf("expected error 'kategori_id_wajib_diisi', got '%s'", errResp["error"])
	}
}

func TestCreateProduct_InvalidKategoriID(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama":        "Produk Test",
		"harga_beli":  10000.0,
		"harga_jual":  15000.0,
		"stok":        50.0,
		"kategori_id": "invalid-uuid",
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "kategori_id_tidak_valid" {
		t.Errorf("expected error 'kategori_id_tidak_valid', got '%s'", errResp["error"])
	}
}

func TestCreateProduct_NegativeHargaBeli(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama":        "Produk Test",
		"harga_beli":  -1000.0,
		"harga_jual":  15000.0,
		"stok":        50.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "harga_beli_tidak_valid" {
		t.Errorf("expected error 'harga_beli_tidak_valid', got '%s'", errResp["error"])
	}
}

func TestCreateProduct_NegativeHargaJual(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama":        "Produk Test",
		"harga_beli":  10000.0,
		"harga_jual":  -5000.0,
		"stok":        50.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "harga_jual_tidak_valid" {
		t.Errorf("expected error 'harga_jual_tidak_valid', got '%s'", errResp["error"])
	}
}

func TestCreateProduct_NegativeStok(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama":        "Produk Test",
		"harga_beli":  10000.0,
		"harga_jual":  15000.0,
		"stok":        -10.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "stok_tidak_valid" {
		t.Errorf("expected error 'stok_tidak_valid', got '%s'", errResp["error"])
	}
}

func TestCreateProduct_InvalidJSON(t *testing.T) {
	repo := newMockProductRepo()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader([]byte("invalid json")))
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "invalid_json" {
		t.Errorf("expected error 'invalid_json', got '%s'", errResp["error"])
	}
}

func TestGetProducts_Success(t *testing.T) {
	repo := newMockProductRepo()
	// Add a product first
	kategoriID := uuid.New()
	repo.CreateProduct(context.Background(), uuid.New().String(), "Test Product", "", "", kategoriID.String(), 10000, 15000, 50)

	req := newRequestWithPath(http.MethodGet, "/api/v1/products?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()

	h := handler.ListProductsHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if data, ok := resp["data"].([]any); !ok {
		t.Error("expected 'data' to be an array")
	} else if len(data) < 1 {
		t.Error("expected at least 1 product")
	}
}

func TestGetProducts_WithPagination(t *testing.T) {
	repo := newMockProductRepo()

	req := newRequestWithPath(http.MethodGet, "/api/v1/products?limit=50&offset=100", nil)
	rec := httptest.NewRecorder()

	h := handler.ListProductsHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if limit, ok := resp["limit"].(float64); !ok || limit != 50 {
		t.Errorf("expected limit 50, got %v", resp["limit"])
	}
	if offset, ok := resp["offset"].(float64); !ok || offset != 100 {
		t.Errorf("expected offset 100, got %v", resp["offset"])
	}
}

func TestGetProductByID_Success(t *testing.T) {
	repo := newMockProductRepo()

	// Create a product first
	kategoriID := uuid.New()
	prodID := uuid.New()
	repo.CreateProduct(context.Background(), prodID.String(), "Test Product", "Description", "", kategoriID.String(), 10000, 15000, 50)

	path := "/api/v1/products/" + prodID.String()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	h := handler.GetProductByIDHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var fetched handler.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if fetched.ID != prodID {
		t.Errorf("expected ID %s, got %s", prodID, fetched.ID)
	}
	if fetched.Nama != "Test Product" {
		t.Errorf("expected nama 'Test Product', got '%s'", fetched.Nama)
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	repo := newMockProductRepo()

	nonExistentID := uuid.New()
	path := "/api/v1/products/" + nonExistentID.String()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	h := handler.GetProductByIDHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "product_not_found" {
		t.Errorf("expected error 'product_not_found', got '%s'", errResp["error"])
	}
}

func TestGetProductByID_InvalidID(t *testing.T) {
	repo := newMockProductRepo()

	path := "/api/v1/products/invalid-uuid"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	h := handler.GetProductByIDHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "id_tidak_valid" {
		t.Errorf("expected error 'id_tidak_valid', got '%s'", errResp["error"])
	}
}

func TestUpdateProduct_Success(t *testing.T) {
	repo := newMockProductRepo()

	// Create a product first
	kategoriID := uuid.New()
	prodID := uuid.New()
	repo.CreateProduct(context.Background(), prodID.String(), "Original Product", "", "", kategoriID.String(), 10000, 15000, 50)

	updateReqBody := map[string]any{
		"nama":        "Updated Product",
		"deskripsi":   "Updated description",
		"harga_beli":  12000.0,
		"harga_jual":  18000.0,
		"stok":        100.0,
		"kategori_id": uuid.New().String(),
	}
	updateBody, _ := json.Marshal(updateReqBody)

	path := "/api/v1/products/" + prodID.String()
	updateReq := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(updateBody))
	updateRec := httptest.NewRecorder()

	h := handler.UpdateProductHandler(repo)
	h(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, updateRec.Code, updateRec.Body.String())
	}

	var updated handler.Product
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if updated.Nama != "Updated Product" {
		t.Errorf("expected nama 'Updated Product', got '%s'", updated.Nama)
	}
	if updated.HargaBeli != 12000 {
		t.Errorf("expected harga_beli 12000, got %d", updated.HargaBeli)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo := newMockProductRepo()

	nonExistentID := uuid.New()
	reqBody := map[string]any{
		"nama":        "Updated Product",
		"deskripsi":   "Updated description",
		"harga_beli":  12000.0,
		"harga_jual":  18000.0,
		"stok":        100.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	path := "/api/v1/products/" + nonExistentID.String()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h := handler.UpdateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateProduct_MissingNama(t *testing.T) {
	repo := newMockProductRepo()

	// Create a product first
	kategoriID := uuid.New()
	prodID := uuid.New()
	repo.CreateProduct(context.Background(), prodID.String(), "Test Product", "", "", kategoriID.String(), 10000, 15000, 50)

	reqBody := map[string]any{
		"deskripsi":   "Updated description",
		"harga_beli":  12000.0,
		"harga_jual":  18000.0,
		"stok":        100.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	path := "/api/v1/products/" + prodID.String()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h := handler.UpdateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "nama_wajib_diisi" {
		t.Errorf("expected error 'nama_wajib_diisi', got '%s'", errResp["error"])
	}
}

func TestDeleteProduct_Success(t *testing.T) {
	repo := newMockProductRepo()

	// Create a product first
	kategoriID := uuid.New()
	prodID := uuid.New()
	repo.CreateProduct(context.Background(), prodID.String(), "Test Product", "", "", kategoriID.String(), 10000, 15000, 50)

	path := "/api/v1/products/" + prodID.String()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	rec := httptest.NewRecorder()

	h := handler.DeleteProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body.String())
	}

	// Verify it's deleted
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+prodID.String(), nil)
	getRec := httptest.NewRecorder()

	getH := handler.GetProductByIDHandler(repo)
	getH(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected status %d after delete, got %d", http.StatusNotFound, getRec.Code)
	}
}

func TestDeleteProduct_NotFound(t *testing.T) {
	repo := newMockProductRepo()

	nonExistentID := uuid.New()
	path := "/api/v1/products/" + nonExistentID.String()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	rec := httptest.NewRecorder()

	h := handler.DeleteProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDeleteProduct_InvalidID(t *testing.T) {
	repo := newMockProductRepo()

	path := "/api/v1/products/invalid-uuid"
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	rec := httptest.NewRecorder()

	h := handler.DeleteProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp["error"] != "id_tidak_valid" {
		t.Errorf("expected error 'id_tidak_valid', got '%s'", errResp["error"])
	}
}

// Router tests for nil DB case

func TestHealthEndpoint(t *testing.T) {
	router := handler.NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
	if resp.Database != "not configured" {
		t.Errorf("expected database 'not configured', got '%s'", resp.Database)
	}
}

func TestRouter_ProductEndpoints_NoDB(t *testing.T) {
	router := handler.NewRouter(nil)

	// Test health endpoint (always works)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("health endpoint: expected %d, got %d", http.StatusOK, rec.Code)
	}

	// Test product routes (expect 503 without DB)
	kategoriID := uuid.New()
	products := []struct {
		method     string
		path       string
		body       string
		expectCode int
	}{
		{http.MethodPost, "/api/v1/products", `{"nama":"Test","harga_beli":10000,"harga_jual":15000,"stok":50,"kategori_id":"` + kategoriID.String() + `"}`, http.StatusServiceUnavailable},
		{http.MethodGet, "/api/v1/products", "", http.StatusServiceUnavailable},
		{http.MethodGet, "/api/v1/products/" + uuid.New().String(), "", http.StatusServiceUnavailable},
		{http.MethodPut, "/api/v1/products/" + uuid.New().String(), `{"nama":"Test"}`, http.StatusServiceUnavailable},
		{http.MethodDelete, "/api/v1/products/" + uuid.New().String(), "", http.StatusServiceUnavailable},
	}

	for _, tc := range products {
		var bodyReader *bytes.Reader
		if tc.body != "" {
			bodyReader = bytes.NewReader([]byte(tc.body))
		} else {
			bodyReader = bytes.NewReader([]byte{})
		}
		req := httptest.NewRequest(tc.method, tc.path, bodyReader)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tc.expectCode {
			t.Errorf("route %s %s: expected %d, got %d", tc.method, tc.path, tc.expectCode, rec.Code)
		}
	}
}

// Test foto_url is optional (not validated)
func TestCreateProduct_OptionalFotoURL(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama":        "Produk Test",
		"deskripsi":   "Optional foto test",
		"harga_beli":  10000.0,
		"harga_jual":  15000.0,
		"stok":        50.0,
		"kategori_id": uuid.New().String(),
		// foto_url is omitted intentionally
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp handler.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify foto_url is empty (optional)
	if resp.FotoURL != "" {
		t.Errorf("expected empty foto_url, got '%s'", resp.FotoURL)
	}
}

// Test empty deskripsi is allowed (optional field)
func TestCreateProduct_OptionalDeskripsi(t *testing.T) {
	repo := newMockProductRepo()

	reqBody := map[string]any{
		"nama": "Produk Test",
		// deskripsi is omitted intentionally
		"harga_beli":  10000.0,
		"harga_jual":  15000.0,
		"stok":        50.0,
		"kategori_id": uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := newRequestWithPath(http.MethodPost, "/api/v1/products", body)
	rec := httptest.NewRecorder()

	h := handler.CreateProductHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
}

// Test max limit for pagination
func TestGetProducts_MaxLimit(t *testing.T) {
	repo := newMockProductRepo()

	req := newRequestWithPath(http.MethodGet, "/api/v1/products?limit=5000&offset=0", nil)
	rec := httptest.NewRecorder()

	h := handler.ListProductsHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Max limit is 1000
	if limit, ok := resp["limit"].(float64); !ok || limit != 1000 {
		t.Errorf("expected max limit 1000, got %v", resp["limit"])
	}
}

// Test negative offset is ignored (defaults to 0)
func TestGetProducts_NegativeOffset(t *testing.T) {
	repo := newMockProductRepo()

	req := newRequestWithPath(http.MethodGet, "/api/v1/products?limit=10&offset=-5", nil)
	rec := httptest.NewRecorder()

	h := handler.ListProductsHandler(repo)
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Offset should default to 0
	if offset, ok := resp["offset"].(float64); !ok || offset != 0 {
		t.Errorf("expected offset 0, got %v", resp["offset"])
	}
}
