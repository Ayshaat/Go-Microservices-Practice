package delivery_test

import (
	"bytes"
	"cart/internal/delivery"
	"cart/internal/models"
	"cart/internal/usecase/mocks"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ListItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validJSON := []byte(`{"userID":1}`)

	decodeResponse := func(body *bytes.Buffer) (delivery.ListResponse, error) {
		var res delivery.ListResponse
		err := json.NewDecoder(body).Decode(&res)

		return res, err
	}

	tests := []struct {
		name           string
		method         string
		body           []byte
		mockSetup      func()
		wantStatusCode int
		wantItems      []models.CartItem
		wantBody       string
	}{
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			body:           nil,
			mockSetup:      func() {},
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "method not allowed\n",
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           []byte(`invalid json`),
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid request body\n",
		},
		{
			name:   "success with items",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().List(gomock.Any(), int64(1)).Return([]models.CartItem{
					{UserID: 1, SKU: 100, Count: 2},
					{UserID: 1, SKU: 101, Count: 3},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantItems: []models.CartItem{
				{UserID: 1, SKU: 100, Count: 2},
				{UserID: 1, SKU: 101, Count: 3},
			},
		},
		{
			name:   "success with nil items returns empty slice",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().List(gomock.Any(), int64(1)).Return(nil, nil)
			},
			wantStatusCode: http.StatusOK,
			wantItems:      []models.CartItem{},
		},
		{
			name:   "internal server error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().List(gomock.Any(), int64(1)).Return(nil, stdErrors.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "db error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(tt.method, "/cart/list", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ListItems(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantStatusCode == http.StatusOK {
				res, err := decodeResponse(rr.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantItems, res.Items)
			} else {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
