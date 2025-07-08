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

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			payload:        map[string]interface{}{"user_id": 1, "sku": 1001, "count": 2},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			payload:        "{invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			case nil:
				body = nil
			default:
				var err error

				body, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/cart/item/add", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestIntegration_ClearCart(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			payload:        map[string]interface{}{"user_id": 1},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			payload:        "{invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			case nil:
				body = nil
			default:
				var err error

				body, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/cart/clear", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestIntegration_DeleteItem(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
		doSetupAdd     bool
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			payload:        map[string]interface{}{"user_id": 1, "sku": 1001},
			expectedStatus: http.StatusOK,
			doSetupAdd:     true,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			payload:        "{invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doSetupAdd {
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
			}

			var body []byte
			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			case nil:
				body = nil
			default:
				var err error

				body, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}
			reqDel := httptest.NewRequest(tt.method, "/cart/item/delete", bytes.NewReader(body))
			reqDel.Header.Set("Content-Type", "application/json")
			recDel := httptest.NewRecorder()
			server.ServeHTTP(recDel, reqDel)

			if recDel.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, recDel.Code, recDel.Body.String())
			}
		})
	}
}

func TestIntegration_ListItems(t *testing.T) {
	skipIfNotIntegration(t)

	db := setupTestDB(t)
	defer db.Close()
	server := setupServer(t, db)

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			payload:        map[string]interface{}{"userID": 1}, // fix key to "user_id" to be consistent
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id",
			method:         http.MethodPost,
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			case nil:
				body = nil
			default:
				var err error

				body, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/cart/list", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
