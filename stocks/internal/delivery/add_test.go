package delivery_test

import (
	"bytes"
	stdErr "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"stocks/internal/delivery"
	"stocks/internal/errors"
	"stocks/internal/usecase/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_AddItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
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
			name:           "method not allowed",
			method:         http.MethodGet,
			body:           nil,
			mockSetup:      func() {},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			body:           []byte(`invalid json`),
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "Invalid request\n", // Updated casing
		},
		{
			name:   "success",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatusCode: http.StatusCreated, // 201 not 200
		},
		{
			name:   "invalid sku error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(errors.ErrInvalidSKU)
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid SKU — not registered\n", // match error
		},
		{
			name:   "ownership violation error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(errors.ErrOwnershipViolation)
			},
			wantStatusCode: http.StatusInternalServerError, // stays 500
			wantBody:       "ownership violation: user does not own this SKU\n",
		},
		{
			name:   "internal server error",
			method: http.MethodPost,
			body:   validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(stdErr.New("some error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "some error\n", // match exact output
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
