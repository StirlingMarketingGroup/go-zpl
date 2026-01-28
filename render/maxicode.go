package render

import (
	"image"
	"image/draw"
	"strings"

	"github.com/ingridhq/maxicode"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// MaxiCode standard dimensions:
// - 1 inch square (approximately 25.4mm)
// - 30 rows, 33 columns of hexagonal modules
// - Center finder pattern (bullseye)

// upsPrimaryLen is the length of the UPS primary message prefix in ZPL.
// This 15-character prefix contains postal/country/service info that is
// redundant with what's in the structured carrier message.
const upsPrimaryLen = 15

// scmHeader is the Structured Carrier Message header for UPS MaxiCodes.
const scmHeader = "[)>\x1e01\x1d"

// drawMaxiCode renders a MaxiCode 2D barcode at the current position.
func (c *canvas) drawMaxiCode(mc *zpl.BarcodeMaxiCode) {
	if mc.Data == "" {
		return
	}

	data := mc.Data

	// For Mode 2/3 (UPS), the ZPL data often has a 15-character primary message
	// prefix before the [)> header. The maxicode library expects the data to
	// start with the SCM header, so we need to strip this prefix.
	if mc.Mode == zpl.MaxiCodeMode2 || mc.Mode == zpl.MaxiCodeMode3 {
		// Find the SCM header position
		if idx := strings.Index(data, scmHeader); idx > 0 && idx <= upsPrimaryLen {
			// Strip the prefix - it's the UPS primary message placeholder
			data = data[idx:]
		}
	}

	// Try encoding with the specified mode
	grid, err := maxicode.Encode(int(mc.Mode), 0, data)
	if err != nil {
		// Mode 2/3 have strict UPS format requirements
		// Fall back to mode 4 (standard symbol) which accepts any data
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
