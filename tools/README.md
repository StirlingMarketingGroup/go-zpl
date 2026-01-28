# Font Extraction Pipeline

Extract pixel-perfect fonts from Zebra printer output - no GUI required.

## Overview

This pipeline extracts actual Zebra printer glyphs and builds a TrueType font:

1. **Print** boxed glyphs on a Zebra printer
2. **Scan** at 600 DPI
3. **Extract** glyphs using automatic grid detection
4. **Trim** to uniform height with proper baseline positioning
5. **Build** TTF font from SVG outlines

## Prerequisites

- Zebra thermal printer (tested with ZP 450)
- Scanner capable of 600 DPI
- ImageMagick (`brew install imagemagick`)
- potrace (`brew install potrace`)
- FontForge (`brew install fontforge`)
- Go 1.21+

## Step 1: Print Glyph Sheets

Print the ZPL files containing boxed glyphs:

```bash
# Page 1: A-Z, a-z, 0-9, basic punctuation
lp -d YOUR_ZEBRA_PRINTER -o raw testdata/font0_boxed_glyphs.zpl

# Page 2: Additional ASCII characters
lp -d YOUR_ZEBRA_PRINTER -o raw testdata/font0_boxed_glyphs_page2.zpl
```

Each character is printed in a bordered box for easy extraction.

## Step 2: Scan

Scan the printed labels at **600 DPI** and save as PNG:

```
testdata/scans/font0_boxed_600dpi.png      # Page 1
testdata/scans/font0_boxed_page2_600dpi.png # Page 2
```

## Step 3: Extract Glyphs

The smart extraction tool automatically detects the grid borders:

```bash
go run tools/smart_extract.go
```

**How it works:**
- Scans horizontally and vertically for dark bands (>70% dark pixels = border line)
- Identifies the grid of boxes from detected lines
- Extracts the interior of each box (inside the borders)
- Converts to PBM and traces to SVG using potrace

**Output:** `testdata/scans/glyphs/*.png`, `*.pbm`, `*.svg`

## Step 4: Trim Glyphs

Normalize all glyphs to uniform height for proper font metrics:

```bash
go run tools/trim_glyphs.go
```

**How it works:**
- Uses 'H' to find cap height and baseline
- Uses 'g' to find descender line
- Trims all glyphs to the same height (cap → descender)
- Preserves horizontal position relative to baseline
- Regenerates SVG files

**Font metrics detected:**
```
Cap line:    Top of uppercase letters
Baseline:    Bottom of 'H' (where most letters sit)
Descender:   Bottom of 'g', 'j', 'p', 'q', 'y'
```

## Step 5: Build Font

Generate TrueType font from SVG outlines:

```bash
fontforge -script tools/build_font.py
```

**Output:** `render/zebra_font0.ttf`

The font is automatically embedded in the Go render package.

## Adding More Characters

### Creating New ZPL Sheets

Use the same box format (90x120 dots, 2-dot border):

```zpl
^XA
^PW800
^LL1200

^FX --- Row: characters ---
^FO10,10^GB90,120,2^FS      ^FX Box at col 0
^FO100,10^GB90,120,2^FS     ^FX Box at col 1
...

^FO25,25^A0N,80,80^FDX^FS   ^FX Character in box
...
^XZ
```

### Updating Extraction Tools

1. Add new character names to `smart_extract.go`:
```go
var chars = []string{
    // existing...
    "plus", "comma", "minus", // new characters
}
```

2. Add Unicode mappings to `build_font.py`:
```python
GLYPH_MAP = {
    # existing...
    "plus": ord("+"),
    "comma": ord(","),
}
```

## File Structure

```
tools/
├── smart_extract.go    # Grid detection and glyph extraction
├── trim_glyphs.go      # Baseline normalization
├── build_font.py       # FontForge TTF builder
├── extract_all.sh      # Legacy manual extraction (deprecated)
└── README.md           # This file

testdata/
├── font0_boxed_glyphs.zpl       # Page 1 ZPL
├── font0_boxed_glyphs_page2.zpl # Page 2 ZPL
└── scans/
    ├── font0_boxed_600dpi.png   # Scanned page 1
    └── glyphs/                   # Extracted glyphs
        ├── upper_A.png
        ├── upper_A.svg
        └── ...

render/
└── zebra_font0.ttf              # Generated font
```

## Troubleshooting

### Grid detection fails
- Ensure scan is at 600 DPI
- Check that box borders are clearly visible
- Adjust `darkThreshold` in smart_extract.go if needed

### Glyphs cut off
- The trim tool uses 'H' and 'g' as references
- Verify these glyphs printed correctly
- Check baseline metrics in trim_glyphs.go output

### Font renders incorrectly
- Verify SVG files look correct (view in browser)
- Check scale factor in build_font.py matches glyph height
- Ensure ascender/descender values match your metrics
