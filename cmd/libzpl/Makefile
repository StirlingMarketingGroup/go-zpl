# Build libzpl for various platforms
#
# Usage:
#   make              # Build for current platform
#   make all          # Build for all platforms
#   make linux        # Build for Linux (amd64)
#   make darwin       # Build for macOS (arm64 + amd64 universal)
#   make windows      # Build for Windows (amd64)

.PHONY: all clean linux darwin windows darwin-arm64 darwin-amd64

OUTPUT_DIR = ../../dist

# Current platform
default:
	go build -buildmode=c-shared -o libzpl$(shell go env GOEXE 2>/dev/null || echo "") ./

all: linux darwin windows

$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

# Linux amd64
linux: $(OUTPUT_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build -buildmode=c-shared -o $(OUTPUT_DIR)/libzpl-linux-amd64.so ./
	@echo "Built $(OUTPUT_DIR)/libzpl-linux-amd64.so"

# macOS arm64 (Apple Silicon)
darwin-arm64: $(OUTPUT_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -buildmode=c-shared -o $(OUTPUT_DIR)/libzpl-darwin-arm64.dylib ./
	@echo "Built $(OUTPUT_DIR)/libzpl-darwin-arm64.dylib"

# macOS amd64 (Intel)
darwin-amd64: $(OUTPUT_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -buildmode=c-shared -o $(OUTPUT_DIR)/libzpl-darwin-amd64.dylib ./
	@echo "Built $(OUTPUT_DIR)/libzpl-darwin-amd64.dylib"

# macOS universal binary (requires both architectures built first)
darwin: darwin-arm64 darwin-amd64
	lipo -create -output $(OUTPUT_DIR)/libzpl.dylib \
		$(OUTPUT_DIR)/libzpl-darwin-arm64.dylib \
		$(OUTPUT_DIR)/libzpl-darwin-amd64.dylib
	@echo "Built $(OUTPUT_DIR)/libzpl.dylib (universal)"

# Windows amd64 (requires mingw-w64)
windows: $(OUTPUT_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
		go build -buildmode=c-shared -o $(OUTPUT_DIR)/libzpl.dll ./
	@echo "Built $(OUTPUT_DIR)/libzpl.dll"

clean:
	rm -f libzpl.so libzpl.dylib libzpl.dll libzpl.h
	rm -f testdata/test_libzpl testdata/test_output.png
	rm -rf $(OUTPUT_DIR)

# Test with the C example
test: default
	gcc -o testdata/test_libzpl testdata/test.c -L. -lzpl -I.
	cd testdata && DYLD_LIBRARY_PATH=.. LD_LIBRARY_PATH=.. ./test_libzpl
