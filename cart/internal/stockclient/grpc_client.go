package stockclient

import (
	"cart/internal/models"
	"context"
	"fmt"
	"strconv"
	"time"

	stockpb "cart/pkg/api/stocks"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout = 5 * time.Second
)

type GRPCClient struct {
	client stockpb.StockServiceClient
	conn   *grpc.ClientConn
}

func NewGRPCClient(addr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("invalid stock service URL: %w", err)
	}

	client := stockpb.NewStockServiceClient(conn)
	return &GRPCClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *GRPCClient) GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	skuStr := strconv.FormatUint(uint64(sku), 10)

	req := &stockpb.GetItemRequest{
		Sku:      skuStr,
		Location: "default-location",
	}

	resp, err := c.client.GetItem(ctx, req)
	if err != nil {
		return models.StockItem{}, fmt.Errorf("create request: %w", err)
	}

	return models.StockItem{
		SKU:      sku,
		Location: resp.Location,
		Count:    int16(resp.Count),
		Price:    float64(resp.Price),
	}, nil

}
