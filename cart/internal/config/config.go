package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	StockServiceURL string
	GRPCPort        string
	HTTPPort        string
	GRPCEndpoint    string
	JaegerEndpoint  string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

func Load(envFile string) (*Config, error) {
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	fmt.Printf("Loaded %s successfully\n", envFile)

	cfg := &Config{
		DBHost:          os.Getenv("DB_HOST"),
		DBPort:          os.Getenv("DB_PORT"),
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		StockServiceURL: os.Getenv("STOCK_SERVICE_URL"),
		GRPCPort:        os.Getenv("GRPC_PORT"),
		HTTPPort:        os.Getenv("HTTP_PORT"),
		GRPCEndpoint:    os.Getenv("GRPC_ENDPOINT"),
		JaegerEndpoint:  os.Getenv("JAEGER_ENDPOINT"),
		ReadTimeout:     ReadTimeout,
		WriteTimeout:    WriteTimeout,
		IdleTimeout:     IdleTimeout,
	}

	log.Println(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return cfg, nil
}

func (c *Config) PostgresConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
