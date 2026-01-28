package zpl

import (
	"fmt"
	"io"
)

// GraphicBox represents a ^GB command for drawing boxes/rectangles.
type GraphicBox struct {
	Width        int
	Height       int
	Thickness    int
	Color        LineColor
	CornerRadius int
}

// NewGraphicBox creates a new graphic box command.
func NewGraphicBox(width, height, thickness int) *GraphicBox {
	return &GraphicBox{
		Width:        width,
		Height:       height,
		Thickness:    thickness,
		Color:        LineColorBlack,
		CornerRadius: 0,
	}
}

// WithColor sets the line color.
func (g *GraphicBox) WithColor(color LineColor) *GraphicBox {
	g.Color = color
	return g
}

// WithCornerRadius sets the corner rounding radius.
func (g *GraphicBox) WithCornerRadius(radius int) *GraphicBox {
	g.CornerRadius = radius
	return g
}

// ZPL returns the ZPL representation.
func (g *GraphicBox) ZPL() string {
	return fmt.Sprintf("^GB%d,%d,%d,%c,%d", g.Width, g.Height, g.Thickness, g.Color, g.CornerRadius)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicBox) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}

// GraphicCircle represents a ^GC command for drawing circles.
type GraphicCircle struct {
	Diameter  int
	Thickness int
	Color     LineColor
}

// NewGraphicCircle creates a new graphic circle command.
func NewGraphicCircle(diameter, thickness int) *GraphicCircle {
	return &GraphicCircle{
		Diameter:  diameter,
		Thickness: thickness,
		Color:     LineColorBlack,
	}
}

// WithColor sets the line color.
func (g *GraphicCircle) WithColor(color LineColor) *GraphicCircle {
	g.Color = color
	return g
}

// ZPL returns the ZPL representation.
func (g *GraphicCircle) ZPL() string {
	return fmt.Sprintf("^GC%d,%d,%c", g.Diameter, g.Thickness, g.Color)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicCircle) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}

// GraphicDiagonalLine represents a ^GD command for drawing diagonal lines.
type GraphicDiagonalLine struct {
	Width       int
	Height      int
	Thickness   int
	Color       LineColor
	Orientation Orientation // R = right-leaning, L = left-leaning
}

// NewGraphicDiagonalLine creates a new graphic diagonal line command.
func NewGraphicDiagonalLine(width, height, thickness int) *GraphicDiagonalLine {
	return &GraphicDiagonalLine{
		Width:       width,
		Height:      height,
		Thickness:   thickness,
		Color:       LineColorBlack,
		Orientation: OrientationRotated90, // R = right-leaning (default)
	}
}

// WithColor sets the line color.
func (g *GraphicDiagonalLine) WithColor(color LineColor) *GraphicDiagonalLine {
	g.Color = color
	return g
}

// WithLeftLeaning sets the diagonal to lean left.
func (g *GraphicDiagonalLine) WithLeftLeaning() *GraphicDiagonalLine {
	g.Orientation = 'L'
	return g
}

// WithRightLeaning sets the diagonal to lean right.
func (g *GraphicDiagonalLine) WithRightLeaning() *GraphicDiagonalLine {
	g.Orientation = 'R'
	return g
}

// ZPL returns the ZPL representation.
func (g *GraphicDiagonalLine) ZPL() string {
	return fmt.Sprintf("^GD%d,%d,%d,%c,%c", g.Width, g.Height, g.Thickness, g.Color, g.Orientation)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicDiagonalLine) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}

// GraphicEllipse represents a ^GE command for drawing ellipses.
type GraphicEllipse struct {
	Width     int
	Height    int
	Thickness int
	Color     LineColor
}

// NewGraphicEllipse creates a new graphic ellipse command.
func NewGraphicEllipse(width, height, thickness int) *GraphicEllipse {
	return &GraphicEllipse{
		Width:     width,
		Height:    height,
		Thickness: thickness,
		Color:     LineColorBlack,
	}
}

// WithColor sets the line color.
func (g *GraphicEllipse) WithColor(color LineColor) *GraphicEllipse {
	g.Color = color
	return g
}

// ZPL returns the ZPL representation.
func (g *GraphicEllipse) ZPL() string {
	return fmt.Sprintf("^GE%d,%d,%d,%c", g.Width, g.Height, g.Thickness, g.Color)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicEllipse) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}

// GraphicSymbol represents a ^GS command for printing symbols.
type GraphicSymbol struct {
	Orientation Orientation
	Height      int
	Width       int
	Symbol      rune
}

// Predefined symbols for ^GS command.
const (
	SymbolRegisteredTrademark = 'A'
	SymbolCopyright           = 'B'
	SymbolTrademark           = 'C'
	SymbolUnderwritersLab     = 'D'
	SymbolCanadianStandards   = 'E'
)

// NewGraphicSymbol creates a new graphic symbol command.
func NewGraphicSymbol(symbol rune, height, width int) *GraphicSymbol {
	return &GraphicSymbol{
		Orientation: OrientationNormal,
		Height:      height,
		Width:       width,
		Symbol:      symbol,
	}
}

// WithOrientation sets the symbol orientation.
func (g *GraphicSymbol) WithOrientation(o Orientation) *GraphicSymbol {
	g.Orientation = o
	return g
}

// ZPL returns the ZPL representation.
func (g *GraphicSymbol) ZPL() string {
	return fmt.Sprintf("^GS%c,%d,%d^FD%c", g.Orientation, g.Height, g.Width, g.Symbol)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicSymbol) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}
