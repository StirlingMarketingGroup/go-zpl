package render

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	imgdraw "golang.org/x/image/draw"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

//go:embed font0.ttf
var font0Data []byte

//go:embed ocr_b.ttf
var ocrBData []byte

//go:embed dejavu_sans_mono.ttf
var dejaVuSansMonoData []byte

//go:embed noto_sans_cjk_sc_bold.otf
var notoCJKData []byte

// fontManager handles font loading and text rendering.
type fontManager struct {
	// Loaded fonts
	font0   *opentype.Font // CG Triumvirate Bold Condensed approximation (scanned from Zebra)
	fontA   *opentype.Font // DejaVu Sans Mono (clean monospace with slashed zeros)
	fontE   *opentype.Font // OCR-B
	fontCJK *opentype.Font // Noto Sans CJK SC Bold (fallback for CJK characters)

	// Cache for font faces at different sizes
	faceCache    map[faceCacheKey]font.Face
	faceCacheCJK map[faceCacheKey]font.Face
	mu           sync.RWMutex
}

type faceCacheKey struct {
	font   zpl.Font
	height int
}

func newFontManager() (*fontManager, error) {
	f0, err := opentype.Parse(font0Data)
	if err != nil {
		return nil, err
	}

	fA, err := opentype.Parse(dejaVuSansMonoData)
	if err != nil {
		return nil, err
	}

	fE, err := opentype.Parse(ocrBData)
	if err != nil {
		return nil, err
	}

	fCJK, err := opentype.Parse(notoCJKData)
	if err != nil {
		return nil, err
	}

	return &fontManager{
		font0:        f0,
		fontA:        fA,
		fontE:        fE,
		fontCJK:      fCJK,
		faceCache:    make(map[faceCacheKey]font.Face),
		faceCacheCJK: make(map[faceCacheKey]font.Face),
	}, nil
}

// getFont returns the opentype.Font for the given ZPL font identifier.
func (fm *fontManager) getFont(f zpl.Font) *opentype.Font {
	switch f {
	case zpl.FontA:
		return fm.fontA
	case zpl.FontE:
		return fm.fontE
	default:
		// Font 0 and all other fonts fall back to the default
		return fm.font0
	}
}

// getFace returns a font face for the given font and size.
// The face is cached for reuse.
func (fm *fontManager) getFace(f zpl.Font, height int) (font.Face, error) {
	key := faceCacheKey{font: f, height: height}

	fm.mu.RLock()
	if face, ok := fm.faceCache[key]; ok {
		fm.mu.RUnlock()
		return face, nil
	}
	fm.mu.RUnlock()

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Double-check after acquiring write lock
	if face, ok := fm.faceCache[key]; ok {
		return face, nil
	}

	// Get the appropriate font
	otFont := fm.getFont(f)

	size := float64(height)

	// Create new face
	face, err := opentype.NewFace(otFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72, // Standard screen DPI for point-to-pixel conversion
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	fm.faceCache[key] = face
	return face, nil
}

// getCJKFace returns the CJK fallback font face for the given size.
func (fm *fontManager) getCJKFace(height int) (font.Face, error) {
	key := faceCacheKey{font: zpl.Font0, height: height} // Use Font0 as key placeholder

	fm.mu.RLock()
	if face, ok := fm.faceCacheCJK[key]; ok {
		fm.mu.RUnlock()
		return face, nil
	}
	fm.mu.RUnlock()

	fm.mu.Lock()
	defer fm.mu.Unlock()

	if face, ok := fm.faceCacheCJK[key]; ok {
		return face, nil
	}

	face, err := opentype.NewFace(fm.fontCJK, &opentype.FaceOptions{
		Size:    float64(height),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	fm.faceCacheCJK[key] = face
	return face, nil
}

// hasGlyph checks if the given face has a glyph for the rune.
func hasGlyph(face font.Face, r rune) bool {
	_, ok := face.GlyphAdvance(r)
	return ok
}

// scaleImageHorizontally scales an image horizontally using bilinear interpolation.
func scaleImageHorizontally(src *image.RGBA, scaleX float64) *image.RGBA {
	if scaleX == 1.0 {
		return src
	}
	srcBounds := src.Bounds()
	destWidth := int(math.Round(float64(srcBounds.Dx()) * scaleX))
	destHeight := srcBounds.Dy()
	if destWidth <= 0 {
		destWidth = 1
	}

	dest := image.NewRGBA(image.Rect(0, 0, destWidth, destHeight))

	// Use affine transformation for horizontal scaling
	// The transform maps destination coords to source coords, so we use 1/scaleX
	imgdraw.ApproxBiLinear.Scale(dest, dest.Bounds(), src, srcBounds, imgdraw.Over, nil)

	return dest
}

// drawText renders text to the image at the given position.
func (fm *fontManager) drawText(img *image.RGBA, text string, x, y int, f zpl.Font, height, width int, orient zpl.Orientation, reverse bool) {
	if text == "" {
		return
	}

	face, err := fm.getFace(f, height)
	if err != nil {
		return
	}

	// Calculate scale factor for width adjustment
	// If width == 0, use natural font proportions (no scaling)
	// If width != 0, scale horizontally to match requested width
	scaleX := 1.0
	if width != 0 {
		scaleX = float64(width) / float64(height)
	}

	// Handle orientation by rendering to a temporary image then rotating
	switch orient {
	case zpl.OrientationNormal:
		fm.drawTextNormal(img, face, text, x, y, height, scaleX, reverse)
	case zpl.OrientationRotated90:
		fm.drawTextRotated90(img, face, text, x, y, height, scaleX, reverse)
	case zpl.OrientationRotated180:
		fm.drawTextRotated180(img, face, text, x, y, height, scaleX, reverse)
	case zpl.OrientationRotated270:
		fm.drawTextRotated270(img, face, text, x, y, height, scaleX, reverse)
	default:
		fm.drawTextNormal(img, face, text, x, y, height, scaleX, reverse)
	}
}

// drawTextNormal draws text in normal orientation (0 degrees).
func (fm *fontManager) drawTextNormal(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool) {
	// Get CJK fallback face
	cjkFace, _ := fm.getCJKFace(height)

	// Calculate actual bounds by scanning glyphs
	// Font metrics can underreport ascent - some glyphs extend above the reported ascent
	metrics := face.Metrics()
	maxAscent := metrics.Ascent
	maxDescent := metrics.Descent
	totalWidth := fixed.Int26_6(0)

	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		bounds, adv, ok := currentFace.GlyphBounds(r)
		if ok {
			totalWidth += adv
			// bounds.Min.Y is negative (above baseline), bounds.Max.Y is positive (below)
			if -bounds.Min.Y > maxAscent {
				maxAscent = -bounds.Min.Y
			}
			if bounds.Max.Y > maxDescent {
				maxDescent = bounds.Max.Y
			}
		}
	}

	ascent := maxAscent.Round()
	textHeight := (maxAscent + maxDescent).Round()
	naturalWidth := totalWidth.Round()

	if naturalWidth == 0 || textHeight == 0 {
		return
	}

	col := color.RGBA{0, 0, 0, 255}
	if reverse {
		col = color.RGBA{255, 255, 255, 255}
	}

	// If scaleX != 1.0, render to temp image at natural width, then scale
	if scaleX != 1.0 {
		// Create temp image at natural width
		tmpImg := image.NewRGBA(image.Rect(0, 0, naturalWidth, textHeight))
		bgCol := color.RGBA{255, 255, 255, 255}
		if reverse {
			bgCol = color.RGBA{0, 0, 0, 255}
		}
		draw.Draw(tmpImg, tmpImg.Bounds(), image.NewUniform(bgCol), image.Point{}, draw.Src)

		drawer := &font.Drawer{
			Dst:  tmpImg,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent)},
		}

		// Draw each character at natural width
		for _, r := range text {
			currentFace := face
			if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
				currentFace = cjkFace
			}
			drawer.Face = currentFace
			drawer.DrawString(string(r))
		}

		// Scale horizontally
		scaled := scaleImageHorizontally(tmpImg, scaleX)

		// Copy to destination, only copying non-background pixels
		for ty := 0; ty < scaled.Bounds().Dy(); ty++ {
			for tx := 0; tx < scaled.Bounds().Dx(); tx++ {
				c := scaled.RGBAAt(tx, ty)
				if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
					destX := x + tx
					destY := y + ty
					if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
						img.Set(destX, destY, c)
					}
				}
			}
		}
	} else {
		// No scaling needed, draw directly
		baselineY := y + ascent
		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(baselineY)},
		}

		for _, r := range text {
			currentFace := face
			if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
				currentFace = cjkFace
			}
			drawer.Face = currentFace
			drawer.DrawString(string(r))
		}
	}
}

// drawTextRotated90 draws text rotated 90 degrees clockwise.
func (fm *fontManager) drawTextRotated90(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool) {
	// Get CJK fallback face
	cjkFace, _ := fm.getCJKFace(height)

	// Calculate actual bounds by scanning glyphs
	metrics := face.Metrics()
	maxAscent := metrics.Ascent
	maxDescent := metrics.Descent
	totalWidth := fixed.Int26_6(0)

	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		bounds, adv, ok := currentFace.GlyphBounds(r)
		if ok {
			totalWidth += adv
			if -bounds.Min.Y > maxAscent {
				maxAscent = -bounds.Min.Y
			}
			if bounds.Max.Y > maxDescent {
				maxDescent = bounds.Max.Y
			}
		}
	}

	ascent := maxAscent.Round()
	textHeight := (maxAscent + maxDescent).Round()
	naturalWidth := totalWidth.Round()

	if naturalWidth == 0 || textHeight == 0 {
		return
	}

	// Create temporary image at NATURAL width (no scaling yet)
	tmpImg := image.NewRGBA(image.Rect(0, 0, naturalWidth, textHeight))
	bgCol := color.RGBA{255, 255, 255, 255}
	if reverse {
		bgCol = color.RGBA{0, 0, 0, 255}
	}
	draw.Draw(tmpImg, tmpImg.Bounds(), image.NewUniform(bgCol), image.Point{}, draw.Src)

	col := color.RGBA{0, 0, 0, 255}
	if reverse {
		col = color.RGBA{255, 255, 255, 255}
	}

	drawer := &font.Drawer{
		Dst:  tmpImg,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent)},
	}

	// Draw each character, using CJK fallback when needed
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		drawer.Face = currentFace
		drawer.DrawString(string(r))
	}

	// Scale horizontally if needed
	scaled := tmpImg
	if scaleX != 1.0 {
		scaled = scaleImageHorizontally(tmpImg, scaleX)
	}
	scaledWidth := scaled.Bounds().Dx()

	// Rotate 90 degrees clockwise and copy to destination
	// (x, y) -> (y, width - 1 - x)
	for ty := 0; ty < textHeight; ty++ {
		for tx := 0; tx < scaledWidth; tx++ {
			c := scaled.RGBAAt(tx, ty)
			if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
				destX := x + ty
				destY := y + scaledWidth - 1 - tx
				if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
					img.Set(destX, destY, c)
				}
			}
		}
	}
}

// drawTextRotated180 draws text rotated 180 degrees.
func (fm *fontManager) drawTextRotated180(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool) {
	// Get CJK fallback face
	cjkFace, _ := fm.getCJKFace(height)

	// Calculate actual bounds by scanning glyphs
	metrics := face.Metrics()
	maxAscent := metrics.Ascent
	maxDescent := metrics.Descent
	totalWidth := fixed.Int26_6(0)

	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		bounds, adv, ok := currentFace.GlyphBounds(r)
		if ok {
			totalWidth += adv
			if -bounds.Min.Y > maxAscent {
				maxAscent = -bounds.Min.Y
			}
			if bounds.Max.Y > maxDescent {
				maxDescent = bounds.Max.Y
			}
		}
	}

	ascent := maxAscent.Round()
	textHeight := (maxAscent + maxDescent).Round()
	naturalWidth := totalWidth.Round()

	if naturalWidth == 0 || textHeight == 0 {
		return
	}

	// Create temporary image at NATURAL width
	tmpImg := image.NewRGBA(image.Rect(0, 0, naturalWidth, textHeight))
	bgCol := color.RGBA{255, 255, 255, 255}
	if reverse {
		bgCol = color.RGBA{0, 0, 0, 255}
	}
	draw.Draw(tmpImg, tmpImg.Bounds(), image.NewUniform(bgCol), image.Point{}, draw.Src)

	col := color.RGBA{0, 0, 0, 255}
	if reverse {
		col = color.RGBA{255, 255, 255, 255}
	}

	drawer := &font.Drawer{
		Dst:  tmpImg,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent)},
	}

	// Draw each character, using CJK fallback when needed
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		drawer.Face = currentFace
		drawer.DrawString(string(r))
	}

	// Scale horizontally if needed
	scaled := tmpImg
	if scaleX != 1.0 {
		scaled = scaleImageHorizontally(tmpImg, scaleX)
	}
	scaledWidth := scaled.Bounds().Dx()

	// Rotate 180 degrees: (x, y) -> (width - 1 - x, height - 1 - y)
	for ty := 0; ty < textHeight; ty++ {
		for tx := 0; tx < scaledWidth; tx++ {
			c := scaled.RGBAAt(tx, ty)
			if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
				destX := x + scaledWidth - 1 - tx
				destY := y + textHeight - 1 - ty
				if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
					img.Set(destX, destY, c)
				}
			}
		}
	}
}

// drawTextRotated270 draws text rotated 270 degrees clockwise (90 counter-clockwise).
func (fm *fontManager) drawTextRotated270(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool) {
	// Get CJK fallback face
	cjkFace, _ := fm.getCJKFace(height)

	// Calculate actual bounds by scanning glyphs
	metrics := face.Metrics()
	maxAscent := metrics.Ascent
	maxDescent := metrics.Descent
	totalWidth := fixed.Int26_6(0)

	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		bounds, adv, ok := currentFace.GlyphBounds(r)
		if ok {
			totalWidth += adv
			if -bounds.Min.Y > maxAscent {
				maxAscent = -bounds.Min.Y
			}
			if bounds.Max.Y > maxDescent {
				maxDescent = bounds.Max.Y
			}
		}
	}

	ascent := maxAscent.Round()
	textHeight := (maxAscent + maxDescent).Round()
	naturalWidth := totalWidth.Round()

	if naturalWidth == 0 || textHeight == 0 {
		return
	}

	// Create temporary image at NATURAL width
	tmpImg := image.NewRGBA(image.Rect(0, 0, naturalWidth, textHeight))
	bgCol := color.RGBA{255, 255, 255, 255}
	if reverse {
		bgCol = color.RGBA{0, 0, 0, 255}
	}
	draw.Draw(tmpImg, tmpImg.Bounds(), image.NewUniform(bgCol), image.Point{}, draw.Src)

	col := color.RGBA{0, 0, 0, 255}
	if reverse {
		col = color.RGBA{255, 255, 255, 255}
	}

	drawer := &font.Drawer{
		Dst:  tmpImg,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent)},
	}

	// Draw each character, using CJK fallback when needed
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		drawer.Face = currentFace
		drawer.DrawString(string(r))
	}

	// Scale horizontally if needed
	scaled := tmpImg
	if scaleX != 1.0 {
		scaled = scaleImageHorizontally(tmpImg, scaleX)
	}
	scaledWidth := scaled.Bounds().Dx()

	// Rotate 270 degrees clockwise (90 counter-clockwise): (x, y) -> (height - 1 - y, x)
	for ty := 0; ty < textHeight; ty++ {
		for tx := 0; tx < scaledWidth; tx++ {
			c := scaled.RGBAAt(tx, ty)
			if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
				destX := x + textHeight - 1 - ty
				destY := y + tx
				if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
					img.Set(destX, destY, c)
				}
			}
		}
	}
}
