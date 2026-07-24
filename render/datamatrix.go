package render

import (
	"image"
	"image/draw"

	"github.com/boombuler/barcode"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/internal/datamatrix"
)

// drawDataMatrix renders a DataMatrix barcode at the current position.
func (c *canvas) drawDataMatrix(bc *zpl.BarcodeDataMatrix) {
	if bc.Data == "" {
		return
	}

	// Encode the DataMatrix. Honor explicit ^BX columns/rows when both are set;
	// fall back to auto-size if the forced size is invalid or data won't fit.
	var dm barcode.Barcode
	var err error
	if bc.Columns > 0 && bc.Rows > 0 {
		dm, err = datamatrix.EncodeWithSize(bc.Data, bc.Columns, bc.Rows)
	}
	if dm == nil || err != nil {
		dm, err = datamatrix.Encode(bc.Data)
		if err != nil {
			// Silently skip if encoding fails
			return
		}
	}

	// Calculate target size based on module height
	// ZPL Height parameter is the size of each module in dots
	moduleSize := bc.Height
	if moduleSize < 1 {
		moduleSize = 3 // Default module size
	}

	// Get the DataMatrix dimensions
	bounds := dm.Bounds()
	modules := bounds.Dx() // DataMatrix is square
	targetSize := modules * moduleSize

	// Scale the barcode to target size
	dm, err = barcode.Scale(dm, targetSize, targetSize)
	if err != nil {
		return
	}

	// Get the scaled image
	dmImg := dm

	// Handle orientation by rotating the image if needed
	var finalImg image.Image = dmImg
	switch bc.Orientation {
	case zpl.OrientationRotated90:
		finalImg = rotateImage90CW(dmImg)
	case zpl.OrientationRotated180:
		finalImg = rotateImage180(dmImg)
	case zpl.OrientationRotated270:
		finalImg = rotateImage270CW(dmImg)
	}

	// Draw onto canvas at current position
	x := c.curX
	y := c.curY
	draw.Draw(c.img, image.Rect(x, y, x+finalImg.Bounds().Dx(), y+finalImg.Bounds().Dy()),
		finalImg, image.Point{0, 0}, draw.Over)
}
