#!/bin/bash
# Build the complete font pipeline from scratch
# Usage: ./tools/build_all.sh

set -euo pipefail
cd "$(dirname "$0")/.."

# Check required dependencies
for cmd in go magick potrace fontforge; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "ERROR: Required command '$cmd' not found"
        exit 1
    fi
done

echo "=== Cleaning old glyphs ==="
rm -f testdata/scans/glyphs/*.png testdata/scans/glyphs/*.pbm testdata/scans/glyphs/*.svg

echo ""
echo "=== Extracting page 1 ==="
go run tools/smart_extract.go

echo ""
echo "=== Extracting page 2 ==="
go run tools/smart_extract_page2.go

echo ""
echo "=== Trimming glyphs ==="
go run tools/trim_glyphs.go

echo ""
echo "=== Building font ==="
fontforge -script tools/build_font.py

echo ""
echo "=== Copying font to render package ==="
cp render/zebra_font0.ttf render/font0.ttf

echo ""
echo "=== Building WASM ==="
cd site/assets/go
GOOS=js GOARCH=wasm go build -o ../../static/lib.wasm .

echo ""
echo "=== Done! ==="
echo "Refresh http://localhost:1400/go-zpl/ to see changes"
