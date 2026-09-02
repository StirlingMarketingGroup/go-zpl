// Libzpl is a C shared library that exposes go-zpl rendering to C and other
// languages via FFI.
//
// Build with:
//
//	go build -buildmode=c-shared -o libzpl.so ./cmd/libzpl     # Linux
//	go build -buildmode=c-shared -o libzpl.dylib ./cmd/libzpl  # macOS
//	go build -buildmode=c-shared -o libzpl.dll ./cmd/libzpl    # Windows
//
// This produces both the shared library and a C header file (libzpl.h).
//
// Exported C functions:
//
//   - zpl_render_png — render ZPL to PNG with DPI and size arguments
//   - zpl_render_png_simple — render ZPL to PNG at 203 DPI with auto size
//   - zpl_free — free PNG buffers returned by the render functions
//   - zpl_version — library version string; do not free
//
// Rust users should use the zpl-rs crate instead of calling this library
// directly.
//
//nolint:gocritic // CGO requires separate import "C" block which triggers false dupImport warnings
package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"unsafe"

	zpl "github.com/StirlingMarketingGroup/go-zpl"
	"github.com/StirlingMarketingGroup/go-zpl/render"
)

// Error codes for C API (intentionally ALL_CAPS for C convention)
//
//nolint:revive,stylecheck // C API uses ALL_CAPS constants
const (
	ZPL_OK           = 0
	ZPL_ERR_PARSE    = -1
	ZPL_ERR_RENDER   = -2
	ZPL_ERR_INTERNAL = -3
)

// zpl_render_png renders ZPL data to a PNG image.
//
// Parameters:
//   - zplData: pointer to ZPL string data (does not need to be null-terminated)
//   - zplLen: length of the ZPL data in bytes
//   - dpi: printer DPI (203, 300, or 600)
//   - width: label width in dots (0 = auto-detect from ZPL)
//   - height: label height in dots (0 = auto-detect from ZPL)
//   - pngOut: output pointer that will be set to the PNG data (caller must free with zpl_free)
//   - pngLen: output pointer that will be set to the PNG data length
//
// Returns:
//   - 0 on success
//   - negative error code on failure
//
//export zpl_render_png
func zpl_render_png( //nolint:revive,stylecheck // exported C function uses snake_case
	zplData *C.char,
	zplLen C.int,
	dpi C.int,
	width C.int,
	height C.int,
	pngOut **C.char,
	pngLen *C.int,
) C.int {
	// Convert C data to Go
	goZPL := C.GoStringN(zplData, zplLen)

	// Parse the ZPL
	label, err := zpl.Parse(goZPL)
	if err != nil {
		return ZPL_ERR_PARSE
	}

	// Zero dimensions pass through so Renderer applies its documented fallback.
	renderer := render.New(zpl.DPI(dpi)).WithSize(int(width), int(height)).WithIgnoreLabelHome(true)

	// Render to PNG
	var buf bytes.Buffer
	if renderErr := renderer.RenderPNG(label, &buf); renderErr != nil {
		return ZPL_ERR_RENDER
	}

	// Allocate C memory for output (caller must free with zpl_free)
	pngBytes := buf.Bytes()
	*pngOut = (*C.char)(C.CBytes(pngBytes))
	*pngLen = C.int(len(pngBytes))

	return ZPL_OK
}

// zpl_render_png_simple is a simpler version that auto-detects dimensions.
//
//export zpl_render_png_simple
func zpl_render_png_simple( //nolint:revive,stylecheck // exported C function uses snake_case
	zplData *C.char,
	zplLen C.int,
	pngOut **C.char,
	pngLen *C.int,
) C.int {
	return zpl_render_png(zplData, zplLen, 203, 0, 0, pngOut, pngLen)
}

// zpl_free frees memory allocated by zpl_render_png.
// Must be called on pngOut pointers returned by render functions.
//
//export zpl_free
func zpl_free(ptr *C.char) { //nolint:revive,stylecheck // exported C function uses snake_case
	C.free(unsafe.Pointer(ptr))
}

// Version is the library version - stored as a package-level variable
// so we only allocate it once.
var cVersion = C.CString("1.0.0")

// zpl_version returns the library version string.
// The returned string is statically allocated and must not be freed.
//
//export zpl_version
func zpl_version() *C.char { //nolint:revive,stylecheck // exported C function uses snake_case
	return cVersion
}

// Required for shared library, but unused
func main() {}
