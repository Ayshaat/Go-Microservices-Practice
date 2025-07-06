package delivery_test

import (
	"bytes"
	"encoding/json"
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

func TestHandler_ListByLocation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validReq := map[string]interface{}{
		"location":    "loc1",
		"pageSize":    10,
		"currentPage": 1,
	}

	validJSON, err := json.Marshal(validReq)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	expectedItems := []models.StockItem{
		{
			UserID:   1,
			SKU:      1001,
			Name:     "T-Shirt",
			Type:     "clothing",
			Price:    15.0,
			Count:    5,
			Location: "loc1",
		},
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
			name:   "method not allowed",
			method: http.MethodGet,
			body:   nil,
			mockSetup: func() {
				// no-op
			},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid JSON",
			method: http.MethodPost,
			body:   []byte(`invalid`),
			mockSetup: func() {
				// no-op
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "internal error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().
					ListByLocation(gomock.Any(), "loc1", int64(10), int64(1)).
					Return(nil, errors.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "success",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().
					ListByLocation(gomock.Any(), "loc1", int64(10), int64(1)).
					Return(expectedItems, nil)
			},
			wantStatusCode: http.StatusOK,
			wantContains:   `"SKU":1001`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(tt.method, "/stocks/location", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ListByLocation(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantContains != "" {
				assert.Contains(t, rr.Body.String(), tt.wantContains)
			}
		})
	}
}
