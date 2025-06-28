package delivery

import (
	"encoding/json"
	"net/http"
	"stocks/internal/models"
)

type listLocationRequest struct {
	Location    string `json:"location"`
	PageSize    int64  `json:"pageSize"`
	CurrentPage int64  `json:"currentPage"`
}

type listLocationResponse struct {
	Items []models.StockItem `json:"items"`
}

func (h *Handler) ListByLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req listLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	items, err := h.usecase.ListByLocation(r.Context(), req.Location, req.PageSize, req.CurrentPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := listLocationResponse{
		Items: items,
	}

	writeJSON(w, res)
}
