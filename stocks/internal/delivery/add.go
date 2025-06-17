package delivery

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var dto StockItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	item := dto.ToModel()

	if err := h.usecase.Add(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
