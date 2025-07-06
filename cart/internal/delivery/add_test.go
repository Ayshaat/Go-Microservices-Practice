package delivery_test

import (
	"bytes"
	"cart/internal/delivery"
	"cart/internal/errors"
	"cart/internal/usecase/mocks"
	stdErr "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_AddItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validJSON := []byte(`{"user_id":1,"sku":100,"count":2}`)

	tests := []struct {
		name           string
		method         string
		body           []byte
		mockSetup      func()
		wantStatusCode int
		wantBody       string
	}{
		{
			name:   "method not allowed",
			method: http.MethodGet,
			body:   nil,
			mockSetup: func() {
				// no usecase call expected
			},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid JSON",
			method: http.MethodPost,
			body:   []byte(`invalid json`),
			mockSetup: func() {
				// no usecase call expected
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid request\n",
		},
		{
			name:   "success",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "invalid sku error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(errors.ErrInvalidSKU)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "Invalid SKU — not registered\n",
		},
		{
			name:   "item exists error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(errors.ErrCartItemExists)
			},
			wantStatusCode: http.StatusConflict,
			wantBody:       "Item already exists in cart\n",
		},
		{
			name:   "internal server error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(stdErr.New("some error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(tt.method, "/add", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.AddItem(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
