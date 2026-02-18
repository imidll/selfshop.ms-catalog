ifneq (,$(wildcard ./.env))
    include ./.env
    TO_EXPORT := $(filter KO_%,$(.VARIABLES))
    export $(TO_EXPORT)
endif


APP_NAME         := selfshop.ms-catalog
CMD_PATH         := ./cmd/$(APP_NAME)
CMD_COMPOSE_PROD := docker compose --project-directory . -f deploy/compose.yml
CMD_COMPOSE_DEV  := docker compose --project-directory . -f deploy/compose.hot-reload.yml

VERSION    ?= $(shell git describe --tags --exact-match 2>/dev/null || echo dev)
COMHASH    ?= $(shell git rev-parse HEAD 2>/dev/null || echo "undefined")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "undefined")
GO_VERSION ?= $(shell go mod edit -json | jq -r '.Go' | cut -d'.' -f1,2)

.PHONY: help
help: ## Показать справку
	@awk 'BEGIN { in_section=0 } \
		/^# ===/ { \
			gsub(/# === | ===/, ""); \
			printf "\n\033[1m%s:\033[0m\n", $$0; \
			in_section=1 \
		} \
		/^[a-zA-Z_-]+:.*##/ { \
			if (in_section) { \
				split($$0, parts, ":.*?## "); \
				printf "  \033[36m%-20s\033[0m %s\n", parts[1], parts[2] \
			} \
		}' $(MAKEFILE_LIST)
	@echo ""

.DEFAULT_GOAL := help

# === Production ===

.PHONY: prod
prod: prod-build prod-up ## Собрать и запустить prod окружение

.PHONY: prod-build
prod-build: ## Собрать prod образ через ko
	ko build $(CMD_PATH) -B -L

.PHONY: prod-up
prod-up: ## Запустить prod окружение
	$(CMD_COMPOSE_PROD) up -d

.PHONY: prod-down
prod-down: ## Остановить prod окружение
	$(CMD_COMPOSE_PROD) down

.PHONY: prod-logs
prod-logs: ## Показать логи prod окружения
	$(CMD_COMPOSE_PROD) logs -f

.PHONY: prod-restart
prod-restart: prod-down prod ## Перезапустить prod окружение

# === Development (hot-reload) ===

.PHONY: dev
dev: dev-build dev-up ## Собрать и запустить dev окружение

.PHONY: dev-build
dev-build: ## Пересобрать dev образ
	$(CMD_COMPOSE_DEV) build --build-arg GO_VERSION=$(GO_VERSION)

.PHONY: dev-up
dev-up: ## Запустить dev окружение (hot-reload)
	$(CMD_COMPOSE_DEV) up -d

.PHONY: dev-down
dev-down: ## Остановить dev окружение
	$(CMD_COMPOSE_DEV) down

.PHONY: dev-logs
dev-logs: ## Показать логи dev окружения
	$(CMD_COMPOSE_DEV) logs -f

.PHONY: dev-restart
dev-restart: dev-down dev ## Перезапустить dev окружение

# === CI/CD ===

.PHONY: lint
lint: ## Запустить линтер
	golangci-lint run --verbose

.PHONY: test
test: ## Запустить тесты с покрытием
	go test -race -coverprofile=coverage.out -covermode=atomic ./internal/...
	sed -i '/_mock\.go:/d' coverage.out || true
	go tool cover -func=coverage.out > coverage.humanize
	go tool cover -func=coverage.out | tail -1
				 
.PHONY: ci
ci: lint test ## Полный CI pipeline локально
	@echo "✅ CI pipeline completed"

# === Утилиты ===

.PHONY: clean
clean: ## Очистить артефакты
	rm -rf bin/ tmp/ sbom/ coverage.*
	$(CMD_COMPOSE_PROD) down -v 2>/dev/null || true
	$(CMD_COMPOSE_DEV)  down -v 2>/dev/null || true

.PHONY: down
down: ## Остановить все окружения
	$(CMD_COMPOSE_PROD) down 2>/dev/null || true
	$(CMD_COMPOSE_DEV)  down 2>/dev/null || true

.PHONY: version
version: ## Показать версию
	@echo "Version:    $(VERSION)"
	@echo "Comhash:    $(COMHASH)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"