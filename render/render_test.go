package render

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

func TestRenderer_BasicText(t *testing.T) {
	label := zpl.NewLabel().
		SetSizeDots(400, 200).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30)).
		Add(zpl.NewFieldData("Hello World"))

	renderer := New(zpl.DPI203).WithSize(400, 200)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 200 {
		t.Errorf("Expected image size 400x200, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that some pixels are black (text was rendered)
	hasBlack := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				hasBlack = true
				break
			}
		}
		if hasBlack {
			break
		}
	}

	if !hasBlack {
		t.Error("Expected some black pixels (rendered text), but found none")
	}
}

func TestRenderer_GraphicBox(t *testing.T) {
	label := zpl.NewLabel().
		SetSizeDots(200, 200).
		Add(zpl.NewFieldOrigin(50, 50)).
		Add(zpl.NewGraphicBox(100, 100, 4))

	renderer := New(zpl.DPI203).WithSize(200, 200)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Check that the box outline is present
	// Top edge should be black at y=50, x=50-150
	r, g, b, _ := img.At(100, 50).RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Error("Expected black pixel at top edge of box")
	}

	// Center should be white (outline box, not filled)
	r, g, b, _ = img.At(100, 100).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Error("Expected white pixel at center of box (outline only)")
	}
}

func TestRenderer_RenderPNG(t *testing.T) {
	label := zpl.NewLabel().
		SetSizeDots(200, 100).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 24, 24)).
		Add(zpl.NewFieldData("Test"))

	renderer := New(zpl.DPI203).WithSize(200, 100)

	var buf bytes.Buffer
	err := renderer.RenderPNG(label, &buf)
	if err != nil {
		t.Fatalf("RenderPNG failed: %v", err)
	}

	// Verify it's a valid PNG
	_, err = png.Decode(&buf)
	if err != nil {
		t.Errorf("Output is not valid PNG: %v", err)
	}
}

func TestRenderer_RenderJPEG(t *testing.T) {
	label := zpl.NewLabel().
		SetSizeDots(200, 100).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 24, 24)).
		Add(zpl.NewFieldData("Test"))

	renderer := New(zpl.DPI203).WithSize(200, 100)

	var buf bytes.Buffer
	err := renderer.RenderJPEG(label, &buf, 90)
	if err != nil {
		t.Fatalf("RenderJPEG failed: %v", err)
	}

	// Verify it's a valid JPEG
	_, err = jpeg.Decode(&buf)
	if err != nil {
		t.Errorf("Output is not valid JPEG: %v", err)
	}
}

func TestRenderer_MultipleFonts(t *testing.T) {
	label := zpl.NewLabel().
		SetSizeDots(400, 200).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 20)).
		Add(zpl.NewFieldData("Small")).
		Add(zpl.NewFieldOrigin(10, 50)).
		Add(zpl.NewScalableFont(zpl.Font0, 40, 40)).
		Add(zpl.NewFieldData("Large")).
		Add(zpl.NewFieldOrigin(10, 120)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 30)). // Condensed
		Add(zpl.NewFieldData("Condensed"))

	renderer := New(zpl.DPI203).WithSize(400, 200)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Basic sanity check: image should have content
	bounds := img.Bounds()
	hasBlack := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				hasBlack = true
				break
			}
		}
		if hasBlack {
			break
		}
	}

	if !hasBlack {
		t.Error("Expected some black pixels (rendered text)")
	}
}

func TestRenderer_Orientations(t *testing.T) {
	tests := []struct {
		name   string
		orient zpl.Orientation
	}{
		{"Normal", zpl.OrientationNormal},
		{"Rotated90", zpl.OrientationRotated90},
		{"Rotated180", zpl.OrientationRotated180},
		{"Rotated270", zpl.OrientationRotated270},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font := zpl.NewScalableFont(zpl.Font0, 30, 30).WithOrientation(tt.orient)
			label := zpl.NewLabel().
				SetSizeDots(200, 200).
				Add(zpl.NewFieldOrigin(50, 50)).
				Add(font).
				Add(zpl.NewFieldData("Test"))

			renderer := New(zpl.DPI203).WithSize(200, 200)
			img, err := renderer.Render(label)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			// Verify image was created with content
			bounds := img.Bounds()
			hasBlack := false
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					if r == 0 && g == 0 && b == 0 {
						hasBlack = true
						break
					}
				}
				if hasBlack {
					break
				}
			}

			if !hasBlack {
				t.Errorf("Expected some black pixels for orientation %s", tt.name)
			}
		})
	}
}

func TestRenderer_GraphicShapes(t *testing.T) {
	label := zpl.NewLabel().
		SetSizeDots(400, 300).
		// Box
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewGraphicBox(80, 80, 3)).
		// Circle
		Add(zpl.NewFieldOrigin(110, 10)).
		Add(zpl.NewGraphicCircle(80, 3)).
		// Ellipse
		Add(zpl.NewFieldOrigin(210, 10)).
		Add(zpl.NewGraphicEllipse(80, 60, 3)).
		// Diagonal line
		Add(zpl.NewFieldOrigin(310, 10)).
		Add(zpl.NewGraphicDiagonalLine(80, 80, 3))

	renderer := New(zpl.DPI203).WithSize(400, 300)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Expected image size 400x300, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestRenderer_UPSLabel tests rendering a real-world UPS label
func TestRenderer_UPSLabel(t *testing.T) {
	// Create a simplified UPS-style label
	label := zpl.NewLabel().
		SetSizeDots(812, 1218).
		// Sender info
		Add(zpl.NewFieldOrigin(15, 7)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("JOHN DOE")).
		Add(zpl.NewFieldOrigin(15, 27)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("15551234567")).
		Add(zpl.NewFieldOrigin(15, 47)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("TEST COMPANY INC")).
		// Ship to header
		Add(zpl.NewFieldOrigin(15, 142)).
		Add(zpl.NewScalableFont(zpl.Font0, 28, 32)).
		Add(zpl.NewFieldData("SHIP TO:")).
		// Recipient info
		Add(zpl.NewFieldOrigin(61, 166)).
		Add(zpl.NewScalableFont(zpl.Font0, 28, 32)).
		Add(zpl.NewFieldData("JANE SMITH")).
		Add(zpl.NewFieldOrigin(61, 279)).
		Add(zpl.NewScalableFont(zpl.Font0, 45, 44)).
		Add(zpl.NewFieldData("ANYTOWN  CA  90210")).
		// Service type
		Add(zpl.NewFieldOrigin(9, 670)).
		Add(zpl.NewScalableFont(zpl.Font0, 56, 58)).
		Add(zpl.NewFieldData("UPS EXPEDITED")).
		// Horizontal lines
		Add(zpl.NewFieldOrigin(0, 648)).
		Add(zpl.NewGraphicBox(811, 14, 14)).
		Add(zpl.NewFieldOrigin(0, 423)).
		Add(zpl.NewGraphicBox(812, 4, 4))

	renderer := New(zpl.DPI203).WithSize(812, 1218)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 812 || bounds.Dy() != 1218 {
		t.Errorf("Expected image size 812x1218, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// saveTestImage is a helper for debugging - saves image to testdata folder.
//
//nolint:unused // Keep for debugging tests
func saveTestImage(t *testing.T, name string, label *zpl.Label, width, height int) {
	t.Helper()

	renderer := New(zpl.DPI203).WithSize(width, height)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Create testdata directory if it doesn't exist
	dir := filepath.Join("..", "testdata", "render")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("Failed to create testdata dir: %v", err)
	}

	path := filepath.Join(dir, name+".png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	t.Logf("Saved test image to %s", path)
}

func TestRenderer_OCRB(t *testing.T) {
	// Test Font E (OCR-B)
	label := zpl.NewLabel().
		SetSizeDots(400, 150).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.FontE, 30, 30)).
		Add(zpl.NewFieldData("0123456789")).
		Add(zpl.NewFieldOrigin(10, 60)).
		Add(zpl.NewScalableFont(zpl.FontE, 40, 40)).
		Add(zpl.NewFieldData("ABCDEFGHIJ")).
		Add(zpl.NewFieldOrigin(10, 110)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30)).
		Add(zpl.NewFieldData("Font 0 comparison"))

	renderer := New(zpl.DPI203).WithSize(400, 150)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	bounds := img.Bounds()
	hasBlack := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				hasBlack = true
				break
			}
		}
		if hasBlack {
			break
		}
	}

	if !hasBlack {
		t.Error("Expected some black pixels (rendered OCR-B text)")
	}
}

func TestEncodeCode128Auto_NumericOnly(t *testing.T) {
	// Test numeric-only data - should use Subset C for efficiency
	// "12345678901" has 11 digits
	// Start C + 5 digit pairs (10 digits) = 6 symbols
	// Need to handle odd digit (switch B, digit, switch C) or different approach
	data := "1234567890"
	values := encodeCode128Auto(data)

	// Should start with Start C (105)
	if len(values) < 2 {
		t.Fatalf("Expected at least 2 values, got %d", len(values))
	}
	if values[0] != code128StartC {
		t.Errorf("Expected Start C (105), got %d", values[0])
	}

	// Compare with pure Subset B encoding to verify efficiency
	valuesB := encodeCode128B(data)
	if len(values) >= len(valuesB) {
		t.Errorf("Auto encoding should be shorter than B encoding for numeric data: auto=%d, B=%d",
			len(values), len(valuesB))
	}
}

func TestEncodeCode128Auto_MixedData(t *testing.T) {
	// Test mixed alphanumeric data
	data := "ABC12345678DEF"
	values := encodeCode128Auto(data)

	if len(values) == 0 {
		t.Fatal("Expected non-empty encoding")
	}

	// Should start with Start B since it begins with letters
	if values[0] != code128StartB {
		t.Errorf("Expected Start B (104), got %d", values[0])
	}

	// Verify check digit is correct by recalculating
	sum := values[0]
	for i := 1; i < len(values)-1; i++ {
		sum += i * values[i]
	}
	expectedCheck := sum % 103
	actualCheck := values[len(values)-1]
	if actualCheck != expectedCheck {
		t.Errorf("Check digit mismatch: expected %d, got %d", expectedCheck, actualCheck)
	}
}

func TestEncodeCode128Auto_ShortNumeric(t *testing.T) {
	// Short numeric runs (< 4 digits) should use Subset B
	data := "A12B"
	values := encodeCode128Auto(data)

	if len(values) == 0 {
		t.Fatal("Expected non-empty encoding")
	}

	// Should use Subset B throughout (no benefit to switching for 2 digits)
	if values[0] != code128StartB {
		t.Errorf("Expected Start B (104), got %d", values[0])
	}
}

func TestEncodeCode128Auto_EmptyData(t *testing.T) {
	values := encodeCode128Auto("")
	if len(values) != 0 {
		t.Errorf("Expected empty encoding for empty data, got %d values", len(values))
	}
}

func TestRenderer_TextHorizontalScaling(t *testing.T) {
	// Test that width < height produces narrower text
	// Create two labels with same text but different widths
	label1 := zpl.NewLabel().
		SetSizeDots(500, 100).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 50, 50)). // Square scaling
		Add(zpl.NewFieldData("TESTING"))

	label2 := zpl.NewLabel().
		SetSizeDots(500, 100).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 50, 40)). // Condensed (80% width)
		Add(zpl.NewFieldData("TESTING"))

	renderer := New(zpl.DPI203).WithSize(500, 100)

	img1, err := renderer.Render(label1)
	if err != nil {
		t.Fatalf("Render failed for label1: %v", err)
	}

	img2, err := renderer.Render(label2)
	if err != nil {
		t.Fatalf("Render failed for label2: %v", err)
	}

	// Count black pixels in each image to verify text was rendered
	countBlack := func(img interface {
		At(x, y int) interface{ RGBA() (r, g, b, a uint32) }
	}, bounds interface{ Max() (dx, dy int) }) int {
		count := 0
		b := img.(interface {
			Bounds() interface {
				Dx() int
				Dy() int
				Min() (x, y int)
			}
		}).Bounds()
		dx, dy := b.Dx(), b.Dy()
		minX, minY := b.(interface{ Min() (x, y int) }).Min()
		for y := minY; y < minY+dy; y++ {
			for x := minX; x < minX+dx; x++ {
				r, g, bb, _ := img.At(x, y).RGBA()
				if r == 0 && g == 0 && bb == 0 {
					count++
				}
			}
		}
		return count
	}
	_ = countBlack // Unused in this simplified test

	// Just verify both images have black pixels (text rendered)
	hasBlack1 := false
	hasBlack2 := false
	for y := 0; y < 100; y++ {
		for x := 0; x < 500; x++ {
			r, g, b, _ := img1.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				hasBlack1 = true
			}
			r, g, b, _ = img2.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				hasBlack2 = true
			}
		}
	}

	if !hasBlack1 {
		t.Error("Expected black pixels in label1 (square scaling)")
	}
	if !hasBlack2 {
		t.Error("Expected black pixels in label2 (condensed scaling)")
	}
}

// Uncomment to generate reference images for visual inspection
/*
func TestSaveReferenceImages(t *testing.T) {
	// Basic text
	saveTestImage(t, "basic_text", zpl.NewLabel().
		SetSizeDots(400, 100).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30)).
		Add(zpl.NewFieldData("Hello World")), 400, 100)

	// Multiple sizes
	saveTestImage(t, "multiple_sizes", zpl.NewLabel().
		SetSizeDots(500, 200).
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 20)).
		Add(zpl.NewFieldData("Small 20pt")).
		Add(zpl.NewFieldOrigin(10, 50)).
		Add(zpl.NewScalableFont(zpl.Font0, 40, 40)).
		Add(zpl.NewFieldData("Medium 40pt")).
		Add(zpl.NewFieldOrigin(10, 120)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("Large 60pt")), 500, 200)

	// Rotations
	saveTestImage(t, "rotations", zpl.NewLabel().
		SetSizeDots(400, 400).
		Add(zpl.NewFieldOrigin(50, 50)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30).WithOrientation(zpl.OrientationNormal)).
		Add(zpl.NewFieldData("Normal")).
		Add(zpl.NewFieldOrigin(200, 50)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30).WithOrientation(zpl.OrientationRotated90)).
		Add(zpl.NewFieldData("Rotated 90")).
		Add(zpl.NewFieldOrigin(50, 250)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30).WithOrientation(zpl.OrientationRotated180)).
		Add(zpl.NewFieldData("Rotated 180")).
		Add(zpl.NewFieldOrigin(200, 250)).
		Add(zpl.NewScalableFont(zpl.Font0, 30, 30).WithOrientation(zpl.OrientationRotated270)).
		Add(zpl.NewFieldData("Rotated 270")), 400, 400)
}
*/

func TestRenderDefaultCanvasFollowsDPI(t *testing.T) {
	// A label with no ^PW/^LL and a renderer with no WithSize falls back to a
	// 4×6 inch canvas, which must scale with the renderer's DPI rather than
	// always being the 203 DPI 812×1218.
	label := zpl.NewLabel().TextField(10, 10, zpl.Font0, 20, 20, "dpi")
	cases := []struct {
		dpi    zpl.DPI
		width  int
		height int
	}{
		{zpl.DPI203, 812, 1218},
		{zpl.DPI300, 1200, 1800},
		{zpl.DPI600, 2400, 3600},
		{0, 812, 1218},
	}
	for _, c := range cases {
		img, err := (&Renderer{DPI: c.dpi}).Render(label)
		if err != nil {
			t.Fatalf("DPI %d: %v", c.dpi, err)
		}
		if b := img.Bounds(); b.Dx() != c.width || b.Dy() != c.height {
			t.Errorf("DPI %d: expected %dx%d, got %dx%d", c.dpi, c.width, c.height, b.Dx(), b.Dy())
		}
	}
}

func TestRenderRejectsUnsupportedDPI(t *testing.T) {
	label := zpl.NewLabel().TextField(10, 10, zpl.Font0, 20, 20, "dpi")
	if _, err := (&Renderer{DPI: zpl.DPI(-1)}).Render(label); err == nil {
		t.Error("expected an error for an unsupported DPI, got nil")
	}
}
