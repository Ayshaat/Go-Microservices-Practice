package tests

import (
	"cart/internal/config"
	"cart/internal/server"
	"cart/tests/mock"
	"context"
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

func mustLoadConfig(t *testing.T) *config.Config {
	cfg, err := config.Load("../.env.local")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	return cfg
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

	cfg := mustLoadConfig(t)

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

	mockStock := mock.StartMockStockServer()
	t.Cleanup(mockStock.Close)

	cfg := mustLoadConfig(t)

	ctx := context.Background()

	handler, err := server.NewGatewayMux(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to build gateway mux: %v", err)
	}

	return handler
}
