package tests

import (
	"net/http"
	"os"
	"testing"

	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/delivery"
	"stocks/internal/repository"
	"stocks/internal/usecase"

	"github.com/jmoiron/sqlx"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	_ "github.com/lib/pq"
)

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TEST not set")
	}
}

func setupTestDB(t *testing.T) *sqlx.DB {
	var envFile string
	if os.Getenv("INTEGRATION_TEST") == "1" {
		envFile = ".env.docker"
	} else {
		envFile = ".env.local"
	}
	cfg, err := config.Load(envFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	sqlDB, err := db.ConnectDB(cfg.PostgresConnStr())
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	dbx := sqlx.NewDb(sqlDB, "postgres")

	if err := db.RunMigrations(dbx.DB); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

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

func setupServer(_ *testing.T, db *sqlx.DB) http.Handler {
	txFactory := trmsqlx.NewDefaultFactory(db.DB)
	txManager := manager.Must(txFactory)
	txCtxGetter := trmsqlx.DefaultCtxGetter

	repo := repository.NewPostgresStockRepo(db, txCtxGetter)
	useCase := usecase.NewStockUsecase(repo, txManager)
	handler := delivery.NewHandler(useCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return mux
}
