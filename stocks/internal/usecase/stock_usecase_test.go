package usecase_test

import (
	"context"
	stdErr "errors"
	"stocks/internal/errors"
	"stocks/internal/models"
	"stocks/internal/repository/mocks"
	"stocks/internal/usecase"
	"testing"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/golang/mock/gomock"
)

func TestStockUseCase_Add(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	txManager := &mockTxManager{}

	uc := usecase.NewStockUsecase(mockRepo, txManager)

	ctx := context.Background()
	item := models.StockItem{
		UserID:   1,
		SKU:      1001,
		Price:    10.0,
		Count:    5,
		Location: "loc1",
	}

	t.Run("success new insert", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().GetSKUInfo(ctx, item.SKU).Return("t-shirt", "apparel", nil)
		mockRepo.EXPECT().GetByUserSKU(ctx, item.UserID, item.SKU).Return(models.StockItem{}, errors.ErrItemNotFound)
		mockRepo.EXPECT().InsertStockItem(ctx, item).Return(nil)

		err := uc.Add(ctx, item)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success update existing", func(t *testing.T) {
		t.Parallel()

		existing := item
		existing.Count = 3

		mockRepo.EXPECT().GetSKUInfo(ctx, item.SKU).Return("t-shirt", "apparel", nil)
		mockRepo.EXPECT().GetByUserSKU(ctx, item.UserID, item.SKU).Return(existing, nil)
		mockRepo.EXPECT().UpdateCount(ctx, item.UserID, item.SKU, existing.Count+item.Count).Return(nil)

		err := uc.Add(ctx, item)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid sku error", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().GetSKUInfo(ctx, item.SKU).Return("", "", stdErr.New("not found"))

		err := uc.Add(ctx, item)
		if !stdErr.Is(err, errors.ErrInvalidSKU) {
			t.Fatalf("expected ErrInvalidSKU, got %v", err)
		}
	})

	t.Run("ownership violation error", func(t *testing.T) {
		t.Parallel()

		existing := item
		existing.UserID = 999

		mockRepo.EXPECT().GetSKUInfo(ctx, item.SKU).Return("t-shirt", "apparel", nil)
		mockRepo.EXPECT().GetByUserSKU(ctx, item.UserID, item.SKU).Return(existing, nil)

		err := uc.Add(ctx, item)
		if !stdErr.Is(err, errors.ErrOwnershipViolation) {
			t.Fatalf("expected ErrOwnershipViolation, got %v", err)
		}
	})

	t.Run("other repo error", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().GetSKUInfo(ctx, item.SKU).Return("t-shirt", "apparel", nil)
		mockRepo.EXPECT().GetByUserSKU(ctx, item.UserID, item.SKU).Return(models.StockItem{}, stdErr.New("some error"))

		err := uc.Add(ctx, item)
		if err == nil || err.Error() != "some error" {
			t.Fatalf("expected some error, got %v", err)
		}
	})
}

func TestStockUseCase_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	txManager := &mockTxManager{}

	uc := usecase.NewStockUsecase(mockRepo, txManager)
	ctx := context.Background()

	t.Run("success delete", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().Delete(ctx, uint32(1001)).Return(nil)

		err := uc.Delete(ctx, 1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().Delete(ctx, uint32(1001)).Return(stdErr.New("delete error"))

		err := uc.Delete(ctx, 1001)
		if err == nil || err.Error() != "delete error" {
			t.Fatalf("expected delete error, got %v", err)
		}
	})
}

func TestStockUseCase_GetBySKU(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	txManager := &mockTxManager{}

	uc := usecase.NewStockUsecase(mockRepo, txManager)
	ctx := context.Background()

	expectedItem := models.StockItem{
		UserID:   1,
		SKU:      1001,
		Name:     "t-shirt",
		Type:     "apparel",
		Price:    10.0,
		Count:    5,
		Location: "loc1",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().GetBySKU(ctx, uint32(1001)).Return(expectedItem, nil)

		item, err := uc.GetBySKU(ctx, 1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if item != expectedItem {
			t.Fatalf("expected %v, got %v", expectedItem, item)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().GetBySKU(ctx, uint32(1001)).Return(models.StockItem{}, stdErr.New("not found"))

		_, err := uc.GetBySKU(ctx, 1001)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestStockUseCase_ListByLocation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	txManager := &mockTxManager{}

	uc := usecase.NewStockUsecase(mockRepo, txManager)
	ctx := context.Background()

	expectedItems := []models.StockItem{
		{
			UserID:   1,
			SKU:      1001,
			Name:     "t-shirt",
			Type:     "apparel",
			Price:    10.0,
			Count:    5,
			Location: "loc1",
		},
		{
			UserID:   2,
			SKU:      2020,
			Name:     "cup",
			Type:     "accessory",
			Price:    5.0,
			Count:    3,
			Location: "loc1",
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().ListByLocation(ctx, "loc1", int64(10), int64(1)).Return(expectedItems, nil)

		items, err := uc.ListByLocation(ctx, "loc1", 10, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(items) != len(expectedItems) {
			t.Fatalf("expected %d items, got %d", len(expectedItems), len(items))
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		mockRepo.EXPECT().ListByLocation(ctx, "loc1", int64(10), int64(1)).Return(nil, stdErr.New("db error"))

		_, err := uc.ListByLocation(ctx, "loc1", 10, 1)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (m *mockTxManager) DoWithSettings(ctx context.Context, settings trm.Settings, f func(ctx context.Context) error) error {
	return f(ctx)
}
