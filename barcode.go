package zpl

import (
	"fmt"
	"io"
)

// BarcodeDefault represents a ^BY command for barcode field defaults.
type BarcodeDefault struct {
	ModuleWidth       int     // Width of narrow bar in dots (1-10)
	WideToNarrowRatio float64 // Ratio of wide to narrow bars (2.0-3.0)
	Height            int     // Barcode height in dots
}

// NewBarcodeDefault creates a new barcode default command.
func NewBarcodeDefault(moduleWidth int, ratio float64, height int) *BarcodeDefault {
	return &BarcodeDefault{
		ModuleWidth:       moduleWidth,
		WideToNarrowRatio: ratio,
		Height:            height,
	}
}

// ZPL returns the ZPL representation.
func (b *BarcodeDefault) ZPL() string {
	return fmt.Sprintf("^BY%d,%.1f,%d", b.ModuleWidth, b.WideToNarrowRatio, b.Height)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeDefault) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// Code128Mode represents the Code 128 subset mode.
type Code128Mode rune

// Code 128 modes.
const (
	Code128Auto    Code128Mode = 'N' // Auto mode (printer selects best subset)
	Code128UCC     Code128Mode = 'U' // UCC/EAN-128 mode
	Code128SubsetA Code128Mode = 'A' // Subset A
	Code128SubsetB Code128Mode = 'B' // Subset B (default for auto)
	Code128SubsetC Code128Mode = 'C' // Subset C (numeric pairs)
	Code128SubsetD Code128Mode = 'D' // Subset D (reader programming)
)

// BarcodeCode128 represents a ^BC command for Code 128 barcodes.
type BarcodeCode128 struct {
	Orientation         Orientation
	Height              int
	PrintInterpretation bool // Print human-readable text below
	InterpretationAbove bool // Print text above barcode
	CheckDigit          bool // UCC check digit
	Mode                Code128Mode
	Data                string
}

// NewBarcodeCode128 creates a new Code 128 barcode command.
func NewBarcodeCode128(data string, height int) *BarcodeCode128 {
	return &BarcodeCode128{
		Orientation:         OrientationNormal,
		Height:              height,
		PrintInterpretation: true,
		InterpretationAbove: false,
		CheckDigit:          false,
		Mode:                Code128Auto,
		Data:                data,
	}
}

// WithOrientation sets the barcode orientation.
func (b *BarcodeCode128) WithOrientation(o Orientation) *BarcodeCode128 {
	b.Orientation = o
	return b
}

// WithInterpretation sets whether to print human-readable text.
func (b *BarcodeCode128) WithInterpretation(printText, above bool) *BarcodeCode128 {
	b.PrintInterpretation = printText
	b.InterpretationAbove = above
	return b
}

// WithMode sets the Code 128 mode.
func (b *BarcodeCode128) WithMode(mode Code128Mode) *BarcodeCode128 {
	b.Mode = mode
	return b
}

// WithUCCCheckDigit enables UCC check digit.
func (b *BarcodeCode128) WithUCCCheckDigit(enabled bool) *BarcodeCode128 {
	b.CheckDigit = enabled
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeCode128) ZPL() string {
	printInterp := 'Y'
	if !b.PrintInterpretation {
		printInterp = 'N'
	}
	interpAbove := 'N'
	if b.InterpretationAbove {
		interpAbove = 'Y'
	}
	checkDigit := 'N'
	if b.CheckDigit {
		checkDigit = 'Y'
	}
	return fmt.Sprintf("^BC%c,%d,%c,%c,%c,%c^FD%s^FS",
		b.Orientation, b.Height, printInterp, interpAbove, checkDigit, b.Mode, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeCode128) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// BarcodeCode39 represents a ^B3 command for Code 39 barcodes.
type BarcodeCode39 struct {
	Orientation         Orientation
	CheckDigit          bool // Mod 43 check digit
	Height              int
	PrintInterpretation bool
	InterpretationAbove bool
	Data                string
}

// NewBarcodeCode39 creates a new Code 39 barcode command.
func NewBarcodeCode39(data string, height int) *BarcodeCode39 {
	return &BarcodeCode39{
		Orientation:         OrientationNormal,
		CheckDigit:          false,
		Height:              height,
		PrintInterpretation: true,
		InterpretationAbove: false,
		Data:                data,
	}
}

// WithOrientation sets the barcode orientation.
func (b *BarcodeCode39) WithOrientation(o Orientation) *BarcodeCode39 {
	b.Orientation = o
	return b
}

// WithCheckDigit enables Mod 43 check digit.
func (b *BarcodeCode39) WithCheckDigit(enabled bool) *BarcodeCode39 {
	b.CheckDigit = enabled
	return b
}

// WithInterpretation sets whether to print human-readable text.
func (b *BarcodeCode39) WithInterpretation(printText, above bool) *BarcodeCode39 {
	b.PrintInterpretation = printText
	b.InterpretationAbove = above
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeCode39) ZPL() string {
	checkDigit := 'N'
	if b.CheckDigit {
		checkDigit = 'Y'
	}
	printInterp := 'Y'
	if !b.PrintInterpretation {
		printInterp = 'N'
	}
	interpAbove := 'N'
	if b.InterpretationAbove {
		interpAbove = 'Y'
	}
	return fmt.Sprintf("^B3%c,%c,%d,%c,%c^FD%s^FS",
		b.Orientation, checkDigit, b.Height, printInterp, interpAbove, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeCode39) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// QRCodeModel represents the QR code model.
type QRCodeModel int

// QR code models.
const (
	QRCodeModel1 QRCodeModel = 1
	QRCodeModel2 QRCodeModel = 2 // Recommended
)

// QRCodeErrorCorrection represents the error correction level.
type QRCodeErrorCorrection rune

// QR code error correction levels.
const (
	QRCodeECHigh     QRCodeErrorCorrection = 'H' // ~30% recovery
	QRCodeECQuartile QRCodeErrorCorrection = 'Q' // ~25% recovery
	QRCodeECMedium   QRCodeErrorCorrection = 'M' // ~15% recovery
	QRCodeECLow      QRCodeErrorCorrection = 'L' // ~7% recovery
)

// BarcodeQR represents a ^BQ command for QR codes.
type BarcodeQR struct {
	Orientation         Orientation
	Model               QRCodeModel
	MagnificationFactor int // 1-10
	ErrorCorrection     QRCodeErrorCorrection
	MaskValue           int // 0-7, typically 7 (auto)
	Data                string
}

// NewBarcodeQR creates a new QR code command.
func NewBarcodeQR(data string, magnification int) *BarcodeQR {
	return &BarcodeQR{
		Orientation:         OrientationNormal,
		Model:               QRCodeModel2,
		MagnificationFactor: magnification,
		ErrorCorrection:     QRCodeECMedium,
		MaskValue:           7,
		Data:                data,
	}
}

// WithOrientation sets the QR code orientation.
func (b *BarcodeQR) WithOrientation(o Orientation) *BarcodeQR {
	b.Orientation = o
	return b
}

// WithModel sets the QR code model.
func (b *BarcodeQR) WithModel(model QRCodeModel) *BarcodeQR {
	b.Model = model
	return b
}

// WithErrorCorrection sets the error correction level.
func (b *BarcodeQR) WithErrorCorrection(ec QRCodeErrorCorrection) *BarcodeQR {
	b.ErrorCorrection = ec
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeQR) ZPL() string {
	// ^BQo,m,s - o=orientation, m=model, s=magnification
	// ^FDerror_correction,data
	return fmt.Sprintf("^BQ%c,%d,%d^FD%cA,%s^FS",
		b.Orientation, b.Model, b.MagnificationFactor, b.ErrorCorrection, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeQR) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// BarcodeDataMatrix represents a ^BX command for DataMatrix barcodes.
type BarcodeDataMatrix struct {
	Orientation  Orientation
	Height       int // Module size (height of individual cells)
	QualityLevel int // 0, 50, 80, 100, 140, 200
	Columns      int // Number of columns (even, 10-144)
	Rows         int // Number of rows (even, 10-144)
	FormatID     int // Format identifier
	EscapeChar   rune
	AspectRatio  int // 1 = square, 2 = rectangular
	Data         string
}

// NewBarcodeDataMatrix creates a new DataMatrix barcode command.
func NewBarcodeDataMatrix(data string, height int) *BarcodeDataMatrix {
	return &BarcodeDataMatrix{
		Orientation:  OrientationNormal,
		Height:       height,
		QualityLevel: 200,
		Columns:      0, // Auto
		Rows:         0, // Auto
		FormatID:     6, // ECC 200
		EscapeChar:   '~',
		AspectRatio:  1,
		Data:         data,
	}
}

// WithOrientation sets the DataMatrix orientation.
func (b *BarcodeDataMatrix) WithOrientation(o Orientation) *BarcodeDataMatrix {
	b.Orientation = o
	return b
}

// WithQualityLevel sets the quality level.
func (b *BarcodeDataMatrix) WithQualityLevel(level int) *BarcodeDataMatrix {
	b.QualityLevel = level
	return b
}

// WithSize sets the explicit size (columns and rows).
func (b *BarcodeDataMatrix) WithSize(columns, rows int) *BarcodeDataMatrix {
	b.Columns = columns
	b.Rows = rows
	return b
}

// WithRectangular sets the aspect ratio to rectangular.
func (b *BarcodeDataMatrix) WithRectangular() *BarcodeDataMatrix {
	b.AspectRatio = 2
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeDataMatrix) ZPL() string {
	return fmt.Sprintf("^BX%c,%d,%d,%d,%d,%d,%c,%d^FD%s^FS",
		b.Orientation, b.Height, b.QualityLevel, b.Columns, b.Rows,
		b.FormatID, b.EscapeChar, b.AspectRatio, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeDataMatrix) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// BarcodePDF417 represents a ^B7 command for PDF417 barcodes.
type BarcodePDF417 struct {
	Orientation   Orientation
	Height        int
	SecurityLevel int  // 0-8
	DataColumns   int  // 1-30
	Rows          int  // 3-90
	Truncate      bool // Truncated PDF417
	Data          string
}

// NewBarcodePDF417 creates a new PDF417 barcode command.
func NewBarcodePDF417(data string, height int) *BarcodePDF417 {
	return &BarcodePDF417{
		Orientation:   OrientationNormal,
		Height:        height,
		SecurityLevel: 0,
		DataColumns:   0, // Auto
		Rows:          0, // Auto
		Truncate:      false,
		Data:          data,
	}
}

// WithOrientation sets the PDF417 orientation.
func (b *BarcodePDF417) WithOrientation(o Orientation) *BarcodePDF417 {
	b.Orientation = o
	return b
}

// WithSecurityLevel sets the error correction level.
func (b *BarcodePDF417) WithSecurityLevel(level int) *BarcodePDF417 {
	b.SecurityLevel = level
	return b
}

// WithColumns sets the number of data columns.
func (b *BarcodePDF417) WithColumns(columns int) *BarcodePDF417 {
	b.DataColumns = columns
	return b
}

// WithRows sets the number of rows.
func (b *BarcodePDF417) WithRows(rows int) *BarcodePDF417 {
	b.Rows = rows
	return b
}

// WithTruncation enables truncated PDF417.
func (b *BarcodePDF417) WithTruncation(truncate bool) *BarcodePDF417 {
	b.Truncate = truncate
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodePDF417) ZPL() string {
	truncate := 'N'
	if b.Truncate {
		truncate = 'Y'
	}
	return fmt.Sprintf("^B7%c,%d,%d,%d,%d,%c^FD%s^FS",
		b.Orientation, b.Height, b.SecurityLevel, b.DataColumns, b.Rows, truncate, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodePDF417) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// BarcodeEAN13 represents a ^BE command for EAN-13 barcodes.
type BarcodeEAN13 struct {
	Orientation         Orientation
	Height              int
	PrintInterpretation bool
	InterpretationAbove bool
	Data                string
}

// NewBarcodeEAN13 creates a new EAN-13 barcode command.
func NewBarcodeEAN13(data string, height int) *BarcodeEAN13 {
	return &BarcodeEAN13{
		Orientation:         OrientationNormal,
		Height:              height,
		PrintInterpretation: true,
		InterpretationAbove: false,
		Data:                data,
	}
}

// WithOrientation sets the barcode orientation.
func (b *BarcodeEAN13) WithOrientation(o Orientation) *BarcodeEAN13 {
	b.Orientation = o
	return b
}

// WithInterpretation sets whether to print human-readable text.
func (b *BarcodeEAN13) WithInterpretation(printText, above bool) *BarcodeEAN13 {
	b.PrintInterpretation = printText
	b.InterpretationAbove = above
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeEAN13) ZPL() string {
	printInterp := 'Y'
	if !b.PrintInterpretation {
		printInterp = 'N'
	}
	interpAbove := 'N'
	if b.InterpretationAbove {
		interpAbove = 'Y'
	}
	return fmt.Sprintf("^BE%c,%d,%c,%c^FD%s^FS",
		b.Orientation, b.Height, printInterp, interpAbove, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeEAN13) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// BarcodeUPCA represents a ^BU command for UPC-A barcodes.
type BarcodeUPCA struct {
	Orientation         Orientation
	Height              int
	PrintInterpretation bool
	InterpretationAbove bool
	CheckDigit          bool
	Data                string
}

// NewBarcodeUPCA creates a new UPC-A barcode command.
func NewBarcodeUPCA(data string, height int) *BarcodeUPCA {
	return &BarcodeUPCA{
		Orientation:         OrientationNormal,
		Height:              height,
		PrintInterpretation: true,
		InterpretationAbove: false,
		CheckDigit:          true,
		Data:                data,
	}
}

// WithOrientation sets the barcode orientation.
func (b *BarcodeUPCA) WithOrientation(o Orientation) *BarcodeUPCA {
	b.Orientation = o
	return b
}

// WithInterpretation sets whether to print human-readable text.
func (b *BarcodeUPCA) WithInterpretation(printText, above bool) *BarcodeUPCA {
	b.PrintInterpretation = printText
	b.InterpretationAbove = above
	return b
}

// WithCheckDigit enables check digit printing.
func (b *BarcodeUPCA) WithCheckDigit(enabled bool) *BarcodeUPCA {
	b.CheckDigit = enabled
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeUPCA) ZPL() string {
	printInterp := 'Y'
	if !b.PrintInterpretation {
		printInterp = 'N'
	}
	interpAbove := 'N'
	if b.InterpretationAbove {
		interpAbove = 'Y'
	}
	checkDigit := 'Y'
	if !b.CheckDigit {
		checkDigit = 'N'
	}
	return fmt.Sprintf("^BU%c,%d,%c,%c,%c^FD%s^FS",
		b.Orientation, b.Height, printInterp, interpAbove, checkDigit, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeUPCA) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// BarcodeInterleaved2of5 represents a ^B2 command for Interleaved 2 of 5 barcodes.
type BarcodeInterleaved2of5 struct {
	Orientation         Orientation
	Height              int
	PrintInterpretation bool
	InterpretationAbove bool
	CheckDigit          bool
	Data                string
}

// NewBarcodeInterleaved2of5 creates a new Interleaved 2 of 5 barcode command.
func NewBarcodeInterleaved2of5(data string, height int) *BarcodeInterleaved2of5 {
	return &BarcodeInterleaved2of5{
		Orientation:         OrientationNormal,
		Height:              height,
		PrintInterpretation: true,
		InterpretationAbove: false,
		CheckDigit:          false,
		Data:                data,
	}
}

// WithOrientation sets the barcode orientation.
func (b *BarcodeInterleaved2of5) WithOrientation(o Orientation) *BarcodeInterleaved2of5 {
	b.Orientation = o
	return b
}

// WithInterpretation sets whether to print human-readable text.
func (b *BarcodeInterleaved2of5) WithInterpretation(printText, above bool) *BarcodeInterleaved2of5 {
	b.PrintInterpretation = printText
	b.InterpretationAbove = above
	return b
}

// WithCheckDigit enables Mod 10 check digit.
func (b *BarcodeInterleaved2of5) WithCheckDigit(enabled bool) *BarcodeInterleaved2of5 {
	b.CheckDigit = enabled
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeInterleaved2of5) ZPL() string {
	checkDigit := 'N'
	if b.CheckDigit {
		checkDigit = 'Y'
	}
	printInterp := 'Y'
	if !b.PrintInterpretation {
		printInterp = 'N'
	}
	interpAbove := 'N'
	if b.InterpretationAbove {
		interpAbove = 'Y'
	}
	return fmt.Sprintf("^B2%c,%d,%c,%c,%c^FD%s^FS",
		b.Orientation, b.Height, printInterp, interpAbove, checkDigit, b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeInterleaved2of5) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}

// MaxiCodeMode represents the MaxiCode encoding mode.
type MaxiCodeMode int

// MaxiCode modes.
const (
	MaxiCodeMode2 MaxiCodeMode = 2 // Structured Carrier Message (US)
	MaxiCodeMode3 MaxiCodeMode = 3 // Structured Carrier Message (International)
	MaxiCodeMode4 MaxiCodeMode = 4 // Standard Symbol (full ECC)
	MaxiCodeMode5 MaxiCodeMode = 5 // Full ECC (secure data)
	MaxiCodeMode6 MaxiCodeMode = 6 // Reader Programming
)

// BarcodeMaxiCode represents a ^BD command for MaxiCode 2D barcodes.
// MaxiCode is a fixed-size 2D barcode primarily used by UPS for package tracking.
type BarcodeMaxiCode struct {
	Mode         MaxiCodeMode
	SymbolNumber int // Symbol number for structured append (1-8)
	SymbolCount  int // Total symbols in structured append (1-8)
	Data         string
	HexIndicator rune // Character used to indicate hex values in data (e.g., '_')
}

// NewBarcodeMaxiCode creates a new MaxiCode barcode command.
func NewBarcodeMaxiCode(data string, mode MaxiCodeMode) *BarcodeMaxiCode {
	return &BarcodeMaxiCode{
		Mode:         mode,
		SymbolNumber: 1,
		SymbolCount:  1,
		Data:         data,
	}
}

// WithStructuredAppend sets the structured append parameters.
func (b *BarcodeMaxiCode) WithStructuredAppend(symbolNumber, symbolCount int) *BarcodeMaxiCode {
	b.SymbolNumber = symbolNumber
	b.SymbolCount = symbolCount
	return b
}

// WithHexIndicator sets the hex escape character used in the data.
func (b *BarcodeMaxiCode) WithHexIndicator(indicator rune) *BarcodeMaxiCode {
	b.HexIndicator = indicator
	return b
}

// ZPL returns the ZPL representation.
func (b *BarcodeMaxiCode) ZPL() string {
	zpl := fmt.Sprintf("^BD%d,%d,%d", b.Mode, b.SymbolNumber, b.SymbolCount)
	if b.HexIndicator != 0 {
		zpl = fmt.Sprintf("^FH%c%s", b.HexIndicator, zpl)
	}
	return zpl + fmt.Sprintf("^FD%s^FS", b.Data)
}

// WriteTo writes the ZPL to the writer.
func (b *BarcodeMaxiCode) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, b.ZPL())
	return int64(n), err
}
