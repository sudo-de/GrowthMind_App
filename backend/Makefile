BIN := bin/api

run:
	go run ./cmd/api

build:
	go build -o $(BIN) ./cmd/api

clean:
	rm -rf bin/

stop:
	@pkill -f "go run ./cmd/api" 2>/dev/null || true
	@pkill -f "$(BIN)" 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@echo "stopped"

fmt:
	gofmt -w ./...

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not installed: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

test:
	go test ./...

test-verbose:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "coverage report: coverage.html"

tidy:
	go mod tidy

.PHONY: run build clean stop fmt vet lint test test-verbose coverage tidy