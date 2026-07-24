package datamatrix

import (
	"image"
	"testing"
)

func TestEncodeWithSize_22x22(t *testing.T) {
	const data = "TBA333092102660"

	// Forced 22×22 (as used by Amazon shipping labels: ^BXN,8,200,22,22)
	dm, err := EncodeWithSize(data, 22, 22)
	if err != nil {
		t.Fatalf("EncodeWithSize(22,22): %v", err)
	}
	b := dm.Bounds()
	if b.Dx() != 22 || b.Dy() != 22 {
		t.Errorf("EncodeWithSize(22,22) Bounds = %dx%d, want 22x22", b.Dx(), b.Dy())
	}

	// Auto-encode picks the minimal symbol (16×16 for this payload)
	auto, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	ab := auto.Bounds()
	if ab.Dx() != 16 || ab.Dy() != 16 {
		t.Errorf("auto Encode Bounds = %dx%d, want 16x16 (documents size divergence)", ab.Dx(), ab.Dy())
	}
}

func TestEncodeWithSize_dataTooLarge(t *testing.T) {
	// 10×10 holds only a few codewords; this payload won't fit.
	_, err := EncodeWithSize("TBA333092102660", 10, 10)
	if err == nil {
		t.Fatal("expected error for data too large for 10x10")
	}
}

func TestEncodeWithSize_unknownSize(t *testing.T) {
	_, err := EncodeWithSize("hi", 17, 17) // odd sizes are not in ECC 200 table
	if err == nil {
		t.Fatal("expected error for unsupported size 17x17")
	}
}

func TestEncodeWithSize_matchesAutoWhenSameSize(t *testing.T) {
	const data = "SFjDW0Y0rv_001_v"

	auto, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	ab := auto.Bounds()
	// Auto chooses 18×18 for this payload
	if ab.Dx() != 18 || ab.Dy() != 18 {
		t.Fatalf("auto Encode Bounds = %dx%d, want 18x18 (test assumption)", ab.Dx(), ab.Dy())
	}

	forced, err := EncodeWithSize(data, 18, 18)
	if err != nil {
		t.Fatalf("EncodeWithSize(18,18): %v", err)
	}
	if !matricesEqual(auto, forced) {
		t.Error("EncodeWithSize at auto-chosen size must produce identical matrix to Encode")
	}
}

func matricesEqual(a, b image.Image) bool {
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
