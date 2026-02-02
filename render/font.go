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

//go:embed letter_gothic_bold.ttf
var letterGothicBoldData []byte

// fontManager handles font loading and text rendering.
type fontManager struct {
	// Loaded fonts
	font0   *opentype.Font // CG Triumvirate Bold Condensed approximation (scanned from Zebra)
	fontA   *opentype.Font // DejaVu Sans Mono (clean monospace with slashed zeros)
	fontD   *opentype.Font // Letter Gothic Bold (typewriter-style for ^Ab, ^Ac, ^Ad)
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

func fontScale(f zpl.Font, height int) float64 { //nolint:revive,unparam // height reserved for future per-size tuning
	switch f {
	case zpl.FontB:
		return 1.5
	case zpl.FontD:
		return 1.15
	default:
		return 1.0
	}
}

func fontSizeAdjust(f zpl.Font, height int) float64 { //nolint:revive,unparam // height reserved for future per-size tuning
	switch f {
	case zpl.FontD:
		return -0.5
	default:
		return 0.0
	}
}

func fontWidthScale(f zpl.Font, height int) float64 { //nolint:revive,unparam // height reserved for future per-size tuning
	switch f {
	case zpl.FontB:
		return 1.3
	case zpl.FontD:
		return 0.95
	default:
		return 1.0
	}
}

func fontBaselineAdjust(f zpl.Font, height int) int { //nolint:revive,unparam // height reserved for future per-size tuning
	switch f {
	case zpl.FontB:
		// Lift small labels slightly so they don't sit on the divider line.
		return -4
	case zpl.FontD:
		// Keep baseline steady when scaling Font D taller.
		return -int(math.Round(float64(height) * 0.15))
	default:
		return 0
	}
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

	fD, err := opentype.Parse(letterGothicBoldData)
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
		fontD:        fD,
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
	case zpl.FontB:
		// Font B should match the small typewriter-style text
		return fm.fontD
	case zpl.FontC, zpl.FontD:
		// Fonts C, D are typewriter-style monospace fonts
		return fm.fontD
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

	size := float64(height)*fontScale(f, height) + fontSizeAdjust(f, height)

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

func fontBoldness(f zpl.Font, height int) int { //nolint:revive,unparam // height reserved for future per-size tuning
	switch f {
	case zpl.FontC:
		return 1
	default:
		return 0
	}
}

func fontInkGain(f zpl.Font, height int) int { //nolint:revive,unparam // height reserved for future per-size tuning
	switch f {
	case zpl.Font0:
		return 1
	default:
		return 0
	}
}

// scaleImageHorizontally scales an image horizontally using nearest-neighbor interpolation.
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

	// Use nearest-neighbor scaling to keep hard pixel edges for ZPL-style text.
	imgdraw.NearestNeighbor.Scale(dest, dest.Bounds(), src, srcBounds, imgdraw.Over, nil)

	return dest
}

func blendTextPixel(dst color.RGBA, src color.RGBA, reverse bool) color.RGBA {
	if reverse {
		if src.R > dst.R {
			dst.R = src.R
		}
		if src.G > dst.G {
			dst.G = src.G
		}
		if src.B > dst.B {
			dst.B = src.B
		}
		dst.A = 255
		return dst
	}

	if src.R < dst.R {
		dst.R = src.R
	}
	if src.G < dst.G {
		dst.G = src.G
	}
	if src.B < dst.B {
		dst.B = src.B
	}
	dst.A = 255
	return dst
}

// drawText renders text to the image at the given position.
// When useBaseline is true (^FT), y is the text baseline.
// When useBaseline is false (^FO), y is the top of the text.
func (fm *fontManager) drawText(img *image.RGBA, text string, x, y int, f zpl.Font, height, width int, orient zpl.Orientation, reverse bool, useBaseline bool) {
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
	scaleX *= fontWidthScale(f, height)

	// Determine boldness (synthetic bold)
	// Font C is a bitmap font on real printers and is quite bold/blocky.
	// Our outline font is too thin, so we smear it.
	boldness := fontBoldness(f, height)
	inkGain := fontInkGain(f, height)
	baselineAdjust := fontBaselineAdjust(f, height)

	// Handle orientation by rendering to a temporary image then rotating
	switch orient {
	case zpl.OrientationNormal:
		fm.drawTextNormal(img, face, text, x, y, height, scaleX, reverse, boldness, inkGain, baselineAdjust, useBaseline)
	case zpl.OrientationRotated90:
		fm.drawTextRotated90(img, face, text, x, y, height, scaleX, reverse, boldness, inkGain, baselineAdjust, useBaseline)
	case zpl.OrientationRotated180:
		fm.drawTextRotated180(img, face, text, x, y, height, scaleX, reverse, boldness, inkGain, baselineAdjust, useBaseline)
	case zpl.OrientationRotated270:
		fm.drawTextRotated270(img, face, text, x, y, height, scaleX, reverse, boldness, inkGain, baselineAdjust, useBaseline)
	default:
		fm.drawTextNormal(img, face, text, x, y, height, scaleX, reverse, boldness, inkGain, baselineAdjust, useBaseline)
	}
}

// measureTextWidth returns the rendered width in pixels for the given text.
func (fm *fontManager) measureTextWidth(text string, f zpl.Font, height, width int) int {
	if text == "" {
		return 0
	}

	face, err := fm.getFace(f, height)
	if err != nil {
		return 0
	}

	cjkFace, _ := fm.getCJKFace(height)

	scaleX := 1.0
	if width != 0 {
		scaleX = float64(width) / float64(height)
	}
	scaleX *= fontWidthScale(f, height)

	penX := fixed.Int26_6(0)
	var minX, maxX fixed.Int26_6
	hasBounds := false
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		bounds, adv, ok := currentFace.GlyphBounds(r)
		if ok {
			glyphMinX := penX + bounds.Min.X
			glyphMaxX := penX + bounds.Max.X
			if !hasBounds {
				minX = glyphMinX
				maxX = glyphMaxX
				hasBounds = true
			} else {
				if glyphMinX < minX {
					minX = glyphMinX
				}
				if glyphMaxX > maxX {
					maxX = glyphMaxX
				}
			}
		}
		penX += adv
	}

	if !hasBounds {
		return 0
	}

	naturalWidth := (maxX - minX).Round()
	if naturalWidth == 0 {
		return 0
	}
	naturalWidth += fontBoldness(f, height)

	return int(math.Round(float64(naturalWidth) * scaleX))
}

// drawTextNormal draws text in normal orientation (0 degrees).
// When useBaseline is true, y is the text baseline (^FT).
// When useBaseline is false, y is the top of the text (^FO).
func (fm *fontManager) drawTextNormal(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool, boldness, inkGain int, baselineAdjust int, useBaseline bool) {
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

	// Calculate the top-left Y position for rendering
	// When useBaseline is true (^FT), y is the baseline, so top = y - ascent
	// When useBaseline is false (^FO), y is already the top
	topY := y
	if useBaseline {
		topY = y - ascent
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
			Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent + baselineAdjust)},
		}

		// Draw text with synthetic bold by drawing at multiple offsets
		// Just horizontal offset (no vertical) for cleaner look
		boldOffsets := []image.Point{{0, 0}}
		if boldness > 0 {
			boldOffsets = []image.Point{{0, 0}, {1, 0}} // Original + 1px right
		}

		for _, offset := range boldOffsets {
			for pass := 0; pass <= inkGain; pass++ {
				drawer.Dot = fixed.Point26_6{X: fixed.I(offset.X), Y: fixed.I(ascent + baselineAdjust + offset.Y)}
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

		// Scale horizontally
		scaled := scaleImageHorizontally(tmpImg, scaleX)

		// Copy to destination, only copying non-background pixels
		for ty := 0; ty < scaled.Bounds().Dy(); ty++ {
			for tx := 0; tx < scaled.Bounds().Dx(); tx++ {
				c := scaled.RGBAAt(tx, ty)
				if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
					destX := x + tx
					destY := topY + ty
					if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
						dst := img.RGBAAt(destX, destY)
						img.Set(destX, destY, blendTextPixel(dst, c, reverse))
					}
				}
			}
		}
	} else {
		// No scaling needed, draw directly
		baselineY := topY + ascent + baselineAdjust
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

			startDot := drawer.Dot
			adv, _ := currentFace.GlyphAdvance(r)

			for pass := 0; pass <= inkGain; pass++ {
				drawer.Dot = startDot
				drawer.DrawString(string(r))

				if boldness > 0 {
					for b := 1; b <= boldness; b++ {
						drawer.Dot = startDot
						drawer.Dot.X += fixed.I(b)
						drawer.DrawString(string(r))
					}
				}
			}

			drawer.Dot = startDot
			drawer.Dot.X += adv
		}
	}
}

// drawTextRotated90 draws text rotated 90 degrees clockwise.
// Text reads top-to-bottom, with letters facing right.
// When useBaseline is true (^FT), y is the baseline (end of text in rotated sense).
// When useBaseline is false (^FO), y is the top of the text.
func (fm *fontManager) drawTextRotated90(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool, boldness, inkGain int, baselineAdjust int, useBaseline bool) {
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
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent + baselineAdjust)},
	}

	// Draw each character, using CJK fallback when needed
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		drawer.Face = currentFace

		startDot := drawer.Dot
		adv, _ := currentFace.GlyphAdvance(r)

		for pass := 0; pass <= inkGain; pass++ {
			drawer.Dot = startDot
			drawer.DrawString(string(r))

			if boldness > 0 {
				for b := 1; b <= boldness; b++ {
					drawer.Dot = startDot
					drawer.Dot.X += fixed.I(b)
					drawer.DrawString(string(r))
				}
			}
		}

		drawer.Dot = startDot
		drawer.Dot.X += adv
	}

	// Scale horizontally if needed
	scaled := tmpImg
	if scaleX != 1.0 {
		scaled = scaleImageHorizontally(tmpImg, scaleX)
	}
	scaledWidth := scaled.Bounds().Dx()

	// Rotate 90 degrees clockwise: text goes DOWN the page
	// First char at top, last char at bottom, letters face right
	//
	// For baseline positioning (^FT):
	// - x is where the baseline should be (baseline runs vertically on left side after rotation)
	// - y is where the first character's baseline is, text extends downward
	// For origin positioning (^FO), (x, y) is the top-left of bounding box.
	for ty := 0; ty < textHeight; ty++ {
		for tx := 0; tx < scaledWidth; tx++ {
			c := scaled.RGBAAt(tx, ty)
			if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
				var destX, destY int
				if useBaseline {
					// Baseline mode: x is the baseline position
					// After 90° clockwise, baseline (at ty=ascent in temp) should be at x
					// The descent area (ty > ascent) goes to the left of baseline
					destX = x + ascent - ty
					// y is start of text, extends downward
					destY = y + tx
				} else {
					// Origin mode: (x, y) is top-left of bounding box
					destX = x + textHeight - 1 - ty
					destY = y + tx
				}
				if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
					dst := img.RGBAAt(destX, destY)
					img.Set(destX, destY, blendTextPixel(dst, c, reverse))
				}
			}
		}
	}
}

// drawTextRotated180 draws text rotated 180 degrees.
// When useBaseline is true (^FT), (x, y) is the baseline of the first character,
// which after 180° rotation appears at the right end of the text.
// When useBaseline is false (^FO), (x, y) is the top-left of the bounding box.
func (fm *fontManager) drawTextRotated180(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool, boldness, inkGain int, baselineAdjust int, useBaseline bool) {
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
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent + baselineAdjust)},
	}

	// Draw each character, using CJK fallback when needed
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		drawer.Face = currentFace

		startDot := drawer.Dot
		adv, _ := currentFace.GlyphAdvance(r)

		for pass := 0; pass <= inkGain; pass++ {
			drawer.Dot = startDot
			drawer.DrawString(string(r))

			if boldness > 0 {
				for b := 1; b <= boldness; b++ {
					drawer.Dot = startDot
					drawer.Dot.X += fixed.I(b)
					drawer.DrawString(string(r))
				}
			}
		}

		drawer.Dot = startDot
		drawer.Dot.X += adv
	}

	// Scale horizontally if needed
	scaled := tmpImg
	if scaleX != 1.0 {
		scaled = scaleImageHorizontally(tmpImg, scaleX)
	}
	scaledWidth := scaled.Bounds().Dx()

	// Rotate 180 degrees: text is upside-down, reading right-to-left
	// For ^FT (useBaseline=true): (x, y) is the baseline of the first character,
	// which after rotation is at the right end. Text extends leftward from x.
	// For ^FO (useBaseline=false): (x, y) is the top-left, text extends rightward.
	for ty := 0; ty < textHeight; ty++ {
		for tx := 0; tx < scaledWidth; tx++ {
			c := scaled.RGBAAt(tx, ty)
			if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
				var destX, destY int
				if useBaseline {
					// Baseline mode: first char baseline at (x, y), text extends left
					// The baseline in temp image is at ty=ascent, so adjust to place it at y
					destX = x - tx
					destY = y + ascent - ty
				} else {
					// Origin mode: top-left at (x, y), text extends right and down
					destX = x + scaledWidth - 1 - tx
					destY = y + textHeight - 1 - ty
				}
				if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
					dst := img.RGBAAt(destX, destY)
					img.Set(destX, destY, blendTextPixel(dst, c, reverse))
				}
			}
		}
	}
}

// drawTextRotated270 draws text rotated 270 degrees clockwise (90 counter-clockwise).
// Text reads bottom-to-top, with letters facing left.
// When useBaseline is true (^FT), y is the baseline (start of text for upward reading).
// When useBaseline is false (^FO), y is the top-left of the bounding box.
func (fm *fontManager) drawTextRotated270(img *image.RGBA, face font.Face, text string, x, y, height int, scaleX float64, reverse bool, boldness, inkGain int, baselineAdjust int, useBaseline bool) {
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
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(ascent + baselineAdjust)},
	}

	// Draw each character, using CJK fallback when needed
	for _, r := range text {
		currentFace := face
		if !hasGlyph(face, r) && cjkFace != nil && hasGlyph(cjkFace, r) {
			currentFace = cjkFace
		}
		drawer.Face = currentFace

		startDot := drawer.Dot
		adv, _ := currentFace.GlyphAdvance(r)

		for pass := 0; pass <= inkGain; pass++ {
			drawer.Dot = startDot
			drawer.DrawString(string(r))

			if boldness > 0 {
				for b := 1; b <= boldness; b++ {
					drawer.Dot = startDot
					drawer.Dot.X += fixed.I(b)
					drawer.DrawString(string(r))
				}
			}
		}

		drawer.Dot = startDot
		drawer.Dot.X += adv
	}

	// Scale horizontally if needed
	scaled := tmpImg
	if scaleX != 1.0 {
		scaled = scaleImageHorizontally(tmpImg, scaleX)
	}
	scaledWidth := scaled.Bounds().Dx()

	// Rotate 270 degrees clockwise (90 counter-clockwise): text goes UP the page
	// First char at bottom, last char at top, letters face left
	//
	// For baseline positioning (^FT):
	// - x is where the baseline should be (baseline runs vertically on right side after rotation)
	// - y is where the first character starts, text extends upward
	// For origin positioning (^FO), (x, y) is the top-left of bounding box.
	for ty := 0; ty < textHeight; ty++ {
		for tx := 0; tx < scaledWidth; tx++ {
			c := scaled.RGBAAt(tx, ty)
			if c.A > 0 && (c.R != bgCol.R || c.G != bgCol.G || c.B != bgCol.B) {
				var destX, destY int
				if useBaseline {
					// Baseline mode: x is the baseline position
					// The baseline in the temp image is at ty=ascent, so we offset by ascent
					// to place the baseline at x
					destX = x + ty - ascent
					// y is start of text, extends upward (adjust by 2 for pixel-perfect alignment)
					destY = y - tx - 2
				} else {
					// Origin mode: (x, y) is top-left of bounding box
					destX = x + ty
					destY = y + scaledWidth - 1 - tx
				}
				if destX >= 0 && destX < img.Bounds().Max.X && destY >= 0 && destY < img.Bounds().Max.Y {
					dst := img.RGBAAt(destX, destY)
					img.Set(destX, destY, blendTextPixel(dst, c, reverse))
				}
			}
		}
	}
}
