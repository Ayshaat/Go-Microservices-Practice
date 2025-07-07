package usecase

import (
	"cart/internal/errors"
	"cart/internal/models"
	"cart/internal/usecase/mocks"
	"context"
	stdErr "errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCartUseCase_Add(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCartRepo := mocks.NewMockCartRepository(ctrl)
	mockStockRepo := mocks.NewMockStockRepository(ctrl)

	u := NewCartUsecase(mockCartRepo, mockStockRepo)

	ctx := context.Background()
	item := models.CartItem{
		UserID: 1,
		SKU:    100,
		Count:  2,
	}

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   error
	}{

		{
			name: "success",
			mockSetup: func() {
				stockItem := models.StockItem{SKU: 100, Count: 10}
				mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(stockItem, nil)
				mockCartRepo.EXPECT().Upsert(ctx, item).Return(nil)
			},
			wantErr: nil,
		},

		{
			name: "invalid sku error",
			mockSetup: func() {
				mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(models.StockItem{}, stdErr.New("not found"))
			},
			wantErr: errors.ErrInvalidSKU,
		},

		{
			name: "not enough stock error",
			mockSetup: func() {
				stockItem := models.StockItem{SKU: 100, Count: 1}
				mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(stockItem, nil)
			},
			wantErr: errors.ErrNotEnoughStock,
		},

		{
			name: "repo error on upsert",
			mockSetup: func() {
				stockItem := models.StockItem{SKU: 100, Count: 10}
				mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(stockItem, nil)
				mockCartRepo.EXPECT().Upsert(ctx, item).Return(stdErr.New("db error"))
			},
			wantErr: stdErr.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()

			err := u.Add(ctx, item)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}

func TestCartUseCase_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCartRepo := mocks.NewMockCartRepository(ctrl)
	mockStockRepo := mocks.NewMockStockRepository(ctrl)

	u := NewCartUsecase(mockCartRepo, mockStockRepo)

	ctx := context.Background()
	userID := int64(1)
	sku := uint32(100)

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockCartRepo.EXPECT().Delete(ctx, userID, sku).Return(nil)
			},
			wantErr: nil,
		},
		{

			name: "repo error",
			mockSetup: func() {
				mockCartRepo.EXPECT().Delete(ctx, userID, sku).Return(stdErr.New("db error"))
			},
			wantErr: stdErr.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()

			err := u.Delete(ctx, userID, sku)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}

func TestCartUseCase_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := int64(1)

	cartItems := []models.CartItem{
		{UserID: userID, SKU: 100, Count: 2},
		{UserID: userID, SKU: 101, Count: 3},
	}

	stockItem1 := models.StockItem{
		SKU:   100,
		Price: 9.99,
		Count: 5,
	}

	stockItem2 := models.StockItem{
		SKU:   101,
		Price: 19.99,
		Count: 7,
	}

	tests := []struct {
		name         string
		mockSetup    func(cartRepo *mocks.MockCartRepository, stockRepo *mocks.MockStockRepository)
		wantLen      int
		wantErr      bool
		wantErrIs    error
		wantPriceSKU map[uint32]float64
	}{

		{
			name: "success",
			mockSetup: func(cartRepo *mocks.MockCartRepository, stockRepo *mocks.MockStockRepository) {
				cartRepo.EXPECT().List(ctx, userID).Return(cartItems, nil)
				stockRepo.EXPECT().GetBySKU(ctx, uint32(100)).Return(stockItem1, nil)
				stockRepo.EXPECT().GetBySKU(ctx, uint32(101)).Return(stockItem2, nil)
			},
			wantLen: 2,
			wantPriceSKU: map[uint32]float64{
				100: 9.99,
				101: 19.99,
			},
		},
		{

			name: "repo list error",
			mockSetup: func(cartRepo *mocks.MockCartRepository, stockRepo *mocks.MockStockRepository) {
				cartRepo.EXPECT().List(ctx, userID).Return(nil, stdErr.New("db error"))
			},
			wantErr:   true,
			wantErrIs: stdErr.New("db error"),
		},
		{
			name: "stock get error",
			mockSetup: func(cartRepo *mocks.MockCartRepository, stockRepo *mocks.MockStockRepository) {
				cartRepo.EXPECT().List(ctx, userID).Return(cartItems, nil)
				stockRepo.EXPECT().GetBySKU(ctx, uint32(100)).Return(models.StockItem{}, stdErr.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCartRepo := mocks.NewMockCartRepository(ctrl)
			mockStockRepo := mocks.NewMockStockRepository(ctrl)

			u := NewCartUsecase(mockCartRepo, mockStockRepo)

			tt.mockSetup(mockCartRepo, mockStockRepo)

			result, err := u.List(ctx, userID)

			if !tt.wantErr {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)

				for _, r := range result {
					assert.InDelta(t, tt.wantPriceSKU[r.SKU], r.Price, 0.0001)
				}
			} else {
				assert.Error(t, err)

				if tt.wantErrIs != nil {
					assert.EqualError(t, err, tt.wantErrIs.Error())
				}
			}
		})
	}
}

func TestCartUseCase_Clear(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCartRepo := mocks.NewMockCartRepository(ctrl)
	mockStockRepo := mocks.NewMockStockRepository(ctrl)

	u := NewCartUsecase(mockCartRepo, mockStockRepo)

	ctx := context.Background()
	userID := int64(1)

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockCartRepo.EXPECT().Clear(ctx, userID).Return(nil)
			},
			wantErr: nil,
		},

		{
			name: "repo error",
			mockSetup: func() {
				mockCartRepo.EXPECT().Clear(ctx, userID).Return(stdErr.New("db error"))
			},
			wantErr: stdErr.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()

			err := u.Clear(ctx, userID)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}
