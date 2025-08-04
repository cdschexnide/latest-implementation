# Proto generation variables
PROTO_DIR := proto
GENERATED_DIR := generated/proto
PROTO_FILES := $(PROTO_DIR)/blade_ingestion.proto

.PHONY: all proto build run test clean help

# Default target
all: proto build

# Help target
help:
	@echo "Available targets:"
	@echo "  make proto    - Generate protobuf files"
	@echo "  make build    - Build the server binary"
	@echo "  make run      - Run the server"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Clean generated files"
	@echo "  make docker   - Build Docker image"
	@echo "  make deps     - Install dependencies"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Generate proto files (we'll implement this later)
proto:
	@echo "Generating proto files..."
	@mkdir -p $(GENERATED_DIR)
	protoc -I $(PROTO_DIR) \
		-I $(PROTO_DIR)/google/api \
		-I $(PROTO_DIR)/google/protobuf \
		-I $(PROTO_DIR)/protoc-gen-openapiv2/options \
		--go_out=$(GENERATED_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GENERATED_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GENERATED_DIR) --grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--openapiv2_out=swagger --openapiv2_opt=allow_merge=true \
		$(PROTO_FILES)
	@echo "Proto generation complete!"

# Build the server
build:
	@echo "Building server..."
	go build -o bin/blade-server ./server/main.go

# Run the server
run: build
	@echo "Starting server..."
	./bin/blade-server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean generated files
proto-clean:
	@echo "Cleaning generated proto files..."
	rm -f $(GENERATED_DIR)/*.pb.go
	rm -f $(GENERATED_DIR)/*.pb.gw.go
	rm -f swagger/*.swagger.json

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t blade-ingestion-service .

# Development mode with hot reload
dev:
	@echo "Starting in development mode..."
	go run ./server/main.go