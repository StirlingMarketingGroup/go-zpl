#!/usr/bin/env fontforge -script
"""
Build a TTF font from trimmed SVG glyphs using FontForge scripting.
No GUI required - run with: fontforge -script build_font.py

All glyphs have uniform height (cap line to descender line = 220px from trim_glyphs.go)
"""

import fontforge
import os

GLYPH_DIR = "testdata/scans/glyphs"
OUTPUT_TTF = "render/zebra_font0.ttf"

# From trim_glyphs.go output:
# Cap line: 35, Baseline: 210, Descender: 274
# Standard height: 239 pixels (descender - cap = 274 - 35)
# Baseline is at y=175 in output frame (baseline - cap = 210 - 35)
GLYPH_HEIGHT = 239  # pixels in source SVGs
BASELINE_Y = 175    # pixels from top in source SVGs (where baseline sits)

# Font metrics (standard em-square)
EM_SIZE = 1000
ASCENDER = int(EM_SIZE * (BASELINE_Y / GLYPH_HEIGHT))  # ~732
DESCENDER = EM_SIZE - ASCENDER  # ~268

# Mapping from filename to Unicode codepoint
GLYPH_MAP = {
    # Uppercase letters
    "upper_A": ord("A"), "upper_B": ord("B"), "upper_C": ord("C"), "upper_D": ord("D"),
    "upper_E": ord("E"), "upper_F": ord("F"), "upper_G": ord("G"), "upper_H": ord("H"),
    "upper_I": ord("I"), "upper_J": ord("J"), "upper_K": ord("K"), "upper_L": ord("L"),
    "upper_M": ord("M"), "upper_N": ord("N"), "upper_O": ord("O"), "upper_P": ord("P"),
    "upper_Q": ord("Q"), "upper_R": ord("R"), "upper_S": ord("S"), "upper_T": ord("T"),
    "upper_U": ord("U"), "upper_V": ord("V"), "upper_W": ord("W"), "upper_X": ord("X"),
    "upper_Y": ord("Y"), "upper_Z": ord("Z"),
    # Lowercase letters
    "lower_a": ord("a"), "lower_b": ord("b"), "lower_c": ord("c"), "lower_d": ord("d"),
    "lower_e": ord("e"), "lower_f": ord("f"), "lower_g": ord("g"), "lower_h": ord("h"),
    "lower_i": ord("i"), "lower_j": ord("j"), "lower_k": ord("k"), "lower_l": ord("l"),
    "lower_m": ord("m"), "lower_n": ord("n"), "lower_o": ord("o"), "lower_p": ord("p"),
    "lower_q": ord("q"), "lower_r": ord("r"), "lower_s": ord("s"), "lower_t": ord("t"),
    "lower_u": ord("u"), "lower_v": ord("v"), "lower_w": ord("w"), "lower_x": ord("x"),
    "lower_y": ord("y"), "lower_z": ord("z"),
    # Digits
    "digit_0": ord("0"), "digit_1": ord("1"), "digit_2": ord("2"), "digit_3": ord("3"),
    "digit_4": ord("4"), "digit_5": ord("5"), "digit_6": ord("6"), "digit_7": ord("7"),
    "digit_8": ord("8"), "digit_9": ord("9"),
    # Special characters - Page 1
    "exclaim": ord("!"), "quote": ord('"'), "hash": ord("#"), "dollar": ord("$"),
    "percent": ord("%"), "ampersand": ord("&"), "apostrophe": ord("'"),
    "lparen": ord("("), "rparen": ord(")"), "asterisk": ord("*"),
    # Special characters - Page 2
    "plus": ord("+"), "comma": ord(","), "minus": ord("-"), "period": ord("."),
    "slash": ord("/"), "colon": ord(":"), "semicolon": ord(";"), "less": ord("<"),
    "equal": ord("="), "greater": ord(">"), "question": ord("?"), "at": ord("@"),
    "lbracket": ord("["), "rbracket": ord("]"), "underscore": ord("_"),
    "backtick": ord("`"), "lbrace": ord("{"), "pipe": ord("|"), "rbrace": ord("}"),
    "cent": 0x00A2,  # ¢ - backslash printed as cent sign
    # Note: ^ (caret) and ~ (tilde) are ZPL control chars and didn't print
}

def main():
    print(f"Building font with em={EM_SIZE}, ascender={ASCENDER}, descender={DESCENDER}")

    # Create a new font
    font = fontforge.font()
    font.fontname = "ZebraFont0"
    font.familyname = "Zebra Font 0"
    font.fullname = "Zebra Font 0 Regular"
    font.copyright = "Extracted from Zebra printer output"
    font.version = "1.0"

    # Set font metrics
    font.em = EM_SIZE
    font.ascent = ASCENDER
    font.descent = DESCENDER

    # First, determine the scale factor and baseline offset by importing reference glyph
    ref_path = os.path.join(GLYPH_DIR, "upper_H.svg")
    ref_glyph = font.createChar(0xFFFF)  # temp
    ref_glyph.importOutlines(ref_path)
    ref_bbox = ref_glyph.boundingBox()
    ref_height = ref_bbox[3] - ref_bbox[1]  # top - bottom
    ref_bottom = ref_bbox[1]  # bottom of H = where baseline should be

    # Scale so that the glyph height matches the ascender (cap height ≈ ascender for this font)
    SCALE = ASCENDER / ref_height if ref_height > 0 else 1.0

    # After scaling, shift so H's bottom (the baseline) is at y=0
    Y_SHIFT = -ref_bottom * SCALE

    font.removeGlyph(ref_glyph)
    print(f"Reference H: height={ref_height:.1f}, bottom={ref_bottom:.1f}")
    print(f"Scale: {SCALE:.4f}, Y shift: {Y_SHIFT:.1f}")

    # Import each glyph
    for name, codepoint in GLYPH_MAP.items():
        svg_path = os.path.join(GLYPH_DIR, f"{name}.svg")
        if not os.path.exists(svg_path):
            print(f"  {name}: SKIP (not found)")
            continue

        # Create glyph at the unicode codepoint
        glyph = font.createChar(codepoint)
        glyph.clear()

        # Import the SVG outline
        glyph.importOutlines(svg_path)

        # Simplify to smooth out chunky traced curves
        # Higher values = smoother but less accurate
        glyph.simplify(3.0)  # More aggressive smoothing
        glyph.round()  # Round to integer coordinates

        # Scale to font units and shift Y to align baseline
        glyph.transform([SCALE, 0, 0, SCALE, 0, Y_SHIFT])

        # Set width based on bounding box
        bbox = glyph.boundingBox()
        if bbox[2] > bbox[0]:
            # width = right edge + small side bearing
            glyph.width = int(bbox[2] + 50)
        else:
            glyph.width = int(EM_SIZE * 0.3)  # default for empty

        print(f"  {name} -> {chr(codepoint)}: width={glyph.width}")

    # Add space character
    space = font.createChar(ord(" "))
    space.width = int(EM_SIZE * 0.3)
    print(f"  space: width={space.width}")

    # Generate the TTF
    font.generate(OUTPUT_TTF)
    print(f"\nGenerated: {OUTPUT_TTF}")

if __name__ == "__main__":
    main()
