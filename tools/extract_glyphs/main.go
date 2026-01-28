//go:build ignore

// extract_glyphs crops individual characters from the font scan for tracing
// Run with: go run tools/extract_glyphs/main.go
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// Glyph defines a character and its bounding box in the scan
type Glyph struct {
	Char string
	X, Y int
	W, H int
}

func main() {
	scanPath := "testdata/scans/font0_600dpi.png"
	outDir := "testdata/scans/glyphs"

	// Open the scan
	f, err := os.Open(scanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening scan: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding PNG: %v\n", err)
		os.Exit(1)
	}

	bounds := img.Bounds()
	fmt.Printf("Scan size: %d x %d\n", bounds.Dx(), bounds.Dy())

	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	// The scan is at 600 DPI, printed at 200 DPI
	// So each printer dot = 3 scan pixels
	// 50pt glyphs = 150 scan pixels tall
	// 70pt glyphs = 210 scan pixels tall

	// Based on the scan layout, define approximate glyph positions
	// These are manually measured from the scan - adjust as needed
	// Format: character, x, y, width, height (in scan pixels)

	// Using the larger 70pt "HAMBURG" and "fontgrp" for better quality
	// Row starts around y=2200 in the scan (measuring from top)

	// Actually, let's extract from the 50pt rows which have all chars
	// Row 1 (ABCDEFGHIJKLM) starts around y=135, each char ~85 wide, ~150 tall
	// Spacing between chars is minimal

	// For now, let's extract the large HAMBURG letters for testing
	// These start around y=2230, are about 210 pixels tall

	glyphs := []Glyph{
		// HAMBURG - 70pt, y starts around 2185
		{"H", 48, 2185, 105, 175},
		{"A", 153, 2185, 105, 175},
		{"M", 255, 2185, 125, 175},
		{"B", 378, 2185, 100, 175},
		{"U", 478, 2185, 105, 175},
		{"R", 583, 2185, 100, 175},
		{"G", 683, 2185, 105, 175},

		// fontgrp - 70pt lowercase, y starts around 2360
		{"f", 52, 2360, 65, 175},
		{"o", 115, 2360, 90, 175},
		{"n", 205, 2360, 90, 175},
		{"t", 293, 2360, 55, 175},
		{"g", 348, 2360, 90, 175},
		{"r", 438, 2360, 60, 175},
		{"p", 498, 2360, 90, 175},
	}

	fmt.Printf("Extracting %d glyphs...\n", len(glyphs))

	for _, g := range glyphs {
		// Crop the glyph
		cropped := cropImage(img, g.X, g.Y, g.W, g.H)

		// Save as PNG
		outPath := filepath.Join(outDir, fmt.Sprintf("glyph_%s.png", sanitizeFilename(g.Char)))
		if err := savePNG(outPath, cropped); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving %s: %v\n", g.Char, err)
			continue
		}
		fmt.Printf("  Extracted: %s -> %s\n", g.Char, outPath)
	}

	fmt.Println("\nDone! Now convert to vectors with:")
	fmt.Println("  cd testdata/scans/glyphs")
	fmt.Println(`  for f in *.png; do potrace -s "$f" -o "${f%.png}.svg"; done`)
}

func cropImage(img image.Image, x, y, w, h int) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	if si, ok := img.(subImager); ok {
		return si.SubImage(image.Rect(x, y, x+w, y+h))
	}

	// Fallback: create new image
	cropped := image.NewRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			cropped.Set(dx, dy, img.At(x+dx, y+dy))
		}
	}
	return cropped
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func sanitizeFilename(s string) string {
	// Handle special characters that can't be in filenames
	switch s {
	case "/":
		return "slash"
	case "\\":
		return "backslash"
	case ":":
		return "colon"
	case "*":
		return "asterisk"
	case "?":
		return "question"
	case "\"":
		return "quote"
	case "<":
		return "less"
	case ">":
		return "greater"
	case "|":
		return "pipe"
	case " ":
		return "space"
	default:
		return s
	}
}
