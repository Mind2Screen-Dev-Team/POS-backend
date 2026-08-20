package handler

import (
	"net/http"
	"strings"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

type SearchHandler struct {
	ProductRepo *repository.ProductRepository
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	query = strings.TrimSpace(query)
	categoryID := r.URL.Query().Get("category_id")

	products, err := h.ProductRepo.SearchProducts(r.Context(), query, categoryID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, products)
}
