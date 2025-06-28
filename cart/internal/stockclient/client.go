package stockclient

import (
	"bytes"
	"cart/internal/errors"
	"cart/internal/models"
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	httpTimeout   = 5 * time.Second
	contentType   = "application/json"
	getBySKUEndpt = "/stocks/item/get"
)

type httpStockClient struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string) (*httpStockClient, error) {
	if err := validateURL(baseURL); err != nil {
		return nil, fmt.Errorf("invalid stock service URL: %w", err)
	}

	return &httpStockClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: httpTimeout,
		},
	}, nil
}

func validateURL(rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return stdErrors.New("only http and https URLs are allowed")
	}

	return nil
}

func (c *httpStockClient) GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error) {
	requestBody, err := json.Marshal(map[string]uint32{"sku": sku})
	if err != nil {
		return models.StockItem{}, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+getBySKUEndpt, bytes.NewReader(requestBody))
	if err != nil {
		return models.StockItem{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return models.StockItem{}, fmt.Errorf("failed to fetch SKU from stocks: %w", err)
	}

	defer resp.Body.Close()

	return decodeSKUResponse(resp)
}

func decodeSKUResponse(resp *http.Response) (models.StockItem, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		var item models.StockItem
		if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
			return models.StockItem{}, fmt.Errorf("decode response: %w", err)
		}

		return item, nil

	case http.StatusBadRequest:
		return models.StockItem{}, errors.ErrInvalidSKU

	default:
		return models.StockItem{}, fmt.Errorf("unexpected response from stock service: %s", resp.Status)
	}
}
