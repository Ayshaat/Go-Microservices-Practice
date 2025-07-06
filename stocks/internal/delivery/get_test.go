package delivery_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"stocks/internal/delivery"
	"stocks/internal/models"
	"stocks/internal/usecase/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validJSON := []byte(`{"sku":1001}`)
	expectedItem := models.StockItem{
		UserID:   1,
		SKU:      1001,
		Name:     "t-shirt",
		Type:     "clothing",
		Price:    15.5,
		Count:    10,
		Location: "loc1",
	}

	tests := []struct {
		name           string
		method         string
		body           []byte
		mockSetup      func()
		wantStatusCode int
		wantContains   string
	}{
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			body:           nil,
			mockSetup:      func() {},
			wantStatusCode: http.StatusMethodNotAllowed,
			wantContains:   "method not allowed",
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           []byte(`invalid`),
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "invalid request body",
		},
		{
			name:   "not found",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().GetBySKU(gomock.Any(), uint32(1001)).
					Return(models.StockItem{}, errors.New("not found"))
			},
			wantStatusCode: http.StatusNotFound,
			wantContains:   "not found",
		},
		{
			name:   "success",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().GetBySKU(gomock.Any(), uint32(1001)).
					Return(expectedItem, nil)
			},
			wantStatusCode: http.StatusOK,
			wantContains:   `"SKU":1001`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(tt.method, "/stocks/item/get", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.GetItem(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantContains != "" {
				assert.Contains(t, rr.Body.String(), tt.wantContains)
			}
		})
	}
}
