#!/bin/bash
# Extract all glyphs from the font scan

cd /Users/brianleishman/go-zpl/testdata/scans
mkdir -p glyphs

# First, let's get a wider crop that has more characters
# and extract from that to make positioning easier

# The full scan positions:
# Row 1 (ABCDEFGHIJKLM): around y=220-330 in full scan
# Row 2 (NOPQRSTUVWXYZ): around y=340-450 in full scan

# Let's extract character by character from the full scan
# using the positions we've identified

SCAN="font0_600dpi.png"
OUT="glyphs"
H=115  # Height of 50pt chars

# Row 1: ABCDEFGHIJKLM (y≈220 in full scan based on crop at y=150 + 70)
Y=220

# Measure each character position (these need fine-tuning)
magick "$SCAN" -crop 72x${H}+130+${Y} +repage PNG24:${OUT}/A.png
magick "$SCAN" -crop 72x${H}+202+${Y} +repage PNG24:${OUT}/B.png
magick "$SCAN" -crop 72x${H}+274+${Y} +repage PNG24:${OUT}/C.png
magick "$SCAN" -crop 75x${H}+346+${Y} +repage PNG24:${OUT}/D.png
magick "$SCAN" -crop 65x${H}+421+${Y} +repage PNG24:${OUT}/E.png
magick "$SCAN" -crop 60x${H}+486+${Y} +repage PNG24:${OUT}/F.png
magick "$SCAN" -crop 78x${H}+546+${Y} +repage PNG24:${OUT}/G.png
magick "$SCAN" -crop 75x${H}+624+${Y} +repage PNG24:${OUT}/H.png
magick "$SCAN" -crop 32x${H}+699+${Y} +repage PNG24:${OUT}/I.png
magick "$SCAN" -crop 55x${H}+731+${Y} +repage PNG24:${OUT}/J.png
magick "$SCAN" -crop 75x${H}+786+${Y} +repage PNG24:${OUT}/K.png
magick "$SCAN" -crop 62x${H}+861+${Y} +repage PNG24:${OUT}/L.png
magick "$SCAN" -crop 90x${H}+923+${Y} +repage PNG24:${OUT}/M.png

echo "Extracted row 1 (A-M)"

# Row 2: NOPQRSTUVWXYZ (y≈350 in full scan)
Y=350

magick "$SCAN" -crop 75x${H}+130+${Y} +repage PNG24:${OUT}/N.png
magick "$SCAN" -crop 80x${H}+205+${Y} +repage PNG24:${OUT}/O.png
magick "$SCAN" -crop 72x${H}+285+${Y} +repage PNG24:${OUT}/P.png
magick "$SCAN" -crop 80x${H}+357+${Y} +repage PNG24:${OUT}/Q.png
magick "$SCAN" -crop 75x${H}+437+${Y} +repage PNG24:${OUT}/R.png
magick "$SCAN" -crop 70x${H}+512+${Y} +repage PNG24:${OUT}/S.png
magick "$SCAN" -crop 70x${H}+582+${Y} +repage PNG24:${OUT}/T.png
magick "$SCAN" -crop 75x${H}+652+${Y} +repage PNG24:${OUT}/U.png
magick "$SCAN" -crop 80x${H}+727+${Y} +repage PNG24:${OUT}/V.png
magick "$SCAN" -crop 100x${H}+807+${Y} +repage PNG24:${OUT}/W.png
magick "$SCAN" -crop 75x${H}+907+${Y} +repage PNG24:${OUT}/X.png
magick "$SCAN" -crop 75x${H}+982+${Y} +repage PNG24:${OUT}/Y.png
magick "$SCAN" -crop 70x${H}+1057+${Y} +repage PNG24:${OUT}/Z.png

echo "Extracted row 2 (N-Z)"

# Count what we have
echo ""
echo "Extracted files:"
ls -la ${OUT}/*.png | wc -l
