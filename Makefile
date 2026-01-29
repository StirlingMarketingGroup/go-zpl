.PHONY: all test lint fmt vet build clean coverage fuzz help setup hooks wasm site site-dev

# Default target
all: lint test build

# Run all tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -s -w .
	goimports -w -local github.com/StirlingMarketingGroup/go-zpl .

# Run go vet
vet:
	go vet ./...

# Build the package
build:
	go build -v ./...

# Run fuzz tests (30 seconds each)
fuzz:
	@echo "Running fuzz tests..."
	@for f in $$(go test -list='Fuzz.*' ./... 2>/dev/null | grep -E '^Fuzz'); do \
		echo "Fuzzing $$f..."; \
		go test -fuzz="^$${f}$$" -fuzztime=30s ./... || true; \
	done

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -testcache

# Install development tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Install git hooks
hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks installed from .githooks/"

# Full development setup (tools + hooks)
setup: tools hooks
	@echo "Development environment ready!"

# Run example
example:
	go run ./examples/...

# Benchmark tests
bench:
	go test -bench=. -benchmem ./...

# Generate documentation
docs:
	@echo "View documentation at: https://pkg.go.dev/github.com/StirlingMarketingGroup/go-zpl"
	go doc -all .

# Build WebAssembly for the site
wasm:
	cd site/assets/go && GOOS=js GOARCH=wasm go build -o ../../static/lib.wasm .

# Build the Hugo site
site: wasm
	cd site && hugo --minify

# Run Hugo dev server
site-dev:
	cd site && hugo server --bind 0.0.0.0

# Run full dev environment (requires npm install first)
dev:
	npm run dev

help:
	@echo "Available targets:"
	@echo "  all       - Run lint, test, and build (default)"
	@echo "  test      - Run all tests with race detection"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  lint      - Run golangci-lint"
	@echo "  fmt       - Format code with gofmt and goimports"
	@echo "  vet       - Run go vet"
	@echo "  build     - Build the package"
	@echo "  fuzz      - Run fuzz tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  tools     - Install development tools"
	@echo "  hooks     - Install git pre-commit hooks"
	@echo "  setup     - Full dev setup (tools + hooks)"
	@echo "  bench     - Run benchmarks"
	@echo "  docs      - View package documentation"
	@echo "  wasm      - Build WebAssembly for the site"
	@echo "  site      - Build the Hugo site (includes wasm)"
	@echo "  site-dev  - Run Hugo dev server"
	@echo "  dev       - Run full dev environment (Hugo + wasm watch)"
	@echo "  help      - Show this help message"
