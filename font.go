package zpl

import (
	"fmt"
	"io"
)

// Font represents a ZPL font identifier.
type Font rune

// Built-in ZPL fonts.
const (
	FontA Font = 'A' // 9x5 matrix
	FontB Font = 'B' // 11x7 matrix
	FontC Font = 'C' // 18x10 matrix
	FontD Font = 'D' // 18x10 matrix
	FontE Font = 'E' // 28x15 matrix (OCR-B)
	FontF Font = 'F' // 26x13 matrix
	FontG Font = 'G' // 60x40 matrix
	FontH Font = 'H' // 21x13 matrix (OCR-A)
	Font0 Font = '0' // 15x12 matrix (default)

	// Additional fonts
	FontGS Font = 'S' // Symbol font
	FontP  Font = 'P' // Standard font
	FontQ  Font = 'Q' // Standard font
	FontR  Font = 'R' // Standard font
	FontS  Font = 'S' // Standard font
	FontT  Font = 'T' // Standard font
	FontU  Font = 'U' // Standard font
	FontV  Font = 'V' // Standard font
)

// ScalableFont represents a ^A command for selecting a font.
type ScalableFont struct {
	Font        Font
	Orientation Orientation
	Height      int // Character height in dots
	Width       int // Character width in dots (0 = proportional)
}

// NewScalableFont creates a new font selection command.
func NewScalableFont(font Font, height, width int) *ScalableFont {
	return &ScalableFont{
		Font:        font,
		Orientation: OrientationNormal,
		Height:      height,
		Width:       width,
	}
}

// WithOrientation sets the font orientation.
func (f *ScalableFont) WithOrientation(o Orientation) *ScalableFont {
	f.Orientation = o
	return f
}

// ZPL returns the ZPL representation.
func (f *ScalableFont) ZPL() string {
	return fmt.Sprintf("^A%c%c,%d,%d", f.Font, f.Orientation, f.Height, f.Width)
}

// WriteTo writes the ZPL to the writer.
func (f *ScalableFont) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// ChangeFont represents a ^CF command for changing the default font.
type ChangeFont struct {
	Font   Font
	Height int
	Width  int
}

// NewChangeFont creates a new change font command.
func NewChangeFont(font Font, height, width int) *ChangeFont {
	return &ChangeFont{
		Font:   font,
		Height: height,
		Width:  width,
	}
}

// ZPL returns the ZPL representation.
func (f *ChangeFont) ZPL() string {
	return fmt.Sprintf("^CF%c,%d,%d", f.Font, f.Height, f.Width)
}

// WriteTo writes the ZPL to the writer.
func (f *ChangeFont) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// CharacterSet represents a ^CI command for character set selection.
type CharacterSet struct {
	CharSet int
}

// Common character sets.
const (
	CharSetUSA          = 0  // USA1
	CharSetUSA2         = 1  // USA2
	CharSetUK           = 2  // UK
	CharSetDutch        = 3  // Holland
	CharSetDanish       = 4  // Denmark
	CharSetSwedish      = 5  // Sweden
	CharSetGerman       = 6  // Germany
	CharSetFrench       = 7  // France1
	CharSetFrench2      = 8  // France2
	CharSetItalian      = 9  // Italy
	CharSetSpanish      = 10 // Spain
	CharSetJIS          = 12 // Japanese
	CharSetUTF8         = 28 // UTF-8
	CharSetUTF16BigEndian = 29 // UTF-16 Big Endian
	CharSetUTF16LittleEndian = 30 // UTF-16 Little Endian
)

// NewCharacterSet creates a new character set command.
func NewCharacterSet(charSet int) *CharacterSet {
	return &CharacterSet{CharSet: charSet}
}

// ZPL returns the ZPL representation.
func (c *CharacterSet) ZPL() string {
	return fmt.Sprintf("^CI%d", c.CharSet)
}

// WriteTo writes the ZPL to the writer.
func (c *CharacterSet) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, c.ZPL())
	return int64(n), err
}
