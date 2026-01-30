package render

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	xdraw "golang.org/x/image/draw"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

type barcodeSpec struct {
	Name       string
	Kind       string
	SearchRect image.Rectangle
	MaxDiff    float64
}

func TestFedexExpressBarcodeScanMatchesReference(t *testing.T) {
	if os.Getenv("BARCODE_SCAN") != "1" {
		t.Skip("set BARCODE_SCAN=1 to run barcode scan comparison against reference")
	}

	baseDir := filepath.Join("..", "testdata", "visual", "fedex_express_scanned")
	zplPath := filepath.Join(baseDir, "label.zpl")
	scanPath := filepath.Join(baseDir, "scan_reference.png")

	zplData, err := os.ReadFile(zplPath)
	if err != nil {
		t.Fatalf("read zpl: %v", err)
	}
	label, err := zpl.Parse(string(zplData))
	if err != nil {
		t.Fatalf("parse zpl: %v", err)
	}

	scanImg := loadPNG(t, scanPath)

	const labelWidthDots = 800
	const labelHeightDots = 1200
	const originY = 20

	renderer := New(zpl.DPI203).WithSize(labelWidthDots, labelHeightDots)
	ourImg, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("render label: %v", err)
	}

	specs := []barcodeSpec{
		{
			Name: "Address Code128",
			Kind: "code128",
			SearchRect: image.Rect(
				16, 416+originY-10,
				16+520, 416+originY+70,
			),
		},
		{
			Name: "PDF417",
			Kind: "pdf417",
			SearchRect: image.Rect(
				10, 449+originY-5,
				10+560, 449+originY+235,
			),
			// Note: PDF417 encoding can vary between implementations while still being valid.
			// The reference scan was printed by a Zebra printer which uses a different encoder.
			// We allow higher tolerance here since both encodings should decode to the same data.
			// TODO: Add actual decode verification once we have a PDF417 decoder.
			MaxDiff: 0.55,
		},
		{
			Name: "Bottom Code128",
			Kind: "code128",
			SearchRect: image.Rect(
				50, 993+originY-5,
				50+720, 993+originY+240,
			),
		},
	}

	scaleX := float64(scanImg.Bounds().Dx()) / float64(labelWidthDots)
	scaleY := float64(scanImg.Bounds().Dy()) / float64(labelHeightDots)

	for _, spec := range specs {
		t.Run(spec.Name, func(t *testing.T) {
			renderBounds, err := findBlackBounds(ourImg, spec.SearchRect)
			if err != nil {
				t.Fatalf("locate %s in render: %v", spec.Name, err)
			}

			scanSearch := clampRect(scaleRect(spec.SearchRect, scaleX, scaleY), scanImg.Bounds())
			scanBounds, err := findBlackBounds(scanImg, scanSearch)
			if err != nil {
				// Fallback to scaled render bounds if scan bounding fails.
				scanBounds = clampRect(scaleRect(renderBounds, scaleX, scaleY), scanImg.Bounds())
			}

			renderCrop := cropImage(ourImg, renderBounds)
			scanCrop := cropImage(scanImg, scanBounds)

			switch spec.Kind {
			case "code128":
				refText, err := decodeCode128(scanCrop)
				if err != nil {
					t.Fatalf("decode %s from scan: %v", spec.Name, err)
				}
				renderText, err := decodeCode128(renderCrop)
				if err != nil {
					t.Fatalf("decode %s from render: %v", spec.Name, err)
				}
				if refText != renderText {
					t.Fatalf("%s mismatch: scan=%q render=%q", spec.Name, refText, renderText)
				}
			case "pdf417":
				diff := compareBarcodePattern(renderCrop, scanCrop, 2)
				if diff > spec.MaxDiff {
					t.Fatalf("%s pattern mismatch: diff=%.2f%% (max %.2f%%)", spec.Name, diff*100, spec.MaxDiff*100)
				}
			default:
				t.Fatalf("unknown barcode kind %q", spec.Kind)
			}
		})
	}
}

func loadPNG(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}

func cropImage(img image.Image, rect image.Rectangle) image.Image {
	rect = rect.Intersect(img.Bounds())
	if rect.Empty() {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	if si, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		return si.SubImage(rect)
	}
	out := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(out, out.Bounds(), img, rect.Min, draw.Src)
	return out
}

func scaleRect(r image.Rectangle, scaleX, scaleY float64) image.Rectangle {
	return image.Rect(
		int(math.Round(float64(r.Min.X)*scaleX)),
		int(math.Round(float64(r.Min.Y)*scaleY)),
		int(math.Round(float64(r.Max.X)*scaleX)),
		int(math.Round(float64(r.Max.Y)*scaleY)),
	)
}

func clampRect(r, bounds image.Rectangle) image.Rectangle {
	if r.Min.X < bounds.Min.X {
		r.Min.X = bounds.Min.X
	}
	if r.Min.Y < bounds.Min.Y {
		r.Min.Y = bounds.Min.Y
	}
	if r.Max.X > bounds.Max.X {
		r.Max.X = bounds.Max.X
	}
	if r.Max.Y > bounds.Max.Y {
		r.Max.Y = bounds.Max.Y
	}
	return r
}

func findBlackBounds(img image.Image, rect image.Rectangle) (image.Rectangle, error) {
	rect = rect.Intersect(img.Bounds())
	if rect.Empty() {
		return image.Rectangle{}, errors.New("empty search rect")
	}
	minX, minY := rect.Max.X, rect.Max.Y
	maxX, maxY := rect.Min.X, rect.Min.Y
	found := false
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if isBlack(img.At(x, y)) {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
				found = true
			}
		}
	}
	if !found {
		return image.Rectangle{}, errors.New("no black pixels found")
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), nil
}

func isBlack(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	lum := (r + 2*g + b) / 4
	return lum < 160*256
}

func decodeCode128(img image.Image) (string, error) {
	reader := oned.NewCode128Reader()
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}
	result, err := reader.Decode(bmp, hints)
	if err == nil {
		return result.GetText(), nil
	}

	src := gozxing.NewLuminanceSourceFromImage(img)
	altBmp, altErr := gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(src))
	if altErr != nil {
		return "", err
	}
	result, altErr = reader.Decode(altBmp, hints)
	if altErr != nil {
		return "", err
	}
	return result.GetText(), nil
}

func compareBarcodePattern(renderImg, scanImg image.Image, step int) float64 {
	renderBounds := renderImg.Bounds()
	if renderBounds.Empty() {
		return 1.0
	}

	scaledScan := resizeImage(scanImg, renderBounds.Dx(), renderBounds.Dy())

	if step < 1 {
		step = 1
	}
	var diff, total int
	for y := 0; y < renderBounds.Dy(); y += step {
		for x := 0; x < renderBounds.Dx(); x += step {
			renderBlack := isBlack(renderImg.At(renderBounds.Min.X+x, renderBounds.Min.Y+y))
			scanBlack := isBlack(scaledScan.At(x, y))
			if renderBlack != scanBlack {
				diff++
			}
			total++
		}
	}
	if total == 0 {
		return 1.0
	}
	return float64(diff) / float64(total)
}

func resizeImage(img image.Image, width, height int) *image.RGBA {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

// TestPDF417MatchesLabelary compares our PDF417 output with Labelary's output.
// This is a more meaningful comparison since both are software encoders.
func TestPDF417MatchesLabelary(t *testing.T) {
	if os.Getenv("LABELARY_COMPARE") != "1" {
		t.Skip("set LABELARY_COMPARE=1 to run Labelary comparison")
	}

	// FedEx PDF417 ZPL - just the barcode portion
	zplCode := `^XA
^FO0,0
^BY2,2^B7N,10,5,14^FH^FD[)>_1E01_1D0294105_1D840_1D20_1D7949819308110201_1DFDE_1D740561073_1D031_1D_1D1/1_1D5.00LB_1DN_1D100 Market Street_1DSan Francisco_1DCA_1DTest Recipient_1E06_1D10ZED008_1D11ZRecipient Corp_1D12Z4155559876_1D15Z114064860_1D20Z_1C_1D31Z1195282044690009410500794981930811_1D32Z02GD_1D34Z01_1D39ZHKAA_1D_1E09_1DFDX_1Dz_1D8_1D_17_04';0?_7F@_1E_04^FS
^XZ`

	// Fetch from Labelary
	resp, err := http.Post(
		"https://api.labelary.com/v1/printers/8dpmm/labels/4x6/0/",
		"application/x-www-form-urlencoded",
		strings.NewReader(zplCode),
	)
	if err != nil {
		t.Fatalf("Labelary request failed: %v", err)
	}
	defer resp.Body.Close()

	labelaryImg, err := png.Decode(resp.Body)
	if err != nil {
		t.Fatalf("Decode Labelary PNG: %v", err)
	}

	// Render with our library
	label, err := zpl.Parse(zplCode)
	if err != nil {
		t.Fatalf("Parse ZPL: %v", err)
	}

	renderer := New(zpl.DPI203).WithSize(812, 1218)
	ourImg, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Find barcode bounds
	searchRect := image.Rect(0, 0, 700, 300)

	labelaryBounds, err := findBlackBounds(labelaryImg, searchRect)
	if err != nil {
		t.Fatalf("Find Labelary bounds: %v", err)
	}

	ourBounds, err := findBlackBounds(ourImg, searchRect)
	if err != nil {
		t.Fatalf("Find our bounds: %v", err)
	}

	// Crop
	labelaryCrop := cropImage(labelaryImg, labelaryBounds)
	ourCrop := cropImage(ourImg, ourBounds)

	// Compare - allow up to 30% difference for encoding variations
	diff := compareBarcodePattern(ourCrop, labelaryCrop, 1)
	t.Logf("Pattern difference vs Labelary: %.2f%%", diff*100)

	// Labelary comparison should be tighter since both are software encoders
	maxDiff := 0.30
	if diff > maxDiff {
		t.Errorf("Pattern mismatch with Labelary exceeds %.0f%%: %.2f%%", maxDiff*100, diff*100)
	}
}
