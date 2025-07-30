package stockclient

import (
	"cart/internal/log"
	"cart/internal/metrics"
	"cart/internal/models"
	"context"
	"fmt"
	"strconv"
	"time"

	stockpb "cart/pkg/api/stocks"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	defaultTimeout = 5 * time.Second
)

type GRPCClient struct {
	client  stockpb.StockServiceClient
	conn    *grpc.ClientConn
	logger  log.Logger
	metrics *metrics.Metrics
}

func NewGRPCClient(addr string, logger log.Logger, m *metrics.Metrics) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to dial stock service", log.String("address", addr), log.Error(err))
		return nil, fmt.Errorf("invalid stock service URL: %w", err)
	}

	logger.Info("connected to stock service", log.String("address", addr))

	client := stockpb.NewStockServiceClient(conn)
	return &GRPCClient{
		client:  client,
		conn:    conn,
		logger:  logger,
		metrics: m,
	}, nil
}

func (c *GRPCClient) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.logger.Error("failed to close stock client connection", log.Error(err))
			return err
		}
		c.logger.Info("stock client connection closed")
	}
	return nil
}

func (c *GRPCClient) GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error) {
	tracer := otel.Tracer("stockclient")
	ctx, span := tracer.Start(ctx, "GetBySKU")
	defer span.End()

	skuStr := strconv.FormatUint(uint64(sku), 10)

	span.SetAttributes(
		attribute.String("stock.sku", skuStr),
	)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	start := time.Now()
	req := &stockpb.GetItemRequest{
		Sku:      skuStr,
		Location: "default-location",
	}

	resp, err := c.client.GetItem(ctx, req)
	duration := time.Since(start).Seconds()

	if c.metrics != nil {
		c.metrics.RequestsTotal.WithLabelValues("stockclient.GetBySKU", "GRPC").Inc()
		c.metrics.RequestDuration.WithLabelValues("stockclient.GetBySKU", "GRPC").Observe(duration)

		if err != nil {
			c.metrics.RequestErrors.WithLabelValues("stockclient.GetBySKU", "GRPC").Inc()
		}
	}

	if err != nil {
		c.logger.Error("failed to get stock item", log.String("sku", skuStr), log.Error(err))
		return models.StockItem{}, fmt.Errorf("create request: %w", err)
	}

	return models.StockItem{
		SKU:      sku,
		Location: resp.Location,
		Count:    int16(resp.Count),
		Price:    float64(resp.Price),
	}, nil

}
