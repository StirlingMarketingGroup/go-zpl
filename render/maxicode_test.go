package render

import (
	"errors"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// maxicodeDecoderPath returns the path to the Rust maxicode-decode binary,
// building it if necessary. Skips the test if cargo is not available.
func maxicodeDecoderPath(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo not found, skipping MaxiCode round-trip test")
	}

	dir := filepath.Join("..", "rust", "maxicode-decode")
	bin := filepath.Join(dir, "target", "release", "maxicode-decode")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	if _, err := os.Stat(bin); err != nil {
		t.Logf("Building maxicode-decode...")
		cmd := exec.Command("cargo", "build", "--release")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("cargo build maxicode-decode: %v\n%s", err, out)
		}
	}

	return bin
}

// decodeMaxiCode saves the image as a PNG and decodes it using the Rust decoder.
// Returns the raw decoded bytes (including control characters).
func decodeMaxiCode(t *testing.T, img image.Image) []byte {
	t.Helper()

	bin := maxicodeDecoderPath(t)

	tmp, err := os.CreateTemp("", "maxicode-*.png")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	if err := png.Encode(tmp, img); err != nil {
		tmp.Close()
		t.Fatalf("encode png: %v", err)
	}
	tmp.Close()

	cmd := exec.Command(bin, tmp.Name())
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("maxicode-decode failed: %v\n%s", err, stderr)
	}

	return out
}

func TestMaxiCodeMode2RoundTrip(t *testing.T) {
	// UPS Mode 2 MaxiCode with valid US data from a real shipping label.
	// Primary prefix: service(086) + country(840) + postal(327920000)
	// The encoding library should receive: [)>\x1e01\x1d96{postal}\x1d{country}\x1d{service}\x1d{tracking}...
	zplCode := "^XA^FO0,0^BD2^FH_^FD086840327920000" +
		"[)>_1E01_1D961Z95861181_1DUPSN_1D3F141E" +
		"_1E07MEE_1CX81M153YT\"*\"ELLG\"" +
		"_1C/E'Z.)0UZLZ'\"'ZF#Z%H(J_0D_1E_04^FS^XZ"

	label, err := zpl.Parse(zplCode)
	if err != nil {
		t.Fatalf("parse ZPL: %v", err)
	}

	renderer := New(zpl.DPI203).WithSize(300, 300)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	decoded := decodeMaxiCode(t, img)

	expected := "[)>\x1e01\x1d96327920000\x1d840\x1d086\x1d" +
		"1Z95861181\x1dUPSN\x1d3F141E" +
		"\x1e07MEE\x1cX81M153YT\"*\"ELLG\"" +
		"\x1c/E'Z.)0UZLZ'\"'ZF#Z%H(J\x0d\x1e\x04"

	if string(decoded) != expected {
		t.Errorf("MaxiCode Mode 2 round-trip mismatch\ngot:  %q\nwant: %q", string(decoded), expected)
	}
}

func TestMaxiCodeMode2DifferentZIP(t *testing.T) {
	// Second Mode 2 test with different postal/service data to verify
	// the primary prefix parsing works for various inputs.
	// Primary: service(065) + country(840) + postal(902100000) = ZIP 90210
	zplCode := "^XA^FO0,0^BD2^FH_^FD065840902100000" +
		"[)>_1E01_1D961Z12345678_1DUPSN_1DABC123" +
		"_1E07TEST_0D_1E_04^FS^XZ"

	label, err := zpl.Parse(zplCode)
	if err != nil {
		t.Fatalf("parse ZPL: %v", err)
	}

	renderer := New(zpl.DPI203).WithSize(300, 300)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	decoded := decodeMaxiCode(t, img)

	expected := "[)>\x1e01\x1d96902100000\x1d840\x1d065\x1d" +
		"1Z12345678\x1dUPSN\x1dABC123" +
		"\x1e07TEST\x0d\x1e\x04"

	if string(decoded) != expected {
		t.Errorf("MaxiCode Mode 2 round-trip mismatch\ngot:  %q\nwant: %q", string(decoded), expected)
	}
}

func TestMaxiCodeMode2FiveDigitZip(t *testing.T) {
	// Mode 2 with 5-digit ZIP. Should be padded with 0000.
	// Primary: service(001) + country(840) + postal(12345)
	zplCode := "^XA^FO0,0^BD2^FH_^FD00184012345" +
		"[)>_1E01_1D961Z12345678_1DUPSN_1DABC123" +
		"_1E07TEST_0D_1E_04^FS^XZ"

	label, err := zpl.Parse(zplCode)
	if err != nil {
		t.Fatalf("parse ZPL: %v", err)
	}

	renderer := New(zpl.DPI203).WithSize(300, 300)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	decoded := decodeMaxiCode(t, img)

	// Expect 12345 padded to 123450000
	expected := "[)>\x1e01\x1d96123450000\x1d840\x1d001\x1d" +
		"1Z12345678\x1dUPSN\x1dABC123" +
		"\x1e07TEST\x0d\x1e\x04"

	if string(decoded) != expected {
		t.Errorf("MaxiCode Mode 2 5-digit ZIP round-trip mismatch\ngot:  %q\nwant: %q", string(decoded), expected)
	}
}

func TestMaxiCodeMode2NonUSCountry(t *testing.T) {
	// Mode 2 with country=000 (dummy data like import_control.zpl).
	// The upstream library rejects this (requires country=840), but our
	// internal fork accepts any country code to match real Zebra behavior.
	zplCode := "^XA^FO0,0^BD2^FH_^FD000000000000000" +
		"[)>_1E01_1D961Z00000001_1DUPSN_1D00A00A" +
		"_1E07Y+0*0A.AA'AA#A0A%'_0DAAA0.00" +
		"_1C*0AAA'A_1C0AA000$&A_0D_1E_04^FS^XZ"

	label, err := zpl.Parse(zplCode)
	if err != nil {
		t.Fatalf("parse ZPL: %v", err)
	}

	renderer := New(zpl.DPI203).WithSize(300, 300)
	img, err := renderer.Render(label)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	decoded := decodeMaxiCode(t, img)
	decodedStr := string(decoded)

	// With the internal fork, Mode 2 encoding succeeds.
	// The decoded data should include the postal/country/service from the primary
	// plus the full secondary data including format 07 section.
	expected := "[)>\x1e01\x1d96000000000\x1d000\x1d000\x1d" +
		"1Z00000001\x1dUPSN\x1d00A00A" +
		"\x1e07Y+0*0A.AA'AA#A0A%'\x0dAAA0.00" +
		"\x1c*0AAA'A\x1c0AA000$&A\x0d\x1e\x04"

	if decodedStr != expected {
		t.Errorf("MaxiCode Mode 2 non-US round-trip mismatch\ngot:  %q\nwant: %q", decodedStr, expected)
	}
}

func TestMaxiCodeReconstructData(t *testing.T) {
	// Unit test for the data reconstruction logic.
	tests := []struct {
		name string
		data string
		mode zpl.MaxiCodeMode
		want string
	}{
		{
			name: "mode 2 valid US",
			data: "086840327920000[)>\x1e01\x1d961Z95861181\x1dUPSN\x1d3F141E\x1e07DATA\x1e\x04",
			mode: zpl.MaxiCodeMode2,
			want: "[)>\x1e01\x1d96327920000\x1d840\x1d086\x1d1Z95861181\x1dUPSN\x1d3F141E\x1e07DATA\x1e\x04",
		},
		{
			name: "mode 2 different zip",
			data: "065840902100000[)>\x1e01\x1d961Z12345678\x1dUPSN\x1dABC123\x1e07TEST\x1e\x04",
			mode: zpl.MaxiCodeMode2,
			want: "[)>\x1e01\x1d96902100000\x1d840\x1d065\x1d1Z12345678\x1dUPSN\x1dABC123\x1e07TEST\x1e\x04",
		},
		{
			name: "mode 2 5-digit zip",
			data: "00184012345[)>\x1e01\x1d961Z12345678\x1dUPSN\x1dABC123\x1e07TEST\x1e\x04",
			mode: zpl.MaxiCodeMode2,
			want: "[)>\x1e01\x1d96123450000\x1d840\x1d001\x1d1Z12345678\x1dUPSN\x1dABC123\x1e07TEST\x1e\x04",
		},
		{
			name: "mode 3 international",
			data: "086999AB1234[)>\x1e01\x1d96TRACK1\x1dSCAC\x1dSHIP01\x1e\x04",
			mode: zpl.MaxiCodeMode3,
			want: "[)>\x1e01\x1d96AB1234\x1d999\x1d086\x1dTRACK1\x1dSCAC\x1dSHIP01\x1e\x04",
		},
		{
			name: "no SCM header - pass through",
			data: "some random data without header",
			mode: zpl.MaxiCodeMode2,
			want: "some random data without header",
		},
		{
			name: "mode 4 - pass through",
			data: "086840327920000[)>\x1e01\x1d96data",
			mode: zpl.MaxiCodeMode4,
			want: "086840327920000[)>\x1e01\x1d96data",
		},
		{
			name: "primary too short",
			data: "0868[)>\x1e01\x1d96data",
			mode: zpl.MaxiCodeMode2,
			want: "0868[)>\x1e01\x1d96data",
		},
		{
			name: "primary too long - pass through",
			data: "0868403279200001234[)>\x1e01\x1d96data",
			mode: zpl.MaxiCodeMode2,
			want: "0868403279200001234[)>\x1e01\x1d96data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconstructMaxiCodeData(tt.data, tt.mode)
			if got != tt.want {
				t.Errorf("reconstructMaxiCodeData()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
