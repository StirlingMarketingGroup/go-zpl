//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
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

	// Optional 5th argument: ignoreLabelHome (default true)
	ignoreLabelHome := true
	if len(args) >= 5 && !args[4].IsUndefined() && !args[4].IsNull() {
		ignoreLabelHome = args[4].Bool()
	}

	// Optional 6th argument: isBase64 (default false)
	if len(args) >= 6 && !args[5].IsUndefined() && !args[5].IsNull() && args[5].Bool() {
		decoded, err := base64.StdEncoding.DecodeString(zplData)
		if err != nil {
			return map[string]interface{}{"error": "base64 decode error: " + err.Error()}
		}
		zplData = string(decoded)
	}

	labels, err := zpl.ParseAll(zplData)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	if len(labels) == 0 {
		return map[string]interface{}{"error": "no labels found in ZPL"}
	}

	renderer := render.New(dpi).WithSize(width, height).WithIgnoreLabelHome(ignoreLabelHome)

	// Render all labels to PNG
	images := make([]interface{}, 0, len(labels))
	for _, label := range labels {
		var buf bytes.Buffer
		if err := renderer.RenderPNG(label, &buf); err != nil {
			return map[string]interface{}{"error": err.Error()}
		}
		images = append(images, base64.StdEncoding.EncodeToString(buf.Bytes()))
	}

	return map[string]interface{}{
		"images": images,
		"width":  width,
		"height": height,
		// Keep backward compatibility - also return first image as "image"
		"image": images[0],
	}
}
