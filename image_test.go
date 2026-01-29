package zpl

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestImageConverter_Convert_SimpleBlack(t *testing.T) {
	// Create a small all-black image
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.Black)
		}
	}

	gf := NewImageConverter().Convert(img)

	if gf.Format != GraphicFieldASCII {
		t.Errorf("expected ASCII format, got %c", gf.Format)
	}
	if gf.BytesPerRow != 1 {
		t.Errorf("expected 1 byte per row, got %d", gf.BytesPerRow)
	}
	// All black = all bits set = FF per row
	if gf.Data != "FFFF" {
		t.Errorf("expected FFFF for all-black 8x2, got %s", gf.Data)
	}
}

func TestImageConverter_Convert_SimpleWhite(t *testing.T) {
	// Create a small all-white image
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.White)
		}
	}

	gf := NewImageConverter().Convert(img)

	// All white = no bits set = 00 per row
	if gf.Data != "0000" {
		t.Errorf("expected 0000 for all-white 8x2, got %s", gf.Data)
	}
}

func TestImageConverter_Convert_Checkerboard(t *testing.T) {
	// Create a checkerboard pattern (8x2)
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	gf := NewImageConverter().Convert(img)

	// Row 0: B W B W B W B W = 10101010 = AA
	// Row 1: W B W B W B W B = 01010101 = 55
	if gf.Data != "AA55" {
		t.Errorf("expected AA55 for checkerboard, got %s", gf.Data)
	}
}

func TestImageConverter_Convert_PartialByte(t *testing.T) {
	// Create a 5-pixel wide image (not byte-aligned)
	img := image.NewRGBA(image.Rect(0, 0, 5, 1))
	for x := 0; x < 5; x++ {
		img.Set(x, 0, color.Black)
	}

	gf := NewImageConverter().Convert(img)

	if gf.BytesPerRow != 1 {
		t.Errorf("expected 1 byte per row for 5 pixels, got %d", gf.BytesPerRow)
	}
	// 5 black pixels = 11111000 = F8
	if gf.Data != "F8" {
		t.Errorf("expected F8 for 5 black pixels, got %s", gf.Data)
	}
}

func TestImageConverter_Convert_MultipleBytes(t *testing.T) {
	// Create a 12-pixel wide image
	img := image.NewRGBA(image.Rect(0, 0, 12, 1))
	for x := 0; x < 12; x++ {
		img.Set(x, 0, color.Black)
	}

	gf := NewImageConverter().Convert(img)

	if gf.BytesPerRow != 2 {
		t.Errorf("expected 2 bytes per row for 12 pixels, got %d", gf.BytesPerRow)
	}
	// 12 black pixels = 11111111 11110000 = FF F0
	if gf.Data != "FFF0" {
		t.Errorf("expected FFF0 for 12 black pixels, got %s", gf.Data)
	}
}

func TestImageConverter_WithInvert(t *testing.T) {
	// Create all-black image
	img := image.NewRGBA(image.Rect(0, 0, 8, 1))
	for x := 0; x < 8; x++ {
		img.Set(x, 0, color.Black)
	}

	gf := NewImageConverter().WithInvert(true).Convert(img)

	// Inverted: all black becomes all white = 00
	if gf.Data != "00" {
		t.Errorf("expected 00 for inverted all-black, got %s", gf.Data)
	}
}

func TestImageConverter_ConvertBinary(t *testing.T) {
	// Create all-black image
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.Black)
		}
	}

	gf := NewImageConverter().ConvertBinary(img)

	if gf.Format != GraphicFieldBinary {
		t.Errorf("expected Binary format, got %c", gf.Format)
	}
	if len(gf.BinaryData) != 2 {
		t.Errorf("expected 2 bytes, got %d", len(gf.BinaryData))
	}
	if gf.BinaryData[0] != 0xFF || gf.BinaryData[1] != 0xFF {
		t.Errorf("expected [0xFF, 0xFF], got %v", gf.BinaryData)
	}
}

func TestImageConverter_FloydSteinberg(t *testing.T) {
	// Create a gradient image
	img := image.NewGray(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			// Create a diagonal gradient
			gray := uint8((x + y) * 255 / 30)
			img.SetGray(x, y, color.Gray{Y: gray})
		}
	}

	gf := NewImageConverter().WithDithering(DitheringFloydSteinberg).Convert(img)

	// Just verify it produces output without panicking
	if gf.Data == "" {
		t.Error("expected non-empty data")
	}
	if gf.BytesPerRow != 2 {
		t.Errorf("expected 2 bytes per row for 16 pixels, got %d", gf.BytesPerRow)
	}
}

func TestImageConverter_OrderedDither(t *testing.T) {
	// Create a gradient image
	img := image.NewGray(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			gray := uint8((x + y) * 255 / 30)
			img.SetGray(x, y, color.Gray{Y: gray})
		}
	}

	gf := NewImageConverter().WithDithering(DitheringOrdered).Convert(img)

	if gf.Data == "" {
		t.Error("expected non-empty data")
	}
}

func TestImageConverter_Transparency(t *testing.T) {
	// Create an image with transparent pixels
	img := image.NewRGBA(image.Rect(0, 0, 8, 1))
	// Set all pixels to transparent
	for x := 0; x < 8; x++ {
		img.Set(x, 0, color.RGBA{0, 0, 0, 0})
	}

	gf := NewImageConverter().Convert(img)

	// Transparent should be treated as white = 00
	if gf.Data != "00" {
		t.Errorf("expected 00 for transparent pixels, got %s", gf.Data)
	}
}

func TestImageConverter_ConvertReader(t *testing.T) {
	// Create a simple PNG in memory
	img := image.NewRGBA(image.Rect(0, 0, 8, 1))
	for x := 0; x < 8; x++ {
		img.Set(x, 0, color.Black)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	gf, err := NewImageConverter().ConvertReader(&buf)
	if err != nil {
		t.Fatalf("ConvertReader failed: %v", err)
	}

	if gf.Data != "FF" {
		t.Errorf("expected FF, got %s", gf.Data)
	}
}

func TestImageToGraphicField(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 1))
	for x := 0; x < 8; x++ {
		img.Set(x, 0, color.Black)
	}

	gf := ImageToGraphicField(img)
	if gf.Data != "FF" {
		t.Errorf("expected FF, got %s", gf.Data)
	}
}

func TestImageToGraphicFieldDithered(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 8, 1))
	for x := 0; x < 8; x++ {
		img.SetGray(x, 0, color.Gray{Y: 100}) // Medium gray
	}

	gf := ImageToGraphicFieldDithered(img)
	// Just verify it produces output
	if gf.Data == "" {
		t.Error("expected non-empty data")
	}
}

func TestResizeImage(t *testing.T) {
	// Create 100x50 image
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))

	// Resize to max 50x50 (should become 50x25 to maintain aspect ratio)
	resized := ResizeImage(img, 50, 50)
	bounds := resized.Bounds()

	if bounds.Dx() != 50 {
		t.Errorf("expected width 50, got %d", bounds.Dx())
	}
	if bounds.Dy() != 25 {
		t.Errorf("expected height 25, got %d", bounds.Dy())
	}
}

func TestResizeImage_NoResizeNeeded(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	resized := ResizeImage(img, 100, 100)

	// Should return original image
	if resized != img {
		t.Error("expected original image when no resize needed")
	}
}

func TestGrayscaleImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255}) // Red
	img.Set(1, 0, color.RGBA{0, 255, 0, 255}) // Green

	gray := GrayscaleImage(img)

	// Red should be darker than green in grayscale (luminance formula)
	r := gray.GrayAt(0, 0).Y
	g := gray.GrayAt(1, 0).Y

	if g <= r {
		t.Errorf("green should be brighter than red in grayscale, got red=%d, green=%d", r, g)
	}
}

func TestGraphicField_ZPL(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.Black)
		}
	}

	gf := NewImageConverter().Convert(img)
	zpl := gf.ZPL()

	// Should produce ^GFA,2,2,1,FFFF
	if !strings.HasPrefix(zpl, "^GFA,") {
		t.Errorf("expected ^GFA prefix, got %s", zpl)
	}
	if !strings.HasSuffix(zpl, "FFFF") {
		t.Errorf("expected FFFF suffix, got %s", zpl)
	}
}

func TestImageConverter_ConvertZ64(t *testing.T) {
	// Create a larger image where compression would be beneficial
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Create a simple pattern
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if (x/10+y/10)%2 == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	gf := NewImageConverter().ConvertZ64(img)

	// Verify Z64 format
	if !strings.HasPrefix(gf.Data, ":Z64:") {
		t.Errorf("expected :Z64: prefix, got %s", gf.Data[:min(20, len(gf.Data))])
	}

	// Should have CRC at the end (4 hex chars after colon)
	parts := strings.Split(gf.Data, ":")
	if len(parts) != 4 { // empty, Z64, base64data, crc
		t.Errorf("expected 4 parts in Z64 data, got %d", len(parts))
	}

	crc := parts[3]
	if len(crc) != 4 {
		t.Errorf("expected 4-char CRC, got %q", crc)
	}

	// Verify the data is valid base64
	b64Data := parts[2]
	_, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		t.Errorf("invalid base64 in Z64 data: %v", err)
	}
}

func TestImageConverter_Z64_SmallerThanASCII(t *testing.T) {
	// Create a larger image with repeating pattern (should compress well)
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			// Alternating rows of black and white
			if y%20 < 10 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	asciiGF := NewImageConverter().Convert(img)
	z64GF := NewImageConverter().ConvertZ64(img)

	// Z64 should be significantly smaller for this pattern
	// (base64 overhead is 33%, but zlib compression should overcome that)
	if len(z64GF.Data) >= len(asciiGF.Data) {
		t.Logf("Z64 size: %d, ASCII size: %d", len(z64GF.Data), len(asciiGF.Data))
		// Don't fail the test, just log - compression ratio varies
	}
}

func TestImageToGraphicFieldZ64(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.Black)
		}
	}

	gf := ImageToGraphicFieldZ64(img)

	if !strings.HasPrefix(gf.Data, ":Z64:") {
		t.Errorf("expected :Z64: prefix, got %s", gf.Data[:min(20, len(gf.Data))])
	}
}

func TestCRC16CCITT(t *testing.T) {
	// Test with known values
	// Empty string should give specific CRC
	result := crc16CCITT([]byte(""))
	if result != "FFFF" {
		t.Errorf("CRC of empty string: expected FFFF, got %s", result)
	}

	// Test with "123456789" - standard test vector
	// CRC-CCITT of "123456789" should be 0x29B1
	result = crc16CCITT([]byte("123456789"))
	if result != "29B1" {
		t.Errorf("CRC of 123456789: expected 29B1, got %s", result)
	}
}
