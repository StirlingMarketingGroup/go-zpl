//go:build ignore

// smart_extract_page2 extracts glyphs from page 2 scan
// Run with: go run tools/smart_extract_page2.go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	scanFile      = "testdata/scans/font0_boxed_page2_600dpi.png"
	outDir        = "testdata/scans/glyphs"
	darkThreshold = 150
)

// Page 2 characters - based on what actually printed
// Row 1: + , - . / : ; <
// Row 2: = > ? @ [ ¢ ] (^ didn't print)
// Row 3: _ ` { | } (~ didn't print)
var chars = []string{
	// Row 1
	"plus", "comma", "minus", "period", "slash", "colon", "semicolon", "less",
	// Row 2
	"equal", "greater", "question", "at", "lbracket", "cent", "rbracket", "caret",
	// Row 3
	"underscore", "backtick", "lbrace", "pipe", "rbrace", "tilde",
}

func main() {
	f, err := os.Open(scanFile)
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

	// Find grid lines
	fmt.Println("Finding horizontal grid lines...")
	hLines := findHorizontalLines(img)
	fmt.Printf("Found %d horizontal lines: %v\n", len(hLines), hLines)

	fmt.Println("Finding vertical grid lines...")
	vLines := findVerticalLines(img)
	fmt.Printf("Found %d vertical lines: %v\n", len(vLines), vLines)

	// Extract each glyph
	cols := 8
	rows := 3
	idx := 0

	for row := 0; row < rows && 2*row+1 < len(hLines); row++ {
		for col := 0; col < cols && col < len(vLines)-1; col++ {
			if idx >= len(chars) {
				break
			}

			char := chars[idx]
			idx++

			boxLeft := vLines[col]
			boxRight := vLines[col+1]
			boxTop := hLines[2*row]
			boxBottom := hLines[2*row+1]

			innerLeft := findWhiteRight(img, boxLeft, boxTop, boxBottom) + 2
			innerRight := findWhiteLeft(img, boxRight, boxTop, boxBottom) - 2
			innerTop := findWhiteDown(img, boxTop, boxLeft, boxRight) + 2
			innerBottom := findWhiteUp(img, boxBottom, boxLeft, boxRight) - 2

			w := innerRight - innerLeft
			h := innerBottom - innerTop

			if w <= 0 || h <= 0 {
				fmt.Printf("  [%d,%d] %s - Invalid dimensions\n", row, col, char)
				continue
			}

			cropped := cropImage(img, innerLeft, innerTop, w, h)

			pngPath := filepath.Join(outDir, fmt.Sprintf("%s.png", char))
			if err := savePNG(pngPath, cropped); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving %s: %v\n", char, err)
				continue
			}

			pbmPath := filepath.Join(outDir, fmt.Sprintf("%s.pbm", char))
			// Use 30% threshold - only very dark pixels become black (eliminates faint scan artifacts)
			cmd := exec.Command("magick", pngPath, "-colorspace", "gray", "-threshold", "30%", pbmPath)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error converting %s to PBM: %v\n", char, err)
				continue
			}

			svgPath := filepath.Join(outDir, fmt.Sprintf("%s.svg", char))
			cmd = exec.Command("potrace", "-s", pbmPath, "-o", svgPath, "--flat", "--turdsize", "200")
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error tracing %s: %v\n", char, err)
				continue
			}

			fmt.Printf("  [%d,%d] %s: %dx%d\n", row, col, char, w, h)
		}
	}

	fmt.Println("\nDone!")
}

func findHorizontalLines(img image.Image) []int {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var lines []int
	startX := width / 4
	endX := 3 * width / 4
	sampleWidth := endX - startX
	threshold := int(float64(sampleWidth) * 0.7)

	inLine := false
	lineStart := 0

	for y := 0; y < height; y++ {
		darkCount := 0
		for x := startX; x < endX; x++ {
			if isDark(img.At(x, y)) {
				darkCount++
			}
		}

		if darkCount > threshold {
			if !inLine {
				inLine = true
				lineStart = y
			}
		} else {
			if inLine {
				lineMiddle := (lineStart + y) / 2
				lines = append(lines, lineMiddle)
				inLine = false
			}
		}
	}

	return lines
}

func findVerticalLines(img image.Image) []int {
	bounds := img.Bounds()
	width := bounds.Dx()

	var lines []int
	// For page 2, use the rows we know exist (roughly y=50 to y=1150)
	startY := 50
	endY := 1150
	sampleHeight := endY - startY
	threshold := int(float64(sampleHeight) * 0.5) // Lower threshold for shorter grid

	inLine := false
	lineStart := 0

	// Start at x=50 to skip image edge artifacts
	for x := 50; x < width; x++ {
		darkCount := 0
		for y := startY; y < endY; y++ {
			if isDark(img.At(x, y)) {
				darkCount++
			}
		}

		if darkCount > threshold {
			if !inLine {
				inLine = true
				lineStart = x
			}
		} else {
			if inLine {
				lineMiddle := (lineStart + x) / 2
				lines = append(lines, lineMiddle)
				inLine = false
			}
		}
	}

	return lines
}

// findWhiteRight finds first white pixel moving right from startX
// Samples multiple Y positions and returns the MAX X to handle varying border thickness
func findWhiteRight(img image.Image, startX, top, bottom int) int {
	bounds := img.Bounds()
	maxX := startX
	for y := top + 10; y < bottom-10; y += 20 {
		endX := startX + 100
		if endX > bounds.Max.X {
			endX = bounds.Max.X
		}
		for x := startX; x < endX; x++ {
			if !isDark(img.At(x, y)) {
				if x > maxX {
					maxX = x
				}
				break
			}
		}
	}
	if maxX == startX {
		return startX + 10
	}
	return maxX
}

// findWhiteLeft finds first white pixel moving left from startX
// Samples multiple Y positions and returns the MIN X to handle varying border thickness
func findWhiteLeft(img image.Image, startX, top, bottom int) int {
	bounds := img.Bounds()
	minX := startX
	for y := top + 10; y < bottom-10; y += 20 {
		endX := startX - 100
		if endX < bounds.Min.X {
			endX = bounds.Min.X
		}
		for x := startX; x > endX; x-- {
			if !isDark(img.At(x, y)) {
				if x < minX {
					minX = x
				}
				break
			}
		}
	}
	if minX == startX {
		return startX - 10
	}
	return minX
}

// findWhiteDown finds first white pixel moving down from startY
// Samples multiple X positions and returns the MAX Y to handle varying border thickness
func findWhiteDown(img image.Image, startY, left, right int) int {
	bounds := img.Bounds()
	maxY := startY
	for x := left + 10; x < right-10; x += 20 {
		endY := startY + 100
		if endY > bounds.Max.Y {
			endY = bounds.Max.Y
		}
		for y := startY; y < endY; y++ {
			if !isDark(img.At(x, y)) {
				if y > maxY {
					maxY = y
				}
				break
			}
		}
	}
	if maxY == startY {
		return startY + 10
	}
	return maxY
}

// findWhiteUp finds first white pixel moving up from startY
// Samples multiple X positions and returns the MIN Y to handle varying border thickness
func findWhiteUp(img image.Image, startY, left, right int) int {
	bounds := img.Bounds()
	minY := startY
	for x := left + 10; x < right-10; x += 20 {
		endY := startY - 100
		if endY < bounds.Min.Y {
			endY = bounds.Min.Y
		}
		for y := startY; y > endY; y-- {
			if !isDark(img.At(x, y)) {
				if y < minY {
					minY = y
				}
				break
			}
		}
	}
	if minY == startY {
		return startY - 10
	}
	return minY
}

func isDark(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	avg := (r>>8 + g>>8 + b>>8) / 3
	return avg < darkThreshold
}

func cropImage(img image.Image, x, y, w, h int) image.Image {
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
