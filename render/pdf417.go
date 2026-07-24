package render

import (
	"image"
	"image/draw"

	xdraw "golang.org/x/image/draw"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	zpdf417 "github.com/StirlingMarketingGroup/go-zpl/internal/pdf417"
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
	code, err := zpdf417.EncodeWithDimensions(bc.Data, byte(bc.SecurityLevel&0xFF), bc.DataColumns, bc.Rows)
	if err != nil {
		// Silently skip if encoding fails
		return
	}

	// Get library output dimensions
	bounds := code.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Module width from ^BY command (default 2)
	if moduleWidth < 1 {
		moduleWidth = 2
	}

	// Row height from ZPL ^B7 parameter (dots per row)
	rowHeight := bc.Height
	if rowHeight < 1 {
		rowHeight = moduleWidth * 3 // Default aspect ratio
	}

	// Calculate target width: scale by module width.
	targetWidth := origWidth * moduleWidth

	// Calculate target height
	// The pdf417 library renders each row at a fixed moduleHeight (currently 2).
	// Convert ZPL rowHeight (dots per row) to the library's pixel rows.
	const pdf417ModuleHeight = 2
	targetHeight := (origHeight*rowHeight + pdf417ModuleHeight - 1) / pdf417ModuleHeight

	// Scale the barcode using NearestNeighbor to stretch to exact dimensions
	// The barcode.Scale function preserves aspect ratio and centers, which causes position issues
	scaledImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.NearestNeighbor.Scale(scaledImg, scaledImg.Bounds(), code, code.Bounds(), xdraw.Over, nil)

	// Get the scaled image
	pdfImg := scaledImg

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
	srcBounds := finalImg.Bounds()
	// Use srcBounds.Min as source point - the scaled image may have non-zero origin
	draw.Draw(c.img, image.Rect(x, y, x+srcBounds.Dx(), y+srcBounds.Dy()),
		finalImg, srcBounds.Min, draw.Over)
}
