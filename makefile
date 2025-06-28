BINARY_NAME=toy-redis
BUILD_DIR=build
GOPATH=$(shell go env GOPATH)
COMPILEDAEMON=$(GOPATH)/bin/CompileDaemon

# Build the application
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) server/main.go

# Run the application
run:
	@go run server/main.go

# Clean build files
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Install CompileDaemon for hot reloading
install-daemon:
	@echo "Installing CompileDaemon..."
	@go install github.com/githubnemo/CompileDaemon@latest

# Run with hot reloading using CompileDaemon
dev: install-daemon
	@$(COMPILEDAEMON) --build="go build -o $(BUILD_DIR)/$(BINARY_NAME) server/main.go" --command="./$(BUILD_DIR)/$(BINARY_NAME)" --color=true -pattern="(.+\.go|.+\.env)$$"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...
