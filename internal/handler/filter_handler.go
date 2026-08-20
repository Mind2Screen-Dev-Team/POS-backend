package handler

import (
	"net/http"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

type FilterHandler struct {
	ProductRepo *repository.ProductRepository
}

func (h *FilterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("category_id")

	products, err := h.ProductRepo.FilterProductsByCategory(r.Context(), categoryID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, products)
}
