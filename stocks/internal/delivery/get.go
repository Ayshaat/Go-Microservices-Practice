package delivery

import (
	"encoding/json"
	"net/http"
)

type getRequest struct {
	SKU uint32 `json:"sku"`
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req getRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := h.usecase.GetBySKU(r.Context(), req.SKU)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, item)
}
