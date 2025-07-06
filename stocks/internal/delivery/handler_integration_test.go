package delivery_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func setupTestDB(t *testing.T) *sqlx.DB {
	cfg, err := config.Load()
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
func TestIntegration_AddItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	server := setupServer(t, db)

	validPayload := map[string]interface{}{
		"userID":   1,
		"sku":      1001,
		"price":    20.5,
		"count":    10,
		"location": "loc1",
	}

	body, err := json.Marshal(validPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/add", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/add", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stocks/item/add", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestIntegration_DeleteItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	server := setupServer(t, db)

	addPayload := map[string]interface{}{
		"userID":   1,
		"sku":      1001,
		"price":    15.5,
		"count":    5,
		"location": "loc1",
	}

	addBody, err := json.Marshal(addPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	addReq := httptest.NewRequest(http.MethodPost, "/stocks/item/add", bytes.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addRec := httptest.NewRecorder()
	server.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusCreated {
		t.Fatalf("setup add failed with status %d", addRec.Code)
	}

	delPayload := map[string]interface{}{
		"userID": 1,
		"sku":    1001,
	}

	delBody, err := json.Marshal(delPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/delete", bytes.NewReader(delBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/delete", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stocks/item/delete", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestIntegration_GetItem(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	server := setupServer(t, db)

	addPayload := map[string]interface{}{
		"userID":   1,
		"sku":      1001,
		"price":    15.5,
		"count":    5,
		"location": "loc1",
	}

	addBody, err := json.Marshal(addPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	addReq := httptest.NewRequest(http.MethodPost, "/stocks/item/add", bytes.NewReader(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addRec := httptest.NewRecorder()
	server.ServeHTTP(addRec, addReq)

	if addRec.Code != http.StatusCreated {
		t.Fatalf("setup add failed with status %d", addRec.Code)
	}

	getPayload := map[string]interface{}{
		"sku": 1001,
	}

	getBody, err := json.Marshal(getPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/get", bytes.NewReader(getBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/get", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		notFoundPayload := map[string]interface{}{
			"sku": 9999,
		}

		notFoundBody, err := json.Marshal(notFoundPayload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/get", bytes.NewReader(notFoundBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stocks/item/get", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestIntegration_ListByLocation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO sku_info (sku, name, type) VALUES
		(1001, 't-shirt', 'apparel'),
		(1002, 'jeans', 'apparel')
		ON CONFLICT (sku) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert sku_info for list test: %v", err)
	}

	server := setupServer(t, db)

	items := []map[string]interface{}{
		{"userID": 1, "sku": 1001, "price": 15.5, "count": 5, "location": "loc1"},
		{"userID": 2, "sku": 1002, "price": 10.0, "count": 3, "location": "loc1"},
	}
	for _, item := range items {
		body, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/stocks/item/add", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("setup add failed with status %d", rec.Code)
		}
	}

	listPayload := map[string]interface{}{
		"location":    "loc1",
		"pageSize":    10,
		"currentPage": 1,
	}

	listBody, err := json.Marshal(listPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/list/location", bytes.NewReader(listBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stocks/list/location", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stocks/list/location", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}
