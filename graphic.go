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

// DiagonalOrientation represents the lean direction of a diagonal line.
type DiagonalOrientation rune

const (
	// DiagonalRightLeaning draws from upper-left to lower-right.
	DiagonalRightLeaning DiagonalOrientation = 'R'
	// DiagonalLeftLeaning draws from lower-left to upper-right.
	DiagonalLeftLeaning DiagonalOrientation = 'L'
)

// GraphicDiagonalLine represents a ^GD command for drawing diagonal lines.
type GraphicDiagonalLine struct {
	Width       int
	Height      int
	Thickness   int
	Color       LineColor
	Orientation DiagonalOrientation
}

// NewGraphicDiagonalLine creates a new graphic diagonal line command.
func NewGraphicDiagonalLine(width, height, thickness int) *GraphicDiagonalLine {
	return &GraphicDiagonalLine{
		Width:       width,
		Height:      height,
		Thickness:   thickness,
		Color:       LineColorBlack,
		Orientation: DiagonalRightLeaning,
	}
}

// WithColor sets the line color.
func (g *GraphicDiagonalLine) WithColor(color LineColor) *GraphicDiagonalLine {
	g.Color = color
	return g
}

// WithLeftLeaning sets the diagonal to lean left.
func (g *GraphicDiagonalLine) WithLeftLeaning() *GraphicDiagonalLine {
	g.Orientation = DiagonalLeftLeaning
	return g
}

// WithRightLeaning sets the diagonal to lean right.
func (g *GraphicDiagonalLine) WithRightLeaning() *GraphicDiagonalLine {
	g.Orientation = DiagonalRightLeaning
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
	return fmt.Sprintf("^GS%c,%d,%d^FD%c^FS", g.Orientation, g.Height, g.Width, g.Symbol)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicSymbol) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}

// GraphicFieldFormat represents the compression format for ^GF commands.
type GraphicFieldFormat rune

const (
	// GraphicFieldASCII is ASCII hexadecimal format (^GFA).
	GraphicFieldASCII GraphicFieldFormat = 'A'
	// GraphicFieldBinary is binary format (^GFB).
	GraphicFieldBinary GraphicFieldFormat = 'B'
	// GraphicFieldCompressed is compressed binary format (^GFC).
	GraphicFieldCompressed GraphicFieldFormat = 'C'
)

// GraphicField represents a ^GF command for embedding bitmap graphics.
type GraphicField struct {
	Format      GraphicFieldFormat
	DataBytes   int    // Total bytes in the data
	TotalBytes  int    // Total bytes comprising the graphic
	BytesPerRow int    // Number of bytes per row
	Data        string // Hex data (for ASCII format)
	BinaryData  []byte // Binary data (for binary format)
}

// NewGraphicFieldASCII creates a new graphic field with ASCII hex data.
func NewGraphicFieldASCII(bytesPerRow int, data string) *GraphicField {
	return &GraphicField{
		Format:      GraphicFieldASCII,
		DataBytes:   len(data) / 2, // 2 hex chars per byte
		TotalBytes:  len(data) / 2,
		BytesPerRow: bytesPerRow,
		Data:        data,
	}
}

// ZPL returns the ZPL representation.
func (g *GraphicField) ZPL() string {
	return fmt.Sprintf("^GF%c,%d,%d,%d,%s", g.Format, g.DataBytes, g.TotalBytes, g.BytesPerRow, g.Data)
}

// WriteTo writes the ZPL to the writer.
func (g *GraphicField) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, g.ZPL())
	return int64(n), err
}
