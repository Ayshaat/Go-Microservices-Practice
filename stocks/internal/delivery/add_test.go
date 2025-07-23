package delivery_test

import (
	"context"
	stdErr "errors"
	"testing"

	"stocks/internal/delivery"
	"stocks/internal/errors"
	"stocks/internal/usecase/mocks"
	stockspb "stocks/pkg/api"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_AddItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	server := delivery.NewStockServer(mockUsecase)

	validReq := &stockspb.AddItemRequest{
		Sku:      "100",
		Location: "warehouse1",
		Count:    2,
	}

	tests := []struct {
		name           string
		req            *stockspb.AddItemRequest
		mockSetup      func()
		expectedResult *stockspb.StockResponse
		expectedErr    string
	}{
		{
			name: "success",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResult: &stockspb.StockResponse{Message: "Item added successfully"},
		},
		{
			name: "invalid sku error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(errors.ErrInvalidSKU)
			},
			expectedErr: "invalid SKU",
		},
		{
			name: "ownership violation error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(errors.ErrOwnershipViolation)
			},
			expectedErr: "ownership violation",
		},
		{
			name: "internal server error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Add(gomock.Any(), gomock.Any()).Return(stdErr.New("some error"))
			},
			expectedErr: "some error",
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
