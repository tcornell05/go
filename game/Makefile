# Makefile for Isometric Game Development

.PHONY: all run build test clean dev watch check validate assets help

# Default target
all: build

# Development run with automatic rebuild on errors
run:
	@echo "🎮 Running game..."
	@go run . || (echo "❌ Build failed. Fix errors and try again." && exit 1)

# Build the game executable
build:
	@echo "🔨 Building game executable..."
	@go build -o game . && echo "✅ Build successful: ./game" || echo "❌ Build failed"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./... -cover

# Run tests with coverage report
test-coverage:
	@echo "📊 Generating coverage report..."
	@go test -v ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f game game.exe game.wasm coverage.out coverage.html
	@echo "✅ Clean complete"

# Development mode with hot reload
dev:
	@echo "🔥 Starting development server with hot reload..."
	@./scripts/dev.sh

# Watch for changes and rebuild
watch:
	@echo "👁️ Watching for changes..."
	@./scripts/watch.sh

# Check for compilation errors without building
check:
	@echo "🔍 Checking for compilation errors..."
	@go build -o /dev/null ./... && echo "✅ No compilation errors" || echo "❌ Compilation errors found"

# Validate code quality
validate:
	@echo "📝 Running validation checks..."
	@echo "→ Checking formatting..."
	@test -z "$$(gofmt -l .)" || (echo "❌ Files need formatting:" && gofmt -l . && exit 1)
	@echo "→ Running go vet..."
	@go vet ./... || exit 1
	@echo "→ Checking dependencies..."
	@go mod verify || exit 1
	@echo "✅ All validation checks passed"

# Quick error check - useful for iterative development
quick:
	@go build -o /dev/null . 2>&1 | head -20 || true

# Build for WebAssembly
wasm:
	@echo "🌐 Building for WebAssembly..."
	@GOOS=js GOARCH=wasm go build -o game.wasm .
	@echo "✅ WebAssembly build: game.wasm"

# Update dependencies
deps:
	@echo "📦 Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies updated"

# Database operations
db-reset:
	@echo "🗄️ Resetting database..."
	@rm -f data/game.db
	@echo "✅ Database reset"

# Asset validation
assets:
	@echo "🎨 Validating assets..."
	@./scripts/check-assets.sh

# Help command
help:
	@echo "🎮 Isometric Game Development Commands"
	@echo ""
	@echo "  make run          - Run the game"
	@echo "  make build        - Build executable"
	@echo "  make test         - Run all tests"
	@echo "  make dev          - Start with hot reload"
	@echo "  make watch        - Watch and auto-rebuild"
	@echo "  make check        - Check for errors"
	@echo "  make validate     - Run code quality checks"
	@echo "  make quick        - Quick error check"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make wasm         - Build for WebAssembly"
	@echo "  make deps         - Update dependencies"
	@echo "  make db-reset     - Reset database"
	@echo "  make assets       - Validate assets"
	@echo "  make help         - Show this help"