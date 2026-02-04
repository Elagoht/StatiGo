.PHONY: build run dev clean help prerender clear-cache

help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build         - Build statigo binary"
	@echo "  make run           - Run statigo"
	@echo "  make dev           - Run development server with hot reload (air)"
	@echo "  make clean         - Remove build artifacts"
	@echo ""
	@echo "Cache Management:"
	@echo "  make prerender     - Pre-render all cacheable pages"
	@echo "  make clear-cache   - Clear all cached files"
	@echo ""
	@echo "  make help          - Show this help message"

build:
	@echo "Building statigo..."
	@go build -o ./statigo .
	@echo "Build complete: ./statigo"

run:
	@./statigo

dev:
	@air

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf statigo tmp/*
	@echo "Clean complete"

prerender: build
	@echo "Pre-rendering all cacheable pages..."
	@./statigo prerender

clear-cache: build
	@echo "Clearing cache..."
	@./statigo clear-cache
