# TODO

Comprehensive task list for the go-zpl library.

## Phase 1: Core Foundation

### Project Setup
- [ ] Set up CI/CD (GitHub Actions)
- [ ] Add linting configuration (golangci-lint)
- [ ] Add pre-commit hooks
- [ ] Set up test coverage reporting

### Core Types & Interfaces
- [ ] Define `Label` type with builder pattern
- [ ] Define `Command` interface for all ZPL commands
- [ ] Implement `fmt.Stringer` for ZPL output
- [ ] Implement `io.WriterTo` for efficient streaming
- [ ] Implement `encoding.TextMarshaler` / `TextUnmarshaler`
- [ ] Define units system (dots, inches, mm, cm)
- [ ] Define coordinate system and positioning

### Basic Label Commands
- [ ] `^XA` - Start format
- [ ] `^XZ` - End format
- [ ] `^PW` - Print width
- [ ] `^LL` - Label length
- [ ] `^LH` - Label home
- [ ] `^PO` - Print orientation
- [ ] `^PQ` - Print quantity

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
- [ ] Implement Font 0 (15x12 matrix)
- [ ] Implement remaining fonts (GS, P, Q, R, S, T, U, V)
- [ ] Font scaling algorithm matching Zebra's implementation
- [ ] Font rotation (0°, 90°, 180°, 270°)

### Text Commands
- [ ] `^A` - Scalable/bitmapped font selection
- [ ] `^CF` - Change default font
- [ ] `^FO` - Field origin
- [ ] `^FT` - Field typeset (baseline positioning)
- [ ] `^FD` - Field data
- [ ] `^FS` - Field separator
- [ ] `^FB` - Field block (text wrapping, justification)
- [ ] `^FP` - Field parameter (direction)
- [ ] `^FR` - Field reverse print
- [ ] `^CI` - Character set selection

### Text Features
- [ ] Left/center/right justification
- [ ] Multi-line text wrapping
- [ ] Line spacing control
- [ ] Hanging indent
- [ ] International character support (Code pages)

---

## Phase 3: Graphics

### Basic Shapes
- [ ] `^GB` - Graphic box (rectangle, with rounding)
- [ ] `^GC` - Graphic circle
- [ ] `^GD` - Graphic diagonal line
- [ ] `^GE` - Graphic ellipse
- [ ] `^GS` - Graphic symbol

### Line Styles
- [ ] Solid lines
- [ ] Line thickness control
- [ ] Corner rounding for boxes

### Bitmap Graphics
- [ ] `^GF` - Graphic field (bitmap data)
- [ ] ASCII hex encoding
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
- [ ] `^BY` - Barcode field default (width, ratio, height)
- [ ] Human-readable text positioning
- [ ] Quiet zone handling
- [ ] Check digit calculation and validation

### Code 128
- [ ] `^BC` - Code 128 barcode
- [ ] Auto mode (automatic subset selection)
- [ ] Subset A, B, C explicit selection
- [ ] UCC/EAN-128 mode
- [ ] Check digit calculation

### Code 39
- [ ] `^B3` - Code 39 barcode
- [ ] Standard Code 39
- [ ] Extended Code 39 (full ASCII)
- [ ] Check digit (mod 43)

### UPC/EAN
- [ ] `^BU` - UPC-A
- [ ] `^B9` - UPC-E
- [ ] `^BE` - EAN-13
- [ ] `^B8` - EAN-8
- [ ] Guard bars and human-readable formatting

### Other 1D Barcodes
- [ ] `^BI` - Industrial 2 of 5
- [ ] `^B2` - Interleaved 2 of 5
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
- [ ] `^BQ` - QR Code
- [ ] Model 1 and Model 2 support
- [ ] Error correction levels (L, M, Q, H)
- [ ] Version selection (1-40)
- [ ] Encoding modes (numeric, alphanumeric, byte, kanji)
- [ ] Structured append

### DataMatrix
- [ ] `^BX` - DataMatrix
- [ ] ECC 200 (standard)
- [ ] ECC 000-140 (legacy, if needed)
- [ ] Square and rectangular formats
- [ ] GS1 DataMatrix support

### PDF417
- [ ] `^B7` - PDF417
- [ ] Standard PDF417
- [ ] Truncated PDF417
- [ ] Macro PDF417
- [ ] Row/column configuration
- [ ] Error correction levels

### Other 2D Barcodes
- [ ] `^BD` - MaxiCode
- [ ] `^BO` - Aztec Code
- [ ] `^BF` - MicroPDF417

---

## Phase 6: Parser

### Lexer
- [ ] Tokenize ZPL commands
- [ ] Handle ^ commands
- [ ] Handle ~ commands
- [ ] Handle parameters and data fields
- [ ] Handle comments
- [ ] Error reporting with line/column

### Parser
- [ ] Build AST from tokens
- [ ] Validate command syntax
- [ ] Handle nested/compound commands
- [ ] Preserve formatting for round-trip

### AST
- [ ] Define AST node types
- [ ] Implement visitor pattern
- [ ] Pretty-print AST
- [ ] AST to ZPL serialization

---

## Phase 7: Renderer

### Core Renderer
- [ ] Define `Renderer` type
- [ ] Implement `image.Image` output
- [ ] Support multiple DPI (203, 300, 600)
- [ ] Label size configuration
- [ ] Background color (white by default)
- [ ] Print color (black by default)

### Text Rendering
- [ ] Render built-in fonts at all sizes
- [ ] Font scaling algorithm
- [ ] Text rotation
- [ ] Text alignment in field blocks
- [ ] Reverse print (white on black)

### Graphics Rendering
- [ ] Rectangle/box rendering
- [ ] Circle rendering
- [ ] Ellipse rendering
- [ ] Diagonal line rendering
- [ ] Bitmap image rendering

### Barcode Rendering
- [ ] 1D barcode rendering with proper bar widths
- [ ] Human-readable text below/above barcodes
- [ ] QR code rendering
- [ ] DataMatrix rendering
- [ ] PDF417 rendering

### Output Formats
- [ ] PNG export
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
- [ ] GoDoc for all exported types
- [ ] Package-level documentation
- [ ] Architecture documentation

### Examples
- [ ] Basic label creation
- [ ] Shipping label example
- [ ] Product label with barcode
- [ ] Label with QR code
- [ ] Label with images
- [ ] Complex multi-element label
- [ ] Parsing existing ZPL
- [ ] Rendering to image

### Tutorials
- [ ] Getting started guide
- [ ] Migration from Labelary
- [ ] Common use cases

---

## Known Issues / Research Needed

- [ ] Document exact font bitmap data for all built-in fonts
- [ ] Research Zebra's exact font scaling algorithm
- [ ] Research exact barcode bar width calculations
- [ ] Determine differences between printer firmware versions
- [ ] Document any known deviations from real printer output

---

## Non-Goals (Out of Scope)

- Printer communication/driver functionality
- Print spooling
- Printer status monitoring
- Firmware updates
- EPL (Eltron Programming Language) support
