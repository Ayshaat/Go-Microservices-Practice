package delivery

import (
	"encoding/json"
	"net/http"
)

type deleteRequest struct {
	UserID int64  `json:"userID"`
	SKU    uint32 `json:"sku"`
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.usecase.Delete(req.SKU); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
