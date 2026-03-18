include .env
export

dev:
	@go run cmd/api/*.go

migrate-up:
	@migrate -path ./migrations -database "${POSTGRES_URL}?sslmode=disable" up

migrate-down:
	@migrate -path ./migrations -database "${POSTGRES_URL}?sslmode=disable" down

swagger:
	@swag init -g cmd/api/main.go -o docs