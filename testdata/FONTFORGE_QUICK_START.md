# Quick FontForge Workflow

## Open FontForge
```bash
fontforge
```

## Step 1: Create New Font
1. File → New

## Step 2: Set Font Info
1. Element → Font Info
2. Set:
   - Family Name: `ZebraFont0`
   - Weight: `Bold`
3. Click OK

## Step 3: Import Scan as Background
1. Double-click on a glyph slot (e.g., "A")
2. File → Import
3. Select `testdata/scans/font0_600dpi.png`
4. The scan appears as background

## Step 4: Position and Scale
1. Use View → Zoom to see the character you want
2. Element → Transform → Move to position the "A" in the scan over the glyph area

## Step 5: Auto-trace
1. Once positioned, Element → Autotrace
2. FontForge will trace the black areas to vector paths

## Step 6: Clean Up
1. Use the pointer tool to select and delete unwanted paths
2. Element → Simplify → Simplify to reduce points

## Step 7: Repeat
Do this for each character you need.

## Alternative: Potrace Individual Glyphs

If you have individual glyph PNGs:
```bash
# Convert to PBM (potrace input format)
magick glyph_A.png -threshold 50% glyph_A.pbm

# Trace to SVG
potrace -s glyph_A.pbm -o glyph_A.svg

# Import SVG into FontForge
```
