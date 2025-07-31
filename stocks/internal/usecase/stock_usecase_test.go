package usecase_test

import (
	"context"
	stdErr "errors"
	"fmt"
	"stocks/internal/errors"
	"stocks/internal/log/zap"
	"stocks/internal/models"
	"stocks/internal/repository/mocks"
	"stocks/internal/usecase"
	mockKafka "stocks/internal/usecase/mocks"
	"testing"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/golang/mock/gomock"
)

func TestStockUseCase_Add(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	item := models.StockItem{
		UserID:   1,
		SKU:      1001,
		Price:    10.0,
		Count:    5,
		Location: "loc1",
	}

	tests := []struct {
		name      string
		mockSetup func(mockRepo *mocks.MockStockRepository, mockProducer *mockKafka.MockProducerInterface)
		wantErr   error
	}{

		{
			name: "success new insert",
			mockSetup: func(mockRepo *mocks.MockStockRepository, mockProducer *mockKafka.MockProducerInterface) {
				mockRepo.EXPECT().GetSKUInfo(gomock.Any(), item.SKU).Return("t-shirt", "apparel", nil)
				mockRepo.EXPECT().GetByUserSKU(gomock.Any(), item.UserID, item.SKU).Return(models.StockItem{}, errors.ErrItemNotFound)
				mockRepo.EXPECT().InsertStockItem(gomock.Any(), item).Return(nil)
				mockProducer.EXPECT().
					SendSKUCreated(gomock.Any(), fmt.Sprint(item.SKU), item.Price, int(item.Count)).
					Return(nil)
			},

			wantErr: nil,
		},

		{
			name: "success update existing",
			mockSetup: func(mockRepo *mocks.MockStockRepository, mockProducer *mockKafka.MockProducerInterface) {
				existing := item
				existing.Count = 3

				mockRepo.EXPECT().GetSKUInfo(gomock.Any(), item.SKU).Return("t-shirt", "apparel", nil)
				mockRepo.EXPECT().GetByUserSKU(gomock.Any(), item.UserID, item.SKU).Return(existing, nil)
				mockRepo.EXPECT().UpdateCount(gomock.Any(), item.UserID, item.SKU, existing.Count+item.Count, item.Price).Return(nil)
				mockProducer.EXPECT().
					SendStockChanged(gomock.Any(), fmt.Sprint(existing.SKU), int(existing.Count+item.Count), existing.Price).
					Return(nil)
			},
			wantErr: nil,
		},

		{
			name: "invalid sku error",
			mockSetup: func(mockRepo *mocks.MockStockRepository, _ *mockKafka.MockProducerInterface) {
				mockRepo.EXPECT().GetSKUInfo(gomock.Any(), item.SKU).Return("", "", stdErr.New("not found"))
			},
			wantErr: errors.ErrInvalidSKU,
		},
		{
			name: "ownership violation error",
			mockSetup: func(mockRepo *mocks.MockStockRepository, _ *mockKafka.MockProducerInterface) {
				existing := item
				existing.UserID = 999
				mockRepo.EXPECT().GetSKUInfo(gomock.Any(), item.SKU).Return("t-shirt", "apparel", nil)
				mockRepo.EXPECT().GetByUserSKU(gomock.Any(), item.UserID, item.SKU).Return(existing, nil)
			},
			wantErr: errors.ErrOwnershipViolation,
		},

		{
			name: "other repo error",
			mockSetup: func(mockRepo *mocks.MockStockRepository, _ *mockKafka.MockProducerInterface) {
				mockRepo.EXPECT().GetSKUInfo(gomock.Any(), item.SKU).Return("t-shirt", "apparel", nil)
				mockRepo.EXPECT().GetByUserSKU(gomock.Any(), item.UserID, item.SKU).Return(models.StockItem{}, stdErr.New("some error"))
			},
			wantErr: stdErr.New("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockStockRepository(ctrl)
			mockProducer := mockKafka.NewMockProducerInterface(ctrl)
			txManager := &mockTxManager{}

			logger, cleanup, err := zap.NewLogger()
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}
			defer cleanup()

			uc := usecase.NewStockUsecase(mockRepo, txManager, mockProducer, logger)
			tt.mockSetup(mockRepo, mockProducer)

			err = uc.Add(ctx, item)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || (!stdErr.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error()) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestStockUseCase_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	mockProducer := mockKafka.NewMockProducerInterface(ctrl)
	txManager := &mockTxManager{}

	logger, cleanup, err := zap.NewLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer cleanup()

	uc := usecase.NewStockUsecase(mockRepo, txManager, mockProducer, logger)
	ctx := context.Background()

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   error
	}{

		{
			name: "success delete",
			mockSetup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), uint32(1001)).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "delete error",
			mockSetup: func() {
				mockRepo.EXPECT().Delete(gomock.Any(), uint32(1001)).Return(stdErr.New("delete error"))
			},
			wantErr: stdErr.New("delete error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()

			err := uc.Delete(ctx, 1001)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr != nil && (err == nil || err.Error() != tt.wantErr.Error()) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestStockUseCase_GetBySKU(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	mockProducer := mockKafka.NewMockProducerInterface(ctrl)
	txManager := &mockTxManager{}

	logger, cleanup, err := zap.NewLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer cleanup()

	uc := usecase.NewStockUsecase(mockRepo, txManager, mockProducer, logger)
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

	tests := []struct {
		name         string
		mockSetup    func()
		wantItem     models.StockItem
		expectErrStr string
	}{

		{
			name: "success",
			mockSetup: func() {
				mockRepo.EXPECT().GetBySKU(ctx, uint32(1001)).Return(expectedItem, nil)
			},
			wantItem: expectedItem,
		},
		{

			name: "not found",
			mockSetup: func() {
				mockRepo.EXPECT().GetBySKU(ctx, uint32(1001)).Return(models.StockItem{}, stdErr.New("not found"))
			},
			expectErrStr: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()

			item, err := uc.GetBySKU(ctx, 1001)
			if tt.expectErrStr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if item != tt.wantItem {
					t.Fatalf("expected %v, got %v", tt.wantItem, item)
				}
			} else {
				if err == nil || err.Error() != tt.expectErrStr {
					t.Fatalf("expected error %q, got %v", tt.expectErrStr, err)
				}
			}
		})
	}
}

func TestStockUseCase_ListByLocation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStockRepository(ctrl)
	mockProducer := mockKafka.NewMockProducerInterface(ctrl)
	txManager := &mockTxManager{}

	logger, cleanup, err := zap.NewLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer cleanup()

	uc := usecase.NewStockUsecase(mockRepo, txManager, mockProducer, logger)
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

	tests := []struct {
		name         string
		mockSetup    func()
		wantItems    []models.StockItem
		expectErrStr string
	}{

		{
			name: "success",
			mockSetup: func() {
				mockRepo.EXPECT().ListByLocation(ctx, "loc1", int64(10), int64(1)).Return(expectedItems, nil)
			},
			wantItems: expectedItems,
		},

		{
			name: "error",
			mockSetup: func() {
				mockRepo.EXPECT().ListByLocation(ctx, "loc1", int64(10), int64(1)).Return(nil, stdErr.New("db error"))
			},
			expectErrStr: "db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()

			items, err := uc.ListByLocation(ctx, "loc1", 10, 1)

			if tt.expectErrStr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(items) != len(tt.wantItems) {
					t.Fatalf("expected %d items, got %d", len(tt.wantItems), len(items))
				}
			} else {
				if err == nil || err.Error() != tt.expectErrStr {
					t.Fatalf("expected error %q, got %v", tt.expectErrStr, err)
				}
			}
		})
	}
}

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func (m *mockTxManager) DoWithSettings(ctx context.Context, settings trm.Settings, f func(ctx context.Context) error) error {
	return f(ctx)
}
