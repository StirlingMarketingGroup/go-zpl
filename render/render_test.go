package render

import (
	"bytes"
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
