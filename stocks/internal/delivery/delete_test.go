package delivery_test

import (
	"context"
	stdErr "errors"
	"stocks/internal/delivery"
	"stocks/internal/usecase/mocks"
	stockspb "stocks/pkg/api/stocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_DeleteItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	server := delivery.NewStockServer(mockUsecase)

	validReq := &stockspb.DeleteItemRequest{
		Sku:      "1001",
		Location: "warehouse1",
	}

	tests := []struct {
		name           string
		req            *stockspb.DeleteItemRequest
		mockSetup      func()
		expectedResult *stockspb.StockResponse
		expectedErr    string
	}{
		{
			name: "success",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), uint32(1001)).Return(nil)
			},
			expectedResult: &stockspb.StockResponse{Message: "Item deleted successfully"},
		},
		{
			name: "internal error from usecase",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().Delete(gomock.Any(), uint32(1001)).Return(stdErr.New("db error"))
			},
			expectedErr: "failed to delete item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.DeleteItem(context.Background(), tt.req)

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
