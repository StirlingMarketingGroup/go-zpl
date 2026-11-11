// Package render converts ZPL labels into bitmap images. The target is
// pixel-for-pixel parity with a real Zebra printer; any deviation is a bug
// (^POI print-orientation inversion is one known gap).
//
// # Usage
//
//	renderer := render.New(zpl.DPI203).WithSize(812, 1218)
//	err := renderer.RenderPNG(label, w)
//
// Render returns an image.Image with a white background and black elements.
// RenderAll renders parsed labels one image each; pair it with zpl.ParseAll,
// which splits multi-page ZPL and skips setup-only blocks.
// IgnoreLabelHome is false by default so ^LH offsets match the printer;
// set it true for cleaner previews that ignore those offsets.
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

	"golang.org/x/image/font"
	"golang.org/x/text/encoding/charmap"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// Renderer holds configuration for rendering ZPL labels to images.
type Renderer struct {
	// DPI is the printer resolution (203, 300, or 600 dots per inch); zero means
	// DPI203. It sizes the default 4×6 inch canvas when neither the renderer nor
	// the label sets a size; ZPL coordinates are already in dots and are not
	// scaled.
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

// New creates a new Renderer with the given DPI (203, 300, or 600; zero means
// DPI203). Width and height will be taken from the label if not set.
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
	// A zero-value Renderer literal has no DPI; treat it as the 203 DPI default
	// so the fallback canvas is never 0×0.
	dpi := r.DPI
	if dpi == 0 {
		dpi = zpl.DPI203
	}
	switch dpi {
	case zpl.DPI203, zpl.DPI300, zpl.DPI600:
	default:
		return nil, fmt.Errorf("unsupported DPI %d: want 203, 300, or 600", dpi)
	}

	width := r.Width
	if width == 0 {
		width = label.Width()
	}
	if width == 0 {
		width = zpl.ToDots(4, zpl.UnitInches, dpi)
	}

	height := r.Height
	if height == 0 {
		height = label.Height()
	}
	if height == 0 {
		height = zpl.ToDots(6, zpl.UnitInches, dpi)
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

// RenderAll renders multiple labels and returns an image for each.
func (r *Renderer) RenderAll(labels []*zpl.Label) ([]image.Image, error) {
	images := make([]image.Image, 0, len(labels))
	for _, label := range labels {
		img, err := r.Render(label)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
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

	// Current position (set by ^FO or ^FT) and that field's z justification.
	curX         int
	curY         int
	fieldJustify zpl.Justification

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
	fieldReverse   bool
	fieldDirection zpl.Orientation // Default rotation for fields (^FW)
	fieldBlock     zpl.Command     // the next field's *zpl.FieldBlock or *zpl.TextBlock; the later command wins
	useBaseline    bool            // true when position set by ^FT (baseline), false for ^FO (top-left)
	charSet        int             // ^CI. 28 is UTF-8, the default when a label sets none.

	// Barcode defaults (set by ^BY)
	barcodeModuleWidth int
	barcodeHeight      int

	// Pending barcode (waiting for ^FD to provide data)
	// In ZPL, barcode commands like ^BC set up the barcode parameters,
	// then the following ^FD provides the data
	pendingBarcode interface{}
}

// The parsed fonts, shared by every render.
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
		img: img,
		fontMgr: &fontManager{
			font0:        fm.font0,
			fontA:        fm.fontA,
			fontD:        fm.fontD,
			fontE:        fm.fontE,
			fontCJK:      fm.fontCJK,
			faceCache:    make(map[faceCacheKey]font.Face),
			faceCacheCJK: make(map[faceCacheKey]font.Face),
		},
		currentFont:        zpl.Font0,
		fontHeight:         30,
		fontWidth:          0, // Zero means proportional width
		fontOrient:         zpl.OrientationNormal,
		fieldDirection:     zpl.OrientationNormal,
		charSet:            28,
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
		c.useBaseline = false
		c.fieldJustify = v.Justification

	case *zpl.FieldTypeset:
		c.curX = v.X + c.homeX
		c.curY = v.Y + c.homeY
		c.useBaseline = true
		c.fieldJustify = v.Justification

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

	case *zpl.FieldDirection:
		c.fieldDirection = v.Orientation

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
		c.fieldBlock = v

	case *zpl.TextBlock:
		c.fieldBlock = v

	case *zpl.CharacterSet:
		c.charSet = v.CharSet
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
	// ^FB applies to a single field, clear after consuming ^FD for barcode
	c.fieldBlock = nil
}

// decodeFieldText maps field bytes to characters for the current ^CI set.
// Sets 0–13 are code page 850, 27 is Windows-1252, and anything else
// (including the default 28, UTF-8) is left as it is.
func decodeFieldText(text string, charSet int) string {
	codePage := charmap.CodePage850
	switch {
	case charSet >= 0 && charSet <= 13:
	case charSet == 27:
		codePage = charmap.Windows1252
	default:
		return text
	}
	// A single-byte code page maps every byte, so decoding cannot fail.
	decoded, _ := codePage.NewDecoder().String(text)
	return decoded
}

// drawText renders text at the current position using the current font settings.
func (c *canvas) drawText(text string) {
	if c.fontMgr == nil || text == "" {
		c.fieldReverse = false
		c.fieldBlock = nil
		return
	}

	text = decodeFieldText(text, c.charSet)

	height := c.fontHeight
	if height == 0 {
		height = 30
	}

	// Font rotation (^A0R) wins. Otherwise the field direction (^FW) applies.
	orient := c.fontOrient
	if orient == zpl.OrientationNormal && c.fieldDirection != zpl.OrientationNormal {
		orient = c.fieldDirection
	}

	switch block := c.fieldBlock.(type) {
	case *zpl.FieldBlock:
		if block.Width > 0 {
			c.drawFieldBlock(text, block, orient, height)
			return
		}
	case *zpl.TextBlock:
		c.drawTextBlock(text, block, orient, height)
		return
	}

	x := c.curX
	y := c.curY
	// z = 1 right-justifies a text field that has no block. A block ignores z.
	if c.fieldJustify == zpl.JustifyRight {
		tw := c.fontMgr.measureTextWidth(text, c.currentFont, height, c.fontWidth)
		switch orient {
		case zpl.OrientationRotated90:
			if c.useBaseline {
				y -= tw
			} else {
				x -= height
			}
		case zpl.OrientationRotated180:
			if c.useBaseline {
				x += tw
			} else {
				x -= tw
			}
		case zpl.OrientationRotated270:
			if c.useBaseline {
				y += tw
			} else {
				x -= height
			}
		default:
			x -= tw
		}
	}

	c.fontMgr.drawText(c.img, text, x, y, c.currentFont, height, c.fontWidth, orient, c.fieldReverse, c.useBaseline)
	c.fieldReverse = false
	c.fieldBlock = nil
}

// drawFieldBlock lays out a ^FB once, then maps each line into the field orientation.
// ^FT anchors the baseline of the last possible line (maxLines-1), not line 0.
func (c *canvas) drawFieldBlock(text string, block *zpl.FieldBlock, orient zpl.Orientation, height int) {
	maxLines := max(1, block.MaxLines)
	lines := c.layoutFieldBlockLines(text, block.Width, height)

	pitch := height + block.LineSpacing
	x := c.curX
	y := c.curY
	for i, line := range lines {
		tw := 0
		if line != "" {
			tw = c.fontMgr.measureTextWidth(line, c.currentFont, height, c.fontWidth)
		}
		u := 0
		switch block.Justification {
		case zpl.JustifyRight:
			u = block.Width - tw
		case zpl.JustifyCenter:
			u = (block.Width - tw) / 2
		}
		// A Zebra prints every line past the last one on top of it.
		i = min(i, maxLines-1)
		back := maxLines - 1 - i
		var lx, ly int
		switch orient {
		case zpl.OrientationRotated90:
			lx = x + back*pitch
			ly = y + u
		case zpl.OrientationRotated180:
			ly = y + back*pitch
			if c.useBaseline {
				lx = x - u
			} else {
				lx = x + block.Width - u - tw
			}
		case zpl.OrientationRotated270:
			if c.useBaseline {
				lx = x - back*pitch
				ly = y - u
			} else {
				lx = x + i*pitch
				ly = y + block.Width - u - tw
			}
		default:
			lx = x + u
			if c.useBaseline {
				ly = y - back*pitch
			} else {
				ly = y + i*pitch
			}
		}
		c.fontMgr.drawText(c.img, line, lx, ly, c.currentFont, height, c.fontWidth, orient, c.fieldReverse, c.useBaseline)
	}

	c.fieldReverse = false
	c.fieldBlock = nil
}

// drawTextBlock lays out a ^TB the way a Zebra ZP 450 prints it: lines sit
// one font height apart, \& is literal text, and whatever passes the block's
// edge is cut off.
func (c *canvas) drawTextBlock(text string, block *zpl.TextBlock, orient zpl.Orientation, height int) {
	if block.Orientation != 0 {
		orient = block.Orientation
	}

	// The block as it sits on the label. ^FO gives its top-left corner;
	// ^FT gives the corner where the last line would start.
	area := image.Rect(0, 0, block.Width, block.Height)
	switch orient {
	case zpl.OrientationRotated90:
		area = image.Rect(0, 0, block.Height, block.Width)
	case zpl.OrientationRotated180:
		if c.useBaseline {
			area = area.Sub(image.Pt(block.Width, 0))
		}
	case zpl.OrientationRotated270:
		area = image.Rect(0, 0, block.Height, block.Width)
		if c.useBaseline {
			area = area.Sub(image.Pt(block.Height, block.Width))
		}
	default:
		if c.useBaseline {
			area = area.Sub(image.Pt(0, block.Height))
		}
		// An upright block prints a tenth of the font height below its origin.
		area = area.Add(image.Pt(0, height/10))
	}
	area = area.Add(image.Pt(c.curX, c.curY))
	clipped := c.img.SubImage(area).(*image.RGBA)

	for i, line := range c.wrapFieldBlockSegment(text, block.Width, height) {
		down := i * height
		if down >= block.Height {
			break
		}
		tw := c.fontMgr.measureTextWidth(line, c.currentFont, height, c.fontWidth)
		var lx, ly int
		switch orient {
		case zpl.OrientationRotated90:
			lx, ly = area.Max.X-height-down, area.Min.Y
		case zpl.OrientationRotated180:
			lx, ly = area.Max.X-tw, area.Max.Y-height-down
		case zpl.OrientationRotated270:
			lx, ly = area.Min.X+down, area.Max.Y-tw
		default:
			lx, ly = area.Min.X, area.Min.Y+down
		}
		c.fontMgr.drawText(clipped, line, lx, ly, c.currentFont, height, c.fontWidth, orient, c.fieldReverse, false)
	}

	c.fieldReverse = false
	c.fieldBlock = nil
}

// layoutFieldBlockLines splits text on ZPL \& hard breaks and word-wraps each
// segment to the field block width.
func (c *canvas) layoutFieldBlockLines(text string, width, fontHeight int) []string {
	// Split on literal \&; a trailing \& yields no trailing empty line.
	segments := strings.Split(text, `\&`)
	if len(segments) > 0 && segments[len(segments)-1] == "" {
		segments = segments[:len(segments)-1]
	}

	lines := make([]string, 0, len(segments))
	for _, seg := range segments {
		wrapped := c.wrapFieldBlockSegment(seg, width, fontHeight)
		lines = append(lines, wrapped...)
	}
	return lines
}

// wrapFieldBlockSegment greedy word-wraps a single segment to maxWidth.
// Breaks on spaces; a single word wider than maxWidth breaks at the last
// character that fits.
func (c *canvas) wrapFieldBlockSegment(text string, maxWidth, fontHeight int) []string {
	if text == "" {
		return []string{""}
	}

	measure := func(s string) int {
		return c.fontMgr.measureTextWidth(s, c.currentFont, fontHeight, c.fontWidth)
	}

	// Fast path: already fits
	if measure(text) <= maxWidth {
		return []string{text}
	}

	words := strings.Split(text, " ")
	var lines []string
	var current string

	flush := func() {
		if current != "" {
			lines = append(lines, current)
			current = ""
		}
	}

	for _, word := range words {
		// Character-break a word that exceeds the width on its own
		if measure(word) > maxWidth {
			flush()
			parts := breakWordToWidth(word, maxWidth, measure)
			for i, part := range parts {
				if i < len(parts)-1 {
					lines = append(lines, part)
				} else {
					current = part
				}
			}
			continue
		}

		if current == "" {
			current = word
			continue
		}

		trial := current + " " + word
		if measure(trial) <= maxWidth {
			current = trial
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	flush()

	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

// breakWordToWidth splits a single word into pieces that each fit maxWidth.
func breakWordToWidth(word string, maxWidth int, measure func(string) int) []string {
	if word == "" {
		return []string{""}
	}
	if measure(word) <= maxWidth {
		return []string{word}
	}

	var parts []string
	var current string
	for _, r := range word {
		trial := current + string(r)
		if current != "" && measure(trial) > maxWidth {
			parts = append(parts, current)
			current = string(r)
		} else {
			current = trial
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	if len(parts) == 0 {
		return []string{word}
	}
	return parts
}

// drawBox renders a graphic box at the current position.
func (c *canvas) drawBox(box *zpl.GraphicBox) {
	x := c.curX
	y := c.curY
	// A box is never narrower or shorter than its border: ^GB812,1,8 is a
	// line 8 dots thick, and a width or height of 0 is a line of the border's thickness.
	t := box.Thickness
	w := max(box.Width, t)
	h := max(box.Height, t)
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
