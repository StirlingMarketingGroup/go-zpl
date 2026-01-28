# Font 0 Extraction Guide

This guide walks through extracting Zebra's Font 0 from printer output to create a pixel-perfect TTF file.

## Equipment Needed
- Zebra thermal printer (any model with Font 0)
- Scanner capable of 1200 DPI
- FontForge (free, open source font editor)

## Step 1: Print the Glyph Charts

Print both ZPL files to your Zebra printer:

```bash
# Option 1: Send directly to printer (adjust for your setup)
lp -d zebra testdata/font0_glyph_chart.zpl
lp -d zebra testdata/font0_isolated_glyphs.zpl

# Option 2: Using netcat to network printer
cat testdata/font0_glyph_chart.zpl | nc printer.local 9100
cat testdata/font0_isolated_glyphs.zpl | nc printer.local 9100

# Option 3: Copy to printer share (Windows)
copy testdata\font0_glyph_chart.zpl \\printer\share
```

**Files:**
- `font0_glyph_chart.zpl` - Overview with all characters, reference text
- `font0_isolated_glyphs.zpl` - Individual glyphs in boxes (easier to trace)

## Step 2: Scan the Printed Labels

**Scanner Settings:**
- Resolution: **1200 DPI** (critical for accurate tracing)
- Color mode: **Grayscale** or **Black & White**
- Format: **PNG** or **TIFF** (lossless)

**Tips:**
- Ensure label is flat and aligned with scanner bed
- Use the corner marks on the label for alignment verification
- Scan both labels

Save as:
- `font0_glyph_chart_scan.png`
- `font0_isolated_glyphs_scan.png`

## Step 3: Calculate Scale Factor

The glyphs were printed at 72 dots at 203 DPI printer resolution.

**Physical size of a 72-dot glyph:**
```
72 dots ÷ 203 DPI = 0.355 inches = 9.01 mm
```

**At 1200 DPI scan:**
```
0.355 inches × 1200 DPI = 426 pixels per glyph
```

So each 72pt character should be approximately **426 pixels tall** in your scan.

## Step 4: Set Up FontForge

1. Install FontForge: https://fontforge.org/
2. Create a new font: File → New
3. Set font properties (Element → Font Info):
   - Family Name: `ZebraFont0`
   - Weight: `Bold`
   - Em Size: `1000` (standard)
   - Ascent: ~800
   - Descent: ~200

## Step 5: Import and Trace Glyphs

For each character:

1. **Open the glyph slot**: Double-click on the character in the font view
2. **Import background image**: File → Import → select your scan
3. **Scale the image**:
   - The scan is at 1200 DPI
   - Scale so the character fits properly in the em-square
   - Scale factor: approximately `1000 / 426 = 2.35`
4. **Auto-trace**: Element → Autotrace
   - Or manually trace with the pen tool for more precision
5. **Clean up**: Simplify paths, remove artifacts
6. **Set width**: Measure from scan, scale appropriately

### Recommended Tracing Order

Start with these reference characters to establish metrics:
1. `H` - Establishes cap height and stem width
2. `x` - Establishes x-height
3. `p` or `g` - Establishes descender depth
4. `o` - Establishes curve thickness and roundness
5. `n` - Establishes arch shape

Then continue with full alphabet, numbers, punctuation.

## Step 6: Set Metrics

**Key measurements to capture from scan:**
- Cap height (top of H)
- x-height (top of x)
- Ascender height (top of h, l)
- Descender depth (bottom of p, g, y)
- Stem width (vertical stroke of H)
- Bowl width (curve of o)

**Spacing:**
- Left/right sidebearings for each glyph
- Kerning pairs (AV, AW, To, etc.)

## Step 7: Test and Iterate

1. Generate TTF: File → Generate Fonts → TrueType
2. Replace `render/font0.ttf` with your new font
3. Run comparison:

```bash
# Generate test image with new font
go run ./examples/render/

# Compare against Labelary or real printer output
```

4. Adjust glyphs as needed
5. Repeat until satisfied

## Step 8: Validate at Multiple Sizes

Test the font at various sizes to ensure it scales correctly:
- 12pt, 18pt, 24pt, 36pt, 48pt, 72pt

The font should look consistent at all sizes.

## Tips for Accurate Results

1. **Consistent stroke width**: Font 0 has a consistent bold weight
2. **Condensed proportions**: Characters are narrower than typical fonts
3. **Clean corners**: No rounded corners on stems
4. **Optical adjustments**: Round characters (O, C, G) slightly exceed cap height

## File Locations

After completion, place the font file at:
```
render/font0.ttf
```

The render package will automatically use it.

## Alternative: Semi-Automated Approach

If manual tracing is too tedious, consider:

1. Use Potrace or similar auto-vectorizer on the scan
2. Import vectors into FontForge
3. Clean up and adjust metrics

Tools:
- Potrace: http://potrace.sourceforge.net/
- Vector Magic: https://vectormagic.com/
