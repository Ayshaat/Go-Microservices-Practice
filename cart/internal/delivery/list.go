package delivery

import (
	"cart/internal/models"
	"encoding/json"
	"net/http"
)

type ListRequest struct {
	UserID int64 `json:"userID"`
}

type ListResponse struct {
	Items []models.CartItem `json:"items"`
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 {
		http.Error(w, "userID is required", http.StatusBadRequest)
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

	res := ListResponse{
		Items: items,
	}

	writeJSON(w, res)
}
