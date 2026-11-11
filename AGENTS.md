# AGENTS.md

Instructions for AI agents working on this codebase.

## Project Philosophy

**This is a production-grade, local replacement for services like Labelary.**

### The Standard of Correctness

The reference is a real Zebra printer. Rendering that differs from what a Zebra prints is a **bug**. We strive for pixel-perfect accuracy.

- **How to check a rule:** print the ZPL on a Zebra-brand printer, scan the label, and compare the scan with our render. `cmd/zplprint` sends a `.zpl` file raw to the printer.
- **Labelary is not the reference.** It is useful for a first look, and it is wrong in places: it spaces `^TB` lines 1.25 font heights apart, and a Zebra spaces them one font height apart. Never settle a rule on Labelary alone, and do not add its pictures as references.
- **Keep the scan.** Save it beside the fixture as `testdata/visual/<label>/scan.png`. Fixture labels hold made-up names, addresses and tracking numbers, never a real one.
- **A barcode must scan.** A test reads it back with a barcode reader and gets the data the ZPL asked for.
- Write down any known difference from the print and treat it as something to fix, not an acceptable tradeoff.

## Project Overview

This is a native Go library for ZPL (Zebra Programming Language) generation, parsing, and **rendering**. The goal is to provide a pure Go implementation without relying on external services like Labelary.

## Architecture Goals

- **No external service dependencies** - everything runs locally
- **Pixel-perfect rendering** - match real Zebra printer output
- **Idiomatic Go** - implement standard interfaces where appropriate
- **Builder pattern** - ergonomic label construction
- **Modular design** - separate packages for generation, parsing, rendering, and validation
- **Comprehensive ZPL support** - cover all commonly used ZPL commands

## Standard Go Interfaces to Implement

Where appropriate, types should implement:

- `fmt.Stringer` - for ZPL string output
- `io.Reader` / `io.Writer` - for streaming ZPL
- `io.WriterTo` / `io.ReaderFrom` - for efficient I/O
- `encoding.TextMarshaler` / `encoding.TextUnmarshaler` - for text serialization
- `encoding/json.Marshaler` / `encoding/json.Unmarshaler` - for JSON serialization
- `image.Image` - for rendered label output
- `draw.Image` - for label rendering target

## Key ZPL Commands to Support

### Label Control
- `^XA` / `^XZ` - Start/end label format
- `^PW` - Print width
- `^LL` - Label length
- `^LH` - Label home position
- `^PQ` - Print quantity
- `^PO` - Print orientation

### Field Positioning & Data
- `^FO` - Field origin (positioning)
- `^FT` - Field typeset (baseline positioning)
- `^FD` / `^FS` - Field data / field separator
- `^FR` - Field reverse print
- `^FN` - Field number (for variable data)

### Fonts & Text
- `^A` - Scalable/bitmapped fonts (A-Z, 0-9)
- `^CF` - Change default font
- `^FB` - Field block (text wrapping)
- `^FP` - Field parameter (text direction)
- `^CI` - Change international character set

### Barcodes (1D)
- `^BC` - Code 128
- `^B3` - Code 39
- `^BE` - EAN-13
- `^BU` - UPC-A
- `^BI` - Industrial 2 of 5
- `^B2` - Interleaved 2 of 5
- `^BY` - Barcode field default

### Barcodes (2D)
- `^BQ` - QR code
- `^BX` - DataMatrix
- `^B7` - PDF417
- `^BD` - MaxiCode

### Graphics
- `^GB` - Graphic box
- `^GC` - Graphic circle
- `^GD` - Graphic diagonal line
- `^GE` - Graphic ellipse
- `^GF` - Graphic field (bitmap images)
- `^GS` - Graphic symbol

### Image Handling
- `^IL` - Image load
- `^IS` - Image save
- `~DG` - Download graphics
- `~DY` - Download objects

## Code Style

- Follow standard Go conventions (gofmt, golint, go vet)
- Use table-driven tests
- Document all exported types and functions
- Keep the public API minimal and intuitive
- Error messages should be actionable
- No panics in library code; always return errors

## Testing

- Unit tests for all ZPL command generators
- Unit tests for all ZPL command parsers
- **Visual regression tests** comparing rendered output to reference images
- Integration tests that validate output against ZPL spec
- Fuzz tests for parser robustness
- Benchmark tests for performance-critical paths
- Example tests for documentation

## File Structure

```
go-zpl/
├── zpl.go              # Main entry point, Label type
├── command.go          # Base command interface and types
├── commands/           # Individual ZPL command implementations
│   ├── field.go        # ^FO, ^FD, ^FS, ^FT, etc.
│   ├── font.go         # ^A, ^CF, ^FB
│   ├── barcode.go      # All barcode commands
│   ├── graphic.go      # ^GB, ^GC, ^GD, ^GE, ^GF
│   └── label.go        # ^XA, ^XZ, ^PW, ^LL, etc.
├── parser/             # ZPL parsing
│   ├── lexer.go        # Tokenization
│   ├── parser.go       # AST construction
│   └── ast.go          # AST types
├── render/             # ZPL rendering
│   ├── render.go       # Main renderer
│   ├── font.go         # Font rendering (built-in ZPL fonts)
│   ├── barcode.go      # Barcode rendering
│   ├── text.go         # Text layout and rendering
│   └── dpi.go          # DPI handling (203, 300, 600)
├── font/               # Built-in ZPL font definitions
│   ├── font_a.go       # Font A bitmap data
│   ├── font_b.go       # Font B bitmap data
│   └── ...
├── validate/           # ZPL validation
├── testdata/           # Reference images and test fixtures
└── examples/           # Usage examples
```

## Rendering Requirements

### DPI Support
- 203 DPI (8 dots/mm) - most common
- 300 DPI (12 dots/mm)
- 600 DPI (24 dots/mm)

### Font Rendering
All built-in ZPL fonts (A-Z, 0-9) must be implemented with exact bitmap data matching real printers. Font scaling must match Zebra's scaling algorithms.

### Barcode Rendering
All barcodes must be rendered to match the exact output of real Zebra printers. This includes:
- Correct bar widths
- Correct quiet zones
- Correct human-readable text positioning (when enabled)
- Support for all barcode-specific parameters

### Image Output
The renderer should produce `image.Image` output that can be:
- Saved as PNG, JPEG, etc.
- Displayed in GUI applications
- Used for visual comparison testing
