package delivery

import (
	"cart/internal/models"
	"encoding/json"
	"net/http"
)

type listRequest struct {
	UserID int64 `json:"userID"`
}

type listResponse struct {
	Items []models.CartItem `json:"items"`
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req listRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	items, err := h.usecase.List(r.Context(), req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []models.CartItem{}
	}

	res := listResponse{
		Items: items,
	}

	writeJSON(w, res)
}
