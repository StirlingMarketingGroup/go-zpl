package zpl

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"io"
	"strconv"
	"strings"
)

// Parse parses a ZPL string and returns the first Label.
// For ZPL with multiple labels (multiple ^XA...^XZ blocks), use ParseAll.
func Parse(zpl string) (*Label, error) {
	labels, err := ParseAll(zpl)
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		return NewLabel(), nil
	}
	return labels[0], nil
}

// ParseAll parses a ZPL string and returns all labels (one per ^XA...^XZ block).
func ParseAll(zpl string) ([]*Label, error) {
	p := &parser{
		input:  zpl,
		pos:    0,
		labels: make([]*Label, 0),
	}
	return p.parseAll()
}

// parser holds the state for parsing ZPL.
type parser struct {
	input  string
	pos    int
	label  *Label   // Current label being parsed
	labels []*Label // All completed labels

	// Current state
	inFormat          bool // Inside ^XA...^XZ block
	fieldHexIndicator rune // ^FH indicator for next ^FD/^FV
}

func (p *parser) parseAll() ([]*Label, error) {
	for p.pos < len(p.input) {
		// Skip whitespace and newlines
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		ch := p.input[p.pos]

		switch ch {
		case '^':
			if err := p.parseCaretCommand(); err != nil {
				return nil, err
			}
		case '~':
			if err := p.parseTildeCommand(); err != nil {
				return nil, err
			}
		default:
			// Skip unknown characters
			p.pos++
		}
	}

	// If we have an unfinished label (no closing ^XZ) with drawable content, add it
	if p.label != nil && p.label.HasDrawableContent() {
		p.labels = append(p.labels, p.label)
	}

	return p.labels, nil
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			p.pos++
		} else {
			break
		}
	}
}

func (p *parser) parseCaretCommand() error {
	p.pos++ // Skip ^
	if p.pos >= len(p.input) {
		return nil
	}

	// Read command name (1-2 characters)
	cmd := p.readCommandName()
	if cmd == "" {
		return nil
	}

	switch cmd {
	case "XA":
		p.inFormat = true
		p.label = NewLabel()    // Start a new label
		p.fieldHexIndicator = 0 // Reset state for new label
		return nil
	case "XZ":
		p.inFormat = false
		if p.label != nil {
			// Only keep labels that produce ink; setup-only blocks (^XA^MCY^XZ) are skipped
			if p.label.HasDrawableContent() {
				p.labels = append(p.labels, p.label)
			}
			p.label = nil
		}
		return nil
	}

	// Skip commands outside ^XA...^XZ block
	if p.label == nil {
		p.skipToNextCommand()
		return nil
	}

	switch cmd {
	case "FO":
		return p.parseFieldOrigin()
	case "FT":
		return p.parseFieldTypeset()
	case "FD":
		return p.parseFieldData()
	case "FV":
		return p.parseFieldData() // FV is like FD but for variable data
	case "FS":
		// Field separator - nothing to do
	case "FR":
		p.label.Add(NewFieldReverse())
	case "FW":
		return p.parseFieldDirection()
	case "FH":
		return p.parseFieldHexIndicator()
	case "A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8", "A9",
		"AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI", "AJ",
		"AK", "AL", "AM", "AN", "AO", "AP", "AQ", "AR", "AS", "AT",
		"AU", "AV", "AW", "AX", "AY", "AZ":
		return p.parseScalableFont(cmd)
	case "CF":
		return p.parseChangeFont()
	case "FB":
		return p.parseFieldBlock()
	case "BY":
		return p.parseBarcodeDefault()
	case "BC":
		return p.parseBarcodeCode128()
	case "GB":
		return p.parseGraphicBox()
	case "GC":
		return p.parseGraphicCircle()
	case "GD":
		return p.parseGraphicDiagonalLine()
	case "GE":
		return p.parseGraphicEllipse()
	case "GF":
		return p.parseGraphicField()
	case "PO":
		return p.parsePrintOrientation()
	case "PW":
		return p.parsePrintWidth()
	case "LL":
		return p.parseLabelLength()
	case "LH":
		return p.parseLabelHome()
	case "CI":
		return p.parseCharacterSet()
	case "FX":
		return p.parseComment()
	case "BD":
		return p.parseMaxiCode()
	case "BQ":
		return p.parseBarcodeQR()
	case "BX":
		return p.parseBarcodeDataMatrix()
	case "B7":
		return p.parseBarcodePDF417()
	case "BO":
		return p.parseBarcodeAztec()
	case "PQ":
		return p.parsePrintQuantity()
	// Commands we recognize but ignore for now
	case "LR", "MN", "MF", "MC", "CV", "DN", "B3":
		p.skipToNextCommand()
	default:
		// Unknown command - skip to next command
		p.skipToNextCommand()
	}

	return nil
}

func (p *parser) parseTildeCommand() error { //nolint:unparam // Error return reserved for future commands
	p.pos++ // Skip ~
	if p.pos >= len(p.input) {
		return nil
	}

	// Read command name
	cmd := p.readCommandName()
	if cmd == "" {
		return nil
	}

	// Most tilde commands are printer control - skip them
	p.skipToNextCommand()
	return nil
}

func (p *parser) readCommandName() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			p.pos++
			if p.pos-start >= 2 {
				break // Command names are max 2 chars
			}
		} else {
			break
		}
	}
	return strings.ToUpper(p.input[start:p.pos])
}

func (p *parser) readParams() []string {
	var params []string
	var current strings.Builder

	for p.pos < len(p.input) {
		ch := p.input[p.pos]

		if ch == '^' || ch == '~' {
			// Start of next command
			break
		}

		if ch == ',' {
			params = append(params, current.String())
			current.Reset()
			p.pos++
			continue
		}

		if ch == '\r' || ch == '\n' {
			p.pos++
			continue
		}

		current.WriteByte(ch)
		p.pos++
	}

	// Add final parameter if non-empty
	if current.Len() > 0 {
		params = append(params, current.String())
	}

	return params
}

// readIntParam reads a numeric parameter from the current position,
// stopping at comma, ^, ~, or end of input. Returns 0 if empty or invalid.
func (p *parser) readIntParam() int {
	var sb strings.Builder
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ',' || ch == '^' || ch == '~' || ch == '\r' || ch == '\n' {
			break
		}
		sb.WriteByte(ch)
		p.pos++
	}
	return parseInt(sb.String(), 0)
}

func (p *parser) skipToNextCommand() {
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '^' || ch == '~' {
			break
		}
		p.pos++
	}
}

func parseInt(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func parseFloat(s string, def float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

// Command parsers

func (p *parser) parseFieldOrigin() error {
	params := p.readParams()
	x := parseInt(getParam(params, 0), 0)
	y := parseInt(getParam(params, 1), 0)
	p.label.Add(NewFieldOrigin(x, y))
	return nil
}

func (p *parser) parseFieldTypeset() error {
	params := p.readParams()
	x := parseInt(getParam(params, 0), 0)
	y := parseInt(getParam(params, 1), 0)
	p.label.Add(NewFieldTypeset(x, y))
	return nil
}

func (p *parser) parseFieldDirection() error {
	params := p.readParams()
	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	p.label.Add(NewFieldDirection(orient))
	return nil
}

func (p *parser) parseFieldHexIndicator() error {
	// ^FH command: optional single-character indicator, default '_'
	if p.pos >= len(p.input) || p.input[p.pos] == '^' || p.input[p.pos] == '~' {
		p.fieldHexIndicator = '_'
		return nil
	}
	p.fieldHexIndicator = rune(p.input[p.pos])
	p.pos++
	return nil
}

func (p *parser) parseFieldData() error {
	// Read until ^FS
	var sb strings.Builder
	for p.pos < len(p.input) {
		if strings.HasPrefix(p.input[p.pos:], "^FS") {
			p.pos += 3
			break
		}
		if p.input[p.pos] == '^' {
			// Another command - stop here
			break
		}
		sb.WriteByte(p.input[p.pos])
		p.pos++
	}
	data := sb.String()
	if data != "" {
		if p.fieldHexIndicator != 0 {
			data = decodeHexEscapes(data, p.fieldHexIndicator)
		}
		p.label.Add(NewFieldData(data))
	}
	// ^FH applies to the current field only; reset after ^FD/^FV
	p.fieldHexIndicator = 0
	return nil
}

func (p *parser) parsePrintQuantity() error {
	params := p.readParams()
	quantity := parseInt(getParam(params, 0), 1)
	if quantity < 1 {
		quantity = 1
	}
	p.label.SetPrintQuantity(quantity, 0, 0, false)
	return nil
}

func (p *parser) parseScalableFont(cmd string) error {
	// Get font identifier from command (A0 = Font0, AD = FontD, etc)
	fontChar := '0'
	if len(cmd) > 1 {
		fontChar = rune(cmd[1])
	}

	params := p.readParams()
	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}

	// Default sizes depend on font type
	// Bitmap fonts (A-Z) have specific default sizes
	defaultHeight := 30
	defaultWidth := 0
	if fontChar >= 'A' && fontChar <= 'Z' {
		// Bitmap font defaults (approximate Zebra standard sizes)
		switch fontChar {
		case 'A':
			defaultHeight, defaultWidth = 9, 5
		case 'B':
			defaultHeight, defaultWidth = 11, 7
		case 'C', 'D':
			// Spec says 18x10, but real Zebra output is wider (~18x20)
			// We also apply synthetic horizontal bold to thicken strokes
			defaultHeight, defaultWidth = 18, 20
		case 'E':
			defaultHeight, defaultWidth = 28, 15
		case 'F':
			defaultHeight, defaultWidth = 26, 13
		case 'G':
			defaultHeight, defaultWidth = 60, 40
		case 'H':
			defaultHeight, defaultWidth = 21, 13
		default:
			defaultHeight, defaultWidth = 18, 10
		}
	}

	height := parseInt(getParam(params, 1), defaultHeight)
	width := parseInt(getParam(params, 2), defaultWidth)

	// If height or width is 0, use defaults
	if height == 0 {
		height = defaultHeight
	}
	if width == 0 {
		width = defaultWidth
	}

	font := Font(fontChar)
	p.label.Add(NewScalableFont(font, height, width).WithOrientation(orient))
	return nil
}

func (p *parser) parseChangeFont() error {
	params := p.readParams()
	fontChar := '0'
	if getParam(params, 0) != "" {
		fontChar = rune(getParam(params, 0)[0])
	}
	height := parseInt(getParam(params, 1), 30)
	width := parseInt(getParam(params, 2), 0)

	p.label.Add(NewChangeFont(Font(fontChar), height, width))
	return nil
}

func (p *parser) parseFieldBlock() error {
	params := p.readParams()
	width := parseInt(getParam(params, 0), 0)
	maxLines := parseInt(getParam(params, 1), 1)
	lineSpacing := parseInt(getParam(params, 2), 0)
	justification := JustifyLeft
	if getParam(params, 3) != "" {
		justification = Justification(getParam(params, 3)[0])
	}
	hangingIndent := parseInt(getParam(params, 4), 0)

	fb := NewFieldBlock(width).
		WithMaxLines(maxLines).
		WithLineSpacing(lineSpacing).
		WithJustification(justification).
		WithHangingIndent(hangingIndent)
	p.label.Add(fb)
	return nil
}

func (p *parser) parseBarcodeDefault() error {
	params := p.readParams()
	moduleWidth := parseInt(getParam(params, 0), 2)
	ratio := parseFloat(getParam(params, 1), 3.0)
	height := parseInt(getParam(params, 2), 10)

	p.label.Add(NewBarcodeDefault(moduleWidth, ratio, height))
	return nil
}

// readBarcodeFieldData consumes an optional ^FH and a ^FD/^FV following a barcode
// command, returning the (hex-decoded) data.
func (p *parser) readBarcodeFieldData() (data string, hexIndicator rune, ok bool) {
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != '^' {
			break
		}

		save := p.pos
		p.pos++
		cmd := p.readCommandName()

		switch cmd {
		case "FH":
			// Field Hexadecimal - read the indicator character
			if p.pos < len(p.input) && p.input[p.pos] != '^' {
				hexIndicator = rune(p.input[p.pos])
				p.pos++
			} else {
				hexIndicator = '_' // Default
			}
			continue
		case "FD", "FV":
			var rawData strings.Builder
			for p.pos < len(p.input) {
				if strings.HasPrefix(p.input[p.pos:], "^FS") {
					p.pos += 3
					break
				}
				if p.input[p.pos] == '^' {
					break
				}
				rawData.WriteByte(p.input[p.pos])
				p.pos++
			}
			if hexIndicator != 0 {
				data = decodeHexEscapes(rawData.String(), hexIndicator)
			} else {
				data = rawData.String()
			}
			return data, hexIndicator, true
		default:
			// Not a field command, restore position
			p.pos = save
			return "", hexIndicator, false
		}
	}
	return "", hexIndicator, false
}

func (p *parser) parseBarcodeCode128() error {
	params := p.readParams()
	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	height := parseInt(getParam(params, 1), 0) // 0 means use ^BY height
	printInterp := getParam(params, 2) != "N"
	interpAbove := getParam(params, 3) == "Y"
	checkDigit := getParam(params, 4) == "Y"
	mode := Code128Auto
	if getParam(params, 5) != "" {
		mode = Code128Mode(getParam(params, 5)[0])
	}

	// The data comes from the next ^FH? ^FD/^FV
	bc := NewBarcodeCode128("", height).
		WithOrientation(orient).
		WithInterpretation(printInterp, interpAbove).
		WithUCCCheckDigit(checkDigit).
		WithMode(mode)

	if data, _, ok := p.readBarcodeFieldData(); ok {
		bc.Data = data
	}

	p.label.Add(bc)
	return nil
}

func (p *parser) parseBarcodeQR() error {
	// ^BQ command for QR Code
	// Format: ^BQa,b,c where a=orientation, b=model (1-2), c=magnification (1-10)
	// Data format: ^FD[error_correction][mode],[data]^FS
	// e.g., ^FDMA,HelloWorld^FS means Medium error correction, Auto mode, data "HelloWorld"
	params := p.readParams()
	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	model := QRCodeModel(parseInt(getParam(params, 1), 2))
	magnification := parseInt(getParam(params, 2), 3)

	// Default error correction
	errorCorrection := QRCodeECMedium

	var data string

	if fieldData, _, ok := p.readBarcodeFieldData(); ok {
		// Parse the QR data format: [error_correction][mode],[data]
		// e.g., "MA,HelloWorld" or "HA,https://example.com"
		if len(fieldData) >= 3 && fieldData[2] == ',' {
			// First char is error correction level
			switch fieldData[0] {
			case 'H':
				errorCorrection = QRCodeECHigh
			case 'Q':
				errorCorrection = QRCodeECQuartile
			case 'M':
				errorCorrection = QRCodeECMedium
			case 'L':
				errorCorrection = QRCodeECLow
			}
			// Second char is mode (A=auto, M=manual) - we ignore this
			// Data starts after the comma
			data = fieldData[3:]
		} else {
			// No prefix, use raw data
			data = fieldData
		}
	}

	bc := NewBarcodeQR(data, magnification).
		WithOrientation(orient).
		WithModel(model).
		WithErrorCorrection(errorCorrection)

	p.label.Add(bc)
	return nil
}

func (p *parser) parseBarcodeDataMatrix() error {
	// ^BX command for DataMatrix barcode
	// Format: ^BXa,b,c,d,e,f,g,h where:
	// a=orientation, b=height (module size), c=quality level,
	// d=columns, e=rows, f=format ID, g=escape char, h=aspect ratio
	params := p.readParams()

	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	height := parseInt(getParam(params, 1), 3)         // Default module size 3
	qualityLevel := parseInt(getParam(params, 2), 200) // Default quality 200
	columns := parseInt(getParam(params, 3), 0)        // Auto
	rows := parseInt(getParam(params, 4), 0)           // Auto
	formatID := parseInt(getParam(params, 5), 6)       // ECC 200
	escapeChar := '~'
	if getParam(params, 6) != "" {
		escapeChar = rune(getParam(params, 6)[0])
	}
	aspectRatio := parseInt(getParam(params, 7), 1) // Square

	data, _, _ := p.readBarcodeFieldData()

	bc := NewBarcodeDataMatrix(data, height).
		WithOrientation(orient).
		WithQualityLevel(qualityLevel)
	if columns > 0 && rows > 0 {
		bc.WithSize(columns, rows)
	}
	bc.FormatID = formatID
	bc.EscapeChar = escapeChar
	bc.AspectRatio = aspectRatio

	p.label.Add(bc)
	return nil
}

func (p *parser) parseBarcodePDF417() error {
	// ^B7 command for PDF417 barcode
	// Format: ^B7o,h,s,c,r,t where:
	// o=orientation, h=height, s=security level (0-8),
	// c=data columns (1-30), r=rows (3-90), t=truncate (Y/N)
	params := p.readParams()

	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	height := parseInt(getParam(params, 1), 50)       // Default height
	securityLevel := parseInt(getParam(params, 2), 0) // Default security 0
	dataColumns := parseInt(getParam(params, 3), 0)   // Auto
	rows := parseInt(getParam(params, 4), 0)          // Auto
	truncate := getParam(params, 5) == "Y"

	data, _, _ := p.readBarcodeFieldData()

	bc := NewBarcodePDF417(data, height).
		WithOrientation(orient).
		WithSecurityLevel(securityLevel).
		WithTruncation(truncate)
	if dataColumns > 0 {
		bc.WithColumns(dataColumns)
	}
	if rows > 0 {
		bc.WithRows(rows)
	}

	p.label.Add(bc)
	return nil
}

func (p *parser) parseBarcodeAztec() error {
	// ^BO command for Aztec barcode
	// Format: ^BOa,b,c,d,e,f,g where:
	// a=orientation, b=magnification (1-10), c=extended channel (Y/N),
	// d=error correction/size, e=menu symbol (Y/N), f=symbols for append, g=ID
	params := p.readParams()

	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	magnification := parseInt(getParam(params, 1), 2)   // Default magnification 2
	ecic := getParam(params, 2) == "Y"                  // Extended channel interpretation
	errorCorrection := parseInt(getParam(params, 3), 0) // Default auto
	menuSymbol := getParam(params, 4) == "Y"            // Menu symbol
	symbolCount := parseInt(getParam(params, 5), 1)     // Structured append count
	symbolID := getParam(params, 6)                     // Structured append ID

	data, _, _ := p.readBarcodeFieldData()

	bc := NewBarcodeAztec(data, magnification).
		WithOrientation(orient).
		WithErrorCorrection(errorCorrection)
	bc.ECICEnabled = ecic
	bc.MenuSymbol = menuSymbol
	bc.SymbolCount = symbolCount
	bc.SymbolID = symbolID

	p.label.Add(bc)
	return nil
}

func (p *parser) parseMaxiCode() error {
	// ^BD command for MaxiCode 2D barcode
	// Format: ^BDm,n,t where m=mode (2-6), n=symbol number, t=total symbols
	params := p.readParams()
	mode := MaxiCodeMode(parseInt(getParam(params, 0), 2))
	symbolNumber := parseInt(getParam(params, 1), 1)
	symbolCount := parseInt(getParam(params, 2), 1)

	data, hexIndicator, ok := p.readBarcodeFieldData()
	if !ok {
		// Preserve prior behavior: only add MaxiCode when field data is present
		return nil
	}

	mc := NewBarcodeMaxiCode(data, mode).
		WithStructuredAppend(symbolNumber, symbolCount)
	if hexIndicator != 0 {
		mc.WithHexIndicator(hexIndicator)
	}
	p.label.Add(mc)
	return nil
}

// decodeHexEscapes converts hex escape sequences to their byte values.
// For example, with indicator '_': "_1E" becomes ASCII 30 (RS).
func decodeHexEscapes(data string, indicator rune) string {
	var result strings.Builder
	indicatorStr := string(indicator)
	i := 0
	for i < len(data) {
		// Check if we have room for indicator + 2 hex chars (i+3 <= len)
		if i+3 <= len(data) && string(data[i]) == indicatorStr {
			// Try to parse next two characters as hex
			hexStr := data[i+1 : i+3]
			// Use ParseUint to handle full byte range 0x00-0xFF (ParseInt fails for > 0x7F)
			if val, err := strconv.ParseUint(hexStr, 16, 8); err == nil {
				result.WriteByte(byte(val))
				i += 3
				continue
			}
		}
		result.WriteByte(data[i])
		i++
	}
	return result.String()
}

func (p *parser) parseGraphicBox() error {
	params := p.readParams()
	width := parseInt(getParam(params, 0), 1)
	height := parseInt(getParam(params, 1), 1)
	thickness := parseInt(getParam(params, 2), 1)
	lineColor := LineColorBlack
	if getParam(params, 3) != "" {
		lineColor = LineColor(getParam(params, 3)[0])
	}
	cornerRadius := parseInt(getParam(params, 4), 0)

	gb := NewGraphicBox(width, height, thickness).
		WithColor(lineColor).
		WithCornerRadius(cornerRadius)
	p.label.Add(gb)
	return nil
}

func (p *parser) parseGraphicCircle() error {
	params := p.readParams()
	diameter := parseInt(getParam(params, 0), 1)
	thickness := parseInt(getParam(params, 1), 1)
	lineColor := LineColorBlack
	if getParam(params, 2) != "" {
		lineColor = LineColor(getParam(params, 2)[0])
	}

	gc := NewGraphicCircle(diameter, thickness).WithColor(lineColor)
	p.label.Add(gc)
	return nil
}

func (p *parser) parseGraphicDiagonalLine() error {
	params := p.readParams()
	width := parseInt(getParam(params, 0), 1)
	height := parseInt(getParam(params, 1), 1)
	thickness := parseInt(getParam(params, 2), 1)
	lineColor := LineColorBlack
	if getParam(params, 3) != "" {
		lineColor = LineColor(getParam(params, 3)[0])
	}
	orient := DiagonalRightLeaning
	if getParam(params, 4) == "L" {
		orient = DiagonalLeftLeaning
	}

	gd := NewGraphicDiagonalLine(width, height, thickness).WithColor(lineColor)
	if orient == DiagonalLeftLeaning {
		gd.WithLeftLeaning()
	}
	p.label.Add(gd)
	return nil
}

func (p *parser) parseGraphicEllipse() error {
	params := p.readParams()
	width := parseInt(getParam(params, 0), 1)
	height := parseInt(getParam(params, 1), 1)
	thickness := parseInt(getParam(params, 2), 1)
	lineColor := LineColorBlack
	if getParam(params, 3) != "" {
		lineColor = LineColor(getParam(params, 3)[0])
	}

	ge := NewGraphicEllipse(width, height, thickness).WithColor(lineColor)
	p.label.Add(ge)
	return nil
}

func (p *parser) parseGraphicField() error {
	// For binary format, we need to handle parameters specially because binary data
	// can contain ^ characters. Read the first 4 parameters manually, then handle
	// the data based on format.

	// Read format character (A, B, or C)
	format := GraphicFieldASCII
	if p.pos < len(p.input) && p.input[p.pos] != ',' {
		format = GraphicFieldFormat(p.input[p.pos])
		p.pos++
	}

	// Skip comma after format
	if p.pos < len(p.input) && p.input[p.pos] == ',' {
		p.pos++
	}

	// Read dataBytes (param 1)
	dataBytes := p.readIntParam()

	// Skip comma
	if p.pos < len(p.input) && p.input[p.pos] == ',' {
		p.pos++
	}

	// Read totalBytes (param 2)
	totalBytes := p.readIntParam()

	// Skip comma
	if p.pos < len(p.input) && p.input[p.pos] == ',' {
		p.pos++
	}

	// Read bytesPerRow (param 3)
	bytesPerRow := p.readIntParam()
	if bytesPerRow == 0 {
		bytesPerRow = 1
	}

	// Skip comma before data
	if p.pos < len(p.input) && p.input[p.pos] == ',' {
		p.pos++
	}

	switch format {
	case GraphicFieldBinary:
		// Binary format: read exactly dataBytes of raw binary data
		if dataBytes <= 0 {
			return nil
		}

		// Read the binary data directly - do NOT stop at ^ or any other character
		binaryData := make([]byte, 0, dataBytes)
		for p.pos < len(p.input) && len(binaryData) < dataBytes {
			binaryData = append(binaryData, p.input[p.pos])
			p.pos++
		}

		// Skip ^FS if present
		if p.pos < len(p.input) && strings.HasPrefix(p.input[p.pos:], "^FS") {
			p.pos += 3
		}

		if len(binaryData) > 0 {
			gf := NewGraphicFieldBinary(bytesPerRow, binaryData)
			if totalBytes > 0 {
				gf.TotalBytes = totalBytes
			}
			p.label.Add(gf)
		}

	case GraphicFieldASCII:
		// ASCII format: read ALL characters until ^FS or next command (preserve
		// compression count letters, ',', '!', ':', and Z64/B64 envelopes).
		var raw strings.Builder
		for p.pos < len(p.input) {
			if strings.HasPrefix(p.input[p.pos:], "^FS") {
				p.pos += 3
				break
			}
			if p.input[p.pos] == '^' {
				break
			}
			raw.WriteByte(p.input[p.pos])
			p.pos++
		}

		rawData := raw.String()
		if rawData == "" {
			return nil
		}

		// Handle :Z64: / :B64: envelopes (base64 [+ zlib] → expand to hex)
		if hex, ok := decompressZ64OrB64(rawData); ok {
			gf := NewGraphicFieldASCII(bytesPerRow, hex)
			if totalBytes > 0 {
				gf.TotalBytes = totalBytes
				gf.DataBytes = dataBytes
			}
			p.label.Add(gf)
			return nil
		}

		// Decompress Zebra alternative ASCII compression to plain hex
		hex := decompressGraphicFieldASCII(rawData, bytesPerRow)
		if hex == "" {
			return nil
		}
		gf := NewGraphicFieldASCII(bytesPerRow, hex)
		if totalBytes > 0 {
			gf.TotalBytes = totalBytes
			gf.DataBytes = dataBytes
		}
		p.label.Add(gf)
	}

	return nil
}

// decompressZ64OrB64 expands a :Z64: or :B64: graphic-field envelope to plain hex.
// Format: :Z64:<base64>:<crc> (zlib-inflated) or :B64:<base64>:<crc> (raw).
// CRC is ignored. Returns ok=false if the data is not such an envelope.
func decompressZ64OrB64(data string) (hex string, ok bool) {
	trimmed := strings.TrimSpace(data)
	var compressed bool
	var rest string
	switch {
	case strings.HasPrefix(trimmed, ":Z64:"):
		compressed = true
		rest = trimmed[5:]
	case strings.HasPrefix(trimmed, ":B64:"):
		compressed = false
		rest = trimmed[5:]
	default:
		return "", false
	}

	// rest is <base64>:<crc> — CRC may be absent
	b64Part := rest
	if idx := strings.LastIndex(rest, ":"); idx >= 0 {
		b64Part = rest[:idx]
	}

	decoded, err := base64.StdEncoding.DecodeString(b64Part)
	if err != nil {
		// Try raw encoding without padding issues
		decoded, err = base64.RawStdEncoding.DecodeString(b64Part)
		if err != nil {
			return "", false
		}
	}

	var binary []byte
	if compressed {
		zr, err := zlib.NewReader(bytes.NewReader(decoded))
		if err != nil {
			return "", false
		}
		binary, err = io.ReadAll(zr)
		_ = zr.Close()
		if err != nil {
			return "", false
		}
	} else {
		binary = decoded
	}

	const hexDigits = "0123456789ABCDEF"
	var sb strings.Builder
	sb.Grow(len(binary) * 2)
	for _, b := range binary {
		sb.WriteByte(hexDigits[b>>4])
		sb.WriteByte(hexDigits[b&0x0F])
	}
	return sb.String(), true
}

// decompressGraphicFieldASCII expands Zebra's alternative data-compression scheme
// used in ^GFA fields into plain hex nibbles. Plain hex input (no compression chars)
// is returned byte-identical (aside from ignored whitespace).
//
// Rules (row length = bytesPerRow*2 nibbles):
//   - whitespace: ignore
//   - hex digit 0-9A-Fa-f: emit it count times (count defaults to 1); reset count
//   - uppercase G..Z: count += (ch - 'F')      // G=1 … Z=20; consecutive accumulate
//   - lowercase g..z: count += 20 * (ch - 'f') // g=20 … z=400
//   - ',' : fill remainder of current row with '0'
//   - '!' : fill remainder of current row with 'F'
//   - ':' : emit a copy of the previous complete row (or a row of '0' if none yet)
func decompressGraphicFieldASCII(data string, bytesPerRow int) string {
	if bytesPerRow <= 0 {
		bytesPerRow = 1
	}
	rowLen := bytesPerRow * 2

	var out strings.Builder
	out.Grow(len(data) * 2)

	row := make([]byte, 0, rowLen)
	var prevRow []byte
	count := 0

	emitNibble := func(ch byte, n int) {
		for i := 0; i < n; i++ {
			row = append(row, ch)
			if len(row) == rowLen {
				out.Write(row)
				prevRow = append(prevRow[:0], row...)
				row = row[:0]
			}
		}
	}

	fillRest := func(fill byte) {
		need := rowLen - len(row)
		if need <= 0 {
			need = rowLen
			row = row[:0]
		}
		for i := 0; i < need; i++ {
			row = append(row, fill)
		}
		out.Write(row)
		prevRow = append(prevRow[:0], row...)
		row = row[:0]
		count = 0
	}

	for i := 0; i < len(data); i++ {
		ch := data[i]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			// ignore whitespace
		case (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f'):
			n := count
			if n == 0 {
				n = 1
			}
			// Preserve original case so plain hex stays byte-identical to prior parsing.
			emitNibble(ch, n)
			count = 0
		case ch >= 'G' && ch <= 'Z':
			count += int(ch - 'F')
		case ch >= 'g' && ch <= 'z':
			count += 20 * int(ch-'f')
		case ch == ',':
			fillRest('0')
		case ch == '!':
			fillRest('F')
		case ch == ':':
			pr := prevRow
			if pr == nil {
				pr = bytes.Repeat([]byte{'0'}, rowLen)
			}
			n := count
			if n == 0 {
				n = 1
			}
			for j := 0; j < n; j++ {
				out.Write(pr)
				prevRow = append(prevRow[:0], pr...)
			}
			row = row[:0]
			count = 0
		default:
			// Unknown character — ignore
		}
	}

	// Append any incomplete trailing row
	if len(row) > 0 {
		out.Write(row)
	}

	return out.String()
}

func (p *parser) parsePrintOrientation() error {
	params := p.readParams()
	if getParam(params, 0) != "" {
		orient := PrintOrientation(getParam(params, 0)[0])
		p.label.SetPrintOrientation(orient)
	}
	return nil
}

func (p *parser) parsePrintWidth() error {
	params := p.readParams()
	width := parseInt(getParam(params, 0), 0)
	if width > 0 {
		// Set width on label (we need a way to set just width)
		// For now, use a reasonable default height
		p.label.SetSizeDots(width, p.label.Height())
	}
	return nil
}

func (p *parser) parseLabelLength() error {
	params := p.readParams()
	length := parseInt(getParam(params, 0), 0)
	if length > 0 {
		p.label.SetSizeDots(p.label.Width(), length)
	}
	return nil
}

func (p *parser) parseLabelHome() error {
	params := p.readParams()
	x := parseInt(getParam(params, 0), 0)
	y := parseInt(getParam(params, 1), 0)
	p.label.SetHomeDots(x, y)
	return nil
}

func (p *parser) parseCharacterSet() error {
	params := p.readParams()
	charSet := parseInt(getParam(params, 0), 0)
	p.label.Add(NewCharacterSet(charSet))
	return nil
}

func (p *parser) parseComment() error {
	// Read until ^FS or next ^
	comment := ""
	for p.pos < len(p.input) {
		if strings.HasPrefix(p.input[p.pos:], "^FS") {
			p.pos += 3
			break
		}
		if p.input[p.pos] == '^' {
			break
		}
		comment += string(p.input[p.pos])
		p.pos++
	}
	if comment != "" {
		p.label.Add(NewComment(comment))
	}
	return nil
}

func getParam(params []string, index int) string {
	if index < len(params) {
		return params[index]
	}
	return ""
}
