// Code derived from github.com/boombuler/barcode/pdf417 (MIT License).
package pdf417

import (
	"fmt"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/utils"
)

const (
	padding_codeword = 900
)

// Encodes the given data and color scheme as PDF417 barcode.
// securityLevel should be between 0 and 8. The higher the number, the more
// additional error-correction codes are added.
func EncodeWithColor(data string, securityLevel byte, color barcode.ColorScheme) (barcode.Barcode, error) {
	return EncodeWithColorAndDimensions(data, securityLevel, 0, 0, color)
}

// EncodeWithColorAndDimensions encodes a PDF417 barcode with optional columns/rows overrides.
// If columns or rows are 0, they are automatically chosen.
func EncodeWithColorAndDimensions(data string, securityLevel byte, columns, rows int, color barcode.ColorScheme) (barcode.Barcode, error) {
	if securityLevel >= 9 {
		return nil, fmt.Errorf("Invalid security level %d", securityLevel)
	}

	sl := securitylevel(securityLevel)

	dataWords, err := highlevelEncode(data)
	if err != nil {
		return nil, err
	}

	columns, rows, err = resolveDimensions(len(dataWords), sl, columns, rows)
	if err != nil {
		return nil, err
	}

	barcode := new(pdfBarcode)
	barcode.data = data
	barcode.color = color

	codeWords, err := encodeDataWithDimensions(dataWords, columns, rows, sl)
	if err != nil {
		return nil, err
	}

	grid := [][]int{}
	for i := 0; i < len(codeWords); i += columns {
		grid = append(grid, codeWords[i:min(i+columns, len(codeWords))])
	}

	codes := [][]int{}

	for rowNum, row := range grid {
		table := rowNum % 3
		rowCodes := make([]int, 0, columns+4)

		rowCodes = append(rowCodes, start_word)
		rowCodes = append(rowCodes, getCodeword(table, getLeftCodeWord(rowNum, rows, columns, securityLevel)))

		for _, word := range row {
			rowCodes = append(rowCodes, getCodeword(table, word))
		}

		rowCodes = append(rowCodes, getCodeword(table, getRightCodeWord(rowNum, rows, columns, securityLevel)))
		rowCodes = append(rowCodes, stop_word)

		codes = append(codes, rowCodes)
	}

	barcode.code = renderBarcode(codes)
	barcode.width = (columns+4)*17 + 1

	return barcode, nil
}

// Encodes the given data as PDF417 barcode.
// securityLevel should be between 0 and 8. The higher the number, the more
// additional error-correction codes are added.
func Encode(data string, securityLevel byte) (barcode.Barcode, error) {
	return EncodeWithColor(data, securityLevel, barcode.ColorScheme16)
}

// EncodeWithDimensions encodes a PDF417 barcode with optional columns/rows overrides.
// If columns or rows are 0, they are automatically chosen.
func EncodeWithDimensions(data string, securityLevel byte, columns, rows int) (barcode.Barcode, error) {
	return EncodeWithColorAndDimensions(data, securityLevel, columns, rows, barcode.ColorScheme16)
}

func resolveDimensions(dataWords int, sl securitylevel, columns, rows int) (int, int, error) {
	ecCount := sl.ErrorCorrectionWordCount()
	switch {
	case columns == 0 && rows == 0:
		cols, r := calcDimensions(dataWords, ecCount)
		if cols < minCols || cols > maxCols || r < minRows || r > maxRows {
			return 0, 0, fmt.Errorf("Unable to fit data in barcode")
		}
		return cols, r, nil
	case columns > 0 && rows == 0:
		if columns < minCols || columns > maxCols {
			return 0, 0, fmt.Errorf("Invalid columns %d", columns)
		}
		// Compute rows based on padding for the given column count.
		totalCount := dataWords + ecCount + 1
		rows = totalCount / columns
		if totalCount%columns != 0 {
			rows++
		}
		if rows < minRows || rows > maxRows {
			return 0, 0, fmt.Errorf("Unable to fit data in barcode")
		}
		return columns, rows, nil
	case columns == 0 && rows > 0:
		if rows < minRows || rows > maxRows {
			return 0, 0, fmt.Errorf("Invalid rows %d", rows)
		}
		totalCount := dataWords + ecCount + 1
		columns = totalCount / rows
		if totalCount%rows != 0 {
			columns++
		}
		if columns < minCols || columns > maxCols {
			return 0, 0, fmt.Errorf("Unable to fit data in barcode")
		}
		return columns, rows, nil
	default:
		if columns < minCols || columns > maxCols || rows < minRows || rows > maxRows {
			return 0, 0, fmt.Errorf("Unable to fit data in barcode")
		}
		return columns, rows, nil
	}
}

func encodeDataWithDimensions(dataWords []int, columns, rows int, sl securitylevel) ([]int, error) {
	if columns <= 0 || rows <= 0 {
		return nil, fmt.Errorf("Invalid PDF417 dimensions")
	}
	dataCount := len(dataWords)
	ecCount := sl.ErrorCorrectionWordCount()
	required := dataCount + ecCount + 1
	total := columns * rows
	if total < required {
		return nil, fmt.Errorf("Unable to fit data in barcode")
	}
	padCount := total - required
	if padCount > 0 {
		padWords := make([]int, padCount)
		for i := 0; i < padCount; i++ {
			padWords[i] = padding_codeword
		}
		dataWords = append(dataWords, padWords...)
	}

	length := len(dataWords) + 1
	dataWords = append([]int{length}, dataWords...)

	ecWords := sl.Compute(dataWords)
	return append(dataWords, ecWords...), nil
}

func encodeData(dataWords []int, columns int, sl securitylevel) ([]int, error) {
	dataCount := len(dataWords)

	ecCount := sl.ErrorCorrectionWordCount()

	padWords := getPadding(dataCount, ecCount, columns)
	dataWords = append(dataWords, padWords...)

	length := len(dataWords) + 1
	dataWords = append([]int{length}, dataWords...)

	ecWords := sl.Compute(dataWords)

	return append(dataWords, ecWords...), nil
}

func getLeftCodeWord(rowNum int, rows int, columns int, securityLevel byte) int {
	tableId := rowNum % 3

	var x int

	switch tableId {
	case 0:
		x = (rows - 1) / 3 // Per ISO 15438, section 5.3.2
	case 1:
		x = int(securityLevel) * 3
		x += (rows - 1) % 3
	case 2:
		x = columns - 1
	}

	return 30*(rowNum/3) + x
}

func getRightCodeWord(rowNum int, rows int, columns int, securityLevel byte) int {
	tableId := rowNum % 3

	var x int

	switch tableId {
	case 0:
		x = columns - 1
	case 1:
		x = (rows - 1) / 3
	case 2:
		x = int(securityLevel) * 3
		x += (rows - 1) % 3
	}

	return 30*(rowNum/3) + x
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func getPadding(dataCount int, ecCount int, columns int) []int {
	totalCount := dataCount + ecCount + 1
	mod := totalCount % columns

	padding := []int{}

	if mod > 0 {
		padCount := columns - mod
		padding = make([]int, padCount)
		for i := 0; i < padCount; i++ {
			padding[i] = padding_codeword
		}
	}

	return padding
}

func renderBarcode(codes [][]int) *utils.BitList {
	bl := new(utils.BitList)
	for _, row := range codes {
		lastIdx := len(row) - 1
		for i, col := range row {
			if i == lastIdx {
				bl.AddBits(col, 18)
			} else {
				bl.AddBits(col, 17)
			}
		}
	}
	return bl
}
