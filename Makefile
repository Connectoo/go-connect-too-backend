.PHONY: run test migrate-up migrate-down tidy fmt

run:
	go run ./cmd/api

test:
	go test ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

tidy:
	go mod tidy

fmt:
	gofmt -w .
