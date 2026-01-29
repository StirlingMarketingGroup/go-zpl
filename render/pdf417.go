package render

import (
	"image"
	"image/draw"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/pdf417"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// drawPDF417 renders a PDF417 barcode at the current position using the canvas barcode module width.
func (c *canvas) drawPDF417(bc *zpl.BarcodePDF417) {
	c.drawPDF417WithModuleWidth(bc, c.barcodeModuleWidth)
}

// drawPDF417WithModuleWidth renders a PDF417 barcode at the current position.
// moduleWidth comes from ^BY command.
func (c *canvas) drawPDF417WithModuleWidth(bc *zpl.BarcodePDF417, moduleWidth int) {
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
	// The PDF417 library outputs 1 pixel per module
	bounds := code.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Module width from ^BY command (default 2)
	if moduleWidth < 1 {
		moduleWidth = 2
	}

	// Row height from ZPL ^B7 parameter (dots per row)
	// PDF417 row height is typically moduleWidth * aspect ratio (usually 3)
	rowHeight := bc.Height
	if rowHeight < 1 {
		rowHeight = moduleWidth * 3 // Default aspect ratio
	}

	// Calculate target dimensions
	targetWidth := origWidth * moduleWidth
	targetHeight := origHeight * rowHeight

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
