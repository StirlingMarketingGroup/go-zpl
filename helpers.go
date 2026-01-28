package zpl

// TextField is a convenience method to add positioned text to the label.
func (l *Label) TextField(x, y int, font Font, height, width int, data string) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewScalableFont(font, height, width)).
		Add(NewFieldData(data))
}

// TextFieldAt is a convenience method using unit conversion.
func (l *Label) TextFieldAt(x, y float64, unit Unit, font Font, height, width int, data string) *Label {
	return l.TextField(ToDots(x, unit, l.dpi), ToDots(y, unit, l.dpi), font, height, width, data)
}

// TextBlock adds a text block with wrapping at the specified position.
func (l *Label) TextBlock(x, y int, font Font, height, width int, blockWidth, maxLines int, justify Justification, data string) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewScalableFont(font, height, width)).
		Add(NewFieldBlock(blockWidth).WithMaxLines(maxLines).WithJustification(justify)).
		Add(NewFieldData(data))
}

// Box adds a graphic box at the specified position.
func (l *Label) Box(x, y, width, height, thickness int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewGraphicBox(width, height, thickness))
}

// BoxAt adds a graphic box using unit conversion.
func (l *Label) BoxAt(x, y, width, height float64, unit Unit, thickness int) *Label {
	return l.Box(
		ToDots(x, unit, l.dpi),
		ToDots(y, unit, l.dpi),
		ToDots(width, unit, l.dpi),
		ToDots(height, unit, l.dpi),
		thickness,
	)
}

// FilledBox adds a filled graphic box at the specified position.
func (l *Label) FilledBox(x, y, width, height int) *Label {
	// A filled box has thickness equal to the smaller dimension
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewGraphicBox(width, height, min(width, height)))
}

// HorizontalLine adds a horizontal line at the specified position.
func (l *Label) HorizontalLine(x, y, length, thickness int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewGraphicBox(length, thickness, thickness))
}

// VerticalLine adds a vertical line at the specified position.
func (l *Label) VerticalLine(x, y, length, thickness int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewGraphicBox(thickness, length, thickness))
}

// Circle adds a circle at the specified position.
func (l *Label) Circle(x, y, diameter, thickness int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewGraphicCircle(diameter, thickness))
}

// Code128 adds a Code 128 barcode at the specified position.
func (l *Label) Code128(x, y int, data string, height int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodeCode128(data, height))
}

// Code128At adds a Code 128 barcode using unit conversion.
func (l *Label) Code128At(x, y float64, unit Unit, data string, height int) *Label {
	return l.Code128(ToDots(x, unit, l.dpi), ToDots(y, unit, l.dpi), data, height)
}

// Code39 adds a Code 39 barcode at the specified position.
func (l *Label) Code39(x, y int, data string, height int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodeCode39(data, height))
}

// QRCode adds a QR code at the specified position.
func (l *Label) QRCode(x, y int, data string, magnification int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodeQR(data, magnification))
}

// QRCodeAt adds a QR code using unit conversion.
func (l *Label) QRCodeAt(x, y float64, unit Unit, data string, magnification int) *Label {
	return l.QRCode(ToDots(x, unit, l.dpi), ToDots(y, unit, l.dpi), data, magnification)
}

// DataMatrix adds a DataMatrix barcode at the specified position.
func (l *Label) DataMatrix(x, y int, data string, moduleSize int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodeDataMatrix(data, moduleSize))
}

// PDF417 adds a PDF417 barcode at the specified position.
func (l *Label) PDF417(x, y int, data string, height int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodePDF417(data, height))
}

// EAN13 adds an EAN-13 barcode at the specified position.
func (l *Label) EAN13(x, y int, data string, height int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodeEAN13(data, height))
}

// UPCA adds a UPC-A barcode at the specified position.
func (l *Label) UPCA(x, y int, data string, height int) *Label {
	return l.Add(NewFieldOrigin(x, y)).
		Add(NewBarcodeUPCA(data, height))
}

// Comment adds a comment to the ZPL (not printed, for documentation).
func (l *Label) Comment(text string) *Label {
	return l.Add(NewComment(text))
}

// ReverseField enables reverse printing for the next field.
func (l *Label) ReverseField() *Label {
	return l.Add(NewFieldReverse())
}

// SetDefaultFont sets the default font for subsequent text fields.
func (l *Label) SetDefaultFont(font Font, height, width int) *Label {
	return l.Add(NewChangeFont(font, height, width))
}

// SetCharacterSet sets the character encoding.
func (l *Label) SetCharacterSet(charSet int) *Label {
	return l.Add(NewCharacterSet(charSet))
}

// SetBarcodeDefaults sets the default barcode parameters.
func (l *Label) SetBarcodeDefaults(moduleWidth int, ratio float64, height int) *Label {
	return l.Add(NewBarcodeDefault(moduleWidth, ratio, height))
}
