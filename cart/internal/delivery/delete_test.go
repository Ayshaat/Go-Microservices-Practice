package delivery_test

import (
	"bytes"
	"cart/internal/delivery"
	"cart/internal/errors"
	"cart/internal/usecase/mocks"
	stdErrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_DeleteItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validJSON := []byte(`{"userID":1,"sku":100}`)

	tests := []struct {
		name           string
		body           []byte
		mockSetup      func()
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "invalid JSON",
			body:           []byte(`invalid json`),
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid request\n",
		},
		{
			name: "success",
			body: validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), int64(1), uint32(100)).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
		{
			name: "not found error",
			body: validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), int64(1), uint32(100)).Return(errors.ErrCartItemNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       errors.ErrCartItemNotFound.Error() + "\n",
		},
		{
			name: "internal server error",
			body: validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), int64(1), uint32(100)).Return(stdErrors.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "db error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/cart/item/delete", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.DeleteItem(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
