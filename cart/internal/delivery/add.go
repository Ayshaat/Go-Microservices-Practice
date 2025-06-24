package delivery

import (
	"cart/internal/errors"
	"encoding/json"
	stdErrors "errors"
	"log"
	"net/http"
)

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var dto CartItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)

		return
	}

	item := dto.ToModel()

	err := h.usecase.Add(r.Context(), item)
	if err != nil {
		if stdErrors.Is(err, errors.ErrInvalidSKU) {
			http.Error(w, "Invalid SKU — not registered", http.StatusBadRequest)
			return
		}

		if stdErrors.Is(err, errors.ErrCartItemExists) {
			http.Error(w, "Item already exists in cart", http.StatusConflict)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
