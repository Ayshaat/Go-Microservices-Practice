package delivery

import (
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"stocks/internal/errors"
)

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var dto StockItemDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	item := dto.ToModel()

	if err := h.usecase.Add(r.Context(), item); err != nil {
		if stdErrors.Is(err, errors.ErrInvalidSKU) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if stdErrors.Is(err, errors.ErrItemExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
}
