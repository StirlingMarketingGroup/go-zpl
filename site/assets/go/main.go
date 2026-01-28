//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"syscall/js"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

func main() {
	js.Global().Set("renderZPL", js.FuncOf(renderZPL))
	// Keep the Go program alive
	select {}
}

func renderZPL(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return map[string]interface{}{"error": "expected 4 arguments: zpl, dpi, width, height"}
	}

	zplData := args[0].String()
	dpi := zpl.DPI(args[1].Int())
	width := args[2].Int()
	height := args[3].Int()

	label, err := zpl.Parse(zplData)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	renderer := render.New(dpi).WithSize(width, height)
	img, err := renderer.Render(label)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"image":  base64.StdEncoding.EncodeToString(buf.Bytes()),
		"width":  img.Bounds().Dx(),
		"height": img.Bounds().Dy(),
	}
}
