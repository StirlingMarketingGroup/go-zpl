// Package render provides image rendering capabilities for ZPL labels.
// It converts ZPL Label objects to bitmap images that approximate
// what a Zebra printer would produce.
package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"sync"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// Renderer holds configuration for rendering ZPL labels to images.
type Renderer struct {
	// DPI is the printer resolution (203, 300, or 600 dots per inch).
	DPI zpl.DPI

	// Width is the label width in dots. If zero, uses the label's configured width.
	Width int

	// Height is the label height in dots. If zero, uses the label's configured height.
	Height int

	// IgnoreLabelHome controls whether ^LH (label home) offsets are applied.
	// When false (default), offsets are applied exactly as a printer would.
	// When true, offsets are ignored for cleaner previews.
	IgnoreLabelHome bool
}

// New creates a new Renderer with the given DPI.
// Width and height will be taken from the label if not set.
func New(dpi zpl.DPI) *Renderer {
	return &Renderer{DPI: dpi}
}

// WithSize sets the label dimensions in dots.
func (r *Renderer) WithSize(width, height int) *Renderer {
	r.Width = width
	r.Height = height
	return r
}

// WithIgnoreLabelHome sets whether to ignore ^LH label home offsets.
// When false (default), offsets are applied exactly as a printer would.
// When true, offsets are ignored for cleaner previews.
func (r *Renderer) WithIgnoreLabelHome(ignore bool) *Renderer {
	r.IgnoreLabelHome = ignore
	return r
}

// Render converts a Label to an image.
// The returned image uses white background with black elements,
// matching thermal label printer output.
func (r *Renderer) Render(label *zpl.Label) (image.Image, error) {
	width := r.Width
	if width == 0 {
		width = label.Width()
	}
	if width == 0 {
		width = 812 // Default 4-inch label at 203 DPI
	}

	height := r.Height
	if height == 0 {
		height = label.Height()
	}
	if height == 0 {
		height = 1218 // Default 6-inch label at 203 DPI
	}

	canvas, err := newCanvas(width, height)
	if err != nil {
		return nil, err
	}

	// Apply label home offset (^LH) unless ignored
	// By default, offsets are applied (IgnoreLabelHome = false)
	if !r.IgnoreLabelHome {
		homeX, homeY := label.Home()
		canvas.homeX = homeX
		canvas.homeY = homeY
	}

	// Process all commands
	for _, cmd := range label.Commands() {
		if err := canvas.processCommand(cmd); err != nil {
			return nil, err
		}
	}

	// Note: ^POI (Print Orientation Inverted) affects how the printer outputs,
	// but for preview rendering we show the label as it appears when viewed.
	// The ZPL coordinates are already laid out for the final appearance.
	return canvas.Image(), nil
}

// RenderPNG renders the label and writes it as a PNG image to the writer.
func (r *Renderer) RenderPNG(label *zpl.Label, w io.Writer) error {
	img, err := r.Render(label)
	if err != nil {
		return err
	}

	// Convert to paletted image (1-bit black/white)
	// This is much faster to encode than RGBA since it's only 2 colors
	rgbaImg := img.(*image.RGBA)
	bounds := rgbaImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a 2-color palette (white=0, black=1)
	palette := color.Palette{
		color.RGBA{255, 255, 255, 255}, // index 0: white
		color.RGBA{0, 0, 0, 255},       // index 1: black
	}
	palettedImg := image.NewPaletted(bounds, palette)

	// Fast conversion: operate directly on pixel arrays
	// RGBA is 4 bytes per pixel, Paletted is 1 byte per pixel
	rgbaPix := rgbaImg.Pix
	palPix := palettedImg.Pix
	rgbaStride := rgbaImg.Stride
	palStride := palettedImg.Stride

	for y := 0; y < height; y++ {
		rgbaRow := y * rgbaStride
		palRow := y * palStride
		for x := 0; x < width; x++ {
			// Check if pixel is dark enough to be black
			// RGBA layout: R, G, B, A for each pixel
			// Use threshold of 128 to capture anti-aliased text (gray pixels from font smoothing)
			if rgbaPix[rgbaRow+x*4] < 128 {
				palPix[palRow+x] = 1 // black
			}
			// else leave as 0 (white) - already initialized to zero
		}
	}

	// Use BestSpeed compression
	encoder := &png.Encoder{CompressionLevel: png.BestSpeed}
	return encoder.Encode(w, palettedImg)
}

// RenderJPEG renders the label and writes it as a JPEG image to the writer.
// Quality should be between 1 and 100, where 100 is best quality.
func (r *Renderer) RenderJPEG(label *zpl.Label, w io.Writer, quality int) error {
	img, err := r.Render(label)
	if err != nil {
		return err
	}

	// Clamp quality to valid range
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
}

// canvas manages the rendering state and image buffer.
type canvas struct {
	img *image.RGBA

	// Current position (set by ^FO)
	curX int
	curY int

	// Label home offset (set by ^LH)
	homeX int
	homeY int

	// Current font settings
	fontMgr     *fontManager
	currentFont zpl.Font
	fontHeight  int
	fontWidth   int
	fontOrient  zpl.Orientation

	// Field state
	fieldReverse bool

	// Barcode defaults (set by ^BY)
	barcodeModuleWidth int
	barcodeHeight      int

	// Pending barcode (waiting for ^FD to provide data)
	// In ZPL, barcode commands like ^BC set up the barcode parameters,
	// then the following ^FD provides the data
	pendingBarcode interface{}
}

// Shared font manager (parsed once, reused across renders)
var (
	sharedFontMgr     *fontManager
	sharedFontMgrOnce sync.Once
	sharedFontMgrErr  error
)

func getSharedFontManager() (*fontManager, error) {
	sharedFontMgrOnce.Do(func() {
		sharedFontMgr, sharedFontMgrErr = newFontManager()
	})
	return sharedFontMgr, sharedFontMgrErr
}

func newCanvas(width, height int) (*canvas, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white background using draw.Draw (faster than pixel-by-pixel)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

	fm, err := getSharedFontManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize fonts: %w", err)
	}

	return &canvas{
		img:                img,
		fontMgr:            fm,
		currentFont:        zpl.Font0,
		fontHeight:         30,
		fontWidth:          0, // Zero means proportional width
		fontOrient:         zpl.OrientationNormal,
		barcodeModuleWidth: 2, // Default module width
		barcodeHeight:      100,
	}, nil
}

// Image returns the rendered image.
func (c *canvas) Image() image.Image {
	return c.img
}

// processCommand handles a single ZPL command.
func (c *canvas) processCommand(cmd zpl.Command) error { //nolint:unparam // Error return reserved for future commands
	switch v := cmd.(type) {
	case *zpl.FieldOrigin:
		c.curX = v.X + c.homeX
		c.curY = v.Y + c.homeY

	case *zpl.FieldTypeset:
		c.curX = v.X + c.homeX
		c.curY = v.Y + c.homeY

	case *zpl.ScalableFont:
		c.currentFont = v.Font
		c.fontHeight = v.Height
		c.fontWidth = v.Width
		c.fontOrient = v.Orientation

	case *zpl.ChangeFont:
		c.currentFont = v.Font
		c.fontHeight = v.Height
		c.fontWidth = v.Width

	case *zpl.FieldData:
		// Check if there's a pending barcode waiting for data
		if c.pendingBarcode != nil {
			c.drawPendingBarcode(v.Data)
		} else {
			c.drawText(v.Data)
		}

	case *zpl.FieldReverse:
		c.fieldReverse = true

	case *zpl.GraphicBox:
		c.drawBox(v)
		c.fieldReverse = false // Reset after drawing

	case *zpl.GraphicCircle:
		c.drawCircle(v)
		c.fieldReverse = false // Reset after drawing

	case *zpl.GraphicDiagonalLine:
		c.drawDiagonalLine(v)
		c.fieldReverse = false // Reset after drawing

	case *zpl.GraphicEllipse:
		c.drawEllipse(v)
		c.fieldReverse = false // Reset after drawing

	case *zpl.BarcodeDefault:
		c.setBarcodeDefault(v)

	case *zpl.BarcodeCode128:
		// If barcode has data, draw immediately; otherwise wait for ^FD
		if v.Data != "" {
			c.drawBarcode128(v, c.barcodeModuleWidth)
		} else {
			c.pendingBarcode = v
		}

	case *zpl.GraphicField:
		c.drawGraphicField(v)

	case *zpl.BarcodeMaxiCode:
		c.drawMaxiCode(v)

	case *zpl.BarcodeQR:
		c.drawQRCode(v)

	case *zpl.BarcodeDataMatrix:
		c.drawDataMatrix(v)

	case *zpl.BarcodePDF417:
		// If barcode has data, draw immediately; otherwise wait for ^FD
		if v.Data != "" {
			c.drawPDF417(v)
		} else {
			c.pendingBarcode = v
		}

	case *zpl.BarcodeAztec:
		c.drawAztec(v)

	// Ignore commands we don't render
	case *zpl.Comment:
		// Comments are ignored

	case *zpl.FieldBlock:
		// Field blocks affect text wrapping - not yet implemented

	case *zpl.CharacterSet:
		// Character set selection - not yet implemented
	}

	return nil
}

// drawPendingBarcode renders a pending barcode with the given data.
func (c *canvas) drawPendingBarcode(data string) {
	if c.pendingBarcode == nil {
		return
	}

	switch bc := c.pendingBarcode.(type) {
	case *zpl.BarcodeCode128:
		bc.Data = data
		c.drawBarcode128(bc, c.barcodeModuleWidth)
	case *zpl.BarcodePDF417:
		bc.Data = data
		c.drawPDF417(bc)
	}

	// Clear pending barcode
	c.pendingBarcode = nil
}

// drawText renders text at the current position using the current font settings.
func (c *canvas) drawText(text string) {
	if c.fontMgr == nil || text == "" {
		return
	}

	height := c.fontHeight
	if height == 0 {
		height = 30
	}

	// Pass fontWidth directly - 0 means proportional (natural font width)
	c.fontMgr.drawText(c.img, text, c.curX, c.curY, c.currentFont, height, c.fontWidth, c.fontOrient, c.fieldReverse)

	// Reset field reverse after drawing
	c.fieldReverse = false
}

// drawBox renders a graphic box at the current position.
func (c *canvas) drawBox(box *zpl.GraphicBox) {
	x := c.curX
	y := c.curY
	w := box.Width
	h := box.Height
	t := box.Thickness
	isWhite := box.Color == zpl.LineColorWhite

	// For filled boxes (thickness >= min dimension / 2)
	if t >= w/2 || t >= h/2 {
		// Draw filled rectangle
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				c.setPixel(x+dx, y+dy, isWhite)
			}
		}
		return
	}

	// Draw outline box
	// Top edge
	for dy := 0; dy < t; dy++ {
		for dx := 0; dx < w; dx++ {
			c.setPixel(x+dx, y+dy, isWhite)
		}
	}
	// Bottom edge
	for dy := h - t; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.setPixel(x+dx, y+dy, isWhite)
		}
	}
	// Left edge
	for dy := t; dy < h-t; dy++ {
		for dx := 0; dx < t; dx++ {
			c.setPixel(x+dx, y+dy, isWhite)
		}
	}
	// Right edge
	for dy := t; dy < h-t; dy++ {
		for dx := w - t; dx < w; dx++ {
			c.setPixel(x+dx, y+dy, isWhite)
		}
	}
}

// setPixel sets a pixel, handling field reverse (^FR) by XOR-ing with existing pixels.
func (c *canvas) setPixel(px, py int, isWhite bool) {
	if px < 0 || px >= c.img.Bounds().Max.X || py < 0 || py >= c.img.Bounds().Max.Y {
		return
	}

	// ^FR inverts the background - black becomes white, white becomes black
	if c.fieldReverse {
		existing := c.img.RGBAAt(px, py)
		if existing.R == 0 && existing.G == 0 && existing.B == 0 {
			// Black -> White
			c.img.Set(px, py, color.RGBA{255, 255, 255, 255})
		} else {
			// White -> Black
			c.img.Set(px, py, color.RGBA{0, 0, 0, 255})
		}
		return
	}

	// Normal drawing
	if isWhite {
		c.img.Set(px, py, color.RGBA{255, 255, 255, 255})
	} else {
		c.img.Set(px, py, color.RGBA{0, 0, 0, 255})
	}
}

// drawCircle renders a graphic circle at the current position.
func (c *canvas) drawCircle(circle *zpl.GraphicCircle) {
	cx := c.curX + circle.Diameter/2
	cy := c.curY + circle.Diameter/2
	r := circle.Diameter / 2
	t := circle.Thickness
	isWhite := circle.Color == zpl.LineColorWhite

	rOuter := r
	rInner := r - t
	if rInner < 0 {
		rInner = 0
	}

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := dx*dx + dy*dy
			if dist <= rOuter*rOuter && dist >= rInner*rInner {
				c.setPixel(cx+dx, cy+dy, isWhite)
			}
		}
	}
}

// drawDiagonalLine renders a diagonal line at the current position.
func (c *canvas) drawDiagonalLine(line *zpl.GraphicDiagonalLine) {
	x := c.curX
	y := c.curY
	w := line.Width
	h := line.Height
	t := line.Thickness
	isWhite := line.Color == zpl.LineColorWhite
	isRight := line.Orientation == zpl.DiagonalRightLeaning

	// Use Bresenham-style line drawing with thickness
	for dy := 0; dy < h; dy++ {
		// Calculate x position along the diagonal
		var lineX int
		if isRight {
			lineX = dy * w / h
		} else {
			lineX = w - 1 - dy*w/h
		}

		// Draw thickness perpendicular to line direction
		for dt := -t / 2; dt <= t/2; dt++ {
			c.setPixel(x+lineX+dt, y+dy, isWhite)
		}
	}
}

// drawEllipse renders an ellipse at the current position.
func (c *canvas) drawEllipse(ellipse *zpl.GraphicEllipse) {
	cx := c.curX + ellipse.Width/2
	cy := c.curY + ellipse.Height/2
	rx := ellipse.Width / 2
	ry := ellipse.Height / 2
	t := ellipse.Thickness
	isWhite := ellipse.Color == zpl.LineColorWhite

	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			// Ellipse equation: (x/rx)^2 + (y/ry)^2 = 1
			outer := float64(dx*dx)/float64(rx*rx) + float64(dy*dy)/float64(ry*ry)
			innerRx := rx - t
			innerRy := ry - t
			if innerRx < 1 {
				innerRx = 1
			}
			if innerRy < 1 {
				innerRy = 1
			}
			inner := float64(dx*dx)/float64(innerRx*innerRx) + float64(dy*dy)/float64(innerRy*innerRy)

			if outer <= 1.0 && inner >= 1.0 {
				c.setPixel(cx+dx, cy+dy, isWhite)
			}
		}
	}
}

// drawGraphicField renders a bitmap graphic field at the current position.
func (c *canvas) drawGraphicField(gf *zpl.GraphicField) {
	x := c.curX
	y := c.curY
	bytesPerRow := gf.BytesPerRow
	col := color.RGBA{0, 0, 0, 255}

	// Each byte represents 8 pixels
	pixelsPerByte := 8
	rowWidthPixels := bytesPerRow * pixelsPerByte

	row := 0
	pixelInRow := 0

	switch gf.Format {
	case zpl.GraphicFieldBinary:
		// Binary format: each byte directly represents 8 pixels
		for _, b := range gf.BinaryData {
			// Each bit is a pixel (MSB first)
			for bit := 7; bit >= 0; bit-- {
				if b&(1<<bit) != 0 {
					px := x + pixelInRow
					py := y + row
					if px >= 0 && px < c.img.Bounds().Max.X && py >= 0 && py < c.img.Bounds().Max.Y {
						c.img.Set(px, py, col)
					}
				}
				pixelInRow++

				if pixelInRow >= rowWidthPixels {
					pixelInRow = 0
					row++
				}
			}
		}

	case zpl.GraphicFieldASCII:
		// ASCII format: each hex char represents 4 pixels (1 nibble)
		data := gf.Data

		// Remove any whitespace/newlines from data
		var cleanData strings.Builder
		for _, r := range data {
			if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f') {
				cleanData.WriteRune(r)
			}
		}

		hexData := cleanData.String()
		for i := 0; i < len(hexData); i++ {
			hexChar := hexData[i]
			var nibble int
			switch {
			case hexChar >= '0' && hexChar <= '9':
				nibble = int(hexChar - '0')
			case hexChar >= 'A' && hexChar <= 'F':
				nibble = int(hexChar-'A') + 10
			case hexChar >= 'a' && hexChar <= 'f':
				nibble = int(hexChar-'a') + 10
			default:
				continue
			}

			// Each nibble is 4 bits = 4 pixels
			for bit := 3; bit >= 0; bit-- {
				if nibble&(1<<bit) != 0 {
					px := x + pixelInRow
					py := y + row
					if px >= 0 && px < c.img.Bounds().Max.X && py >= 0 && py < c.img.Bounds().Max.Y {
						c.img.Set(px, py, col)
					}
				}
				pixelInRow++

				if pixelInRow >= rowWidthPixels {
					pixelInRow = 0
					row++
				}
			}
		}
	}
}
