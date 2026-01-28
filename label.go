package zpl

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Label represents a ZPL label with all its commands.
// It implements fmt.Stringer, io.WriterTo, and encoding.TextMarshaler.
//
// A Label is not safe for concurrent use.
type Label struct {
	commands []Command

	// Label dimensions in dots
	widthDots  int
	heightDots int

	// Configuration
	dpi              DPI
	homeX, homeY     int
	printOrientation PrintOrientation
	printQuantity    int
	pauseAndCut      int
	replicates       int
	overridePause    bool
}

// NewLabel creates a new empty label with default settings.
func NewLabel() *Label {
	return &Label{
		commands:         make([]Command, 0),
		dpi:              DPI203,
		printOrientation: PrintOrientationNormal,
		printQuantity:    1,
	}
}

// SetDPI sets the printer DPI for unit conversions.
func (l *Label) SetDPI(dpi DPI) *Label {
	l.dpi = dpi
	return l
}

// SetSize sets the label dimensions.
func (l *Label) SetSize(width, height float64, unit Unit) *Label {
	l.widthDots = ToDots(width, unit, l.dpi)
	l.heightDots = ToDots(height, unit, l.dpi)
	return l
}

// SetSizeDots sets the label dimensions in dots.
func (l *Label) SetSizeDots(width, height int) *Label {
	l.widthDots = width
	l.heightDots = height
	return l
}

// SetHome sets the label home position (^LH).
func (l *Label) SetHome(x, y float64, unit Unit) *Label {
	l.homeX = ToDots(x, unit, l.dpi)
	l.homeY = ToDots(y, unit, l.dpi)
	return l
}

// SetHomeDots sets the label home position in dots (^LH).
func (l *Label) SetHomeDots(x, y int) *Label {
	l.homeX = x
	l.homeY = y
	return l
}

// SetPrintOrientation sets the print orientation (^PO).
func (l *Label) SetPrintOrientation(orientation PrintOrientation) *Label {
	l.printOrientation = orientation
	return l
}

// SetPrintQuantity sets the print quantity (^PQ).
// quantity: total number of labels to print
// pauseAndCut: pause and cut after this many labels (0 = no pause)
// replicates: number of replicates of each serial number
// overridePause: override pause count ('Y' or 'N')
func (l *Label) SetPrintQuantity(quantity, pauseAndCut, replicates int, overridePause bool) *Label {
	l.printQuantity = quantity
	l.pauseAndCut = pauseAndCut
	l.replicates = replicates
	l.overridePause = overridePause
	return l
}

// Add appends a command to the label.
// Panics if cmd is nil.
func (l *Label) Add(cmd Command) *Label {
	if cmd == nil {
		panic("zpl: cannot add nil command to label")
	}
	l.commands = append(l.commands, cmd)
	return l
}

// AddAll appends multiple commands to the label.
// Panics if any command is nil.
func (l *Label) AddAll(cmds ...Command) *Label {
	for i, cmd := range cmds {
		if cmd == nil {
			panic(fmt.Sprintf("zpl: cannot add nil command at index %d to label", i))
		}
	}
	l.commands = append(l.commands, cmds...)
	return l
}

// Width returns the label width in dots.
func (l *Label) Width() int {
	return l.widthDots
}

// Height returns the label height in dots.
func (l *Label) Height() int {
	return l.heightDots
}

// DPISetting returns the configured DPI.
func (l *Label) DPISetting() DPI {
	return l.dpi
}

// PrintOrientationSetting returns the configured print orientation.
func (l *Label) PrintOrientationSetting() PrintOrientation {
	return l.printOrientation
}

// Home returns the label home offset in dots (^LH).
func (l *Label) Home() (x, y int) {
	return l.homeX, l.homeY
}

// Commands returns a copy of the label's commands.
func (l *Label) Commands() []Command {
	result := make([]Command, len(l.commands))
	copy(result, l.commands)
	return result
}

// String returns the complete ZPL code for this label.
// Implements fmt.Stringer.
func (l *Label) String() string {
	var buf strings.Builder
	_, _ = l.WriteTo(&buf)
	return buf.String()
}

// ZPL returns the complete ZPL code for this label.
// This is an alias for String() for API consistency.
func (l *Label) ZPL() string {
	return l.String()
}

// WriteTo writes the complete ZPL code to the given writer.
// Implements io.WriterTo.
func (l *Label) WriteTo(w io.Writer) (int64, error) {
	var total int64

	// Start format
	n, err := io.WriteString(w, "^XA\n")
	total += int64(n)
	if err != nil {
		return total, err
	}

	// Print width (^PW)
	if l.widthDots > 0 {
		n, err = fmt.Fprintf(w, "^PW%d\n", l.widthDots)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	// Label length (^LL)
	if l.heightDots > 0 {
		n, err = fmt.Fprintf(w, "^LL%d\n", l.heightDots)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	// Label home (^LH)
	if l.homeX != 0 || l.homeY != 0 {
		n, err = fmt.Fprintf(w, "^LH%d,%d\n", l.homeX, l.homeY)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	// Print orientation (^PO)
	if l.printOrientation != PrintOrientationNormal {
		n, err = fmt.Fprintf(w, "^PO%c\n", l.printOrientation)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	// Write all commands
	for _, cmd := range l.commands {
		written, err := cmd.WriteTo(w)
		total += written
		if err != nil {
			return total, err
		}
		n, err = io.WriteString(w, "\n")
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	// Print quantity (^PQ)
	if l.printQuantity > 1 || l.pauseAndCut > 0 || l.replicates > 0 || l.overridePause {
		override := 'N'
		if l.overridePause {
			override = 'Y'
		}
		n, err = fmt.Fprintf(w, "^PQ%d,%d,%d,%c\n", l.printQuantity, l.pauseAndCut, l.replicates, override)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	// End format
	n, err = io.WriteString(w, "^XZ\n")
	total += int64(n)
	if err != nil {
		return total, err
	}

	return total, nil
}

// MarshalText returns the ZPL code as a byte slice.
// Implements encoding.TextMarshaler.
func (l *Label) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	_, err := l.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Bytes returns the ZPL code as a byte slice.
func (l *Label) Bytes() []byte {
	text, _ := l.MarshalText()
	return text
}
