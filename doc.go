// Package zpl provides a native Go library for generating and parsing ZPL
// (Zebra Programming Language) labels for thermal printers.
//
// This is a production-grade, local replacement for services like Labelary.
// Rendering differences from real printers are bugs, not features.
//
// # Basic Usage
//
// Create a label using the fluent builder API:
//
//	label := zpl.NewLabel().
//		SetDPI(zpl.DPI203).
//		SetSize(4, 6, zpl.UnitInches).
//		TextField(50, 50, zpl.Font0, 30, 30, "Hello, World!").
//		Code128(50, 150, "123456789", 100)
//
//	fmt.Println(label)
//
// # Parsing
//
// Parse reads one label. ParseAll returns one Label per printable ^XA...^XZ
// block — setup-only blocks are skipped and a final block missing its ^XZ is
// kept — which is how multi-page ZPL such as USPS APO continuation sheets is
// handled:
//
//	label, err := zpl.Parse(onePage)
//	labels, err := zpl.ParseAll(multiPage)
//	fmt.Println(len(labels))
//
// # Rendering
//
// The [github.com/StirlingMarketingGroup/go-zpl/render] package turns a Label
// into an image:
//
//	err := render.New(zpl.DPI203).RenderPNG(label, w)
//
// # Units and DPI
//
// SetDPI selects the printer resolution used by unit conversion (203, 300, or
// 600). SetSize takes physical units; SetSizeDots takes raw dots. The *At
// helpers (TextFieldAt, Code128At, QRCodeAt, BoxAt) convert positions with
// ToDots:
//
//	label.SetDPI(zpl.DPI203).SetSize(4, 6, zpl.UnitInches)
//	label.TextFieldAt(0.5, 0.5, zpl.UnitInches, zpl.Font0, 30, 30, "Text")
//	dots := zpl.ToDots(1, zpl.UnitInches, zpl.DPI203) // 203
//
// # Images
//
// ImageConverter turns an image.Image into a ^GF graphic field. Configure
// threshold, dithering, and invert, then Add the result to a label:
//
//	gf := zpl.NewImageConverter().WithThreshold(128).Convert(img)
//	label.Add(gf)
//
// # Low-Level Commands
//
// Add and AddAll append commands built by constructors such as NewFieldOrigin,
// NewScalableFont, and NewFieldData. Commands returns a copy of a parsed or
// built label's command list:
//
//	label.Add(zpl.NewFieldOrigin(100, 200)).
//		Add(zpl.NewScalableFont(zpl.FontE, 50, 40).WithOrientation(zpl.OrientationRotated90)).
//		Add(zpl.NewFieldData("Rotated Text"))
//	cmds := label.Commands()
//
// # Command-Line Tools
//
//	go install github.com/StirlingMarketingGroup/go-zpl/cmd/zplrender@latest
//
// zplrender renders ZPL files to PNG or JPEG. zplprint sends ZPL to a USB
// Zebra printer on macOS. libzpl builds a C shared library for FFI.
//
// # Security
//
// When using untrusted input (e.g., user-provided text), use EscapeFieldData
// to prevent ZPL injection:
//
//	label.TextField(x, y, zpl.Font0, 30, 30, zpl.EscapeFieldData(userInput))
//
// # Thread Safety
//
// Label instances are not safe for concurrent use. Create separate labels
// for each goroutine or use appropriate synchronization.
package zpl
