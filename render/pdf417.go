package render

import (
	"image"
	"image/draw"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/pdf417"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// drawPDF417 renders a PDF417 barcode at the current position.
func (c *canvas) drawPDF417(bc *zpl.BarcodePDF417) {
	if bc.Data == "" {
		return
	}

	// Encode the PDF417
	// The pdf417 library uses security level 0-8 like ZPL
	code, err := pdf417.Encode(bc.Data, byte(bc.SecurityLevel))
	if err != nil {
		// Silently skip if encoding fails
		return
	}

	// Calculate target size
	// ZPL Height parameter is the height of the barcode in dots
	// PDF417 has a specific aspect ratio based on columns and rows
	bounds := code.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Target height from ZPL parameter
	targetHeight := bc.Height
	if targetHeight < 1 {
		targetHeight = 50 // Default height
	}

	// Scale to match target height while preserving aspect ratio
	scale := float64(targetHeight) / float64(origHeight)
	targetWidth := int(float64(origWidth) * scale)

	if targetWidth < 1 {
		targetWidth = origWidth
	}
	if targetHeight < 1 {
		targetHeight = origHeight
	}

	// Scale the barcode
	code, err = barcode.Scale(code, targetWidth, targetHeight)
	if err != nil {
		return
	}

	// Get the scaled image
	pdfImg := code

	// Handle orientation by rotating the image if needed
	var finalImg image.Image = pdfImg
	switch bc.Orientation {
	case zpl.OrientationRotated90:
		finalImg = rotateImage90CW(pdfImg)
	case zpl.OrientationRotated180:
		finalImg = rotateImage180(pdfImg)
	case zpl.OrientationRotated270:
		finalImg = rotateImage270CW(pdfImg)
	}

	// Draw onto canvas at current position
	x := c.curX
	y := c.curY
	draw.Draw(c.img, image.Rect(x, y, x+finalImg.Bounds().Dx(), y+finalImg.Bounds().Dy()),
		finalImg, image.Point{0, 0}, draw.Over)
}
