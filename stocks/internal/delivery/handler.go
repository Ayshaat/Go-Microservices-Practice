package delivery

import (
	"net/http"
	"stocks/internal/usecase"
)

type Handler struct {
	usecase usecase.StockUseCase
}

func NewHandler(u usecase.StockUseCase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/stocks/item/add", h.AddItem)
	mux.HandleFunc("/stocks/item/delete", h.DeleteItem)
	mux.HandleFunc("/stocks/item/get", h.GetItem)
	mux.HandleFunc("/stocks/list/location", h.ListByLocation)
}
