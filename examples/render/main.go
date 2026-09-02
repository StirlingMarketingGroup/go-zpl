// Render builds a UPS-style shipping label and writes it to label.png in the
// current directory, then prints the ZPL to stdout.
//
//	go run ./examples/render
package main

import (
	"fmt"
	"os"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

func main() {
	// Create a sample UPS-style label
	label := zpl.NewLabel().
		SetSizeDots(812, 600).
		// Sender info
		Add(zpl.NewFieldOrigin(15, 10)).
		Add(zpl.NewScalableFont(zpl.Font0, 24, 28)).
		Add(zpl.NewFieldData("JOHN DOE")).
		Add(zpl.NewFieldOrigin(15, 35)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("123 SENDER STREET")).
		Add(zpl.NewFieldOrigin(15, 57)).
		Add(zpl.NewScalableFont(zpl.Font0, 20, 24)).
		Add(zpl.NewFieldData("NEW YORK, NY 10001")).
		// Divider line
		Add(zpl.NewFieldOrigin(0, 85)).
		Add(zpl.NewGraphicBox(812, 3, 3)).
		// Ship To header
		Add(zpl.NewFieldOrigin(15, 100)).
		Add(zpl.NewScalableFont(zpl.Font0, 28, 32)).
		Add(zpl.NewFieldData("SHIP TO:")).
		// Recipient info
		Add(zpl.NewFieldOrigin(40, 135)).
		Add(zpl.NewScalableFont(zpl.Font0, 32, 36)).
		Add(zpl.NewFieldData("JANE SMITH")).
		Add(zpl.NewFieldOrigin(40, 170)).
		Add(zpl.NewScalableFont(zpl.Font0, 28, 32)).
		Add(zpl.NewFieldData("456 RECEIVER ROAD")).
		Add(zpl.NewFieldOrigin(40, 205)).
		Add(zpl.NewScalableFont(zpl.Font0, 48, 52)).
		Add(zpl.NewFieldData("LOS ANGELES CA 90210")).
		// Another divider
		Add(zpl.NewFieldOrigin(0, 265)).
		Add(zpl.NewGraphicBox(812, 12, 12)).
		// Service type
		Add(zpl.NewFieldOrigin(20, 290)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 64)).
		Add(zpl.NewFieldData("UPS GROUND")).
		// Tracking info
		Add(zpl.NewFieldOrigin(20, 360)).
		Add(zpl.NewScalableFont(zpl.Font0, 24, 28)).
		Add(zpl.NewFieldData("TRACKING #: 1Z 999 AA1 01 2345 6784")).
		// Divider
		Add(zpl.NewFieldOrigin(0, 400)).
		Add(zpl.NewGraphicBox(812, 4, 4)).
		// Weight and package info
		Add(zpl.NewFieldOrigin(20, 420)).
		Add(zpl.NewScalableFont(zpl.Font0, 22, 26)).
		Add(zpl.NewFieldData("WEIGHT: 5.2 LBS")).
		Add(zpl.NewFieldOrigin(300, 420)).
		Add(zpl.NewScalableFont(zpl.Font0, 22, 26)).
		Add(zpl.NewFieldData("PKG: 1 OF 1")).
		// Large zone indicator
		Add(zpl.NewFieldOrigin(650, 420)).
		Add(zpl.NewScalableFont(zpl.Font0, 120, 100)).
		Add(zpl.NewFieldData("5"))

	// Create renderer at 203 DPI (standard Zebra resolution)
	renderer := render.New(zpl.DPI203)

	// Render to file
	f, err := os.Create("label.png")
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

	fmt.Println("Label rendered to label.png")

	// Also print the ZPL code
	fmt.Println("\nZPL Code:")
	fmt.Println(label.ZPL())
}
