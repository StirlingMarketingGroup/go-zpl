package zpl

import (
	"io"
)

// Command represents a single ZPL command that can be serialized to ZPL format.
type Command interface {
	// ZPL returns the ZPL representation of this command.
	ZPL() string

	// WriteTo writes the ZPL representation to the given writer.
	WriteTo(w io.Writer) (int64, error)
}

// Orientation represents the rotation of an element.
type Orientation rune

const (
	// OrientationNormal is 0° rotation (normal).
	OrientationNormal Orientation = 'N'
	// OrientationRotated90 is 90° clockwise rotation.
	OrientationRotated90 Orientation = 'R'
	// OrientationRotated180 is 180° rotation (inverted).
	OrientationRotated180 Orientation = 'I'
	// OrientationRotated270 is 270° clockwise rotation.
	OrientationRotated270 Orientation = 'B'
)

// Justification represents text justification in field blocks.
type Justification rune

const (
	// JustifyLeft aligns text to the left.
	JustifyLeft Justification = 'L'
	// JustifyCenter centers the text.
	JustifyCenter Justification = 'C'
	// JustifyRight aligns text to the right.
	JustifyRight Justification = 'R'
	// JustifyJustified stretches text to fill the width.
	JustifyJustified Justification = 'J'
)

// PrintOrientation represents the label print orientation.
type PrintOrientation rune

const (
	// PrintOrientationNormal prints the label normally.
	PrintOrientationNormal PrintOrientation = 'N'
	// PrintOrientationInverted prints the label inverted (180°).
	PrintOrientationInverted PrintOrientation = 'I'
)

// Unit represents a unit of measurement.
type Unit int

const (
	// UnitDots represents raw printer dots.
	UnitDots Unit = iota
	// UnitInches represents inches.
	UnitInches
	// UnitMillimeters represents millimeters.
	UnitMillimeters
	// UnitCentimeters represents centimeters.
	UnitCentimeters
)

// DPI represents the printer resolution in dots per inch.
type DPI int

const (
	// DPI203 is 203 dots per inch (8 dots/mm), the most common resolution.
	DPI203 DPI = 203
	// DPI300 is 300 dots per inch (12 dots/mm).
	DPI300 DPI = 300
	// DPI600 is 600 dots per inch (24 dots/mm).
	DPI600 DPI = 600
)

// ToDots converts a measurement in the given unit to dots at the specified DPI.
func ToDots(value float64, unit Unit, dpi DPI) int {
	switch unit {
	case UnitDots:
		return int(value)
	case UnitInches:
		return int(value * float64(dpi))
	case UnitMillimeters:
		return int(value * float64(dpi) / 25.4)
	case UnitCentimeters:
		return int(value * float64(dpi) / 2.54)
	default:
		return int(value)
	}
}

// LineColor represents the color of lines and shapes.
type LineColor rune

const (
	// LineColorBlack draws in black.
	LineColorBlack LineColor = 'B'
	// LineColorWhite draws in white (erases).
	LineColorWhite LineColor = 'W'
)
