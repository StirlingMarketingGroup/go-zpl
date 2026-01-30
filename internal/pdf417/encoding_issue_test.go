package pdf417

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	xdraw "golang.org/x/image/draw"
)

// TestIncrementalComplexity tests encoding with increasing complexity
// to find where things break
func TestIncrementalComplexity(t *testing.T) {
	if os.Getenv("PDF417_DEBUG") != "1" {
		t.Skip("set PDF417_DEBUG=1 to run encoding complexity tests (writes images to Desktop)")
	}
	homeDir, _ := os.UserHomeDir()

	testCases := []struct {
		name     string
		data     string
		secLevel byte
		cols     int
	}{
		// Simple cases that should work
		{"01_text_only", "Hello World", 2, 0},
		{"02_numbers_only", "12345678901234567890", 2, 0},

		// Add one control character
		{"03_text_with_RS", "Hello\x1EWorld", 2, 0},
		{"04_text_with_GS", "Hello\x1DWorld", 2, 0},

		// Start of FedEx format
		{"05_fedex_header", "[)>\x1E01", 2, 0},
		{"06_fedex_header_gs", "[)>\x1E01\x1D", 2, 0},

		// First data segment
		{"07_fedex_first_segment", "[)>\x1E01\x1D0294105", 2, 0},

		// Two segments
		{"08_fedex_two_segments", "[)>\x1E01\x1D0294105\x1D840", 2, 0},

		// More segments
		{"09_fedex_more", "[)>\x1E01\x1D0294105\x1D840\x1D20", 2, 0},

		// Long numeric (triggers numeric mode)
		{"10_long_numeric", "7949819308110201", 2, 0},
		{"11_with_long_numeric", "[)>\x1E01\x1D7949819308110201", 2, 0},

		// Mixed with text after control chars
		{"12_mixed_text", "[)>\x1E01\x1DFDE", 2, 0},

		// Partial FedEx (first ~50 bytes)
		{"13_partial_fedex", fedexData[:50], 3, 0},

		// Partial FedEx (first ~100 bytes)
		{"14_partial_fedex_100", fedexData[:100], 4, 0},

		// Full FedEx
		{"15_full_fedex", fedexData, 5, 14},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := EncodeWithDimensions(tc.data, tc.secLevel, tc.cols, 0)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			bounds := code.Bounds()
			t.Logf("Data length: %d bytes, Barcode: %dx%d", len(tc.data), bounds.Dx(), bounds.Dy())

			// Show codewords for debugging
			dataWords, _ := highlevelEncode(tc.data)
			t.Logf("Data codewords: %d", len(dataWords))
			if len(dataWords) <= 20 {
				t.Logf("Codewords: %v", dataWords)
			} else {
				t.Logf("First 20 codewords: %v", dataWords[:20])
			}

			// Save barcode
			img := saveTestBarcode(code, 3, 20)
			outPath := filepath.Join(homeDir, "Desktop", "pdf417_test_"+tc.name+".png")
			f, err := os.Create(outPath)
			if err != nil {
				t.Logf("Could not save to Desktop, using temp: %v", err)
				outPath = filepath.Join(os.TempDir(), "pdf417_test_"+tc.name+".png")
				f, _ = os.Create(outPath)
			}
			png.Encode(f, img)
			f.Close()
			t.Logf("Saved to: %s", outPath)
		})
	}
}

func saveTestBarcode(code image.Image, scaleFactor, quietZone int) *image.RGBA {
	bounds := code.Bounds()

	// Create raw image with quiet zone
	rawWidth := bounds.Dx() + 2*quietZone
	rawHeight := bounds.Dy() + 2*quietZone
	rawImg := image.NewRGBA(image.Rect(0, 0, rawWidth, rawHeight))

	// Fill with white
	for y := 0; y < rawHeight; y++ {
		for x := 0; x < rawWidth; x++ {
			rawImg.Set(x, y, color.White)
		}
	}

	// Draw barcode
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rawImg.Set(x+quietZone, y+quietZone, code.At(x, y))
		}
	}

	// Scale up
	scaledWidth := rawWidth * scaleFactor
	scaledHeight := rawHeight * scaleFactor
	img := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))
	xdraw.NearestNeighbor.Scale(img, img.Bounds(), rawImg, rawImg.Bounds(), xdraw.Over, nil)

	return img
}

// TestNumericEncoding specifically tests numeric mode encoding
func TestNumericEncoding(t *testing.T) {
	// PDF417 numeric mode encodes digits very efficiently
	// 13+ consecutive digits triggers numeric mode

	testData := "7949819308110201" // 16 digits from FedEx data

	codewords, err := highlevelEncode(testData)
	if err != nil {
		t.Fatalf("highlevelEncode failed: %v", err)
	}

	t.Logf("Input: %q (%d digits)", testData, len(testData))
	t.Logf("Codewords: %v", codewords)

	// First codeword should be 902 (latch to numeric)
	if len(codewords) > 0 && codewords[0] != 902 {
		t.Errorf("Expected latch to numeric (902), got %d", codewords[0])
	}

	// Verify by encoding the numeric value
	// 44 digits -> 15 codewords, so 16 digits should be ~6 codewords
	t.Logf("Codeword count: %d (expected ~6 for 16 digits)", len(codewords))
}

// TestModeTransitions tests transitions between encoding modes
func TestModeTransitions(t *testing.T) {
	// This tests the specific pattern in FedEx data
	// Text -> Binary -> Text -> Binary -> Numeric

	data := "[)>\x1E01\x1D0294105\x1D7949819308110201"

	codewords, err := highlevelEncode(data)
	if err != nil {
		t.Fatalf("highlevelEncode failed: %v", err)
	}

	t.Logf("Input: %d bytes", len(data))
	t.Logf("Codewords (%d total): %v", len(codewords), codewords)

	// Identify mode latches in the output
	for i, cw := range codewords {
		switch cw {
		case 900:
			t.Logf("  [%d] = 900 (latch to text)", i)
		case 901:
			t.Logf("  [%d] = 901 (latch to byte padded)", i)
		case 902:
			t.Logf("  [%d] = 902 (latch to numeric)", i)
		case 913:
			t.Logf("  [%d] = 913 (shift to byte)", i)
		case 924:
			t.Logf("  [%d] = 924 (latch to byte)", i)
		}
	}
}
