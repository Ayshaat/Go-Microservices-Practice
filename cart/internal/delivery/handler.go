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
	http.HandleFunc("/cart/item/add", h.AddItem)
	http.HandleFunc("/cart/item/delete", h.DeleteItem)
	http.HandleFunc("/cart/list", h.ListItems)
	http.HandleFunc("/cart/clear", h.ClearCart)
}
