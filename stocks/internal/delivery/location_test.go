package delivery_test

import (
	"context"
	"errors"
	"stocks/internal/delivery"
	"stocks/internal/models"
	"stocks/internal/usecase/mocks"
	stockspb "stocks/pkg/api"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ListByLocation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockStockUseCase(ctrl)
	server := delivery.NewStockServer(mockUsecase)

	validReq := &stockspb.ListByLocationRequest{
		Location: "loc1",
	}

	expectedItems := []models.StockItem{
		{
			UserID:   1,
			SKU:      1001,
			Name:     "T-Shirt",
			Type:     "clothing",
			Price:    15.0,
			Count:    5,
			Location: "loc1",
		},
	}

	tests := []struct {
		name           string
		req            *stockspb.ListByLocationRequest
		mockSetup      func()
		expectedResult *stockspb.ListByLocationResponse
		expectedErr    string
	}{
		{
			name: "internal error",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					ListByLocation(gomock.Any(), "loc1", int64(10), int64(1)).
					Return(nil, errors.New("db error"))
			},
			expectedErr: "db error",
		},
		{
			name: "success",
			req:  validReq,
			mockSetup: func() {
				mockUsecase.EXPECT().
					ListByLocation(gomock.Any(), "loc1", int64(10), int64(1)).
					Return(expectedItems, nil)
			},
			expectedResult: &stockspb.ListByLocationResponse{
				Location: "loc1",
				Items: []*stockspb.StockItem{
					{
						Sku:      "1001",
						Location: "loc1",
						Count:    5,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.ListByLocation(context.Background(), tt.req)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Location, resp.Location)
				assert.Len(t, resp.Items, len(tt.expectedResult.Items))
			}
		})
	}
}
