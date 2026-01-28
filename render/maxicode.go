package render

import (
	"image"
	"image/draw"

	"github.com/ingridhq/maxicode"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// MaxiCode standard dimensions:
// - 1 inch square (approximately 25.4mm)
// - 30 rows, 33 columns of hexagonal modules
// - Center finder pattern (bullseye)

// upsPrimaryPadding is the 15-character padding UPS uses in their ZPL
// for the primary message encoding in Mode 2/3 MaxiCodes.
const upsPrimaryPadding = "000000000000000"

// drawMaxiCode renders a MaxiCode 2D barcode at the current position.
func (c *canvas) drawMaxiCode(mc *zpl.BarcodeMaxiCode) {
	if mc.Data == "" {
		return
	}

	// Encode the MaxiCode data
	// Try the specified mode first
	grid, err := maxicode.Encode(int(mc.Mode), 0, mc.Data)
	if err != nil {
		// Mode 2/3 have strict UPS format requirements
		// Fall back to mode 4 (standard symbol) which accepts any data
		grid, err = maxicode.Encode(4, 0, mc.Data)
		if err != nil {
			// Still failing - try stripping the UPS primary message padding
			data := mc.Data
			if len(data) > len(upsPrimaryPadding) && data[:len(upsPrimaryPadding)] == upsPrimaryPadding {
				data = data[len(upsPrimaryPadding):]
			}
			grid, err = maxicode.Encode(4, 0, data)
			if err != nil {
				// If encoding still fails, silently skip (data may be malformed)
				return
			}
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
