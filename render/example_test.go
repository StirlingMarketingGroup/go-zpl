package render_test

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

// ExampleRenderer_Render converts a label into an image.
func ExampleRenderer_Render() {
	label := zpl.NewLabel().
		SetSizeDots(400, 200).
		TextField(10, 10, zpl.Font0, 30, 30, "Hello")
	img, err := render.New(zpl.DPI203).Render(label)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(img.Bounds())
	// Output:
	// (0,0)-(400,200)
}

// ExampleRenderer_RenderPNG writes a PNG and reports its decoded size.
func ExampleRenderer_RenderPNG() {
	label := zpl.NewLabel().
		SetSizeDots(400, 200).
		TextField(10, 10, zpl.Font0, 30, 30, "Hello")
	var buf bytes.Buffer
	if err := render.New(zpl.DPI203).RenderPNG(label, &buf); err != nil {
		fmt.Println(err)
		return
	}
	cfg, err := png.DecodeConfig(&buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg.Width, cfg.Height)
	// Output:
	// 400 200
}

// ExampleRenderer_WithSize uses the renderer dimensions instead of the label's.
func ExampleRenderer_WithSize() {
	label := zpl.NewLabel().
		SetSizeDots(100, 80).
		TextField(10, 10, zpl.Font0, 20, 20, "Hi")
	img, err := render.New(zpl.DPI203).WithSize(812, 1218).Render(label)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(img.Bounds())
	// Output:
	// (0,0)-(812,1218)
}

// ExampleRenderer_RenderAll renders each printable ^XA...^XZ block.
func ExampleRenderer_RenderAll() {
	src := "^XA^FO10,10^FDOne^FS^XZ\n^XA^FO10,10^FDTwo^FS^XZ"
	labels, err := zpl.ParseAll(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	imgs, err := render.New(zpl.DPI203).WithSize(200, 100).RenderAll(labels)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(imgs))
	for _, img := range imgs {
		fmt.Println(img.Bounds())
	}
	// Output:
	// 2
	// (0,0)-(200,100)
	// (0,0)-(200,100)
}

// ExampleRenderer_WithIgnoreLabelHome shows ^LH offsets vs cleaner previews.
func ExampleRenderer_WithIgnoreLabelHome() {
	label := zpl.NewLabel().
		SetSizeDots(100, 100).
		SetHomeDots(50, 50).
		FilledBox(0, 0, 10, 10)

	defaultImg, err := render.New(zpl.DPI203).Render(label)
	if err != nil {
		fmt.Println(err)
		return
	}
	ignoredImg, err := render.New(zpl.DPI203).WithIgnoreLabelHome(true).Render(label)
	if err != nil {
		fmt.Println(err)
		return
	}

	isBlack := func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		return r == 0 && g == 0 && b == 0
	}
	fmt.Println(isBlack(defaultImg.At(0, 0)), isBlack(defaultImg.At(50, 50)))
	fmt.Println(isBlack(ignoredImg.At(0, 0)), isBlack(ignoredImg.At(50, 50)))
	// Output:
	// false true
	// true false
}
