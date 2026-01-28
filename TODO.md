# TODO

Comprehensive task list for the go-zpl library.

## Phase 1: Core Foundation

### Project Setup

- [x] Set up CI/CD (GitHub Actions)
- [x] Add linting configuration (golangci-lint)
- [x] Add pre-commit hooks
- [x] Set up test coverage reporting

### Core Types & Interfaces

- [x] Define `Label` type with builder pattern
- [x] Define `Command` interface for all ZPL commands
- [x] Implement `fmt.Stringer` for ZPL output
- [x] Implement `io.WriterTo` for efficient streaming
- [x] Implement `encoding.TextMarshaler` / `TextUnmarshaler`
- [x] Define units system (dots, inches, mm, cm)
- [x] Define coordinate system and positioning

### Basic Label Commands

- [x] `^XA` - Start format
- [x] `^XZ` - End format
- [x] `^PW` - Print width
- [x] `^LL` - Label length
- [x] `^LH` - Label home
- [x] `^PO` - Print orientation
- [x] `^PQ` - Print quantity

---

## Phase 2: Text & Fonts

### Built-in Font Implementation

- [ ] Research and document exact font bitmaps for all ZPL built-in fonts
- [ ] Implement Font A (9x5 matrix, uppercase only)
- [ ] Implement Font B (11x7 matrix)
- [ ] Implement Font C (18x10 matrix)
- [ ] Implement Font D (18x10 matrix)
- [ ] Implement Font E (28x15 matrix)
- [ ] Implement Font F (26x13 matrix)
- [ ] Implement Font G (60x40 matrix)
- [ ] Implement Font H (21x13 matrix)
- [x] Implement Font 0 (scanned from real Zebra printer output, traced to TTF)
- [ ] Implement remaining fonts (GS, P, Q, R, S, T, U, V)
- [ ] Font scaling algorithm matching Zebra's implementation
- [x] Font rotation (0°, 90°, 180°, 270°) - implemented in renderer

### Text Commands (ZPL Generation)

- [x] `^A` - Scalable/bitmapped font selection
- [x] `^CF` - Change default font
- [x] `^FO` - Field origin
- [x] `^FT` - Field typeset (baseline positioning)
- [x] `^FD` - Field data
- [x] `^FS` - Field separator
- [x] `^FB` - Field block (text wrapping, justification)
- [ ] `^FP` - Field parameter (direction)
- [x] `^FR` - Field reverse print
- [x] `^CI` - Character set selection

### Text Features

- [x] Left/center/right justification (in ^FB)
- [x] Multi-line text wrapping (in ^FB)
- [x] Line spacing control (in ^FB)
- [x] Hanging indent (in ^FB)
- [ ] International character support (Code pages)

---

## Phase 3: Graphics

### Basic Shapes (ZPL Generation)

- [x] `^GB` - Graphic box (rectangle, with rounding)
- [x] `^GC` - Graphic circle
- [x] `^GD` - Graphic diagonal line
- [x] `^GE` - Graphic ellipse
- [x] `^GS` - Graphic symbol

### Line Styles

- [x] Solid lines
- [x] Line thickness control
- [x] Corner rounding for boxes

### Bitmap Graphics

- [x] `^GF` - Graphic field (bitmap data)
- [x] ASCII hex encoding
- [ ] Binary encoding
- [ ] Compressed binary encoding (Z64, LZ77)
- [ ] PNG to ZPL conversion
- [ ] JPEG to ZPL conversion
- [ ] Dithering algorithms (Floyd-Steinberg, ordered, etc.)

### Image Management

- [ ] `~DG` - Download graphics to printer memory
- [ ] `^IL` - Image load from printer memory
- [ ] `^IS` - Image save to printer memory
- [ ] `~DY` - Download objects

---

## Phase 4: Barcodes (1D)

### Barcode Infrastructure

- [x] `^BY` - Barcode field default (width, ratio, height)
- [x] Human-readable text positioning
- [ ] Quiet zone handling
- [ ] Check digit calculation and validation

### Code 128

- [x] `^BC` - Code 128 barcode
- [x] Auto mode (automatic subset selection)
- [x] Subset A, B, C explicit selection
- [x] UCC/EAN-128 mode
- [ ] Check digit calculation

### Code 39

- [x] `^B3` - Code 39 barcode
- [x] Standard Code 39
- [ ] Extended Code 39 (full ASCII)
- [x] Check digit (mod 43)

### UPC/EAN

- [x] `^BU` - UPC-A
- [ ] `^B9` - UPC-E
- [x] `^BE` - EAN-13
- [ ] `^B8` - EAN-8
- [ ] Guard bars and human-readable formatting

### Other 1D Barcodes

- [ ] `^BI` - Industrial 2 of 5
- [x] `^B2` - Interleaved 2 of 5
- [ ] `^BJ` - Standard 2 of 5
- [ ] `^BK` - ANSI Codabar
- [ ] `^BL` - LOGMARS (Code 39 variant)
- [ ] `^BM` - MSI
- [ ] `^BP` - Plessey
- [ ] `^BS` - UPC/EAN extensions
- [ ] `^B1` - Code 11
- [ ] `^B4` - Code 49
- [ ] `^B5` - Planet Code
- [ ] `^BZ` - POSTNET
- [ ] `^BA` - GS1-128 (formerly UCC/EAN-128)

---

## Phase 5: Barcodes (2D)

### QR Code

- [x] `^BQ` - QR Code
- [x] Model 1 and Model 2 support
- [x] Error correction levels (L, M, Q, H)
- [ ] Version selection (1-40)
- [ ] Encoding modes (numeric, alphanumeric, byte, kanji)
- [ ] Structured append

### DataMatrix

- [x] `^BX` - DataMatrix
- [x] ECC 200 (standard)
- [ ] ECC 000-140 (legacy, if needed)
- [x] Square and rectangular formats
- [ ] GS1 DataMatrix support

### PDF417

- [x] `^B7` - PDF417
- [x] Standard PDF417
- [x] Truncated PDF417
- [ ] Macro PDF417
- [x] Row/column configuration
- [x] Error correction levels

### Other 2D Barcodes

- [x] `^BD` - MaxiCode (using ingridhq/maxicode library)
- [ ] `^BO` - Aztec Code
- [ ] `^BF` - MicroPDF417

---

## Phase 6: Parser

### Lexer/Parser (Implemented)

- [x] Tokenize ZPL commands
- [x] Handle ^ commands
- [x] Handle ~ commands
- [x] Handle parameters and data fields
- [x] Handle comments (`^FX`)
- [x] Handle hex escape sequences (`^FH`)
- [ ] Error reporting with line/column

### Parser Features

- [x] Parse ZPL string to Label object (`zpl.Parse()`)
- [x] Support all basic commands (^XA, ^XZ, ^FO, ^FT, ^FD, ^FS, etc.)
- [x] Support font commands (^A0-^A9, ^CF)
- [x] Support graphic commands (^GB, ^GC, ^GD, ^GE, ^GFA)
- [x] Support barcode commands (^BY, ^BC, ^BD)
- [x] Support label configuration (^PO, ^PW, ^LL, ^LH, ^CI)
- [ ] Build formal AST from tokens
- [ ] Validate command syntax
- [ ] Preserve formatting for round-trip

### AST (Future)

- [ ] Define AST node types
- [ ] Implement visitor pattern
- [ ] Pretty-print AST
- [ ] AST to ZPL serialization

---

## Phase 7: Renderer

### Core Renderer

- [x] Define `Renderer` type
- [x] Implement `image.Image` output
- [x] Support multiple DPI (203, 300, 600)
- [x] Label size configuration
- [x] Background color (white by default)
- [x] Print color (black by default)

### Text Rendering

- [x] Render Font 0 (scanned from real Zebra printer, traced to TTF)
- [x] Font scaling algorithm
- [x] Text rotation (0°, 90°, 180°, 270°)
- [ ] Text alignment in field blocks
- [x] Reverse print (white on black) - partial (needs black background fill)
- [x] CJK character support via Noto Sans CJK fallback
- [ ] CJK fallback for rotated text orientations (R, I, B)

### Graphics Rendering

- [x] Rectangle/box rendering
- [x] Circle rendering
- [x] Ellipse rendering
- [x] Diagonal line rendering
- [x] Bitmap image rendering (`^GFA` ASCII hex format)

### Barcode Rendering

- [x] 1D barcode rendering with proper bar widths (Code 128)
- [x] Human-readable text below/above barcodes
- [x] MaxiCode rendering (using ingridhq/maxicode library)
- [ ] QR code rendering
- [ ] DataMatrix rendering
- [ ] PDF417 rendering

### Output Formats

- [x] PNG export
- [ ] JPEG export
- [ ] PDF export (optional)
- [ ] Raw bitmap for printing

---

## Phase 8: Validation

### Syntax Validation

- [ ] Valid command detection
- [ ] Parameter count validation
- [ ] Parameter type validation
- [ ] Parameter range validation

### Semantic Validation

- [ ] Field positioning bounds checking
- [ ] Barcode data validation (valid characters)
- [ ] Label size consistency
- [ ] Required command presence (^XA, ^XZ)

### Warnings

- [ ] Fields extending beyond label bounds
- [ ] Overlapping fields
- [ ] Unused downloaded graphics
- [ ] Deprecated commands

---

## Phase 9: Advanced Features

### Variable Data

- [ ] `^FN` - Field number
- [ ] `^SN` - Serialization
- [ ] Field linking

### RFID Commands (if applicable)

- [ ] Basic RFID tag commands

### Printer Configuration

- [ ] `^MD` - Media darkness
- [ ] `^PR` - Print rate (speed)
- [ ] `^MM` - Print mode
- [ ] `^MN` - Media tracking

### Network/Communication

- [ ] `^NC` - Network connect
- [ ] `~HS` - Host status return
- [ ] `~HI` - Host identification

---

## Phase 10: Testing & Quality

### Unit Tests

- [ ] 100% coverage for all ZPL commands
- [ ] 100% coverage for parser
- [ ] 100% coverage for renderer

### Visual Regression Tests

- [ ] Generate reference images from Labelary
- [ ] Pixel-by-pixel comparison testing
- [ ] Tolerance configuration for anti-aliasing differences
- [ ] Test matrix: all fonts × all sizes × all rotations
- [ ] Test matrix: all barcodes × common configurations

### Integration Tests

- [ ] Round-trip: ZPL → Parse → Generate → Compare
- [ ] Real printer output comparison (if available)
- [ ] Labelary output comparison

### Fuzz Testing

- [ ] Parser fuzz testing
- [ ] Command parameter fuzz testing

### Benchmarks

- [ ] Parser benchmarks
- [ ] Renderer benchmarks
- [ ] Large label benchmarks

---

## Phase 11: Documentation & Examples

### API Documentation

- [x] GoDoc for all exported types
- [ ] Package-level documentation
- [ ] Architecture documentation

### Examples

- [x] Basic label creation
- [x] Shipping label example
- [ ] Product label with barcode
- [ ] Label with QR code
- [ ] Label with images
- [ ] Complex multi-element label
- [x] Parsing existing ZPL (via `zpl.Parse()`)
- [x] Rendering to image

### Tutorials

- [ ] Getting started guide
- [ ] Migration from Labelary
- [ ] Common use cases

---

## Known Issues / Research Needed

- [x] Document exact font bitmap data for all built-in fonts - Font 0 extracted from real printer
- [ ] Research Zebra's exact font scaling algorithm
- [ ] Research exact barcode bar width calculations
- [ ] Determine differences between printer firmware versions
- [ ] Document any known deviations from real printer output
- [ ] FieldReverse (^FR) needs black background fill behind text, not just inverted pixels

---

## Internal Tools (for font development)

These tools are in the `tools/` directory with `//go:build ignore` tags:

- `tools/smart_extract.go` - Extract glyphs from scanned font grid (page 1)
- `tools/smart_extract_page2.go` - Extract glyphs from page 2 scan
- `tools/trim_glyphs.go` - Normalize glyph heights using reference metrics
- `tools/extract_glyphs/main.go` - Original manual glyph extraction tool
- `tools/build_font.py` - FontForge script to build TTF from SVG glyphs

---

## Phase 12: Tools & Web UI

### CLI Tool

- [ ] `zplrender` command-line tool for converting ZPL files to images
- [ ] Accept ZPL file path or stdin input
- [ ] Output formats: PNG, JPEG
- [ ] Options for DPI, label dimensions
- [ ] Batch processing multiple files

### Web UI (WebAssembly)

- [ ] Compile renderer to WebAssembly using TinyGo or standard Go WASM
- [ ] Create simple web interface for ZPL preview
- [ ] Live rendering as user types ZPL
- [ ] Auto-deploy to GitHub Pages via GitHub Actions
- [ ] Reference implementation: [go2ts](https://github.com/StirlingMarketingGroup/go2ts) pattern
- [ ] No server required - fully client-side rendering

---

## Non-Goals (Out of Scope)

- Printer communication/driver functionality
- Print spooling
- Printer status monitoring
- Firmware updates
- EPL (Eltron Programming Language) support
