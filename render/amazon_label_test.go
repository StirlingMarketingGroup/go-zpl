package render

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// TestAmazonShippingLabel covers the four renderer gaps fixed for GitHub issue #17:
//  1. Setup-only ^XA…^XZ blocks are not emitted as blank pages
//  2. ^FB field blocks honor \& hard breaks (and wrap)
//  3. ^FH before a barcode ^FD is consumed (mid DataMatrix)
//  4. ^GFA compressed ASCII decompresses to plain hex at parse time
func TestAmazonShippingLabel(t *testing.T) {
	zplPath := filepath.Join("..", "testdata", "amazon-shipping-label.zpl")
	zplData, err := os.ReadFile(zplPath)
	if err != nil {
		t.Fatalf("read ZPL: %v", err)
	}

	labels, err := zpl.ParseAll(string(zplData))
	if err != nil {
		t.Fatalf("ParseAll: %v", err)
	}

	// Fix 1: setup-only block filtered — exactly one label
	if len(labels) != 1 {
		t.Fatalf("expected 1 label (setup block filtered), got %d", len(labels))
	}
	label := labels[0]

	// Structural asserts on parsed commands
	var (
		dataMatrices []*zpl.BarcodeDataMatrix
		graphics     []*zpl.GraphicField
		addressFD    *zpl.FieldData
	)
	for _, cmd := range label.Commands() {
		switch v := cmd.(type) {
		case *zpl.BarcodeDataMatrix:
			dataMatrices = append(dataMatrices, v)
		case *zpl.GraphicField:
			graphics = append(graphics, v)
		case *zpl.FieldData:
			// Address field: multi-line with \& hard breaks after hex decode
			if strings.Contains(v.Data, `\&`) {
				addressFD = v
			}
		}
	}

	// Fix 3: 5 DataMatrix barcodes — four plain + one hex-decoded tracking
	if len(dataMatrices) != 5 {
		t.Fatalf("expected 5 DataMatrix barcodes, got %d", len(dataMatrices))
	}
	var (
		plainCount int
		tbaCount   int
	)
	for _, dm := range dataMatrices {
		switch dm.Data {
		case "SFjDW0Y0rv_001_v":
			plainCount++
		case "TBA333092102660":
			tbaCount++
		default:
			t.Errorf("unexpected DataMatrix data %q", dm.Data)
		}
	}
	if plainCount != 4 {
		t.Errorf("expected 4 DataMatrix with SFjDW0Y0rv_001_v, got %d", plainCount)
	}
	if tbaCount != 1 {
		t.Errorf("expected 1 DataMatrix with TBA333092102660 (hex-decoded via ^FH), got %d", tbaCount)
	}

	// Fix 4: 3 GraphicFields with expanded ASCII hex of TotalBytes*2 nibbles
	if len(graphics) != 3 {
		t.Fatalf("expected 3 GraphicFields, got %d", len(graphics))
	}
	wantHexLens := map[int]bool{2068 * 2: true, 510 * 2: true, 516 * 2: true}
	gotHexLens := make(map[int]int)
	for _, gf := range graphics {
		n := len(gf.Data)
		gotHexLens[n]++
		if !wantHexLens[n] {
			t.Errorf("GraphicField expanded hex length %d not in {4136,1020,1032}; TotalBytes=%d DataBytes=%d BytesPerRow=%d",
				n, gf.TotalBytes, gf.DataBytes, gf.BytesPerRow)
		}
		// Data must be plain hex (no compression count letters remaining)
		for _, r := range gf.Data {
			isHex := (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
			if !isHex {
				t.Errorf("GraphicField Data still contains non-hex rune %q", r)
				break
			}
		}
	}
	for want := range wantHexLens {
		if gotHexLens[want] != 1 {
			t.Errorf("expected exactly one GraphicField with hex length %d, got %d", want, gotHexLens[want])
		}
	}

	// Fix 2: address FieldData retains literal \& so render-side split is exercised
	if addressFD == nil {
		t.Fatal("expected an address FieldData containing literal \\& after parse")
	}
	if !strings.Contains(addressFD.Data, `\&`) {
		t.Error("address FieldData should retain literal \\& for render-side line splitting")
	}
	// Soft check: decoded address should include expected content
	for _, want := range []string{"Capitol Theatre", "John Warner", "149 Westchester", "Port Chester"} {
		if !strings.Contains(addressFD.Data, want) {
			t.Errorf("address FieldData missing %q; got %q", want, addressFD.Data)
		}
	}

	// Render
	renderer := New(zpl.DPI203).WithIgnoreLabelHome(true)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 812 || bounds.Dy() != 1218 {
		t.Errorf("expected image size 812x1218, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	blackPixels := countBlackPixels(img)
	if blackPixels < 10000 {
		t.Errorf("rendered image looks blank: only %d black pixels (want > 10000)", blackPixels)
	}

	// Mid DataMatrix at ^FO575,545^BXN,8,200,22,22 is 22 modules × 8 = 176×176 px.
	// Assert black in the right-edge region past the old auto 16×8=128 px footprint
	// (old right edge at x≈575+128=703; forced size extends to x≈575+176=751).
	if !regionHasBlack(img, 640, 750, 545, 721) {
		t.Error("mid DataMatrix right-edge region [640,750]×[545,721] has no black pixels; " +
			"expected 22×22 symbol (176px) at ^FO575,545 to extend past old 128px footprint")
	}

	// Golden-file check
	goldenPath := filepath.Join("..", "testdata", "visual", "amazon-shipping-label.png")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := writePNG(goldenPath, img); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated golden %s", goldenPath)
		return
	}

	goldenFile, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("open golden %s: %v (set UPDATE_GOLDEN=1 to create)", goldenPath, err)
	}
	defer goldenFile.Close()

	golden, err := png.Decode(goldenFile)
	if err != nil {
		t.Fatalf("decode golden: %v", err)
	}

	if !imagesEqual(img, golden) {
		// Write actual for debugging
		actualPath := filepath.Join("..", "testdata", "visual", "amazon-shipping-label-actual.png")
		_ = writePNG(actualPath, img)
		t.Errorf("rendered image differs from golden %s (wrote actual to %s)", goldenPath, actualPath)
	}
}

func countBlackPixels(img image.Image) int {
	bounds := img.Bounds()
	n := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				n++
			}
		}
	}
	return n
}

// regionHasBlack reports whether any black pixel exists in [x0,x1]×[y0,y1] (inclusive).
func regionHasBlack(img image.Image, x0, x1, y0, y1 int) bool {
	bounds := img.Bounds()
	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	if x1 >= bounds.Max.X {
		x1 = bounds.Max.X - 1
	}
	if y1 >= bounds.Max.Y {
		y1 = bounds.Max.Y - 1
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				return true
			}
		}
	}
	return false
}

func imagesEqual(a, b image.Image) bool {
	ab, bb := a.Bounds(), b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return false
	}
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			aR, aG, aB, aA := a.At(ab.Min.X+x, ab.Min.Y+y).RGBA()
			bR, bG, bB, bA := b.At(bb.Min.X+x, bb.Min.Y+y).RGBA()
			if aR != bR || aG != bG || aB != bB || aA != bA {
				return false
			}
		}
	}
	return true
}

func writePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
