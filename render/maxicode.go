package render

import (
	"image"
	"image/draw"
	"strings"

	"github.com/StirlingMarketingGroup/go-zpl/internal/maxicode"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// MaxiCode standard dimensions:
// - 1 inch square (approximately 25.4mm)
// - 30 rows, 33 columns of hexagonal modules
// - Center finder pattern (bullseye)

// scmHeader is the Structured Carrier Message header for UPS MaxiCodes.
const scmHeader = "[)>\x1e01\x1d"

// scmHeaderWithFormat is the full SCM header including the UPS format ID "96".
// The internal maxicode encoder expects data after this 9-byte prefix to be
// GS-delimited fields: postcode, country, service, tracking, SCAC, shipper...
const scmHeaderWithFormat = scmHeader + "96"

// drawMaxiCode renders a MaxiCode 2D barcode at the current position.
func (c *canvas) drawMaxiCode(mc *zpl.BarcodeMaxiCode) {
	if mc.Data == "" {
		return
	}

	data := mc.Data

	// For Mode 2/3 (UPS), the ZPL ^FD data has a primary message prefix before
	// the [)> header containing service class, country code, and postal code.
	// The encoding library expects these as GS-delimited fields WITHIN the SCM
	// message (after [)>\x1e01\x1d96). We extract them from the prefix and
	// reconstruct the data in the format the library expects.
	if mc.Mode == zpl.MaxiCodeMode2 || mc.Mode == zpl.MaxiCodeMode3 {
		data = reconstructMaxiCodeData(data, mc.Mode)
	}

	// Try encoding with the specified mode.
	// If Mode 2/3 fails (strict UPS format requirements), fall back to Mode 4
	// (standard symbol) which accepts any data. This matches Zebra printer behavior.
	grid, err := maxicode.Encode(int(mc.Mode), 0, data)
	if err != nil {
		grid, err = maxicode.Encode(4, 0, data)
		if err != nil {
			// If encoding still fails, silently skip (data may be malformed)
			return
		}
	}

	// Use the library's Draw function which renders proper hexagonal modules
	// UPS labels have ~224 pixels width for MaxiCode (X=20 to divider at X=244)
	// MaxiCode is 33 columns, so scale = 224/33 ≈ 6.8, use 7.5 for good fill
	ctx := grid.Draw(7.5)

	// Get the rendered image from the context
	maxiImg := ctx.Image()

	// Draw the MaxiCode image onto our canvas at the current position
	x := c.curX
	y := c.curY

	// The library renders with a white background, so we can just draw it directly
	draw.Draw(c.img, image.Rect(x, y, x+maxiImg.Bounds().Dx(), y+maxiImg.Bounds().Dy()),
		maxiImg, image.Point{0, 0}, draw.Over)
}

// reconstructMaxiCodeData transforms ZPL Mode 2/3 data into the format expected
// by the internal maxicode encoding library.
//
// In ZPL, the ^FD data for ^BD Mode 2/3 has this layout:
//
//	{service:3}{country:3}{postal:9 or 6}[)>\x1e01\x1d96{tracking}\x1d{SCAC}\x1d{shipper}\x1e07{data}\x1e\x04
//
// The encoding library expects:
//
//	[)>\x1e01\x1d96{postal}\x1d{country}\x1d{service}\x1d{tracking}\x1d{SCAC}\x1d{shipper}\x1e07{data}\x1e\x04
//
// This function extracts postal/country/service from the primary prefix and
// inserts them as GS-delimited fields after the [)>\x1e01\x1d96 header.
func reconstructMaxiCodeData(data string, mode zpl.MaxiCodeMode) string {
	idx := strings.Index(data, scmHeader)
	if idx <= 0 {
		return data
	}

	primary := data[:idx]
	secondary := data[idx:]

	var postcode, country, service string

	switch mode {
	case zpl.MaxiCodeMode2:
		// Mode 2 (US): primary = service(3) + country(3) + postal(9 or 5) chars
		switch len(primary) {
		case 15: // 9-digit postal code
			service = primary[0:3]
			country = primary[3:6]
			postcode = primary[6:15]
		case 11: // 5-digit postal code, pad to 9
			service = primary[0:3]
			country = primary[3:6]
			postcode = primary[6:11] + "0000"
		default:
			return data
		}
	case zpl.MaxiCodeMode3:
		// Mode 3 (international): primary = service(3) + country(3) + postal(6) = 12 chars
		if len(primary) != 12 {
			return data
		}
		service = primary[0:3]
		country = primary[3:6]
		postcode = primary[6:12]
	default:
		return data
	}

	// The secondary message must start with [)>\x1e01\x1d96 (9 bytes).
	if !strings.HasPrefix(secondary, scmHeaderWithFormat) {
		return data
	}

	rest := secondary[len(scmHeaderWithFormat):]

	// Reconstruct: header + postcode + GS + country + GS + service + GS + rest
	return scmHeaderWithFormat + postcode + "\x1d" + country + "\x1d" + service + "\x1d" + rest
}
