.PHONY: build run dev prod lint clean deps docker-up docker-down docker-down-volumes docker-ps docker-logs docker-build

APP_NAME = arlo-admin
BUILD_DIR = ./build
MAIN_PATH = ./cmd/server/main.go
COMPOSE = docker compose -f ../deployments/docker/docker-compose.yml --env-file ../deployments/docker/.env

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	GOPROXY=https://goproxy.cn,direct go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

dev:
	@echo "Starting in development mode..."
	go run $(MAIN_PATH)

prod:
	@echo "Starting in production mode..."
	APP_ENV=prod $(BUILD_DIR)/$(APP_NAME)

lint:
	@echo "Running linter..."
	go vet ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf logs/

deps:
	@echo "Downloading dependencies..."
	GOPROXY=https://goproxy.cn,direct go mod download
	GOPROXY=https://goproxy.cn,direct go mod tidy

docker-up:
	@test -f ../deployments/docker/.env || (echo "Missing deployments/docker/.env — copy from .env.example first" && exit 1)
	$(COMPOSE) up -d --build

docker-build:
	@test -f ../deployments/docker/.env || (echo "Missing deployments/docker/.env — copy from .env.example first" && exit 1)
	$(COMPOSE) build

docker-down:
	$(COMPOSE) down

docker-down-volumes:
	$(COMPOSE) down -v

docker-ps:
	$(COMPOSE) ps

docker-logs:
	$(COMPOSE) logs -f --tail=200
