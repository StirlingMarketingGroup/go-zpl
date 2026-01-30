package pdf417

import (
	"fmt"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateLabelaryBarcodes generates the same test cases from Labelary
// for comparison scanning
func TestGenerateLabelaryBarcodes(t *testing.T) {
	if os.Getenv("GENERATE_LABELARY") != "1" {
		t.Skip("set GENERATE_LABELARY=1 to generate Labelary comparison barcodes")
	}

	homeDir, _ := os.UserHomeDir()

	testCases := []struct {
		name string
		data string // The actual data (we'll convert to ZPL hex escapes)
	}{
		{"03_text_with_RS", "Hello\x1EWorld"},
		{"05_fedex_header", "[)>\x1E01"},
		{"06_fedex_header_gs", "[)>\x1E01\x1D"},
		{"07_fedex_first_segment", "[)>\x1E01\x1D0294105"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert data to ZPL with hex escapes
			zplData := ""
			for _, b := range []byte(tc.data) {
				if b < 32 || b > 126 {
					zplData += fmt.Sprintf("_%02X", b)
				} else {
					zplData += string(b)
				}
			}

			zpl := fmt.Sprintf("^XA^FO50,50^BY2,2^B7N,8,2,0^FH^FD%s^FS^XZ", zplData)
			t.Logf("ZPL: %s", zpl)

			// Fetch from Labelary
			resp, err := http.Post(
				"https://api.labelary.com/v1/printers/8dpmm/labels/4x2/0/",
				"application/x-www-form-urlencoded",
				strings.NewReader(zpl),
			)
			if err != nil {
				t.Fatalf("Labelary request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				t.Fatalf("Labelary returned status %d", resp.StatusCode)
			}

			img, err := png.Decode(resp.Body)
			if err != nil {
				t.Fatalf("Decode PNG: %v", err)
			}

			// Save
			outPath := filepath.Join(homeDir, "Desktop", "labelary_"+tc.name+".png")
			f, err := os.Create(outPath)
			if err != nil {
				t.Fatalf("Create file: %v", err)
			}
			png.Encode(f, img)
			f.Close()
			t.Logf("Saved Labelary barcode to: %s", outPath)
		})
	}
}
