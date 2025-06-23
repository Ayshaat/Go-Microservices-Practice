package delivery

import (
	"cart/internal/usecase"
	"net/http"
)

type Handler struct {
	usecase usecase.CartUseCase
}

func NewHandler(u usecase.CartUseCase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/cart/item/add", h.AddItem)
	mux.HandleFunc("/cart/item/delete", h.DeleteItem)
	mux.HandleFunc("/cart/list", h.ListItems)
	mux.HandleFunc("/cart/clear", h.ClearCart)
}
