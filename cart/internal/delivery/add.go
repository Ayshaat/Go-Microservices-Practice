package delivery

import (
	"cart/internal/models"
	"encoding/json"
	"net/http"
)

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var item models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.usecase.Add(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
