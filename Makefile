.PHONY: swagger run build tidy

## Generate Swagger docs
swagger:
	swag init -g cmd/server/main.go -o docs

## Run the development server
run:
	go run ./cmd/server/main.go

## Build the binary
build:
	go build -o bin/server ./cmd/server/main.go

## Tidy go modules
tidy:
	go mod tidy

## Generate swagger + run
dev: swagger run
