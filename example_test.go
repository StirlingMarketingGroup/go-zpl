package zpl_test

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
)

// ExampleNewLabel builds a 4×6 inch label with text, Code 128, and a QR code.
func ExampleNewLabel() {
	label := zpl.NewLabel().
		SetDPI(zpl.DPI203).
		SetSize(4, 6, zpl.UnitInches).
		TextField(50, 50, zpl.Font0, 30, 30, "Hello, World!").
		Code128(50, 150, "123456789", 100).
		QRCode(50, 300, "https://example.com", 5)
	fmt.Println(label)
	// Output:
	// ^XA
	// ^PW812
	// ^LL1218
	// ^FO50,50
	// ^A0N,30,30
	// ^FDHello, World!^FS
	// ^FO50,150
	// ^BCN,100,Y,N,N,N^FD123456789^FS
	// ^FO50,300
	// ^BQN,2,5^FDMA,https://example.com^FS
	// ^XZ
}

// ExampleLabel_TextBlock wraps text in a centered field block.
func ExampleLabel_TextBlock() {
	label := zpl.NewLabel().
		TextBlock(50, 50, zpl.Font0, 30, 30, 300, 3, zpl.JustifyCenter, "This text wraps and is centered.")
	fmt.Println(label)
	// Output:
	// ^XA
	// ^FO50,50
	// ^A0N,30,30
	// ^FB300,3,0,C,0
	// ^FDThis text wraps and is centered.^FS
	// ^XZ
}

// ExampleLabel_TextFieldAt converts inch positions to printer dots.
func ExampleLabel_TextFieldAt() {
	label := zpl.NewLabel().
		SetDPI(zpl.DPI203).
		TextFieldAt(0.5, 0.5, zpl.UnitInches, zpl.Font0, 30, 30, "Text")
	fmt.Println(label)
	// Output:
	// ^XA
	// ^FO101,101
	// ^A0N,30,30
	// ^FDText^FS
	// ^XZ
}

// ExampleToDots converts inches and millimeters to dots at a given DPI.
func ExampleToDots() {
	fmt.Println(zpl.ToDots(1, zpl.UnitInches, zpl.DPI203))
	fmt.Println(zpl.ToDots(25.4, zpl.UnitMillimeters, zpl.DPI300))
	// Output:
	// 203
	// 300
}

// ExampleLabel_Add uses command constructors for rotated text.
func ExampleLabel_Add() {
	label := zpl.NewLabel().
		Add(zpl.NewFieldOrigin(100, 200)).
		Add(zpl.NewScalableFont(zpl.FontE, 50, 40).WithOrientation(zpl.OrientationRotated90)).
		Add(zpl.NewFieldData("Rotated Text"))
	fmt.Println(label)
	// Output:
	// ^XA
	// ^FO100,200
	// ^AER,50,40
	// ^FDRotated Text^FS
	// ^XZ
}

// ExampleParse parses a ZPL string into a Label.
func ExampleParse() {
	src := `^XA
^FO50,50^A0N,30,30^FDHello World^FS
^FO50,100^BQN,2,5^FDMA,https://example.com^FS
^XZ`
	label, err := zpl.Parse(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(label.Commands()))
	// Output:
	// 5
}

// ExampleParseAll parses multi-page ZPL into one Label per printable block.
func ExampleParseAll() {
	src := "^XA^FO0,0^FDPage 1^FS^XZ\n^XA^FO0,0^FDPage 2^FS^XZ"
	labels, err := zpl.ParseAll(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(labels))
	// Output:
	// 2
}

// ExampleEscapeFieldData escapes ^ and ~ so untrusted input cannot inject ZPL commands.
func ExampleEscapeFieldData() {
	fmt.Println(zpl.EscapeFieldData(`price^~and\more`))
	// Output:
	// price_5E_7Eand\more
}

// ExampleImageConverter converts an 8×8 image into a ^GF graphic field.
func ExampleImageConverter() {
	img := image.NewGray(image.Rect(0, 0, 8, 8))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, 4, 4), image.Black, image.Point{}, draw.Src)
	gf := zpl.NewImageConverter().WithThreshold(128).Convert(img)
	label := zpl.NewLabel().Add(gf)
	fmt.Println(label)
	// Output:
	// ^XA
	// ^GFA,8,8,1,F0F0F0F000000000
	// ^XZ
}

// ExampleLabel_WriteTo writes ZPL to an io.Writer.
func ExampleLabel_WriteTo() {
	label := zpl.NewLabel().TextField(10, 10, zpl.Font0, 20, 20, "Hi")
	var buf bytes.Buffer
	n, err := label.WriteTo(&buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(n)
	fmt.Print(buf.String())
	// Output:
	// 37
	// ^XA
	// ^FO10,10
	// ^A0N,20,20
	// ^FDHi^FS
	// ^XZ
}
