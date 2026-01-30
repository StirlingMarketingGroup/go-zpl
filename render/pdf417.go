package render

import (
	"image"
	"image/draw"

	"github.com/boombuler/barcode/pdf417"
	xdraw "golang.org/x/image/draw"

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

	// Calculate target width
	// If DataColumns is specified, calculate expected width based on PDF417 structure:
	// - Start pattern: 17 modules
	// - Left row indicator: 17 modules
	// - Data columns: N × 17 modules
	// - Right row indicator: 17 modules
	// - Stop pattern: 18 modules
	// Total = 69 + (N × 17) modules
	var targetWidth int
	if bc.DataColumns > 0 {
		expectedModules := 69 + (bc.DataColumns * 17)
		targetWidth = expectedModules * moduleWidth
	} else {
		// No columns specified, scale by module width
		targetWidth = origWidth * moduleWidth
	}

	// Calculate target height
	// The library outputs 1 pixel per symbol row.
	// The ZPL row height parameter specifies dots per row, but Labelary and real
	// Zebra printers use a more compact row height (about 1.5x module width).
	effectiveRowHeight := rowHeight
	maxRowHeight := (moduleWidth * 3) / 2 // Cap at 1.5x module width (matches Labelary behavior)
	if maxRowHeight < 1 {
		maxRowHeight = 1
	}
	if effectiveRowHeight > maxRowHeight {
		effectiveRowHeight = maxRowHeight
	}
	targetHeight := origHeight * effectiveRowHeight

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
