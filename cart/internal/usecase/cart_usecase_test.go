package usecase

import (
	custErr "cart/internal/errors"
	"cart/internal/models"
	"cart/internal/usecase/mocks"
	"context"
	stdErr "errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCartUseCase_Add(t *testing.T) {
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

	t.Run("success", func(t *testing.T) {
		stockItem := models.StockItem{
			SKU:   100,
			Count: 10,
		}

		mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(stockItem, nil)
		mockCartRepo.EXPECT().Upsert(ctx, item).Return(nil)

		err := u.Add(ctx, item)
		assert.NoError(t, err)
	})

	t.Run("invalid sku error", func(t *testing.T) {
		mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(models.StockItem{}, stdErr.New("not found"))

		err := u.Add(ctx, item)
		assert.ErrorIs(t, err, custErr.ErrInvalidSKU)
	})

	t.Run("not enough stock error", func(t *testing.T) {
		stockItem := models.StockItem{
			SKU:   100,
			Count: 1,
		}

		mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(stockItem, nil)

		err := u.Add(ctx, item)
		assert.ErrorIs(t, err, custErr.ErrNotEnoughStock)
	})

	t.Run("repo error on upsert", func(t *testing.T) {
		stockItem := models.StockItem{
			SKU:   100,
			Count: 10,
		}

		mockStockRepo.EXPECT().GetBySKU(ctx, item.SKU).Return(stockItem, nil)
		mockCartRepo.EXPECT().Upsert(ctx, item).Return(stdErr.New("db error"))

		err := u.Add(ctx, item)
		assert.Error(t, err)
	})
}

func TestCartUseCase_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCartRepo := mocks.NewMockCartRepository(ctrl)
	mockStockRepo := mocks.NewMockStockRepository(ctrl)

	u := NewCartUsecase(mockCartRepo, mockStockRepo)

	ctx := context.Background()
	userID := int64(1)
	sku := uint32(100)

	t.Run("success", func(t *testing.T) {
		mockCartRepo.EXPECT().Delete(ctx, userID, sku).Return(nil)

		err := u.Delete(ctx, userID, sku)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		mockCartRepo.EXPECT().Delete(ctx, userID, sku).Return(stdErr.New("db error"))

		err := u.Delete(ctx, userID, sku)
		assert.Error(t, err)
	})
}

func TestCartUseCase_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCartRepo := mocks.NewMockCartRepository(ctrl)
	mockStockRepo := mocks.NewMockStockRepository(ctrl)

	u := NewCartUsecase(mockCartRepo, mockStockRepo)

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

	t.Run("success", func(t *testing.T) {
		mockCartRepo.EXPECT().List(ctx, userID).Return(cartItems, nil)
		mockStockRepo.EXPECT().GetBySKU(ctx, uint32(100)).Return(stockItem1, nil)
		mockStockRepo.EXPECT().GetBySKU(ctx, uint32(101)).Return(stockItem2, nil)

		result, err := u.List(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, stockItem1.Price, result[0].Price)
		assert.Equal(t, stockItem2.Price, result[1].Price)
		assert.Equal(t, stockItem1.Count, result[0].Count)
		assert.Equal(t, stockItem2.Count, result[1].Count)
	})

	t.Run("repo list error", func(t *testing.T) {
		mockCartRepo.EXPECT().List(ctx, userID).Return(nil, stdErr.New("db error"))

		_, err := u.List(ctx, userID)
		assert.Error(t, err)
	})

	t.Run("stock get error", func(t *testing.T) {
		mockCartRepo.EXPECT().List(ctx, userID).Return(cartItems, nil)
		mockStockRepo.EXPECT().GetBySKU(ctx, uint32(100)).Return(models.StockItem{}, stdErr.New("not found"))

		_, err := u.List(ctx, userID)
		assert.Error(t, err)
	})
}

func TestCartUseCase_Clear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCartRepo := mocks.NewMockCartRepository(ctrl)
	mockStockRepo := mocks.NewMockStockRepository(ctrl)

	u := NewCartUsecase(mockCartRepo, mockStockRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("success", func(t *testing.T) {
		mockCartRepo.EXPECT().Clear(ctx, userID).Return(nil)

		err := u.Clear(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		mockCartRepo.EXPECT().Clear(ctx, userID).Return(stdErr.New("db error"))

		err := u.Clear(ctx, userID)
		assert.Error(t, err)
	})
}
