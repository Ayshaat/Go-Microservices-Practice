package delivery_test

import (
	"cart/internal/delivery"
	"cart/internal/usecase/mocks"
	cart "cart/pkg/api/cart"
	"context"
	stdErrors "errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ClearCart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	server := delivery.NewCartServer(mockUsecase)

	validReq := &cart.ClearCartRequest{
		UserId: "1",
	}

	tests := []struct {
		name           string
		req            *cart.ClearCartRequest
		mockSetup      func()
		expectedResult *cart.CartResponse
		expectedErr    string
	}{
		{
			name: "success",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Clear(gomock.Any(), int64(1)).Return(nil)
			},
			expectedResult: &cart.CartResponse{Message: "Cart cleared successfully"},
		},
		{
			name: "internal server error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Clear(gomock.Any(), int64(1)).Return(stdErrors.New("db error"))
			},
			expectedErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.ClearCart(context.Background(), tt.req)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, resp)
			}
		})
	}
}
