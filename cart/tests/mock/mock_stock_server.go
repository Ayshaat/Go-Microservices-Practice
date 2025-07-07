package mock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

const (
	testSKU   = 1001
	testPrice = 20.5
	testCount = 100
)

func StartMockStockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/stocks/item/get", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SKU uint32 `json:"sku"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.SKU == testSKU {
			resp := map[string]interface{}{
				"sku":   testSKU,
				"name":  "t-shirt",
				"type":  "clothing",
				"price": testPrice,
				"count": testCount,
			}

			w.WriteHeader(http.StatusOK)

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
				return
			}

			return
		}

		http.Error(w, "SKU not found", http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}
