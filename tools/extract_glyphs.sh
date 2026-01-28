#!/bin/bash
# Extract all glyphs from the boxed scan using calculated positions

cd /Users/brianleishman/go-zpl/testdata/scans
mkdir -p glyphs

SCAN="font0_boxed_600dpi.png"

# Grid parameters (measured from scan)
START_X=112
START_Y=62
BOX_W=250
BOX_H=340
SPACING_X=270
SPACING_Y=390

# Characters in row-major order
CHARS=(
  A B C D E F G H
  I J K L M N O P
  Q R S T U V W X
  Y Z a b c d e f
  g h i j k l m n
  o p q r s t u v
  w x y z 0 1 2 3
  4 5 6 7 8 9 exclaim quote
  hash dollar percent ampersand apostrophe lparen rparen asterisk
)

# Extract each glyph
idx=0
for row in {0..8}; do
  for col in {0..7}; do
    if [ $idx -ge ${#CHARS[@]} ]; then
      break 2
    fi

    char="${CHARS[$idx]}"
    x=$((START_X + col * SPACING_X))
    y=$((START_Y + row * SPACING_Y))

    # Extract PNG
    magick "$SCAN" -crop ${BOX_W}x${BOX_H}+${x}+${y} +repage "glyphs/${char}.png"

    # Convert to PBM for potrace (threshold and negate)
    magick "glyphs/${char}.png" -colorspace gray -threshold 70% -negate "glyphs/${char}.pbm"

    # Trace to SVG
    potrace -s "glyphs/${char}.pbm" -o "glyphs/${char}.svg" --flat

    echo "[$row,$col] $char: +${x}+${y}"

    ((idx++))
  done
done

echo ""
echo "Done! Extracted ${idx} glyphs to glyphs/"
ls glyphs/*.svg | wc -l
