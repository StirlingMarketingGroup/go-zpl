#!/bin/bash
# Visual regression test utilities
# Run from the repository root

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

case "${1:-help}" in
    baseline)
        echo "Updating visual regression baselines from current renderer output..."
        UPDATE_VISUAL_BASELINE=1 go test ./render/... -v -run TestVisualRegression
        echo ""
        echo "Baselines updated. Review the changes before committing."
        ;;

    labelary)
        echo "Fetching Labelary reference images for comparison..."
        echo ""

        for dir in "$SCRIPT_DIR"/*/; do
            name=$(basename "$dir")
            if [ -f "$dir/label.zpl" ]; then
                echo "  Fetching $name..."
                curl -s "https://api.labelary.com/v1/printers/8dpmm/labels/4x6/0/" \
                    --data-binary "@$dir/label.zpl" \
                    -o "$dir/labelary.png"
            fi
        done

        echo ""
        echo "Done! Labelary images saved as labelary.png in each test directory."
        echo "Compare with baseline.png and actual.png to see differences."
        ;;

    test)
        echo "Running visual regression tests..."
        go test ./render/... -v -run TestVisualRegression
        ;;

    *)
        echo "Visual Regression Test Utilities"
        echo ""
        echo "Usage: $0 <command>"
        echo ""
        echo "Commands:"
        echo "  baseline  - Update baselines from current renderer (after intentional changes)"
        echo "  labelary  - Fetch Labelary's output for comparison"
        echo "  test      - Run visual regression tests"
        echo ""
        echo "Workflow:"
        echo "  1. Run tests: $0 test"
        echo "  2. If tests fail unexpectedly, investigate the diff.png files"
        echo "  3. If changes are intentional: $0 baseline"
        echo "  4. To compare with Labelary: $0 labelary"
        ;;
esac
