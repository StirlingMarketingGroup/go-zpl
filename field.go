package zpl

import (
	"fmt"
	"io"
)

// FieldOrigin represents a ^FO command for positioning fields.
type FieldOrigin struct {
	X, Y         int
	Justification Justification // Z = use default
}

// NewFieldOrigin creates a new field origin command.
func NewFieldOrigin(x, y int) *FieldOrigin {
	return &FieldOrigin{X: x, Y: y}
}

// WithJustification sets the field justification.
func (f *FieldOrigin) WithJustification(j Justification) *FieldOrigin {
	f.Justification = j
	return f
}

// ZPL returns the ZPL representation.
func (f *FieldOrigin) ZPL() string {
	if f.Justification != 0 {
		return fmt.Sprintf("^FO%d,%d,%c", f.X, f.Y, f.Justification)
	}
	return fmt.Sprintf("^FO%d,%d", f.X, f.Y)
}

// WriteTo writes the ZPL to the writer.
func (f *FieldOrigin) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// FieldData represents a ^FD command containing field data.
type FieldData struct {
	Data string
}

// NewFieldData creates a new field data command.
func NewFieldData(data string) *FieldData {
	return &FieldData{Data: data}
}

// ZPL returns the ZPL representation.
func (f *FieldData) ZPL() string {
	return fmt.Sprintf("^FD%s^FS", f.Data)
}

// WriteTo writes the ZPL to the writer.
func (f *FieldData) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// FieldTypeset represents a ^FT command for baseline positioning.
type FieldTypeset struct {
	X, Y          int
	Justification Justification
}

// NewFieldTypeset creates a new field typeset command.
func NewFieldTypeset(x, y int) *FieldTypeset {
	return &FieldTypeset{X: x, Y: y}
}

// WithJustification sets the field justification.
func (f *FieldTypeset) WithJustification(j Justification) *FieldTypeset {
	f.Justification = j
	return f
}

// ZPL returns the ZPL representation.
func (f *FieldTypeset) ZPL() string {
	if f.Justification != 0 {
		return fmt.Sprintf("^FT%d,%d,%c", f.X, f.Y, f.Justification)
	}
	return fmt.Sprintf("^FT%d,%d", f.X, f.Y)
}

// WriteTo writes the ZPL to the writer.
func (f *FieldTypeset) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// FieldReverse represents a ^FR command for reverse printing.
type FieldReverse struct{}

// NewFieldReverse creates a new field reverse command.
func NewFieldReverse() *FieldReverse {
	return &FieldReverse{}
}

// ZPL returns the ZPL representation.
func (f *FieldReverse) ZPL() string {
	return "^FR"
}

// WriteTo writes the ZPL to the writer.
func (f *FieldReverse) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// FieldBlock represents a ^FB command for text block formatting.
type FieldBlock struct {
	Width            int
	MaxLines         int
	LineSpacing      int
	Justification    Justification
	HangingIndent    int
}

// NewFieldBlock creates a new field block command.
func NewFieldBlock(width int) *FieldBlock {
	return &FieldBlock{
		Width:         width,
		MaxLines:      1,
		LineSpacing:   0,
		Justification: JustifyLeft,
		HangingIndent: 0,
	}
}

// WithMaxLines sets the maximum number of lines.
func (f *FieldBlock) WithMaxLines(lines int) *FieldBlock {
	f.MaxLines = lines
	return f
}

// WithLineSpacing sets the line spacing adjustment.
func (f *FieldBlock) WithLineSpacing(spacing int) *FieldBlock {
	f.LineSpacing = spacing
	return f
}

// WithJustification sets the text justification.
func (f *FieldBlock) WithJustification(j Justification) *FieldBlock {
	f.Justification = j
	return f
}

// WithHangingIndent sets the hanging indent.
func (f *FieldBlock) WithHangingIndent(indent int) *FieldBlock {
	f.HangingIndent = indent
	return f
}

// ZPL returns the ZPL representation.
func (f *FieldBlock) ZPL() string {
	return fmt.Sprintf("^FB%d,%d,%d,%c,%d", f.Width, f.MaxLines, f.LineSpacing, f.Justification, f.HangingIndent)
}

// WriteTo writes the ZPL to the writer.
func (f *FieldBlock) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.ZPL())
	return int64(n), err
}

// Comment represents a ^FX command for adding comments.
type Comment struct {
	Text string
}

// NewComment creates a new comment command.
func NewComment(text string) *Comment {
	return &Comment{Text: text}
}

// ZPL returns the ZPL representation.
func (c *Comment) ZPL() string {
	return fmt.Sprintf("^FX %s", c.Text)
}

// WriteTo writes the ZPL to the writer.
func (c *Comment) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, c.ZPL())
	return int64(n), err
}
