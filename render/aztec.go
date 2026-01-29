package render

import (
	"image"
	"image/draw"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// drawAztec renders an Aztec barcode at the current position.
func (c *canvas) drawAztec(bc *zpl.BarcodeAztec) {
	if bc.Data == "" {
		return
	}

	// Determine minimum error correction level
	// ZPL error correction: 0=default (~23%), 1-99=percentage
	// aztec.Encode takes minECCPercent (0-100)
	minECC := 23 // Default ~23% error correction
	if bc.ErrorCorrection > 0 && bc.ErrorCorrection <= 99 {
		minECC = bc.ErrorCorrection
	}

	// Encode the Aztec barcode
	az, err := aztec.Encode([]byte(bc.Data), minECC, aztec.DEFAULT_LAYERS)
	if err != nil {
		// Silently skip if encoding fails
		return
	}

	// Calculate target size based on magnification
	// Magnification is the size of each module in dots
	moduleSize := bc.Magnification
	if moduleSize < 1 {
		moduleSize = 2 // Default magnification
	}
	if moduleSize > 10 {
		moduleSize = 10
	}

	// Get the Aztec dimensions
	bounds := az.Bounds()
	modules := bounds.Dx() // Aztec is square
	targetSize := modules * moduleSize

	// Scale the barcode to target size
	az, err = barcode.Scale(az, targetSize, targetSize)
	if err != nil {
		return
	}

	// Handle orientation by rotating the image if needed
	var finalImg image.Image = az
	switch bc.Orientation {
	case zpl.OrientationRotated90:
		finalImg = rotateImage90CW(az)
	case zpl.OrientationRotated180:
		finalImg = rotateImage180(az)
	case zpl.OrientationRotated270:
		finalImg = rotateImage270CW(az)
	}

	// Draw onto canvas at current position
	x := c.curX
	y := c.curY
	draw.Draw(c.img, image.Rect(x, y, x+finalImg.Bounds().Dx(), y+finalImg.Bounds().Dy()),
		finalImg, image.Point{0, 0}, draw.Over)
}
