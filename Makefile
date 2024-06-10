build:
	mkdir -p ./bin
	go build -o ./bin/ ./...

setup:
	docker compose build
