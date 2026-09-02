BINARY_NAME=hydravault
BUILD_DIR=bin

.PHONY: all build run test clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/hydravault

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v -race ./...

clean:
	rm -rf $(BUILD_DIR)
