// Example demonstrating different fonts in the render package.
package main

import (
	"fmt"
	"os"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

func main() {
	// Create a label comparing Font 0 and Font E (OCR-B)
	label := zpl.NewLabel().
		SetSizeDots(500, 350).
		// Header
		Add(zpl.NewFieldOrigin(10, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 24, 28)).
		Add(zpl.NewFieldData("Font Comparison")).
		Add(zpl.NewFieldOrigin(0, 40)).
		Add(zpl.NewGraphicBox(500, 2, 2)).
		// Font 0 section
		Add(zpl.NewFieldOrigin(10, 55)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("Font 0 (Roboto Condensed Bold):")).
		Add(zpl.NewFieldOrigin(10, 80)).
		Add(zpl.NewScalableFont(zpl.Font0, 32, 36)).
		Add(zpl.NewFieldData("ABCDEFGHIJKLMNOP")).
		Add(zpl.NewFieldOrigin(10, 115)).
		Add(zpl.NewScalableFont(zpl.Font0, 32, 36)).
		Add(zpl.NewFieldData("0123456789")).
		// Divider
		Add(zpl.NewFieldOrigin(0, 160)).
		Add(zpl.NewGraphicBox(500, 2, 2)).
		// Font E (OCR-B) section
		Add(zpl.NewFieldOrigin(10, 175)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("Font E (OCR-B):")).
		Add(zpl.NewFieldOrigin(10, 200)).
		Add(zpl.NewScalableFont(zpl.FontE, 32, 32)).
		Add(zpl.NewFieldData("ABCDEFGHIJKLMNOP")).
		Add(zpl.NewFieldOrigin(10, 240)).
		Add(zpl.NewScalableFont(zpl.FontE, 32, 32)).
		Add(zpl.NewFieldData("0123456789")).
		// Divider
		Add(zpl.NewFieldOrigin(0, 285)).
		Add(zpl.NewGraphicBox(500, 2, 2)).
		// Mixed usage example
		Add(zpl.NewFieldOrigin(10, 300)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("Barcode digits: ")).
		Add(zpl.NewFieldOrigin(180, 295)).
		Add(zpl.NewScalableFont(zpl.FontE, 28, 28)).
		Add(zpl.NewFieldData("1234567890128"))

	// Create renderer
	renderer := render.New(zpl.DPI203)

	// Render to file
	f, err := os.Create("fonts.png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}

	if err := renderer.RenderPNG(label, f); err != nil {
		f.Close()
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
		os.Exit(1)
	}
	f.Close()

	fmt.Println("Font comparison rendered to fonts.png")
}
