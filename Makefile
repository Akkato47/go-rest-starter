include .env
export

dev:
	@go run cmd/api/*.go

migrate-up:
	@migrate -path ./migrations -database "${POSTGRES_URL}?sslmode=disable" up

migrate-down:
	@migrate -path ./migrations -database "${POSTGRES_URL}?sslmode=disable" down