// Package render provides image rendering capabilities for ZPL labels.
// It converts ZPL Label objects to bitmap images that approximate
// what a Zebra printer would produce.
package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
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

	// Apply label home offset
	homeX, homeY := label.Home()
	canvas.homeX = homeX
	canvas.homeY = homeY

	// Process all commands
	for _, cmd := range label.Commands() {
		if err := canvas.processCommand(cmd); err != nil {
			return nil, err
		}
	}

	// Apply print orientation
	img := canvas.Image()
	if label.PrintOrientationSetting() == zpl.PrintOrientationInverted {
		img = rotateImage180(img)
	}

	return img, nil
}

// RenderPNG renders the label and writes it as a PNG image to the writer.
func (r *Renderer) RenderPNG(label *zpl.Label, w io.Writer) error {
	img, err := r.Render(label)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
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
		c.drawText(v.Data)

	case *zpl.FieldReverse:
		c.fieldReverse = true

	case *zpl.GraphicBox:
		c.drawBox(v)

	case *zpl.GraphicCircle:
		c.drawCircle(v)

	case *zpl.GraphicDiagonalLine:
		c.drawDiagonalLine(v)

	case *zpl.GraphicEllipse:
		c.drawEllipse(v)

	case *zpl.BarcodeDefault:
		c.setBarcodeDefault(v)

	case *zpl.BarcodeCode128:
		c.drawBarcode128(v, c.barcodeModuleWidth)

	case *zpl.GraphicField:
		c.drawGraphicField(v)

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

// drawText renders text at the current position using the current font settings.
func (c *canvas) drawText(text string) {
	if c.fontMgr == nil || text == "" {
		return
	}

	height := c.fontHeight
	if height == 0 {
		height = 30
	}

	width := c.fontWidth
	if width == 0 {
		// Proportional: width is approximately 0.6 * height for most fonts
		width = height * 6 / 10
	}

	c.fontMgr.drawText(c.img, text, c.curX, c.curY, c.currentFont, height, width, c.fontOrient, c.fieldReverse)

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

	col := color.RGBA{0, 0, 0, 255}
	if isWhite {
		col = color.RGBA{255, 255, 255, 255}
	}

	// For filled boxes (thickness >= min dimension / 2)
	if t >= w/2 || t >= h/2 {
		// Draw filled rectangle
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				c.img.Set(x+dx, y+dy, col)
			}
		}
		return
	}

	// Draw outline box
	// Top edge
	for dy := 0; dy < t; dy++ {
		for dx := 0; dx < w; dx++ {
			c.img.Set(x+dx, y+dy, col)
		}
	}
	// Bottom edge
	for dy := h - t; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.img.Set(x+dx, y+dy, col)
		}
	}
	// Left edge
	for dy := t; dy < h-t; dy++ {
		for dx := 0; dx < t; dx++ {
			c.img.Set(x+dx, y+dy, col)
		}
	}
	// Right edge
	for dy := t; dy < h-t; dy++ {
		for dx := w - t; dx < w; dx++ {
			c.img.Set(x+dx, y+dy, col)
		}
	}
}

// drawCircle renders a graphic circle at the current position.
func (c *canvas) drawCircle(circle *zpl.GraphicCircle) {
	cx := c.curX + circle.Diameter/2
	cy := c.curY + circle.Diameter/2
	r := circle.Diameter / 2
	t := circle.Thickness
	isWhite := circle.Color == zpl.LineColorWhite

	col := color.RGBA{0, 0, 0, 255}
	if isWhite {
		col = color.RGBA{255, 255, 255, 255}
	}

	rOuter := r
	rInner := r - t
	if rInner < 0 {
		rInner = 0
	}

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := dx*dx + dy*dy
			if dist <= rOuter*rOuter && dist >= rInner*rInner {
				c.img.Set(cx+dx, cy+dy, col)
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

	col := color.RGBA{0, 0, 0, 255}
	if isWhite {
		col = color.RGBA{255, 255, 255, 255}
	}

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
			px := x + lineX + dt
			py := y + dy
			if px >= 0 && px < c.img.Bounds().Max.X && py >= 0 && py < c.img.Bounds().Max.Y {
				c.img.Set(px, py, col)
			}
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

	col := color.RGBA{0, 0, 0, 255}
	if isWhite {
		col = color.RGBA{255, 255, 255, 255}
	}

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
				c.img.Set(cx+dx, cy+dy, col)
			}
		}
	}
}

// drawGraphicField renders a bitmap graphic field at the current position.
func (c *canvas) drawGraphicField(gf *zpl.GraphicField) {
	if gf.Format != zpl.GraphicFieldASCII {
		// Only ASCII format is supported for now
		return
	}

	x := c.curX
	y := c.curY
	bytesPerRow := gf.BytesPerRow
	data := gf.Data

	// Remove any whitespace/newlines from data
	cleanData := ""
	for _, r := range data {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f') {
			cleanData += string(r)
		}
	}

	col := color.RGBA{0, 0, 0, 255}

	// Each hex character represents 4 pixels
	pixelsPerByte := 8
	rowWidthPixels := bytesPerRow * pixelsPerByte

	row := 0
	pixelInRow := 0

	for i := 0; i < len(cleanData); i++ {
		hexChar := cleanData[i]
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

// rotateImage180 rotates an image 180 degrees.
func rotateImage180(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Map (x, y) to (w-1-x, h-1-y)
			dst.Set(w-1-x, h-1-y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return dst
}
