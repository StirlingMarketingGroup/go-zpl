#!/bin/bash
set -e

cd /Users/brianleishman/go-zpl/testdata/scans
rm -f glyphs/*.png glyphs/*.pbm glyphs/*.svg 2>/dev/null || true
mkdir -p glyphs

SCAN="font0_boxed_600dpi.png"

# Grid parameters - tested to include i/j dots, asterisk, and "1" flag
# Box exterior: 270 wide x 390 tall, starting at ~112,62
# Must start very close to top border to capture asterisk and 1's serif
START_X=132  # 112 + 20
START_Y=63   # 62 + 1 (almost at top border, needed for asterisk)
BOX_W=190    # Narrower to avoid right edge shadow
BOX_H=235    # Adjusted for new start position
SPACING_X=270
SPACING_Y=390

# All characters - using unique names for case-insensitive filesystem
declare -a CHARS=(
  upper_A upper_B upper_C upper_D upper_E upper_F upper_G upper_H
  upper_I upper_J upper_K upper_L upper_M upper_N upper_O upper_P
  upper_Q upper_R upper_S upper_T upper_U upper_V upper_W upper_X
  upper_Y upper_Z lower_a lower_b lower_c lower_d lower_e lower_f
  lower_g lower_h lower_i lower_j lower_k lower_l lower_m lower_n
  lower_o lower_p lower_q lower_r lower_s lower_t lower_u lower_v
  lower_w lower_x lower_y lower_z digit_0 digit_1 digit_2 digit_3
  digit_4 digit_5 digit_6 digit_7 digit_8 digit_9 exclaim quote
  hash dollar percent ampersand apostrophe lparen rparen asterisk
)

idx=0
for row in 0 1 2 3 4 5 6 7 8; do
  for col in 0 1 2 3 4 5 6 7; do
    if [ $idx -ge ${#CHARS[@]} ]; then
      break 2
    fi

    char="${CHARS[$idx]}"
    x=$((START_X + col * SPACING_X))
    y=$((START_Y + row * SPACING_Y))

    # Extract PNG
    magick "$SCAN" -crop ${BOX_W}x${BOX_H}+${x}+${y} +repage "glyphs/${char}.png"

    # Convert to PBM for potrace (black letters on white background)
    # 40% threshold: only very dark pixels become black, eliminates faint border shadows
    magick "glyphs/${char}.png" -colorspace gray -threshold 40% "glyphs/${char}.pbm"
    # --turdsize 20 ignores tiny noise but keeps i/j dots (default is 2)
    potrace -s "glyphs/${char}.pbm" -o "glyphs/${char}.svg" --flat --turdsize 20 2>/dev/null || true

    echo "[$row,$col] $char"
    idx=$((idx + 1))
  done
done

echo ""
echo "Extracted $idx glyphs"
echo "PNGs: $(ls glyphs/*.png 2>/dev/null | wc -l)"
echo "SVGs: $(ls glyphs/*.svg 2>/dev/null | wc -l)"
