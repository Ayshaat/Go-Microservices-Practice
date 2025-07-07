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

	server := setupServer(t, db)

	defer db.Close()

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:   "success",
			method: http.MethodPost,
			payload: map[string]interface{}{
				"userID":   1,
				"sku":      1001,
				"price":    20.5,
				"count":    10,
				"location": "loc1",
			},
			expectedStatus: http.StatusCreated,
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

			req := httptest.NewRequest(tt.method, "/stocks/item/add", bytes.NewReader(body))
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
			name:   "success",
			method: http.MethodPost,
			payload: map[string]interface{}{
				"userID": 1,
				"sku":    1001,
			},
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
			req := httptest.NewRequest(tt.method, "/stocks/item/delete", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestIntegration_GetItem(t *testing.T) {
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
			name:   "success",
			method: http.MethodPost,
			payload: map[string]interface{}{
				"sku": 1001,
			},
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
			name:   "not found",
			method: http.MethodPost,
			payload: map[string]interface{}{
				"sku": 9999,
			},
			expectedStatus: http.StatusNotFound,
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

			req := httptest.NewRequest(tt.method, "/stocks/item/get", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestIntegration_ListByLocation(t *testing.T) {
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
			name:   "success",
			method: http.MethodPost,
			payload: map[string]interface{}{
				"location":    "loc1",
				"pageSize":    10,
				"currentPage": 1,
			},
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
				_, err := db.Exec(`
					INSERT INTO sku_info (sku, name, type) VALUES
					(1001, 't-shirt', 'apparel'),
					(1002, 'jeans', 'apparel')
					ON CONFLICT (sku) DO NOTHING
				`)
				if err != nil {
					t.Fatalf("failed to insert sku_info for list test: %v", err)
				}

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

			req := httptest.NewRequest(tt.method, "/stocks/list/location", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d, body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
