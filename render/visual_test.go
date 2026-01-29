package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// visualTestCase defines a visual regression test case.
type visualTestCase struct {
	Name   string
	Dir    string // Directory under testdata/visual/
	Width  int    // Label width in inches
	Height int    // Label height in inches
	DPI    zpl.DPI
}

var visualTestCases = []visualTestCase{
	{
		Name:   "Labelary Demo",
		Dir:    "labelary",
		Width:  4,
		Height: 6,
		DPI:    zpl.DPI203,
	},
	{
		Name:   "UPS Expedited",
		Dir:    "ups_expedited",
		Width:  4,
		Height: 6,
		DPI:    zpl.DPI203,
	},
}

// TestVisualRegression compares our rendered output against our own baseline images.
// This detects regressions in our renderer, not differences from Labelary.
// Use UPDATE_VISUAL_BASELINE=1 to update baselines after intentional changes.
func TestVisualRegression(t *testing.T) {
	updateBaseline := os.Getenv("UPDATE_VISUAL_BASELINE") == "1"

	for _, tc := range visualTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			baseDir := filepath.Join("..", "testdata", "visual", tc.Dir)

			// Read the ZPL file
			zplPath := filepath.Join(baseDir, "label.zpl")
			zplData, err := os.ReadFile(zplPath)
			if err != nil {
				t.Fatalf("Failed to read ZPL file: %v", err)
			}

			// Parse the ZPL
			label, err := zpl.Parse(string(zplData))
			if err != nil {
				t.Fatalf("Failed to parse ZPL: %v", err)
			}

			// Calculate dimensions in dots
			widthDots := tc.Width * int(tc.DPI)
			heightDots := tc.Height * int(tc.DPI)

			// Render with our renderer
			renderer := New(tc.DPI).WithSize(widthDots, heightDots)
			ourImg, err := renderer.Render(label)
			if err != nil {
				t.Fatalf("Failed to render: %v", err)
			}

			// Baseline is our own known-good output (not Labelary)
			baselinePath := filepath.Join(baseDir, "baseline.png")

			// If updating baselines, just save and skip comparison
			if updateBaseline {
				if err := saveImage(ourImg, baselinePath); err != nil {
					t.Fatalf("Failed to save baseline: %v", err)
				}
				t.Logf("Updated baseline: %s", baselinePath)
				return
			}

			// Load baseline image
			baselineFile, err := os.Open(baselinePath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("Baseline not found: %s\nRun with UPDATE_VISUAL_BASELINE=1 to create it", baselinePath)
				}
				t.Fatalf("Failed to open baseline image: %v", err)
			}
			defer baselineFile.Close()

			baselineImg, err := png.Decode(baselineFile)
			if err != nil {
				t.Fatalf("Failed to decode baseline image: %v", err)
			}

			// Compare images (our current output vs our baseline)
			diff, diffPct := compareImages(ourImg, baselineImg)

			// Report results - 0% tolerance for regression testing
			// Any difference from our baseline is a regression
			const maxDiffPercent = 0.1 // Allow tiny tolerance for floating point/platform differences
			if diffPct > maxDiffPercent {
				// Only save actual and diff images on failure to avoid dirtying workspace
				actualPath := filepath.Join(baseDir, "actual.png")
				if err := saveImage(ourImg, actualPath); err != nil {
					t.Logf("Warning: failed to save actual image: %v", err)
				}
				diffPath := filepath.Join(baseDir, "diff.png")
				if diff != nil {
					if err := saveImage(diff, diffPath); err != nil {
						t.Logf("Warning: failed to save diff image: %v", err)
					}
				}
				t.Errorf("Visual regression detected: %.2f%% difference from baseline\n"+
					"  Baseline: %s\n"+
					"  Actual: %s\n"+
					"  Diff: %s\n"+
					"If this change is intentional, run: UPDATE_VISUAL_BASELINE=1 go test -run TestVisualRegression",
					diffPct, baselinePath, actualPath, diffPath)
			} else if diffPct > 0 {
				t.Logf("Minor difference: %.4f%% (within tolerance)", diffPct)
			}
		})
	}
}

// compareImages compares two images with alignment tolerance.
// It checks if pixels match within a small radius to account for minor positioning differences.
func compareImages(img1, img2 image.Image) (diffImage image.Image, pctDifferent float64) {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Check dimensions match
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		// Create a diff image showing the size mismatch
		maxW := max(bounds1.Dx(), bounds2.Dx())
		maxH := max(bounds1.Dy(), bounds2.Dy())
		diff := image.NewRGBA(image.Rect(0, 0, maxW, maxH))
		// Fill with red to indicate error
		for y := 0; y < maxH; y++ {
			for x := 0; x < maxW; x++ {
				diff.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
		return diff, 100.0 // 100% different if sizes don't match
	}

	width := bounds1.Dx()
	height := bounds1.Dy()
	totalPixels := width * height
	differentPixels := 0

	diff := image.NewRGBA(image.Rect(0, 0, width, height))

	// Alignment tolerance: check within this radius for matching pixels
	const alignRadius = 2
	// Luminance tolerance for anti-aliasing
	const lumTolerance uint32 = 32

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			lum1 := getLuminance(img1, x+bounds1.Min.X, y+bounds1.Min.Y)
			lum2 := getLuminance(img2, x+bounds2.Min.X, y+bounds2.Min.Y)

			// First check: direct pixel match (with luminance tolerance)
			if absDiff(lum1, lum2) <= lumTolerance {
				gray := uint8(lum1)
				diff.Set(x, y, color.RGBA{gray, gray, gray, 255})
				continue
			}

			// Second check: alignment tolerance - look for matching pixel nearby
			found := false
			for dy := -alignRadius; dy <= alignRadius && !found; dy++ {
				for dx := -alignRadius; dx <= alignRadius && !found; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						nearLum := getLuminance(img2, nx+bounds2.Min.X, ny+bounds2.Min.Y)
						if absDiff(lum1, nearLum) <= lumTolerance {
							found = true
						}
					}
				}
			}

			if found {
				// Match found within alignment tolerance - keep original grayscale
				gray := uint8(lum1)
				diff.Set(x, y, color.RGBA{gray, gray, gray, 255})
			} else {
				differentPixels++
				// No match found - mark as red (significant diff)
				diff.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
	}

	pctDifferent = (float64(differentPixels) / float64(totalPixels)) * 100
	return diff, pctDifferent
}

// getLuminance returns the luminance (0-255) of a pixel.
func getLuminance(img image.Image, x, y int) uint32 {
	r, g, b, _ := img.At(x, y).RGBA()
	return ((r >> 8) + (g >> 8) + (b >> 8)) / 3
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// TestVisualRegressionFetchLabelary fetches Labelary's output for comparison.
// This is useful for comparing our renderer against Labelary during development.
// Run with: FETCH_LABELARY=1 go test -run TestVisualRegressionFetchLabelary
func TestVisualRegressionFetchLabelary(t *testing.T) {
	if os.Getenv("FETCH_LABELARY") != "1" {
		t.Skip("Skipping Labelary fetch (set FETCH_LABELARY=1 to run)")
	}

	t.Log("To fetch Labelary reference images, run these commands:")
	t.Log("")

	for _, tc := range visualTestCases {
		baseDir := filepath.Join("..", "testdata", "visual", tc.Dir)
		zplPath := filepath.Join(baseDir, "label.zpl")
		labelaryPath := filepath.Join(baseDir, "labelary.png")

		t.Logf("curl -s 'https://api.labelary.com/v1/printers/8dpmm/labels/%dx%d/0/' --data-binary @%s -o %s",
			tc.Width, tc.Height, zplPath, labelaryPath)
	}
}

// BenchmarkVisualComparison benchmarks the image comparison algorithm.
func BenchmarkVisualComparison(b *testing.B) {
	// Create two test images
	img1 := image.NewRGBA(image.Rect(0, 0, 812, 1218))
	img2 := image.NewRGBA(image.Rect(0, 0, 812, 1218))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compareImages(img1, img2)
	}
}
