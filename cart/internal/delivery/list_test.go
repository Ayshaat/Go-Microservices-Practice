package delivery_test

import (
	"cart/internal/delivery"
	"cart/internal/models"
	"cart/internal/usecase/mocks"
	cart "cart/pkg/api/cart"
	"context"
	stdErrors "errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ListItems(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	server := delivery.NewCartServer(mockUsecase)
	validReq := &cart.ListCartRequest{
		UserId: "1",
	}

	tests := []struct {
		name           string
		req            *cart.ListCartRequest
		mockSetup      func()
		wantStatusCode int
		expectedResult *cart.ListCartResponse
		expectedErr    string
	}{
		{
			name: "success with items",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().List(gomock.Any(), int64(1)).Return([]models.CartItem{
					{UserID: 1, SKU: 100, Count: 2},
					{UserID: 1, SKU: 101, Count: 3},
				}, nil)
			},
			expectedResult: &cart.ListCartResponse{
				UserId: "1",
				Items: []*cart.CartItem{
					{Sku: "100", Count: 2},
					{Sku: "101", Count: 3},
				},
			},
		},
		{
			name: "success with nil items returns empty slice",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().List(gomock.Any(), int64(1)).Return(nil, nil)
			},
			expectedResult: &cart.ListCartResponse{
				UserId: "1",
				Items:  []*cart.CartItem{},
			},
		},
		{
			name: "internal server error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().List(gomock.Any(), int64(1)).Return(nil, stdErrors.New("db error"))
			},
			expectedErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.ListCart(context.Background(), tt.req)

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
