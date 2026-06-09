.PHONY: build run dev test test-integration test-e2e lint fmt clean

# Binary name
BINARY ?= razad-daemon
BUILD_DIR ?= build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go flags
GO ?= go
GOFLAGS ?= -ldflags="-X main.Version=$(VERSION)"
CGO_ENABLED ?= 1
MODULE = github.com/razad/razad

# Build the daemon binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/razad-daemon/

# Run the daemon locally
run: build
	RAZAD_DEBUG=true RAZAD_PORT=8080 RAZAD_DATA_DIR=./tmp/razad \
	./$(BUILD_DIR)/$(BINARY)

# Development with frontend proxy (start frontend dev server separately)
dev:
	@echo "Run the frontend dev server in another terminal:"
	@echo "  cd web && npm run dev -- --port 5173"
	@echo ""
	RAZAD_DEBUG=true RAZAD_PORT=8080 RAZAD_DATA_DIR=./tmp/razad \
	$(GO) run $(GOFLAGS) ./cmd/razad-daemon/

# Tests (use module prefix to avoid node_modules confusion)
test:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(MODULE)/... -count=1 -short

test-integration:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(MODULE)/... -count=1 -run Integration

test-e2e:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(MODULE)/tests/e2e/... -count=1

# Linting
lint:
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) tmp/

# Install dependencies
deps:
	$(GO) mod tidy
	$(GO) mod download

# Database
migrate-up:
	$(GO) run ./cmd/razad-daemon/ -migrate

migrate-down:
	@echo "Not implemented yet"

# Frontend setup
web-setup:
	cd web && npm install

web-build:
	cd web && npm run build

web-dev:
	cd web && npm run dev

# Full build with embedded frontend
release: web-build
	$(GO) build $(GOFLAGS) -tags release -o $(BUILD_DIR)/$(BINARY) ./cmd/razad-daemon/
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY)"

# Help
help:
	@echo "Targets:"
	@echo "  build          Build the daemon binary"
	@echo "  run            Build and run locally"
	@echo "  dev            Run in development mode"
	@echo "  test           Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  lint           Run golangci-lint"
	@echo "  fmt            Format and vet code"
	@echo "  clean          Remove build artifacts"
	@echo "  deps           Download Go dependencies"
	@echo "  web-setup      Install frontend dependencies"
	@echo "  web-build      Build frontend static assets"
	@echo "  release        Build frontend + daemon binary"
