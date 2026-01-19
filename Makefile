BINARY_NAME=crawler
MAIN_PATH=cmd/crawler/main.go

.PHONY: all build run test clean docker-up docker-down

all: build

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
