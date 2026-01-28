//go:build ignore

// smart_extract detects box borders dynamically and extracts glyph interiors
// Run with: go run tools/smart_extract.go
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
	scanFile = "testdata/scans/font0_boxed_600dpi.png"
	outDir   = "testdata/scans/glyphs"

	// Grid layout
	cols = 8
	rows = 9

	// Border detection threshold
	darkThreshold = 150
)

// Characters in order (row by row)
var chars = []string{
	"upper_A", "upper_B", "upper_C", "upper_D", "upper_E", "upper_F", "upper_G", "upper_H",
	"upper_I", "upper_J", "upper_K", "upper_L", "upper_M", "upper_N", "upper_O", "upper_P",
	"upper_Q", "upper_R", "upper_S", "upper_T", "upper_U", "upper_V", "upper_W", "upper_X",
	"upper_Y", "upper_Z", "lower_a", "lower_b", "lower_c", "lower_d", "lower_e", "lower_f",
	"lower_g", "lower_h", "lower_i", "lower_j", "lower_k", "lower_l", "lower_m", "lower_n",
	"lower_o", "lower_p", "lower_q", "lower_r", "lower_s", "lower_t", "lower_u", "lower_v",
	"lower_w", "lower_x", "lower_y", "lower_z", "digit_0", "digit_1", "digit_2", "digit_3",
	"digit_4", "digit_5", "digit_6", "digit_7", "digit_8", "digit_9", "exclaim", "quote",
	"hash", "dollar", "percent", "ampersand", "apostrophe", "lparen", "rparen", "asterisk",
}

func main() {
	// Load scan
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

	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	// Clear old files
	for _, ext := range []string{"*.png", "*.pbm", "*.svg"} {
		oldFiles, _ := filepath.Glob(filepath.Join(outDir, ext))
		for _, f := range oldFiles {
			os.Remove(f)
		}
	}

	// Step 1: Find horizontal grid lines (top/bottom of each row)
	// These are continuous dark horizontal bands
	fmt.Println("Finding horizontal grid lines...")
	hLines := findHorizontalLines(img)
	fmt.Printf("Found %d horizontal lines\n", len(hLines))

	// Step 2: Find vertical grid lines (left/right of each column)
	fmt.Println("Finding vertical grid lines...")
	vLines := findVerticalLines(img)
	fmt.Printf("Found %d vertical lines\n", len(vLines))

	// We expect 10 horizontal lines (top of row 0, bottom of row 0 = top of row 1, ..., bottom of row 8)
	// and 9 vertical lines (left of col 0, ..., right of col 7)
	if len(hLines) < rows+1 || len(vLines) < cols+1 {
		fmt.Printf("Warning: Expected %d h-lines and %d v-lines, got %d and %d\n",
			rows+1, cols+1, len(hLines), len(vLines))
	}

	// Print detected lines for debugging
	fmt.Printf("H-lines: %v\n", hLines)
	fmt.Printf("V-lines: %v\n", vLines)

	// Horizontal lines come in pairs: top border, then a gap, then bottom border
	// But since boxes share borders, we have: top0, bottom0=top1, bottom1=top2, etc.
	// Actually examining the data: we see pairs like (52), (392,426), (771,806)...
	// Line 0 is the top of the first row
	// Lines 1,2 are the gap between rows 0 and 1 (bottom of row 0 / top of row 1)
	// So row i uses: top=hLines[2*i], bottom=hLines[2*i+1]

	// Step 3: Extract each glyph using the detected grid
	for row := 0; row < rows && 2*row+1 < len(hLines); row++ {
		for col := 0; col < cols && col < len(vLines)-1; col++ {
			idx := row*cols + col
			if idx >= len(chars) {
				break
			}

			char := chars[idx]

			// Box boundaries from grid lines
			boxLeft := vLines[col]
			boxRight := vLines[col+1]
			boxTop := hLines[2*row]
			boxBottom := hLines[2*row+1]

			// Find the inner edge (where white begins)
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

			// Crop the glyph
			cropped := cropImage(img, innerLeft, innerTop, w, h)

			// Save as PNG
			pngPath := filepath.Join(outDir, fmt.Sprintf("%s.png", char))
			if err := savePNG(pngPath, cropped); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving %s: %v\n", char, err)
				continue
			}

			// Convert to PBM for potrace
			pbmPath := filepath.Join(outDir, fmt.Sprintf("%s.pbm", char))
			cmd := exec.Command("magick", pngPath, "-colorspace", "gray", "-threshold", "40%", pbmPath)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error converting %s to PBM: %v\n", char, err)
				continue
			}

			// Trace to SVG
			svgPath := filepath.Join(outDir, fmt.Sprintf("%s.svg", char))
			cmd = exec.Command("potrace", "-s", pbmPath, "-o", svgPath, "--flat", "--turdsize", "15")
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error tracing %s: %v\n", char, err)
				continue
			}

			fmt.Printf("  [%d,%d] %s: %dx%d\n", row, col, char, w, h)
		}
	}

	fmt.Println("\nDone!")
	fmt.Printf("PNGs: %d\n", countFiles(outDir, "*.png"))
	fmt.Printf("SVGs: %d\n", countFiles(outDir, "*.svg"))
}

// findHorizontalLines finds Y coordinates of horizontal dark bands
func findHorizontalLines(img image.Image) []int {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Scan each row and count dark pixels
	// A grid line will have many consecutive dark pixels across most of the width
	var lines []int

	// Use a column range that should cross all grid lines (middle third of image)
	startX := width / 4
	endX := 3 * width / 4
	sampleWidth := endX - startX

	threshold := int(float64(sampleWidth) * 0.7) // At least 70% dark = real border line

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
				// End of dark band - use the middle
				lineMiddle := (lineStart + y) / 2
				lines = append(lines, lineMiddle)
				inLine = false
			}
		}
	}

	return lines
}

// findVerticalLines finds X coordinates of vertical dark bands
func findVerticalLines(img image.Image) []int {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var lines []int

	// Use a row range that should cross all grid lines
	startY := height / 4
	endY := 3 * height / 4
	sampleHeight := endY - startY

	threshold := int(float64(sampleHeight) * 0.7)

	inLine := false
	lineStart := 0

	for x := 0; x < width; x++ {
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
func findWhiteRight(img image.Image, startX, top, bottom int) int {
	midY := (top + bottom) / 2
	for x := startX; x < startX+100; x++ {
		if !isDark(img.At(x, midY)) {
			return x
		}
	}
	return startX + 10
}

// findWhiteLeft finds first white pixel moving left from startX
func findWhiteLeft(img image.Image, startX, top, bottom int) int {
	midY := (top + bottom) / 2
	for x := startX; x > startX-100; x-- {
		if !isDark(img.At(x, midY)) {
			return x
		}
	}
	return startX - 10
}

// findWhiteDown finds first white pixel moving down from startY
func findWhiteDown(img image.Image, startY, left, right int) int {
	midX := (left + right) / 2
	for y := startY; y < startY+100; y++ {
		if !isDark(img.At(midX, y)) {
			return y
		}
	}
	return startY + 10
}

// findWhiteUp finds first white pixel moving up from startY
func findWhiteUp(img image.Image, startY, left, right int) int {
	midX := (left + right) / 2
	for y := startY; y > startY-100; y-- {
		if !isDark(img.At(midX, y)) {
			return y
		}
	}
	return startY - 10
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

func countFiles(dir, pattern string) int {
	files, _ := filepath.Glob(filepath.Join(dir, pattern))
	return len(files)
}
