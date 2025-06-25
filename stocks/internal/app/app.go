package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/delivery"
	"stocks/internal/repository"
	"stocks/internal/usecase"

	"github.com/jmoiron/sqlx"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.ConnectDB(cfg.PostgresConnStr())
	if err != nil {
		log.Fatalf("failed to connect and migrate db: %v", err)
	}
	defer database.Close()

	dbx := sqlx.NewDb(database, "postgres")

	defer dbx.Close()

	txFactory := trmsqlx.NewDefaultFactory(dbx.DB)
	txManager := manager.Must(txFactory)
	txCtxGetter := trmsqlx.DefaultCtxGetter

	repo := repository.NewPostgresStockRepo(dbx, txCtxGetter)
	useCase := usecase.NewStockUsecase(repo, txManager)
	handler := delivery.NewHandler(useCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting server on port", port)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("stocks server failed: %v", err)
		}
	}()

	<-stop

	log.Println("Shutting down stocks server...")

	ctx, cancel := context.WithTimeout(context.Background(), config.WriteTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Stocks Server Shutdown Failed: %v", err)
	}

	log.Println("Stocks server gracefully stopped")

	return nil
}
