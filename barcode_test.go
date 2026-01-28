package zpl

import (
	"strings"
	"testing"
)

func TestBarcodeCode128(t *testing.T) {
	bc := NewBarcodeCode128("ABC123", 100)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^BC") {
		t.Error("should contain ^BC command")
	}
	if !strings.Contains(zpl, "^FDABC123^FS") {
		t.Error("should contain data")
	}
	if !strings.Contains(zpl, ",100,") {
		t.Error("should contain height")
	}
}

func TestBarcodeCode128WithOptions(t *testing.T) {
	bc := NewBarcodeCode128("TEST", 80).
		WithOrientation(OrientationRotated90).
		WithInterpretation(false, false).
		WithMode(Code128SubsetB)

	zpl := bc.ZPL()

	if !strings.HasPrefix(zpl, "^BCR") {
		t.Errorf("should start with ^BCR for rotated, got: %s", zpl)
	}
	if !strings.Contains(zpl, ",N,N,") {
		t.Error("should have interpretation disabled")
	}
}

func TestBarcodeCode39(t *testing.T) {
	bc := NewBarcodeCode39("HELLO123", 75)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^B3") {
		t.Error("should contain ^B3 command")
	}
	if !strings.Contains(zpl, "^FDHELLO123^FS") {
		t.Error("should contain data")
	}
}

func TestBarcodeCode39WithCheckDigit(t *testing.T) {
	bc := NewBarcodeCode39("ABC", 75).WithCheckDigit(true)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, ",Y,") {
		t.Error("should have check digit enabled")
	}
}

func TestBarcodeQR(t *testing.T) {
	bc := NewBarcodeQR("https://example.com", 5)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^BQ") {
		t.Error("should contain ^BQ command")
	}
	if !strings.Contains(zpl, "^FDMA,https://example.com^FS") {
		t.Errorf("should contain QR data with error correction, got: %s", zpl)
	}
}

func TestBarcodeQRWithOptions(t *testing.T) {
	bc := NewBarcodeQR("DATA", 10).
		WithModel(QRCodeModel1).
		WithErrorCorrection(QRCodeECHigh)

	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^BQN,1,10") {
		t.Errorf("should have model 1 and magnification 10, got: %s", zpl)
	}
	if !strings.Contains(zpl, "^FDHA,DATA") {
		t.Error("should have high error correction")
	}
}

func TestBarcodeDataMatrix(t *testing.T) {
	bc := NewBarcodeDataMatrix("TEST123", 4)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^BX") {
		t.Error("should contain ^BX command")
	}
	if !strings.Contains(zpl, "^FDTEST123^FS") {
		t.Error("should contain data")
	}
}

func TestBarcodePDF417(t *testing.T) {
	bc := NewBarcodePDF417("TESTDATA", 10)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^B7") {
		t.Error("should contain ^B7 command")
	}
	if !strings.Contains(zpl, "^FDTESTDATA^FS") {
		t.Error("should contain data")
	}
}

func TestBarcodePDF417WithOptions(t *testing.T) {
	bc := NewBarcodePDF417("DATA", 8).
		WithSecurityLevel(5).
		WithColumns(3).
		WithTruncation(true)

	zpl := bc.ZPL()

	if !strings.Contains(zpl, ",5,3,") {
		t.Error("should have security level and columns")
	}
	if !strings.Contains(zpl, ",Y^FD") {
		t.Error("should have truncation enabled")
	}
}

func TestBarcodeEAN13(t *testing.T) {
	bc := NewBarcodeEAN13("5901234123457", 100)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^BE") {
		t.Error("should contain ^BE command")
	}
	if !strings.Contains(zpl, "^FD5901234123457^FS") {
		t.Error("should contain data")
	}
}

func TestBarcodeUPCA(t *testing.T) {
	bc := NewBarcodeUPCA("012345678905", 100)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^BU") {
		t.Error("should contain ^BU command")
	}
	if !strings.Contains(zpl, "^FD012345678905^FS") {
		t.Error("should contain data")
	}
}

func TestBarcodeInterleaved2of5(t *testing.T) {
	bc := NewBarcodeInterleaved2of5("12345678", 80)
	zpl := bc.ZPL()

	if !strings.Contains(zpl, "^B2") {
		t.Error("should contain ^B2 command")
	}
	if !strings.Contains(zpl, "^FD12345678^FS") {
		t.Error("should contain data")
	}
}

func TestBarcodeDefault(t *testing.T) {
	bc := NewBarcodeDefault(2, 3.0, 100)
	zpl := bc.ZPL()

	if zpl != "^BY2,3.0,100" {
		t.Errorf("expected ^BY2,3.0,100, got: %s", zpl)
	}
}
