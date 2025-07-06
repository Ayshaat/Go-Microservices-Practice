package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestIntegration_AddItem(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	t.Run("success", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id": 1,
			"sku":     1001,
			"count":   2,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/cart/item/add", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/cart/item/add", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cart/item/add", nil)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestIntegration_ClearCart(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	t.Run("success", func(t *testing.T) {
		payload := map[string]interface{}{"user_id": 1}

		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/cart/clear", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/cart/clear", bytes.NewReader([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cart/clear", nil)

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestIntegration_DeleteItem(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	t.Run("success", func(t *testing.T) {
		addPayload := map[string]interface{}{
			"user_id": 1,
			"sku":     1001,
			"count":   2,
		}

		bodyAdd, err := json.Marshal(addPayload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		reqAdd := httptest.NewRequest(http.MethodPost, "/cart/item/add", bytes.NewReader(bodyAdd))
		reqAdd.Header.Set("Content-Type", "application/json")
		recAdd := httptest.NewRecorder()
		server.ServeHTTP(recAdd, reqAdd)

		if recAdd.Code != http.StatusOK {
			t.Fatalf("setup add failed with status %d, body: %s", recAdd.Code, recAdd.Body.String())
		}

		delPayload := map[string]interface{}{
			"user_id": 1,
			"sku":     1001,
		}

		bodyDel, err := json.Marshal(delPayload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		reqDel := httptest.NewRequest(http.MethodPost, "/cart/item/delete", bytes.NewReader(bodyDel))
		reqDel.Header.Set("Content-Type", "application/json")
		recDel := httptest.NewRecorder()
		server.ServeHTTP(recDel, reqDel)

		if recDel.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d, body: %s", recDel.Code, recDel.Body.String())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/cart/item/delete", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cart/item/delete", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestIntegration_ListItems(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	t.Run("success", func(t *testing.T) {
		payload := map[string]interface{}{
			"userID": 1,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/cart/list", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing userID", func(t *testing.T) {
		payload := map[string]interface{}{}

		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/cart/list", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cart/list", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}
