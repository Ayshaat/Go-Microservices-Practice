package delivery

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var dto CartItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)

		return
	}

	item := dto.ToModel()

	if err := h.usecase.Add(r.Context(), item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
