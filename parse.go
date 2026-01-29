package zpl

import (
	"strconv"
	"strings"
)

// Parse parses a ZPL string and returns a Label with all commands.
func Parse(zpl string) (*Label, error) {
	p := &parser{
		input: zpl,
		pos:   0,
		label: NewLabel(),
	}
	return p.parse()
}

// parser holds the state for parsing ZPL.
type parser struct {
	input string
	pos   int
	label *Label

	// Current state
	inFormat bool // Inside ^XA...^XZ block
}

func (p *parser) parse() (*Label, error) {
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

	return p.label, nil
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
	case "XZ":
		p.inFormat = false
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
	case "A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8", "A9":
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
	// Commands we recognize but ignore for now
	case "LR", "MN", "MF", "MC", "CV", "FH", "DN", "PQ", "B3":
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

func (p *parser) parseFieldData() error {
	// Read until ^FS
	data := ""
	for p.pos < len(p.input) {
		if strings.HasPrefix(p.input[p.pos:], "^FS") {
			p.pos += 3
			break
		}
		if p.input[p.pos] == '^' {
			// Another command - stop here
			break
		}
		data += string(p.input[p.pos])
		p.pos++
	}
	if data != "" {
		p.label.Add(NewFieldData(data))
	}
	return nil
}

func (p *parser) parseScalableFont(cmd string) error {
	// Get font number from command (A0 = Font0, etc)
	fontNum := 0
	if len(cmd) > 1 {
		fontNum = int(cmd[1] - '0')
	}

	params := p.readParams()
	orient := OrientationNormal
	if getParam(params, 0) != "" {
		orient = Orientation(getParam(params, 0)[0])
	}
	height := parseInt(getParam(params, 1), 30)
	width := parseInt(getParam(params, 2), 0)

	font := Font(rune('0' + fontNum))
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

	// The data comes from the next ^FD or ^FV command
	// We'll read it separately and create the barcode with empty data for now
	bc := NewBarcodeCode128("", height).
		WithOrientation(orient).
		WithInterpretation(printInterp, interpAbove).
		WithUCCCheckDigit(checkDigit).
		WithMode(mode)

	// Look for ^FD or ^FV
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '^' {
		save := p.pos
		p.pos++
		cmd := p.readCommandName()
		if cmd == "FD" || cmd == "FV" {
			data := ""
			for p.pos < len(p.input) {
				if strings.HasPrefix(p.input[p.pos:], "^FS") {
					p.pos += 3
					break
				}
				if p.input[p.pos] == '^' {
					break
				}
				data += string(p.input[p.pos])
				p.pos++
			}
			bc.Data = data
		} else {
			p.pos = save // Restore position
		}
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

	// Look for ^FD or ^FV
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '^' {
		save := p.pos
		p.pos++
		cmd := p.readCommandName()
		if cmd == "FD" || cmd == "FV" {
			// Read the raw field data
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

			// Parse the QR data format: [error_correction][mode],[data]
			// e.g., "MA,HelloWorld" or "HA,https://example.com"
			fieldData := rawData.String()
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
		} else {
			p.pos = save // Restore position
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

	var data string

	// Look for ^FD or ^FV
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '^' {
		save := p.pos
		p.pos++
		cmd := p.readCommandName()
		if cmd == "FD" || cmd == "FV" {
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
			data = rawData.String()
		} else {
			p.pos = save
		}
	}

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

	var data string

	// Look for ^FD or ^FV
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '^' {
		save := p.pos
		p.pos++
		cmd := p.readCommandName()
		if cmd == "FD" || cmd == "FV" {
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
			data = rawData.String()
		} else {
			p.pos = save
		}
	}

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

func (p *parser) parseMaxiCode() error {
	// ^BD command for MaxiCode 2D barcode
	// Format: ^BDm,n,t where m=mode (2-6), n=symbol number, t=total symbols
	params := p.readParams()
	mode := MaxiCodeMode(parseInt(getParam(params, 0), 2))
	symbolNumber := parseInt(getParam(params, 1), 1)
	symbolCount := parseInt(getParam(params, 2), 1)

	var hexIndicator rune
	var data string

	// Look for ^FH (field hexadecimal indicator) and then ^FD or ^FV
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != '^' {
			break
		}

		// Peek at the next command
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
			// Read the field data using strings.Builder for efficiency
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
			// Decode hex escapes if hex indicator is set
			if hexIndicator != 0 {
				data = decodeHexEscapes(rawData.String(), hexIndicator)
			} else {
				data = rawData.String()
			}

			// Create and add the MaxiCode barcode
			mc := NewBarcodeMaxiCode(data, mode).
				WithStructuredAppend(symbolNumber, symbolCount)
			if hexIndicator != 0 {
				mc.WithHexIndicator(hexIndicator)
			}
			p.label.Add(mc)
			return nil
		default:
			// Not a field command, restore position and exit
			p.pos = save
			return nil
		}
	}
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
	params := p.readParams()
	format := GraphicFieldASCII
	if getParam(params, 0) != "" {
		format = GraphicFieldFormat(getParam(params, 0)[0])
	}
	dataBytes := parseInt(getParam(params, 1), 0)   // Total bytes in the data
	bytesPerRow := parseInt(getParam(params, 3), 1) // Bytes per row

	switch format {
	case GraphicFieldBinary:
		// Binary format: read exactly dataBytes of raw binary data
		if dataBytes <= 0 {
			return nil
		}

		// Read the binary data directly
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
			p.label.Add(gf)
		}

	case GraphicFieldASCII:
		// ASCII format: read hex characters
		// Remaining data is the hex string (may span multiple lines)
		var data strings.Builder
		if len(params) > 4 {
			data.WriteString(getParam(params, 4))
		}

		// Continue reading hex data until ^FS or next command
		for p.pos < len(p.input) {
			if strings.HasPrefix(p.input[p.pos:], "^FS") {
				p.pos += 3
				break
			}
			if p.input[p.pos] == '^' {
				break
			}
			ch := p.input[p.pos]
			if (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f') {
				data.WriteByte(ch)
			}
			p.pos++
		}

		if data.Len() > 0 {
			gf := NewGraphicFieldASCII(bytesPerRow, data.String())
			p.label.Add(gf)
		}
	}

	return nil
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
