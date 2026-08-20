# ==============================================================================
# Go Start Project — Makefile
# ==============================================================================
# Uso: `make <target>`. Rode `make help` para ver todos os comandos.
# ==============================================================================

# ---- Configuration -----------------------------------------------------------
# Carrega variáveis do .env (se existir) para dentro do Makefile.
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# Nome do binário e caminho do entrypoint
BINARY_NAME := api
MAIN_PATH   := ./cmd/api
MIGRATIONS  := ./migrations

# String de conexão do banco montada a partir do .env
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# .PHONY: declara alvos que NÃO são arquivos (evita conflito com nomes de arquivo)
.PHONY: help run build clean tidy fmt vet test \
        docker-up docker-down docker-logs \
        migrate-up migrate-down migrate-create migrate-force \
        swag

# ---- Default -----------------------------------------------------------------
# `make` sem argumento mostra a ajuda.
.DEFAULT_GOAL := help

help: ## Mostra esta ajuda
	@echo "Go Start Project — comandos disponiveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ---- Application -------------------------------------------------------------
run: ## Sobe a aplicacao (go run)
	go run $(MAIN_PATH)

build: ## Compila o binario em ./bin
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

clean: ## Remove artefatos de build
	rm -rf bin/
	go clean

# ---- Quality -----------------------------------------------------------------
tidy: ## Organiza as dependencias (go mod tidy)
	go mod tidy

fmt: ## Formata o codigo (gofmt)
	go fmt ./...

vet: ## Analise estatica (go vet)
	go vet ./...

test: ## Roda os testes
	go test -v ./...

check: fmt vet test ## Roda fmt + vet + test em sequencia

# ---- Docker ------------------------------------------------------------------
docker-up: ## Sobe o PostgreSQL via docker-compose
	docker-compose up -d postgres

docker-down: ## Derruba os containers
	docker-compose down

docker-logs: ## Acompanha os logs do PostgreSQL
	docker-compose logs -f postgres

# ---- Migrations (golang-migrate) ---------------------------------------------
migrate-up: ## Aplica todas as migrations pendentes
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

migrate-down: ## Reverte a ultima migration
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down 1

migrate-create: ## Cria uma nova migration. Uso: make migrate-create name=create_x_table
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

migrate-force: ## Forca uma versao (destrava migration suja). Uso: make migrate-force version=3
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" force $(version)

# ---- Docs --------------------------------------------------------------------
swag: ## Regenera a documentacao Swagger
	swag init -g $(MAIN_PATH)/main.go
