package tests

import (
	"cart/internal/config"
	"cart/internal/delivery"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"cart/tests/mock"
	"database/sql"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

type noopProducer struct{}

func (n *noopProducer) SendCartItemAdded(cartId, sku string, count int, status string) error {
	return nil
}

func (n *noopProducer) SendCartItemFailed(cartId, sku string, count int, status, reason string) error {
	return nil
}

func (n *noopProducer) Close() error {
	return nil
}

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TEST not set")
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TEST not set")
	}

	cfg, err := config.Load("../.env.local")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.PostgresConnStr())
	if err != nil {
		t.Fatalf("failed to open db connection: %v", err)
	}

	_, err = db.Exec("DELETE FROM cart_items")
	if err != nil {
		t.Fatalf("failed to clear cart_items table: %v", err)
	}

	return db
}

func setupServer(t *testing.T, db *sql.DB) http.Handler {
	cartRepo := repository.NewPostgresCartRepo(db)

	mockStock := mock.StartMockStockServer()
	t.Cleanup(mockStock.Close)

	stockCli, err := stockclient.New(mockStock.URL)
	if err != nil {
		t.Fatalf("failed to create stock client: %v", err)
	}

	producer := &noopProducer{}

	cartUsecase := usecase.NewCartUsecase(cartRepo, stockCli, producer)
	handler := delivery.NewHandler(cartUsecase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return mux
}
