.PHONY: run test migrate-up migrate-down install-tools tidy fmt

MIGRATE := $(shell go env GOPATH)/bin/migrate

ifneq (,$(wildcard .env))
include .env
export
endif

run:
	go run ./cmd/api

TEST_PKGS := $(shell go list ./... | grep -v node_modules)

test:
	go test $(TEST_PKGS)

install-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up: $(MIGRATE)
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down: $(MIGRATE)
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

$(MIGRATE):
	$(MAKE) install-tools

tidy:
	go mod tidy

fmt:
	gofmt -w .
