// zplrender converts ZPL files to images
package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

var (
	dpi            = flag.Int("dpi", 203, "Printer DPI (203, 300, or 600)")
	width          = flag.Int("width", 0, "Label width in dots (0 = auto from ZPL)")
	height         = flag.Int("height", 0, "Label height in dots (0 = auto from ZPL)")
	format         = flag.String("format", "png", "Output format: png or jpeg")
	quality        = flag.Int("quality", 90, "JPEG quality (1-100)")
	output         = flag.String("o", "", "Output file (default: input.png or stdout if input is stdin)")
	applyLabelHome = flag.Bool("apply-label-home", false, "Apply label home offsets (^LH) - default ignores for cleaner previews")
	showHelp       = flag.Bool("help", false, "Show help")
	showHelpH      = flag.Bool("h", false, "Show help")
)

func usage() {
	fmt.Fprintf(os.Stderr, `zplrender - Convert ZPL files to images

Usage:
  zplrender [options] <input.zpl>
  zplrender [options] < input.zpl
  cat input.zpl | zplrender [options]

Options:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  zplrender label.zpl                     # Creates label.png
  zplrender -dpi 300 label.zpl            # Render at 300 DPI
  zplrender -format jpeg label.zpl        # Output as JPEG
  zplrender -o output.png label.zpl       # Specify output filename
  zplrender -width 812 -height 1218 -     # Read from stdin with explicit size
  cat label.zpl | zplrender -o out.png    # Pipe input
  zplrender -apply-label-home label.zpl   # Include ^LH label home offsets

Supported DPI values: 203 (default), 300, 600
Note: By default, label home offsets (^LH) are ignored for cleaner previews.
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *showHelp || *showHelpH {
		usage()
		os.Exit(0)
	}

	// Validate DPI
	dpiVal := zpl.DPI(*dpi)
	switch dpiVal {
	case zpl.DPI203, zpl.DPI300, zpl.DPI600:
		// Valid
	default:
		fmt.Fprintf(os.Stderr, "Error: Invalid DPI %d. Use 203, 300, or 600.\n", *dpi)
		os.Exit(1)
	}

	// Validate format
	*format = strings.ToLower(*format)
	if *format != "png" && *format != "jpeg" && *format != "jpg" {
		fmt.Fprintf(os.Stderr, "Error: Invalid format %q. Use png or jpeg.\n", *format)
		os.Exit(1)
	}
	if *format == "jpg" {
		*format = "jpeg"
	}

	// Determine input source
	var inputData []byte
	var err error

	args := flag.Args()
	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		// Read from stdin
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Read from file(s)
		for _, inputFile := range args {
			if err := processFile(inputFile, dpiVal); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", inputFile, err)
				os.Exit(1)
			}
		}
		return
	}

	// Process stdin input
	if err := processData(inputData, dpiVal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processFile(inputFile string, dpiVal zpl.DPI) error {
	data, err := os.ReadFile(inputFile) //nolint:gosec // User-provided filename is intended
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Determine output filename
	outputFile := *output
	if outputFile == "" {
		ext := filepath.Ext(inputFile)
		base := strings.TrimSuffix(inputFile, ext)
		outputFile = base + "." + *format
	}

	return processDataToFile(data, outputFile, dpiVal)
}

func processData(data []byte, dpiVal zpl.DPI) error {
	// For stdin, write to stdout or specified output file
	outputFile := *output
	if outputFile == "" {
		// Write to stdout
		return processDataToStdout(data, dpiVal)
	}
	return processDataToFile(data, outputFile, dpiVal)
}

func processDataToFile(data []byte, outputFile string, dpiVal zpl.DPI) (err error) {
	// Parse the ZPL
	label, err := zpl.Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing ZPL: %w", err)
	}

	// Determine dimensions
	w := *width
	h := *height
	if w == 0 {
		w = label.Width()
		if w == 0 {
			w = 812 // Default 4" at 203 DPI
		}
	}

	// Create renderer (ignore margins by default for cleaner previews)
	renderer := render.New(dpiVal).WithSize(w, h).WithIgnoreLabelHome(!*applyLabelHome)

	// Create output file
	f, err := os.Create(outputFile) //nolint:gosec // User-provided filename is intended
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing output file: %w", cerr)
		}
	}()

	// Render
	if *format == "jpeg" {
		img, err := renderer.Render(label)
		if err != nil {
			return fmt.Errorf("rendering: %w", err)
		}
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: *quality}); err != nil {
			return fmt.Errorf("encoding JPEG: %w", err)
		}
	} else {
		if err := renderer.RenderPNG(label, f); err != nil {
			return fmt.Errorf("rendering PNG: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Rendered to %s\n", outputFile)
	return nil
}

func processDataToStdout(data []byte, dpiVal zpl.DPI) error {
	// Parse the ZPL
	label, err := zpl.Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing ZPL: %w", err)
	}

	// Determine dimensions
	w := *width
	h := *height
	if w == 0 {
		w = label.Width()
		if w == 0 {
			w = 812 // Default 4" at 203 DPI
		}
	}

	// Create renderer (ignore margins by default for cleaner previews)
	renderer := render.New(dpiVal).WithSize(w, h).WithIgnoreLabelHome(!*applyLabelHome)

	// Render to stdout
	if *format == "jpeg" {
		img, err := renderer.Render(label)
		if err != nil {
			return fmt.Errorf("rendering: %w", err)
		}
		if err := jpeg.Encode(os.Stdout, img, &jpeg.Options{Quality: *quality}); err != nil {
			return fmt.Errorf("encoding JPEG: %w", err)
		}
	} else {
		if err := renderer.RenderPNG(label, os.Stdout); err != nil {
			return fmt.Errorf("rendering PNG: %w", err)
		}
	}

	return nil
}
