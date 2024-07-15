build:
	mkdir -p ./bin
	go build -o ./bin/ ./...

setup:
	docker compose up -d --build

lint:
	golangci-lint run ./...
