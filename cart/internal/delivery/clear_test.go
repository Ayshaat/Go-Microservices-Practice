package delivery_test

import (
	"bytes"
	"cart/internal/delivery"
	"cart/internal/usecase/mocks"
	stdErrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ClearCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validJSON := []byte(`{"userID":1}`)

	tests := []struct {
		name           string
		method         string
		body           []byte
		mockSetup      func()
		wantStatusCode int
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
			name:   "success",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Clear(gomock.Any(), int64(1)).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
		{
			name:   "internal server error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Clear(gomock.Any(), int64(1)).Return(stdErrors.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "db error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(tt.method, "/cart/clear", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ClearCart(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
