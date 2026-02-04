.PHONY: build run dev clean help bake prerender warm prepare cache-all cache-clear invalidate

help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build        - Build the application to bin/landing-page"
	@echo "  make run          - Run the compiled application"
	@echo "  make dev          - Run development server with hot reload (air)"
	@echo "  make clean        - Remove build artifacts from bin/"
	@echo ""
	@echo "Cache Management:"
	@echo "  make bake         - Pre-render and cache all cacheable pages"
	@echo "  make prerender    - Alias for bake"
	@echo "  make warm         - Alias for bake"
	@echo "  make prepare      - Alias for bake"
	@echo "  make cache-all    - Alias for bake"
	@echo "  make cache-clear  - Clear all cached files"
	@echo "  make invalidate   - Alias for cache-clear"
	@echo ""
	@echo "  make help         - Show this help message"

build:
	@echo "Building application..."
	@mkdir -p bin
	@go build -o ./bin/landing-page .
	@echo "Build complete: bin/landing-page"

run:
	@./bin/landing-page

dev:
	@air

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/*
	@echo "Clean complete"

# Cache management commands
bake: build
	@echo "Pre-rendering all cacheable pages..."
	@./bin/landing-page bake

prerender: bake

warm: bake

prepare: bake

cache-all: bake

cache-clear: build
	@echo "Clearing cache..."
	@./bin/landing-page clear-cache

invalidate: cache-clear
