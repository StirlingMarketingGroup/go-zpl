package render

import (
	"image"
	"image/draw"
	"strings"

	"github.com/boombuler/barcode"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/internal/datamatrix"
)

// drawDataMatrix renders a DataMatrix barcode at the current position.
func (c *canvas) drawDataMatrix(bc *zpl.BarcodeDataMatrix) {
	if bc.Data == "" {
		return
	}

	data := applyDataMatrixEscapes(bc.Data, bc.EscapeChar)

	// Encode the DataMatrix. Honor explicit ^BX columns/rows when both are set;
	// fall back to auto-size if the forced size is invalid or data won't fit.
	var dm barcode.Barcode
	var err error
	if bc.Columns > 0 && bc.Rows > 0 {
		dm, err = datamatrix.EncodeWithSize(data, bc.Columns, bc.Rows)
	}
	if dm == nil || err != nil {
		dm, err = datamatrix.Encode(data)
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

// applyDataMatrixEscapes translates ^BX escapes in a GS1 symbol.
// Only data that begins with e1 is translated: e1 becomes FNC1 and ee becomes
// one literal e. Every other sequence, and data that does not begin with e1,
// is left unchanged.
func applyDataMatrixEscapes(data string, escape rune) string {
	if escape == 0 || data == "" {
		return data
	}
	esc := string(escape)
	if !strings.HasPrefix(data, esc+"1") {
		return data
	}
	var b strings.Builder
	b.Grow(len(data))
	for i := 0; i < len(data); {
		if strings.HasPrefix(data[i:], esc) {
			rest := data[i+len(esc):]
			if strings.HasPrefix(rest, "1") {
				b.WriteByte(datamatrix.FNC1)
				i += len(esc) + 1
				continue
			}
			if strings.HasPrefix(rest, esc) {
				b.WriteString(esc)
				i += 2 * len(esc)
				continue
			}
		}
		b.WriteByte(data[i])
		i++
	}
	return b.String()
}
