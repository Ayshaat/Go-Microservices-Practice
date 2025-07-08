package delivery

import (
	"cart/internal/errors"
	"encoding/json"
	stdErrors "errors"
	"net/http"
)

type deleteRequest struct {
	UserID int64  `json:"userID"`
	SKU    uint32 `json:"sku"`
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.usecase.Delete(r.Context(), req.UserID, req.SKU); err != nil {
		if stdErrors.Is(err, errors.ErrCartItemNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
