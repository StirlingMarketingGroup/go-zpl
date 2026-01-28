//go:build ignore

// trim_glyphs normalizes all glyphs to uniform height using reference metrics
// Cap line from H, baseline from H, descender from underscore
// Run with: go run tools/trim_glyphs.go
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
	glyphDir      = "testdata/scans/glyphs"
	darkThreshold = 200
	edgeSkip      = 15 // Skip first/last N rows to avoid grid line artifacts
)

func main() {
	fmt.Println("=== Finding metrics from reference glyphs ===")

	// Find cap line by scanning ALL uppercase letters for the minimum top Y
	// This ensures we don't cut off any letter's top
	capLine := 9999
	baseline := 0
	for _, letter := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"} {
		img := loadGlyph("upper_" + letter)
		if img == nil {
			continue
		}
		bounds := findGlyphBoundsNoSkip(img)
		if bounds.Empty() {
			continue
		}
		if bounds.Min.Y < capLine {
			capLine = bounds.Min.Y
			fmt.Printf("  New cap line from '%s': %d\n", letter, capLine)
		}
		if bounds.Max.Y > baseline {
			baseline = bounds.Max.Y
		}
	}
	fmt.Printf("Cap line: %d (from scanning all uppercase)\n", capLine)
	fmt.Printf("Baseline: %d\n", baseline)

	// Find digit cap line - the minimum top Y across all digits
	// This ensures all digits align to the same top position
	digitCapLine := 9999
	digitTops := make(map[string]int) // Store each digit's actual top for adjustment
	for _, digit := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		img := loadGlyph("digit_" + digit)
		if img == nil {
			continue
		}
		bounds := findGlyphBoundsNoSkip(img)
		if bounds.Empty() {
			continue
		}
		digitTops["digit_"+digit] = bounds.Min.Y
		if bounds.Min.Y < digitCapLine {
			digitCapLine = bounds.Min.Y
			fmt.Printf("  New digit cap line from '%s': %d\n", digit, digitCapLine)
		}
	}
	fmt.Printf("Digit cap line: %d\n", digitCapLine)
	// Calculate how much to adjust each digit upward to align with digit cap line
	digitAdjust := make(map[string]int)
	for name, top := range digitTops {
		digitAdjust[name] = top - digitCapLine
		if digitAdjust[name] != 0 {
			fmt.Printf("  %s needs adjustment: %d pixels up\n", name, digitAdjust[name])
		}
	}

	// Load underscore to find the absolute bottom
	// Underscore sits ON the baseline, so its bottom pixel is our descender line
	// Use NoSkip because underscore is at the bottom of the cell
	var descenderLine int
	underscoreImg := loadGlyph("underscore")
	if underscoreImg == nil {
		fmt.Println("WARNING: Could not load underscore, using baseline as bottom")
		descenderLine = baseline
	} else {
		underscoreBounds := findGlyphBoundsNoSkip(underscoreImg)
		descenderLine = underscoreBounds.Max.Y
		fmt.Printf("From '_': bottom=%d\n", descenderLine)
	}

	// If underscore is above baseline (shouldn't happen), use baseline
	if descenderLine < baseline {
		descenderLine = baseline
		fmt.Println("WARNING: underscore above baseline, using baseline")
	}

	standardHeight := descenderLine - capLine
	fmt.Printf("\nStandard frame: cap=%d to descender=%d (height=%d)\n", capLine, descenderLine, standardHeight)
	fmt.Printf("Baseline is at y=%d in output frame\n", baseline-capLine)

	// Process ALL glyphs
	fmt.Println("\n=== Trimming glyphs ===")

	glyphs := getAllGlyphNames()
	for _, name := range glyphs {
		img := loadGlyph(name)
		if img == nil {
			fmt.Printf("  %s - SKIP (not found)\n", name)
			continue
		}

		// Use NoSkip since extraction already removed grid borders
		bounds := findGlyphBoundsNoSkip(img)
		if bounds.Empty() {
			fmt.Printf("  %s - SKIP (empty)\n", name)
			continue
		}

		// Horizontal trim: tight to content with 2px padding
		leftPad := 2
		rightPad := 2
		contentLeft := bounds.Min.X - leftPad
		if contentLeft < 0 {
			contentLeft = 0
		}
		contentRight := bounds.Max.X + rightPad
		if contentRight > img.Bounds().Max.X {
			contentRight = img.Bounds().Max.X
		}
		w := contentRight - contentLeft

		// Vertical: use standard height from cap to descender
		h := standardHeight

		if w <= 0 || h <= 0 {
			fmt.Printf("  %s - SKIP (invalid dims w=%d h=%d)\n", name, w, h)
			continue
		}

		// Create output image
		trimmed := image.NewRGBA(image.Rect(0, 0, w, h))

		// Fill with white
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				trimmed.Set(x, y, color.White)
			}
		}

		// Copy glyph pixels at correct vertical position
		// Position in output = position in source - capLine
		// For digits, apply additional adjustment to align them all to the same top
		yAdjust := 0
		if adj, ok := digitAdjust[name]; ok {
			yAdjust = adj
		}
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			outY := y - capLine - yAdjust
			if outY < 0 || outY >= h {
				continue
			}
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				outX := x - contentLeft
				if outX < 0 || outX >= w {
					continue
				}
				if isDark(img.At(x, y)) {
					trimmed.Set(outX, outY, img.At(x, y))
				}
			}
		}

		// Save
		pngPath := filepath.Join(glyphDir, fmt.Sprintf("%s.png", name))
		if err := savePNG(pngPath, trimmed); err != nil {
			fmt.Printf("  %s - ERROR: %v\n", name, err)
			continue
		}

		// Regenerate PBM and SVG
		pbmPath := filepath.Join(glyphDir, fmt.Sprintf("%s.pbm", name))
		cmd := exec.Command("magick", pngPath, "-colorspace", "gray", "-threshold", "40%", pbmPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s - ERROR PBM: %v\n", name, err)
			continue
		}

		svgPath := filepath.Join(glyphDir, fmt.Sprintf("%s.svg", name))
		cmd = exec.Command("potrace", "-s", pbmPath, "-o", svgPath, "--flat", "--turdsize", "15")
		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s - ERROR SVG: %v\n", name, err)
			continue
		}

		if yAdjust > 0 {
			fmt.Printf("  %s: %dx%d (adjusted %d px up)\n", name, w, h, yAdjust)
		} else {
			fmt.Printf("  %s: %dx%d\n", name, w, h)
		}
	}

	fmt.Println("\n=== Done! ===")
	fmt.Printf("All glyphs normalized to height=%d (cap→descender)\n", standardHeight)
	fmt.Printf("Baseline at y=%d\n", baseline-capLine)
}

func loadGlyph(name string) image.Image {
	path := filepath.Join(glyphDir, fmt.Sprintf("%s.png", name))
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil
	}
	return img
}

func findGlyphBounds(img image.Image) image.Rectangle {
	return findGlyphBoundsWithSkip(img, edgeSkip)
}

// findGlyphBoundsNoSkip finds bounds without skipping edge rows
// Use this for glyphs near edges (like underscore at bottom)
func findGlyphBoundsNoSkip(img image.Image) image.Rectangle {
	return findGlyphBoundsWithSkip(img, 0)
}

func findGlyphBoundsWithSkip(img image.Image, skip int) image.Rectangle {
	if img == nil {
		return image.Rectangle{}
	}

	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	// Skip first/last skip rows to avoid grid line artifacts
	startY := bounds.Min.Y + skip
	endY := bounds.Max.Y - skip
	if startY >= endY {
		startY = bounds.Min.Y
		endY = bounds.Max.Y
	}

	for y := startY; y < endY; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isDark(img.At(x, y)) {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if minX > maxX || minY > maxY {
		return image.Rectangle{}
	}

	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func isDark(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	avg := (r>>8 + g>>8 + b>>8) / 3
	return avg < darkThreshold
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func getAllGlyphNames() []string {
	return []string{
		// Page 1
		"upper_A", "upper_B", "upper_C", "upper_D", "upper_E", "upper_F", "upper_G", "upper_H",
		"upper_I", "upper_J", "upper_K", "upper_L", "upper_M", "upper_N", "upper_O", "upper_P",
		"upper_Q", "upper_R", "upper_S", "upper_T", "upper_U", "upper_V", "upper_W", "upper_X",
		"upper_Y", "upper_Z", "lower_a", "lower_b", "lower_c", "lower_d", "lower_e", "lower_f",
		"lower_g", "lower_h", "lower_i", "lower_j", "lower_k", "lower_l", "lower_m", "lower_n",
		"lower_o", "lower_p", "lower_q", "lower_r", "lower_s", "lower_t", "lower_u", "lower_v",
		"lower_w", "lower_x", "lower_y", "lower_z", "digit_0", "digit_1", "digit_2", "digit_3",
		"digit_4", "digit_5", "digit_6", "digit_7", "digit_8", "digit_9", "exclaim", "quote",
		"hash", "dollar", "percent", "ampersand", "apostrophe", "lparen", "rparen", "asterisk",
		// Page 2
		"plus", "comma", "minus", "period", "slash", "colon", "semicolon", "less",
		"equal", "greater", "question", "at", "lbracket", "cent", "rbracket",
		"underscore", "backtick", "lbrace", "pipe", "rbrace",
	}
}
