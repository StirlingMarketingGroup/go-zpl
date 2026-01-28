package zpl

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLabel(t *testing.T) {
	label := NewLabel()
	if label == nil {
		t.Fatal("NewLabel returned nil")
	}
	if label.dpi != DPI203 {
		t.Errorf("expected default DPI %d, got %d", DPI203, label.dpi)
	}
	if label.printOrientation != PrintOrientationNormal {
		t.Errorf("expected default orientation %c, got %c", PrintOrientationNormal, label.printOrientation)
	}
}

func TestLabelSetSize(t *testing.T) {
	tests := []struct {
		name          string
		width, height float64
		unit          Unit
		dpi           DPI
		wantWidth     int
		wantHeight    int
	}{
		{"4x6 inches at 203 DPI", 4, 6, UnitInches, DPI203, 812, 1218},
		{"4x6 inches at 300 DPI", 4, 6, UnitInches, DPI300, 1200, 1800},
		{"100x150 mm at 203 DPI", 100, 150, UnitMillimeters, DPI203, 799, 1198},
		{"dots direct", 800, 1200, UnitDots, DPI203, 800, 1200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := NewLabel().SetDPI(tt.dpi).SetSize(tt.width, tt.height, tt.unit)
			if label.Width() != tt.wantWidth {
				t.Errorf("Width() = %d, want %d", label.Width(), tt.wantWidth)
			}
			if label.Height() != tt.wantHeight {
				t.Errorf("Height() = %d, want %d", label.Height(), tt.wantHeight)
			}
		})
	}
}

func TestLabelBasicOutput(t *testing.T) {
	label := NewLabel().
		SetSizeDots(800, 1200)

	zpl := label.String()

	if !strings.HasPrefix(zpl, "^XA\n") {
		t.Error("ZPL should start with ^XA")
	}
	if !strings.HasSuffix(zpl, "^XZ\n") {
		t.Error("ZPL should end with ^XZ")
	}
	if !strings.Contains(zpl, "^PW800") {
		t.Error("ZPL should contain print width")
	}
	if !strings.Contains(zpl, "^LL1200") {
		t.Error("ZPL should contain label length")
	}
}

func TestLabelWithText(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		TextField(50, 50, Font0, 30, 30, "Hello World")

	zpl := label.String()

	if !strings.Contains(zpl, "^FO50,50") {
		t.Error("ZPL should contain field origin")
	}
	if !strings.Contains(zpl, "^A0N,30,30") {
		t.Error("ZPL should contain font command")
	}
	if !strings.Contains(zpl, "^FDHello World^FS") {
		t.Error("ZPL should contain field data")
	}
}

func TestLabelWithBarcode(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		Code128(50, 100, "12345678", 50)

	zpl := label.String()

	if !strings.Contains(zpl, "^FO50,100") {
		t.Error("ZPL should contain field origin")
	}
	if !strings.Contains(zpl, "^BC") {
		t.Error("ZPL should contain Code 128 barcode command")
	}
	if !strings.Contains(zpl, "^FD12345678^FS") {
		t.Error("ZPL should contain barcode data")
	}
}

func TestLabelWithQRCode(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 400).
		QRCode(50, 50, "https://example.com", 5)

	zpl := label.String()

	if !strings.Contains(zpl, "^BQ") {
		t.Error("ZPL should contain QR code command")
	}
	if !strings.Contains(zpl, "https://example.com") {
		t.Error("ZPL should contain QR data")
	}
}

func TestLabelWithGraphics(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		Box(10, 10, 380, 280, 2).
		HorizontalLine(10, 150, 380, 2).
		Circle(200, 200, 50, 2)

	zpl := label.String()

	if !strings.Contains(zpl, "^GB380,280,2") {
		t.Error("ZPL should contain graphic box")
	}
	if !strings.Contains(zpl, "^GC50,2") {
		t.Error("ZPL should contain graphic circle")
	}
}

func TestLabelWriteTo(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		TextField(50, 50, Font0, 30, 30, "Test")

	var buf bytes.Buffer
	n, err := label.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if n != int64(buf.Len()) {
		t.Errorf("WriteTo returned %d, but buffer has %d bytes", n, buf.Len())
	}
	if buf.String() != label.String() {
		t.Error("WriteTo output differs from String()")
	}
}

func TestLabelMarshalText(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300)

	text, err := label.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if string(text) != label.String() {
		t.Error("MarshalText output differs from String()")
	}
}

func TestLabelPrintQuantity(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		SetPrintQuantity(5, 0, 0, false)

	zpl := label.String()

	if !strings.Contains(zpl, "^PQ5,0,0,N") {
		t.Errorf("ZPL should contain print quantity, got: %s", zpl)
	}
}

func TestLabelHome(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		SetHomeDots(10, 20)

	zpl := label.String()

	if !strings.Contains(zpl, "^LH10,20") {
		t.Error("ZPL should contain label home")
	}
}

func TestLabelPrintOrientation(t *testing.T) {
	label := NewLabel().
		SetSizeDots(400, 300).
		SetPrintOrientation(PrintOrientationInverted)

	zpl := label.String()

	if !strings.Contains(zpl, "^POI") {
		t.Error("ZPL should contain inverted print orientation")
	}
}

func TestLabelCommands(t *testing.T) {
	label := NewLabel().
		Add(NewFieldOrigin(10, 20)).
		Add(NewFieldData("test"))

	cmds := label.Commands()
	if len(cmds) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cmds))
	}
}
