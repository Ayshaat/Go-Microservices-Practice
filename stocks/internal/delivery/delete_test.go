package delivery_test

import (
	"bytes"
	stdErr "errors"
	"net/http"
	"net/http/httptest"
	"stocks/internal/delivery"
	"stocks/internal/usecase/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_DeleteItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	handler := delivery.NewHandler(mockUsecase)

	validJSON := []byte(`{"userID":1,"sku":1001}`)

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
			wantBody:       "Invalid request\n",
		},
		{
			name: "success",
			body: validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), uint32(1001)).Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "internal error from usecase",
			body: validJSON,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), uint32(1001)).Return(stdErr.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "db error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.DeleteItem(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
