// GlyphPreview renders the Font 0 glyph chart used to calibrate fonts against
// real printer output. It writes glyph_chart_preview.png in the current
// directory.
//
//	go run ./examples/glyph_preview
package main

import (
	"fmt"
	"os"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

func main() {
	// Create preview of the glyph chart
	label := zpl.NewLabel().
		SetSizeDots(812, 1000).
		// Header
		Add(zpl.NewFieldOrigin(20, 20)).
		Add(zpl.NewScalableFont(zpl.Font0, 24, 28)).
		Add(zpl.NewFieldData("FONT 0 GLYPH CHART - 60pt")).
		Add(zpl.NewFieldOrigin(20, 50)).
		Add(zpl.NewScalableFont(zpl.Font0, 18, 20)).
		Add(zpl.NewFieldData("Print on Zebra, scan at 1200 DPI")).
		Add(zpl.NewFieldOrigin(0, 80)).
		Add(zpl.NewGraphicBox(812, 2, 2)).
		// Uppercase Row 1
		Add(zpl.NewFieldOrigin(20, 100)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("A B C D E F G")).
		Add(zpl.NewFieldOrigin(20, 170)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("H I J K L M N")).
		// Uppercase Row 2
		Add(zpl.NewFieldOrigin(20, 240)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("O P Q R S T U")).
		Add(zpl.NewFieldOrigin(20, 310)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("V W X Y Z")).
		Add(zpl.NewFieldOrigin(0, 385)).
		Add(zpl.NewGraphicBox(812, 2, 2)).
		// Lowercase Row 1
		Add(zpl.NewFieldOrigin(20, 405)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("a b c d e f g")).
		Add(zpl.NewFieldOrigin(20, 475)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("h i j k l m n")).
		// Lowercase Row 2
		Add(zpl.NewFieldOrigin(20, 545)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("o p q r s t u")).
		Add(zpl.NewFieldOrigin(20, 615)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("v w x y z")).
		Add(zpl.NewFieldOrigin(0, 690)).
		Add(zpl.NewGraphicBox(812, 2, 2)).
		// Numbers
		Add(zpl.NewFieldOrigin(20, 710)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("0 1 2 3 4 5 6 7 8 9")).
		Add(zpl.NewFieldOrigin(0, 785)).
		Add(zpl.NewGraphicBox(812, 2, 2)).
		// Punctuation
		Add(zpl.NewFieldOrigin(20, 805)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("! \" # $ % ^ ' (")).
		Add(zpl.NewFieldOrigin(20, 875)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData(") * + , - . / :")).
		Add(zpl.NewFieldOrigin(20, 945)).
		Add(zpl.NewScalableFont(zpl.Font0, 60, 60)).
		Add(zpl.NewFieldData("; < = > ? @ [ ]"))

	renderer := render.New(zpl.DPI203)

	f, err := os.Create("glyph_chart_preview.png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := renderer.RenderPNG(label, f); err != nil {
		f.Close()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	f.Close()

	fmt.Println("Preview saved to glyph_chart_preview.png")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Print testdata/font0_glyph_chart.zpl on your Zebra")
	fmt.Println("2. Print testdata/font0_isolated_glyphs.zpl on your Zebra")
	fmt.Println("3. Scan both at 1200 DPI")
	fmt.Println("4. Follow testdata/FONT_EXTRACTION_GUIDE.md")
}
