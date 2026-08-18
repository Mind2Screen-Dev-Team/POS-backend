APP_NAME := pos-backend
BUILD_DIR := bin
MAIN_PATH := ./cmd/server

.PHONY: build run test vet fmt clean lint verify

verify: fmt vet test

build:
	go build -o $(BUILD_DIR)/server $(MAIN_PATH)

run: build
	./$(BUILD_DIR)/server

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

fmt:
	@test -z "$$(gofmt -l .)" || { gofmt -l . ; echo "files above need gofmt"; exit 1; }

clean:
	rm -rf $(BUILD_DIR)

lint: fmt vet test
	@echo "All checks passed."
