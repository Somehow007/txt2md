BINARY_NAME=txt2md
VERSION?=dev
BUILD_DIR=build

.PHONY: build test lint clean cross-build install

build:
	go build -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/txt2md

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

cross-build:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/txt2md
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/txt2md
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/txt2md
	GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/txt2md
	GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/txt2md

install:
	go install -ldflags "-X github.com/Somehow007/txt2md/cmd/txt2md/commands.version=$(VERSION)" ./cmd/txt2md
