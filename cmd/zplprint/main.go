// Zplprint sends a .zpl file raw to a USB Zebra printer on macOS through CUPS.
//
// Usage:
//
//	zplprint <file.zpl> [printer-name]
//
// Printing always invokes:
//
//	lp -d ZebraRaw -o raw
//
// The optional printer name (or auto-detection via lpinfo -v when omitted) is
// only echoed in the "Sending <file> to <uri>" line; it does not select the
// destination queue.
//
// If the ZebraRaw queue is missing, the program prints these CUPS setup steps:
//
//  1. Open http://localhost:631 in your browser
//  2. Go to Administration > Add Printer
//  3. Select your Zebra USB printer
//  4. Name it 'ZebraRaw'
//  5. For driver, choose 'Raw' or 'Generic Text-Only'
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: zplprint <file.zpl> [printer-name]")
		fmt.Println("\nIf no printer specified, will try to find a Zebra printer.")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read the ZPL file
	data, err := os.ReadFile(filename) //nolint:gosec // User-provided filename is intended
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Find printer
	printerURI := ""
	if len(os.Args) >= 3 {
		printerURI = os.Args[2]
	} else {
		// Auto-detect Zebra printer
		out, err := exec.Command("lpinfo", "-v").Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing printers: %v\n", err)
			os.Exit(1)
		}

		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(strings.ToLower(line), "zebra") && strings.Contains(line, "usb://") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					printerURI = parts[1]
					break
				}
			}
		}

		if printerURI == "" {
			fmt.Fprintln(os.Stderr, "No Zebra USB printer found. Available printers:")
			_ = exec.Command("lpinfo", "-v").Run()
			os.Exit(1)
		}
	}

	fmt.Printf("Sending %s to %s\n", filename, printerURI)

	// Use lp command with the printer URI
	// First, we need to create a temporary printer or use an existing one

	// Try using lp with -h to specify the backend directly
	// Actually, let's use a different approach - pipe through lp

	cmd := exec.Command("lp", "-d", "ZebraRaw", "-o", "raw")
	cmd.Stdin = strings.NewReader(string(data))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		// If ZebraRaw doesn't exist, try creating via CUPS web admin
		fmt.Fprintf(os.Stderr, "\nCouldn't print to ZebraRaw queue.\n")
		fmt.Fprintln(os.Stderr, "\nTry this instead:")
		fmt.Fprintln(os.Stderr, "1. Open http://localhost:631 in your browser")
		fmt.Fprintln(os.Stderr, "2. Go to Administration > Add Printer")
		fmt.Fprintln(os.Stderr, "3. Select your Zebra USB printer")
		fmt.Fprintln(os.Stderr, "4. Name it 'ZebraRaw'")
		fmt.Fprintln(os.Stderr, "5. For driver, choose 'Raw' or 'Generic Text-Only'")
		fmt.Fprintln(os.Stderr, "\nOr use: cat yourfile.zpl | lp -d PrinterName -o raw")
		os.Exit(1)
	}

	fmt.Println("Sent successfully!")
}
