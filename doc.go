// Package zpl provides a native Go library for generating ZPL (Zebra Programming Language)
// commands for thermal label printers.
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
