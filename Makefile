.PHONY: run test migrate-up migrate-down install-tools tidy fmt dev-up dev-down dev-vpc dev-vpc-down seed db-clean db-refresh

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

dev-up:
	@chmod +x scripts/dev-up.sh scripts/dev-down.sh
	@./scripts/dev-up.sh

dev-down:
	@chmod +x scripts/dev-up.sh scripts/dev-down.sh
	@./scripts/dev-down.sh

dev-vpc:
	@chmod +x scripts/dev-vpc.sh scripts/dev-vpc-down.sh
	@./scripts/dev-vpc.sh

dev-vpc-down:
	@chmod +x scripts/dev-vpc-down.sh
	@./scripts/dev-vpc-down.sh

seed:
	go run ./cmd/seed

db-clean:
	go run ./cmd/seed -clean-only

db-refresh:
	go run ./cmd/seed
