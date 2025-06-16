package delivery

import (
	"encoding/json"
	"net/http"
	"stocks/internal/models"
)

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var item models.StockItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.usecase.Add(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
