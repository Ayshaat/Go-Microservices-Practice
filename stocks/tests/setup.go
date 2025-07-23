package tests

import (
	"net/http"
	"os"
	"testing"

	"context"
	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/repository"
	"stocks/internal/server"
	"stocks/internal/usecase"

	"github.com/jmoiron/sqlx"

	mockKafka "stocks/internal/usecase/mocks"

	"github.com/golang/mock/gomock"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	_ "github.com/lib/pq"
)

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TEST not set")
	}
}

func mustLoadConfig(t *testing.T) *config.Config {
	cfg, err := config.Load("../.env.local")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	return cfg
}

func setupTestDB(t *testing.T) *sqlx.DB {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TEST not set")
	}

	cfg := mustLoadConfig(t)

	sqlDB, err := db.ConnectDB(cfg.PostgresConnStr(), "../internal/db/migrations")
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	dbx := sqlx.NewDb(sqlDB, "postgres")
	// Clean tables before tests
	if _, err := dbx.Exec("DELETE FROM stock_items"); err != nil {
		t.Fatalf("failed to clear stock_items: %v", err)
	}

	if _, err := dbx.Exec("DELETE FROM sku_info"); err != nil {
		t.Fatalf("failed to clear sku_info: %v", err)
	}

	// Insert test SKU info
	_, err = dbx.Exec(`INSERT INTO sku_info (sku, name, type) VALUES (1001, 't-shirt', 'clothing')`)
	if err != nil {
		t.Fatalf("failed to insert sku_info: %v", err)
	}

	return dbx
}

func setupServer(t *testing.T, db *sqlx.DB, ctrl *gomock.Controller) http.Handler {
	txFactory := trmsqlx.NewDefaultFactory(db.DB)
	txManager := manager.Must(txFactory)
	txCtxGetter := trmsqlx.DefaultCtxGetter

	repo := repository.NewPostgresStockRepo(db, txCtxGetter)
	mockProducer := mockKafka.NewMockProducerInterface(ctrl)
	useCase := usecase.NewStockUsecase(repo, txManager, mockProducer)

	cfg := mustLoadConfig(t)

	ctx := context.Background()

	mux, err := server.NewGatewayMux(ctx, cfg, useCase)
	if err != nil {
		t.Fatalf("failed to create gateway mux: %v", err)
	}

	return mux
}
