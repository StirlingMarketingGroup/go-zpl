package render

import (
	"image/color"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// Code 128 bar patterns: each symbol is 6 elements (bars and spaces alternating)
// Values represent the width of each element in modules (1-4)
// Pattern starts with bar, then space, bar, space, bar, space
var code128Patterns = [][]int{
	{2, 1, 2, 2, 2, 2}, // 0: space (Subset B: space, Subset A: space, Subset C: 00)
	{2, 2, 2, 1, 2, 2}, // 1
	{2, 2, 2, 2, 2, 1}, // 2
	{1, 2, 1, 2, 2, 3}, // 3
	{1, 2, 1, 3, 2, 2}, // 4
	{1, 3, 1, 2, 2, 2}, // 5
	{1, 2, 2, 2, 1, 3}, // 6
	{1, 2, 2, 3, 1, 2}, // 7
	{1, 3, 2, 2, 1, 2}, // 8
	{2, 2, 1, 2, 1, 3}, // 9
	{2, 2, 1, 3, 1, 2}, // 10
	{2, 3, 1, 2, 1, 2}, // 11
	{1, 1, 2, 2, 3, 2}, // 12
	{1, 2, 2, 1, 3, 2}, // 13
	{1, 2, 2, 2, 3, 1}, // 14
	{1, 1, 3, 2, 2, 2}, // 15
	{1, 2, 3, 1, 2, 2}, // 16
	{1, 2, 3, 2, 2, 1}, // 17
	{2, 2, 3, 2, 1, 1}, // 18
	{2, 2, 1, 1, 3, 2}, // 19
	{2, 2, 1, 2, 3, 1}, // 20
	{2, 1, 3, 2, 1, 2}, // 21
	{2, 2, 3, 1, 1, 2}, // 22
	{3, 1, 2, 1, 3, 1}, // 23
	{3, 1, 1, 2, 2, 2}, // 24
	{3, 2, 1, 1, 2, 2}, // 25
	{3, 2, 1, 2, 2, 1}, // 26
	{3, 1, 2, 2, 1, 2}, // 27
	{3, 2, 2, 1, 1, 2}, // 28
	{3, 2, 2, 2, 1, 1}, // 29
	{2, 1, 2, 1, 2, 3}, // 30
	{2, 1, 2, 3, 2, 1}, // 31
	{2, 3, 2, 1, 2, 1}, // 32
	{1, 1, 1, 3, 2, 3}, // 33
	{1, 3, 1, 1, 2, 3}, // 34
	{1, 3, 1, 3, 2, 1}, // 35
	{1, 1, 2, 3, 1, 3}, // 36
	{1, 3, 2, 1, 1, 3}, // 37
	{1, 3, 2, 3, 1, 1}, // 38
	{2, 1, 1, 3, 1, 3}, // 39
	{2, 3, 1, 1, 1, 3}, // 40
	{2, 3, 1, 3, 1, 1}, // 41
	{1, 1, 2, 1, 3, 3}, // 42
	{1, 1, 2, 3, 3, 1}, // 43
	{1, 3, 2, 1, 3, 1}, // 44
	{1, 1, 3, 1, 2, 3}, // 45
	{1, 1, 3, 3, 2, 1}, // 46
	{1, 3, 3, 1, 2, 1}, // 47
	{3, 1, 3, 1, 2, 1}, // 48
	{2, 1, 1, 3, 3, 1}, // 49
	{2, 3, 1, 1, 3, 1}, // 50
	{2, 1, 3, 1, 1, 3}, // 51
	{2, 1, 3, 3, 1, 1}, // 52
	{2, 1, 3, 1, 3, 1}, // 53
	{3, 1, 1, 1, 2, 3}, // 54
	{3, 1, 1, 3, 2, 1}, // 55
	{3, 3, 1, 1, 2, 1}, // 56
	{3, 1, 2, 1, 1, 3}, // 57
	{3, 1, 2, 3, 1, 1}, // 58
	{3, 3, 2, 1, 1, 1}, // 59
	{3, 1, 4, 1, 1, 1}, // 60
	{2, 2, 1, 4, 1, 1}, // 61
	{4, 3, 1, 1, 1, 1}, // 62
	{1, 1, 1, 2, 2, 4}, // 63
	{1, 1, 1, 4, 2, 2}, // 64
	{1, 2, 1, 1, 2, 4}, // 65
	{1, 2, 1, 4, 2, 1}, // 66
	{1, 4, 1, 1, 2, 2}, // 67
	{1, 4, 1, 2, 2, 1}, // 68
	{1, 1, 2, 2, 1, 4}, // 69
	{1, 1, 2, 4, 1, 2}, // 70
	{1, 2, 2, 1, 1, 4}, // 71
	{1, 2, 2, 4, 1, 1}, // 72
	{1, 4, 2, 1, 1, 2}, // 73
	{1, 4, 2, 2, 1, 1}, // 74
	{2, 4, 1, 2, 1, 1}, // 75
	{2, 2, 1, 1, 1, 4}, // 76
	{4, 1, 3, 1, 1, 1}, // 77
	{2, 4, 1, 1, 1, 2}, // 78
	{1, 3, 4, 1, 1, 1}, // 79
	{1, 1, 1, 2, 4, 2}, // 80
	{1, 2, 1, 1, 4, 2}, // 81
	{1, 2, 1, 2, 4, 1}, // 82
	{1, 1, 4, 2, 1, 2}, // 83
	{1, 2, 4, 1, 1, 2}, // 84
	{1, 2, 4, 2, 1, 1}, // 85
	{4, 1, 1, 2, 1, 2}, // 86
	{4, 2, 1, 1, 1, 2}, // 87
	{4, 2, 1, 2, 1, 1}, // 88
	{2, 1, 2, 1, 4, 1}, // 89
	{2, 1, 4, 1, 2, 1}, // 90
	{4, 1, 2, 1, 2, 1}, // 91
	{1, 1, 1, 1, 4, 3}, // 92
	{1, 1, 1, 3, 4, 1}, // 93
	{1, 3, 1, 1, 4, 1}, // 94
	{1, 1, 4, 1, 1, 3}, // 95
	{1, 1, 4, 3, 1, 1}, // 96
	{4, 1, 1, 1, 1, 3}, // 97
	{4, 1, 1, 3, 1, 1}, // 98
	{1, 1, 3, 1, 4, 1}, // 99
	{1, 1, 4, 1, 3, 1}, // 100 (Code B)
	{3, 1, 1, 1, 4, 1}, // 101 (Code A)
	{4, 1, 1, 1, 3, 1}, // 102 (FNC1)
	{2, 1, 1, 4, 1, 2}, // 103 (Start A)
	{2, 1, 1, 2, 1, 4}, // 104 (Start B)
	{2, 1, 1, 2, 3, 2}, // 105 (Start C)
}

// Stop pattern: 2-3-3-1-1-1-2 (7 elements, ends with final bar)
var code128Stop = []int{2, 3, 3, 1, 1, 1, 2}

// Code 128 character value mapping for Subset B (most common)
var code128BValues = map[rune]int{
	' ': 0, '!': 1, '"': 2, '#': 3, '$': 4, '%': 5, '&': 6, '\'': 7,
	'(': 8, ')': 9, '*': 10, '+': 11, ',': 12, '-': 13, '.': 14, '/': 15,
	'0': 16, '1': 17, '2': 18, '3': 19, '4': 20, '5': 21, '6': 22, '7': 23,
	'8': 24, '9': 25, ':': 26, ';': 27, '<': 28, '=': 29, '>': 30, '?': 31,
	'@': 32, 'A': 33, 'B': 34, 'C': 35, 'D': 36, 'E': 37, 'F': 38, 'G': 39,
	'H': 40, 'I': 41, 'J': 42, 'K': 43, 'L': 44, 'M': 45, 'N': 46, 'O': 47,
	'P': 48, 'Q': 49, 'R': 50, 'S': 51, 'T': 52, 'U': 53, 'V': 54, 'W': 55,
	'X': 56, 'Y': 57, 'Z': 58, '[': 59, '\\': 60, ']': 61, '^': 62, '_': 63,
	'`': 64, 'a': 65, 'b': 66, 'c': 67, 'd': 68, 'e': 69, 'f': 70, 'g': 71,
	'h': 72, 'i': 73, 'j': 74, 'k': 75, 'l': 76, 'm': 77, 'n': 78, 'o': 79,
	'p': 80, 'q': 81, 'r': 82, 's': 83, 't': 84, 'u': 85, 'v': 86, 'w': 87,
	'x': 88, 'y': 89, 'z': 90, '{': 91, '|': 92, '}': 93, '~': 94,
}

// Code 128 Subset C values (digit pairs 00-99)
var code128CValues = map[string]int{
	"00": 0, "01": 1, "02": 2, "03": 3, "04": 4, "05": 5, "06": 6, "07": 7, "08": 8, "09": 9,
	"10": 10, "11": 11, "12": 12, "13": 13, "14": 14, "15": 15, "16": 16, "17": 17, "18": 18, "19": 19,
	"20": 20, "21": 21, "22": 22, "23": 23, "24": 24, "25": 25, "26": 26, "27": 27, "28": 28, "29": 29,
	"30": 30, "31": 31, "32": 32, "33": 33, "34": 34, "35": 35, "36": 36, "37": 37, "38": 38, "39": 39,
	"40": 40, "41": 41, "42": 42, "43": 43, "44": 44, "45": 45, "46": 46, "47": 47, "48": 48, "49": 49,
	"50": 50, "51": 51, "52": 52, "53": 53, "54": 54, "55": 55, "56": 56, "57": 57, "58": 58, "59": 59,
	"60": 60, "61": 61, "62": 62, "63": 63, "64": 64, "65": 65, "66": 66, "67": 67, "68": 68, "69": 69,
	"70": 70, "71": 71, "72": 72, "73": 73, "74": 74, "75": 75, "76": 76, "77": 77, "78": 78, "79": 79,
	"80": 80, "81": 81, "82": 82, "83": 83, "84": 84, "85": 85, "86": 86, "87": 87, "88": 88, "89": 89,
	"90": 90, "91": 91, "92": 92, "93": 93, "94": 94, "95": 95, "96": 96, "97": 97, "98": 98, "99": 99,
}

// Code 128 special codes
const (
	code128CodeA  = 101 // Switch to Subset A
	code128CodeB  = 100 // Switch to Subset B
	code128CodeC  = 99  // Switch to Subset C
	code128FNC1   = 102 // FNC1
	code128StartA = 103 // Start A
	code128StartB = 104 // Start B
	code128StartC = 105 // Start C
)

// encodeCode128B encodes data using Code 128 Subset B
func encodeCode128B(data string) []int {
	values := []int{code128StartB}

	for _, r := range data {
		if val, ok := code128BValues[r]; ok {
			values = append(values, val)
		}
	}

	// Calculate check digit
	sum := values[0]
	for i := 1; i < len(values); i++ {
		sum += i * values[i]
	}
	checkDigit := sum % 103
	values = append(values, checkDigit)

	return values
}

// isDigit checks if a character is a digit
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// encodeCode128Auto encodes data using Code 128 Auto mode
// Uses Subset C for numeric runs of 4+ digits, Subset B otherwise
func encodeCode128Auto(data string) []int {
	if data == "" {
		return []int{}
	}

	// Find runs of digits
	type segment struct {
		start    int
		end      int
		isDigits bool
	}

	runes := []rune(data)
	var segments []segment

	i := 0
	for i < len(runes) {
		if isDigit(runes[i]) {
			// Count consecutive digits
			start := i
			for i < len(runes) && isDigit(runes[i]) {
				i++
			}
			segments = append(segments, segment{start: start, end: i, isDigits: true})
		} else {
			// Count consecutive non-digits
			start := i
			for i < len(runes) && !isDigit(runes[i]) {
				i++
			}
			segments = append(segments, segment{start: start, end: i, isDigits: false})
		}
	}

	// Decide which segments to encode with Subset C
	// Use Subset C for digit runs of 4+ (because switching costs 1 symbol)
	// Also use Subset C if we can start with digits (saves start code switching)
	for idx := range segments {
		if segments[idx].isDigits {
			digitLen := segments[idx].end - segments[idx].start
			// Use Subset C if:
			// - At least 4 digits, OR
			// - First segment and at least 2 digits (no switch cost for start)
			if digitLen >= 4 || (idx == 0 && digitLen >= 2 && digitLen%2 == 0) {
				// Keep isDigits true to use Subset C
			} else {
				// Convert to non-digits to use Subset B
				segments[idx].isDigits = false
			}
		}
	}

	// Merge adjacent non-digit segments
	var merged []segment
	for _, seg := range segments {
		if len(merged) > 0 && !merged[len(merged)-1].isDigits && !seg.isDigits {
			merged[len(merged)-1].end = seg.end
		} else {
			merged = append(merged, seg)
		}
	}
	segments = merged

	// Build symbol values
	var values []int
	position := 0 // For check digit calculation
	inSubsetC := false

	// Determine starting subset
	if len(segments) > 0 && segments[0].isDigits {
		values = append(values, code128StartC)
		inSubsetC = true
	} else {
		values = append(values, code128StartB)
		inSubsetC = false
	}
	position++

	for _, seg := range segments {
		segData := runes[seg.start:seg.end]

		if seg.isDigits {
			// Switch to Subset C if needed
			if !inSubsetC {
				values = append(values, 99) // Code C (value 99 in Subset B switches to C)
				position++
				inSubsetC = true
			}
			// Encode digit pairs
			// If odd number of digits, encode first digit in Subset B
			startIdx := 0
			if len(segData)%2 == 1 {
				// Switch to B for single digit, then back to C
				values = append(values, code128CodeB)
				position++
				if val, ok := code128BValues[segData[0]]; ok {
					values = append(values, val)
					position++
				}
				startIdx = 1
				// Switch back to C
				values = append(values, 99)
				position++
				// inSubsetC stays true since we're back in C
			}
			for j := startIdx; j < len(segData); j += 2 {
				pair := string(segData[j : j+2])
				if val, ok := code128CValues[pair]; ok {
					values = append(values, val)
					position++
				}
			}
		} else {
			// Switch to Subset B if needed
			if inSubsetC {
				values = append(values, code128CodeB)
				position++
				inSubsetC = false
			}
			// Encode each character
			for _, r := range segData {
				if val, ok := code128BValues[r]; ok {
					values = append(values, val)
					position++
				}
			}
		}
	}

	// Calculate check digit
	sum := values[0] // Start code
	for i := 1; i < len(values); i++ {
		sum += i * values[i]
	}
	checkDigit := sum % 103
	values = append(values, checkDigit)

	return values
}

// encodeCode128C encodes data using Code 128 Subset C only (numeric pairs)
// Returns nil if data contains non-digit characters or has odd length
func encodeCode128C(data string) []int {
	// Validate: must be all digits with even length
	if len(data)%2 != 0 {
		return nil
	}
	for _, r := range data {
		if !isDigit(r) {
			return nil
		}
	}

	values := []int{code128StartC}
	for i := 0; i < len(data); i += 2 {
		pair := data[i : i+2]
		if val, ok := code128CValues[pair]; ok {
			values = append(values, val)
		}
	}

	// Calculate check digit
	sum := values[0]
	for i := 1; i < len(values); i++ {
		sum += i * values[i]
	}
	checkDigit := sum % 103
	values = append(values, checkDigit)

	return values
}

// encodeCode128WithEscapes handles ZPL escape sequences in Code 128 data
// ZPL escape sequences start with '>':
//
//	>; - Switch to/start with Code C
//	>: - Switch to/start with Code A
//	>5 - FNC1
//	>6 - FNC2
//	>7 - FNC3
//	>8 - FNC4
//	>> - Literal '>' character
func encodeCode128WithEscapes(data string) []int {
	if data == "" {
		return []int{}
	}

	var values []int
	currentSubset := byte('B') // Default to Subset B
	started := false

	i := 0
	for i < len(data) {
		// Check for escape sequence
		if data[i] == '>' && i+1 < len(data) {
			switch data[i+1] {
			case ';': // Code C
				if !started {
					values = append(values, code128StartC)
					currentSubset = 'C'
					started = true
				} else if currentSubset != 'C' {
					values = append(values, code128CodeC)
					currentSubset = 'C'
				}
				i += 2
				continue
			case ':': // Code A
				if !started {
					values = append(values, code128StartA)
					currentSubset = 'A'
					started = true
				} else if currentSubset != 'A' {
					values = append(values, code128CodeA)
					currentSubset = 'A'
				}
				i += 2
				continue
			case '5': // FNC1
				if !started {
					values = append(values, code128StartC, code128FNC1) // FNC1 often used with GS1/Code C
					currentSubset = 'C'
					started = true
				} else {
					values = append(values, code128FNC1)
				}
				i += 2
				continue
			case '>': // Literal '>'
				i++ // Skip first '>', process second as normal character
			}
		}

		// Start with default subset if not started
		if !started {
			values = append(values, code128StartB)
			currentSubset = 'B'
			started = true
		}

		// Encode character based on current subset
		switch currentSubset {
		case 'C':
			// Code C encodes digit pairs
			if i+1 < len(data) && isDigit(rune(data[i])) && isDigit(rune(data[i+1])) {
				pair := data[i : i+2]
				if val, ok := code128CValues[pair]; ok {
					values = append(values, val)
				}
				i += 2
			} else {
				// Can't encode as pair, switch to B
				values = append(values, code128CodeB)
				currentSubset = 'B'
				// Don't advance i, re-process in Subset B
			}
		case 'B':
			if val, ok := code128BValues[rune(data[i])]; ok {
				values = append(values, val)
			}
			i++
		case 'A':
			// Subset A is similar to B but with different character mapping
			// For simplicity, use B values for printable ASCII
			if val, ok := code128BValues[rune(data[i])]; ok {
				values = append(values, val)
			}
			i++
		}
	}

	if len(values) == 0 {
		return []int{}
	}

	// Calculate check digit
	sum := values[0] // Start code
	for j := 1; j < len(values); j++ {
		sum += j * values[j]
	}
	checkDigit := sum % 103
	values = append(values, checkDigit)

	return values
}

// drawBarcode128 renders a Code 128 barcode on the canvas
func (c *canvas) drawBarcode128(bc *zpl.BarcodeCode128, moduleWidth int) {
	if bc.Data == "" {
		return
	}

	// Check if data contains ZPL escape sequences (starts with > followed by special char)
	hasEscapes := false
	for i := 0; i < len(bc.Data)-1; i++ {
		if bc.Data[i] == '>' {
			next := bc.Data[i+1]
			if next == ';' || next == ':' || next == '5' || next == '6' || next == '7' || next == '8' || next == '>' {
				hasEscapes = true
				break
			}
		}
	}

	// Select encoding based on escape sequences and mode
	// If data contains escape sequences (>;, >:, etc.), use escape-aware encoder
	// Otherwise use mode-based selection:
	//   Mode 'A' = Auto mode (optimize with Subset C for numeric runs)
	//   Mode 'N', 'B', or unspecified = Subset B (standard behavior)
	//   Mode 'C' = Subset C only (numeric pairs)
	var values []int
	if hasEscapes {
		values = encodeCode128WithEscapes(bc.Data)
	} else {
		switch bc.Mode {
		case zpl.Code128SubsetA: // 'A' is actually Auto mode in ZPL
			values = encodeCode128Auto(bc.Data)
		case zpl.Code128SubsetC:
			values = encodeCode128C(bc.Data)
			if values == nil {
				// Fall back to Subset B if data isn't valid for Subset C
				values = encodeCode128B(bc.Data)
			}
		default: // Mode N, B, or anything else uses Subset B
			values = encodeCode128B(bc.Data)
		}
	}

	// Calculate total width for positioning
	totalModules := 0
	for _, v := range values {
		for _, w := range code128Patterns[v] {
			totalModules += w
		}
	}
	for _, w := range code128Stop {
		totalModules += w
	}

	x := c.curX
	y := c.curY
	height := bc.Height
	if height == 0 {
		height = c.barcodeHeight // Use ^BY height
	}
	if height == 0 {
		height = 100 // Default height
	}

	col := color.RGBA{0, 0, 0, 255}

	// Draw each symbol
	currentX := x
	for _, v := range values {
		pattern := code128Patterns[v]
		isBar := true
		for _, width := range pattern {
			if isBar {
				// Draw bar
				for dx := 0; dx < width*moduleWidth; dx++ {
					for dy := 0; dy < height; dy++ {
						c.img.Set(currentX+dx, y+dy, col)
					}
				}
			}
			currentX += width * moduleWidth
			isBar = !isBar
		}
	}

	// Draw stop pattern
	isBar := true
	for _, width := range code128Stop {
		if isBar {
			for dx := 0; dx < width*moduleWidth; dx++ {
				for dy := 0; dy < height; dy++ {
					c.img.Set(currentX+dx, y+dy, col)
				}
			}
		}
		currentX += width * moduleWidth
		isBar = !isBar
	}

	// Draw human-readable text if enabled
	if bc.PrintInterpretation && c.fontMgr != nil {
		// Calculate barcode width for centering
		barcodeWidth := totalModules * moduleWidth

		textY := y + height + 5
		if bc.InterpretationAbove {
			textY = y - 25
		}

		// Use a font size proportional to barcode width (approx 1/4 of barcode height, capped)
		textHeight := max(height/6, 18)
		if textHeight > 40 {
			textHeight = 40
		}
		textWidth := textHeight

		// Calculate text width to center it
		textLen := len(bc.Data)
		estimatedTextWidth := textLen * textWidth * 6 / 10 // Approximate character width

		// Center the text under the barcode
		textX := x + (barcodeWidth-estimatedTextWidth)/2
		c.fontMgr.drawText(c.img, bc.Data, textX, textY, zpl.FontA, textHeight, textWidth, zpl.OrientationNormal, false, false)
	}
}

// drawBarcodeDefault stores barcode defaults on the canvas
func (c *canvas) setBarcodeDefault(bd *zpl.BarcodeDefault) {
	c.barcodeModuleWidth = bd.ModuleWidth
	c.barcodeHeight = bd.Height
}
