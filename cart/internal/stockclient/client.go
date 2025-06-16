package stockclient

import (
	"cart/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const httpTimeout = 5 * time.Second

type httpStockClient struct {
	baseURL string
}

func New(baseURL string) StockRepository {
	return &httpStockClient{baseURL: baseURL}
}

func validateURL(rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http and https URLs are allowed")
	}

	return nil
}

func (c *httpStockClient) GetBySKU(sku uint32) (models.StockItem, error) {
	url := fmt.Sprintf("%s/stocks/item/get?sku=%d", c.baseURL, sku)

	if err := validateURL(url); err != nil {
		return models.StockItem{}, fmt.Errorf("invalid URL: %w", err)
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return models.StockItem{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return models.StockItem{}, fmt.Errorf("failed to fetch SKU from stocks: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.StockItem{}, fmt.Errorf("stocks service returned status %d", resp.StatusCode)
	}

	var item models.StockItem

	err = json.NewDecoder(resp.Body).Decode(&item)
	if err != nil {
		return models.StockItem{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return item, nil
}
