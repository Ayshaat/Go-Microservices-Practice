package delivery_test

import (
	"context"
	"errors"
	"stocks/internal/delivery"
	"stocks/internal/models"
	"stocks/internal/usecase/mocks"
	stockspb "stocks/pkg/api/stocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetItem(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	server := delivery.NewStockServer(mockUsecase, mockLogger)

	validReq := &stockspb.GetItemRequest{
		Sku:      "1001",
		Location: "loc1",
	}

	expectedItem := models.StockItem{
		UserID:   1,
		SKU:      1001,
		Name:     "t-shirt",
		Type:     "clothing",
		Price:    15.5,
		Count:    10,
		Location: "loc1",
	}

	tests := []struct {
		name           string
		req            *stockspb.GetItemRequest
		mockSetup      func()
		expectedResult *stockspb.StockItem
		expectedErr    string
	}{
		{
			name: "not found",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().GetBySKU(gomock.Any(), uint32(1001)).
					Return(models.StockItem{}, errors.New("not found"))
				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			expectedErr: "not found",
		},
		{
			name: "success",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().GetBySKU(gomock.Any(), uint32(1001)).
					Return(expectedItem, nil)
				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			},
			expectedResult: &stockspb.StockItem{
				Sku:      "1001",
				Location: "loc1",
				Count:    int32(expectedItem.Count),
				Price:    float32(expectedItem.Price),
			},
		},
		{
			name: "internal error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetBySKU(gomock.Any(), uint32(1001)).
					Return(models.StockItem{}, errors.New("db error"))
				mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			expectedErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.GetItem(context.Background(), tt.req)

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
