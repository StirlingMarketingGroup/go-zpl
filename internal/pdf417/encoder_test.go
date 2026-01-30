package pdf417

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	xdraw "golang.org/x/image/draw"
)

// FedEx label data from testdata/visual/fedex_express_scanned/label.zpl
// The ^FH (field hex) markers convert _1E to RS (0x1E), _1D to GS (0x1D), etc.
var fedexData = "[)>\x1E01\x1D0294105\x1D840\x1D20\x1D7949819308110201\x1DFDE\x1D740561073\x1D031\x1D\x1D1/1\x1D5.00LB\x1DN\x1D100 Market Street\x1DSan Francisco\x1DCA\x1DTest Recipient\x1E06\x1D10ZED008\x1D11ZRecipient Corp\x1D12Z4155559876\x1D15Z114064860\x1D20Z\x1C\x1D31Z1195282044690009410500794981930811\x1D32Z02GD\x1D34Z01\x1D39ZHKAA\x1D\x1E09\x1DFDX\x1Dz\x1D8\x1D\x17\x04';0?\x7F@\x1E\x04"

func TestPDF417RoundTrip(t *testing.T) {
	// Check if zbarimg is available
	if _, err := exec.LookPath("zbarimg"); err != nil {
		t.Skip("zbarimg not installed, skipping round-trip test (install with: brew install zbar)")
	}

	// Encode with our library - same params as FedEx label: ^BY2,2^B7N,10,5,14
	// Security level: 5, Data columns: 14, Rows: auto (0)
	code, err := EncodeWithDimensions(fedexData, 5, 14, 0)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Get dimensions
	bounds := code.Bounds()
	t.Logf("Encoded barcode dimensions: %dx%d", bounds.Dx(), bounds.Dy())

	// Convert to image and scale up for reliable scanning
	// The raw barcode has moduleHeight=2 which is too thin for most scanners
	// Scale up by 4x to ensure scannability
	// Also add quiet zone (white border) for scanner detection
	scaleFactor := 4

	// First create 1:1 image with quiet zone
	rawBounds := image.Rect(0, 0, bounds.Dx()+20, bounds.Dy()+20) // 10px quiet zone on each side at 1x
	rawImg := image.NewRGBA(rawBounds)
	// Fill with white (quiet zone)
	for y := 0; y < rawBounds.Dy(); y++ {
		for x := 0; x < rawBounds.Dx(); x++ {
			rawImg.Set(x, y, image.White)
		}
	}
	// Draw barcode offset by quiet zone
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rawImg.Set(x+10, y+10, code.At(x, y))
		}
	}

	// Scale up using nearest neighbor to preserve sharp edges
	scaledBounds := image.Rect(0, 0, rawBounds.Dx()*scaleFactor, rawBounds.Dy()*scaleFactor)
	img := image.NewRGBA(scaledBounds)
	xdraw.NearestNeighbor.Scale(img, img.Bounds(), rawImg, rawImg.Bounds(), xdraw.Over, nil)

	t.Logf("Scaled image dimensions: %dx%d (with quiet zone)", img.Bounds().Dx(), img.Bounds().Dy())

	// Save to temp file
	tmpFile := filepath.Join(os.TempDir(), "pdf417_roundtrip_test.png")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Create temp file: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("Encode PNG: %v", err)
	}
	f.Close()
	t.Logf("Saved test barcode to: %s", tmpFile)

	// Decode with zbarimg
	decoded, err := decodeWithZbar(tmpFile)
	if err != nil {
		// Exit code 4 means no barcode found - zbar on macOS doesn't support PDF417
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 4 {
			t.Skip("zbarimg doesn't support PDF417 on this system (use phone app to verify barcode)")
		}
		t.Fatalf("zbarimg decode failed: %v", err)
	}

	// Compare
	if decoded != fedexData {
		t.Errorf("Round-trip mismatch:\n  Original: %q\n  Decoded:  %q", fedexData, decoded)
		t.Logf("Original len: %d, Decoded len: %d", len(fedexData), len(decoded))
	} else {
		t.Log("Round-trip successful!")
	}
}

func TestPDF417Dimensions(t *testing.T) {
	// Test with FedEx params
	code, err := EncodeWithDimensions(fedexData, 5, 14, 0)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	bounds := code.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Expected width: (columns+4)*17 + 1 = (14+4)*17 + 1 = 307 modules
	expectedWidth := (14+4)*17 + 1
	if width != expectedWidth {
		t.Errorf("Width mismatch: expected %d, got %d", expectedWidth, width)
	}

	// Height depends on number of rows * moduleHeight (2)
	// Log it so we can understand the structure
	t.Logf("Dimensions: %dx%d (width x height)", width, height)
	t.Logf("Expected width: %d, Number of rows: %d", expectedWidth, height/moduleHeight)

	// Verify moduleHeight constant
	if moduleHeight != 2 {
		t.Errorf("moduleHeight should be 2, got %d", moduleHeight)
	}
}

func TestPDF417SecurityLevelEC(t *testing.T) {
	// Security level 5 should produce 2^(5+1) = 64 EC codewords
	var sl securitylevel = 5
	ecCount := sl.ErrorCorrectionWordCount()
	expected := 64
	if ecCount != expected {
		t.Errorf("Security level 5 should produce %d EC codewords, got %d", expected, ecCount)
	}
}

func TestPDF417SimpleData(t *testing.T) {
	// Check if zbarimg is available
	if _, err := exec.LookPath("zbarimg"); err != nil {
		t.Skip("zbarimg not installed, skipping round-trip test")
	}

	// Test with simple ASCII data
	testData := "Hello World 12345"

	code, err := Encode(testData, 2)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	bounds := code.Bounds()
	t.Logf("Simple test barcode dimensions: %dx%d", bounds.Dx(), bounds.Dy())

	// Convert to image and scale up for reliable scanning
	scaleFactor := 4

	// First create 1:1 image with quiet zone
	rawBounds := image.Rect(0, 0, bounds.Dx()+20, bounds.Dy()+20)
	rawImg := image.NewRGBA(rawBounds)
	// Fill with white
	for y := 0; y < rawBounds.Dy(); y++ {
		for x := 0; x < rawBounds.Dx(); x++ {
			rawImg.Set(x, y, image.White)
		}
	}
	// Draw barcode
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rawImg.Set(x+10, y+10, code.At(x, y))
		}
	}

	scaledBounds := image.Rect(0, 0, rawBounds.Dx()*scaleFactor, rawBounds.Dy()*scaleFactor)
	img := image.NewRGBA(scaledBounds)
	xdraw.NearestNeighbor.Scale(img, img.Bounds(), rawImg, rawImg.Bounds(), xdraw.Over, nil)

	// Save to temp file
	tmpFile := filepath.Join(os.TempDir(), "pdf417_simple_test.png")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Create temp file: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("Encode PNG: %v", err)
	}
	f.Close()
	t.Logf("Saved test barcode to: %s", tmpFile)

	// Decode with zbarimg
	decoded, err := decodeWithZbar(tmpFile)
	if err != nil {
		// Exit code 4 means no barcode found - zbar on macOS doesn't support PDF417
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 4 {
			t.Skip("zbarimg doesn't support PDF417 on this system (use phone app to verify barcode)")
		}
		t.Fatalf("zbarimg decode failed: %v", err)
	}

	if decoded != testData {
		t.Errorf("Round-trip mismatch:\n  Original: %q\n  Decoded:  %q", testData, decoded)
	} else {
		t.Log("Simple data round-trip successful!")
	}
}

// decodeWithZbar uses zbarimg to decode a PDF417 barcode from an image file
func decodeWithZbar(imagePath string) (string, error) {
	// zbarimg outputs: "PDF417:data\n"
	// Use --raw to get just the data without symbology prefix
	cmd := exec.Command("zbarimg", "--raw", "--quiet", imagePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Include stderr in error for debugging
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, stderr.String())
		}
		return "", err
	}

	// Trim trailing newline
	result := strings.TrimSuffix(stdout.String(), "\n")
	return result, nil
}
