// Basic builds a 4×6 inch shipping label at 203 DPI: a border, company header,
// separators, a ship-to block, a Code 128 tracking barcode, a QR code, and
// weight/service fields. It prints the ZPL to stdout.
//
//	go run ./examples/basic
//
// Pipe the output to zplrender, or paste it into the web demo at
// https://stirlingmarketinggroup.github.io/go-zpl/ to preview the label.
package main

import (
	"fmt"

	"github.com/StirlingMarketingGroup/go-zpl"
)

func main() {
	// Create a 4x6 inch shipping label at 203 DPI
	label := zpl.NewLabel().
		SetDPI(zpl.DPI203).
		SetSize(4, 6, zpl.UnitInches)

	// Add a border around the label
	label.Box(10, 10, 792, 1198, 2)

	// Add company name
	label.TextField(50, 50, zpl.Font0, 40, 40, "ACME Shipping Co.")

	// Add horizontal separator
	label.HorizontalLine(10, 120, 792, 2)

	// Add shipping information
	label.TextField(50, 150, zpl.Font0, 30, 30, "Ship To:")
	label.TextField(50, 190, zpl.Font0, 25, 25, "John Doe")
	label.TextField(50, 225, zpl.Font0, 25, 25, "123 Main Street")
	label.TextField(50, 260, zpl.Font0, 25, 25, "Anytown, ST 12345")

	// Add another separator
	label.HorizontalLine(10, 320, 792, 2)

	// Add a Code 128 barcode for tracking
	label.TextField(50, 350, zpl.Font0, 20, 20, "Tracking Number:")
	label.Code128(50, 380, "1Z999AA10123456784", 80)

	// Add a QR code with shipment details
	label.TextField(500, 350, zpl.Font0, 20, 20, "Scan for details:")
	label.QRCode(500, 380, "https://track.example.com/1Z999AA10123456784", 4)

	// Add horizontal separator
	label.HorizontalLine(10, 520, 792, 2)

	// Add weight and dimensions
	label.TextField(50, 550, zpl.Font0, 25, 25, "Weight: 5.2 lbs")
	label.TextField(50, 590, zpl.Font0, 25, 25, "Dims: 12x8x6 in")

	// Add service type
	label.TextField(400, 550, zpl.Font0, 35, 35, "PRIORITY")
	label.Box(390, 540, 200, 60, 2)

	// Print the ZPL
	fmt.Println(label)
}
