SHELL := /bin/bash

.PHONY: all build test deps deps-cleancache migrate-up migrate-down migrate-create migrate-version migrate-force seed

GOCMD=go
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATE_CMD=migrate -path pkg/db/migrations -database "$(DATABASE_URL)"
BUILD_DIR=build
BINARY_DIR=$(BUILD_DIR)/bin
CODE_COVERAGE=code-coverage

all: test build

${BINARY_DIR}:
	mkdir -p $(BINARY_DIR)

build: ${BINARY_DIR} ## Compile the code, build Executable File
	$(GOCMD) build -o $(BINARY_DIR) -v ./cmd/api
# 	GOARCH=amd64 $(GOCMD) build -o $(BINARY_DIR)/api-linux-amd64 -v ./cmd/api/main.go

build-run: build ## run project build file if not exist build it
#	./$(BINARY_DIR)/api-linux-amd64
	./$(BINARY_DIR)/api

run: ## Start application
	$(GOCMD) run ./cmd/api/main.go

seed: ## Seed platform admin accounts (idempotent — skips existing emails)
	$(GOCMD) run ./cmd/seed/main.go

run-no-lint: ## Start application without lint checks
	$(GOCMD) run ./cmd/api/main.go

test: ## Run tests
	$(GOCMD) test ./... -cover

test-coverage: ## Run tests and generate coverage file
	$(GOCMD) test ./... -coverprofile=$(CODE_COVERAGE).out
	$(GOCMD) tool cover -html=$(CODE_COVERAGE).out

deps: ## Install dependencies
#	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
#	$(GOCMD) get -u -t -d -v ./...
	$(GOCMD) mod tidy
#	$(GOCMD) mod vendor

deps-cleancache: ## Clear cache in Go module
	$(GOCMD) clean -modcache

wire: ## Generate wire_gen.go
	cd pkg/di && wire

swagger: ## install swagger and its dependencies for generate swagger using swag
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest 
	$(GOCMD) get -u github.com/swaggo/swag/cmd/swag 
	$(GOCMD) get -u github.com/swaggo/gin-swagger 
	$(GOCMD) get -u github.com/swaggo/files

swag: ## Generate swagger docs
	swag init -g pkg/api/server.go -o ./cmd/api/docs

check: ## To check the code standard violations and errors
	golangci-lint run

mockgen: # Generate mock files for the test
	mockgen -source=pkg/repository/interfaces/auth.go -destination=pkg/mock/mockrepo/auth_mock.go -package=mockrepo
	mockgen -source=pkg/repository/interfaces/user.go -destination=pkg/mock/mockrepo/user_mock.go -package=mockrepo
	mockgen -source=pkg/repository/interfaces/platform_user.go -destination=pkg/mock/mockrepo/platform_user_mock.go -package=mockrepo
	mockgen -source=pkg/service/token/token.go -destination=pkg/mock/mockservice/token_mock.go -package=mockservice
	mockgen -source=pkg/usecase/interfaces/auth.go -destination=pkg/mock/mockusecase/auth_mock.go -package=mockusecase
	mockgen -source=pkg/usecase/interfaces/platform_user.go -destination=pkg/mock/mockusecase/platform_user_mock.go -package=mockusecase

docker-up: ## To up the docker compose file
	docker-compose up 

docker-down: ## To down the docker compose file
	docker-compose down

docker-build: ## To build newdocker file for this project
	docker build -t rohit221990/mandi . 

migrate-up: ## Apply all up migrations
	$(MIGRATE_CMD) up

migrate-down: ## Roll back one migration
	$(MIGRATE_CMD) down 1

migrate-create: ## Create a new migration: make migrate-create name=foo
	migrate create -ext sql -dir pkg/db/migrations -seq $(name)

migrate-version: ## Print current migration version
	$(MIGRATE_CMD) version

migrate-force: ## Force a version: make migrate-force version=N
	$(MIGRATE_CMD) force $(version)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'