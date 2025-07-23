package delivery_test

import (
	"cart/internal/delivery"
	"cart/internal/errors"
	"cart/internal/usecase/mocks"
	cart "cart/pkg/api/cart"
	"context"
	stdErr "errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_AddItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockCartUseCase(ctrl)
	server := delivery.NewCartServer(mockUsecase)

	validReq := &cart.AddItemRequest{
		UserId: "1",
		Sku:    "100",
		Count:  2,
	}

	tests := []struct {
		name           string
		req            *cart.AddItemRequest
		mockSetup      func()
		expectedResult *cart.CartResponse
		expectedErr    string
	}{
		{
			name: "success",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedResult: &cart.CartResponse{Message: "Item added successfully"},
		},
		{
			name: "invalid sku error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(errors.ErrInvalidSKU)
			},
			expectedErr: "Invalid SKU",
		},
		{
			name: "item exists error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(errors.ErrCartItemExists)
			},
			expectedErr: "Item already exists",
		},
		{
			name: "internal server error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(stdErr.New("some error"))
			},
			expectedErr: "rpc error: code = Internal desc = some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.AddItem(context.Background(), tt.req)

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
