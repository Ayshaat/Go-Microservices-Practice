APP_NAME := stocks
BIN_DIR := bin
PORT := 8081
DOCKER_IMAGE := ayshaat/stocks:hw7
COMPOSE_FILE := docker-compose.test.yml

.PHONY: build run test clean start-test-env stop-test-env test-integration unit-test docker-build docker-push

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/main.go

run: build
	@echo "Running $(APP_NAME) on port $(PORT)..."
	@./$(BIN_DIR)/$(APP_NAME) -port $(PORT)

test:
	@echo "Running all tests..."
	@go test -v ./...

unit-test:
	@echo "Running unit tests only..."
	@go test -v -tags=unit ./...

start-test-env:
	@echo "Starting test environment..."
	docker-compose -f $(COMPOSE_FILE) up -d --build

stop-test-env:
	@echo "Stopping test environment..."
	docker-compose -f $(COMPOSE_FILE) down

ifeq ($(OS),Windows_NT)
export_env = powershell -Command "$$env:INTEGRATION_TEST='1'; go test -v ./tests"
sleep_cmd = timeout /t 10 /nobreak >nul
else
export_env = INTEGRATION_TEST=1 go test -v ./tests
sleep_cmd = sleep 10
endif

test-integration: start-test-env
	@echo "Waiting for services to be healthy..."
	@$(sleep_cmd)
	@echo "Running integration tests..."
	@$(export_env)
	$(MAKE) stop-test-env

clean:
	@echo "Cleaning $(APP_NAME)..."
	@rm -rf $(BIN_DIR)
	@go clean

docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .

docker-push:
	@echo "Pushing Docker image $(DOCKER_IMAGE)..."
	docker push $(DOCKER_IMAGE)

MIGRATIONS_DIR=internal/db/migrations

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(GOOSE_DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(GOOSE_DB_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(GOOSE_DB_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql
