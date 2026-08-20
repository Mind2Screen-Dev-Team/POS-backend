package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

type mockProductRepo struct{}

func (m *mockProductRepo) SearchProducts(ctx context.Context, query string, categoryID string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *mockProductRepo) FilterProductsByCategory(ctx context.Context, categoryID string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func TestSearchHandler_EmptyQuery(t *testing.T) {
	repo := &repository.ProductRepository{}
	handler := &SearchHandler{ProductRepo: repo}

	req := httptest.NewRequest("GET", "/api/v1/products/search?q=", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestFilterHandler_CategoryFilter(t *testing.T) {
	repo := &repository.ProductRepository{}
	handler := &FilterHandler{ProductRepo: repo}

	req := httptest.NewRequest("GET", "/api/v1/products/filter?category_id=test-uuid", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
