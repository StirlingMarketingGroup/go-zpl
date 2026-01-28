package zpl

import (
	"testing"
)

func TestToDots(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		unit  Unit
		dpi   DPI
		want  int
	}{
		{"1 inch at 203 DPI", 1.0, UnitInches, DPI203, 203},
		{"1 inch at 300 DPI", 1.0, UnitInches, DPI300, 300},
		{"1 inch at 600 DPI", 1.0, UnitInches, DPI600, 600},
		{"25.4 mm at 203 DPI", 25.4, UnitMillimeters, DPI203, 203},
		{"2.54 cm at 203 DPI", 2.54, UnitCentimeters, DPI203, 203},
		{"100 dots", 100, UnitDots, DPI203, 100},
		{"4 inches at 203 DPI", 4.0, UnitInches, DPI203, 812},
		{"6 inches at 203 DPI", 6.0, UnitInches, DPI203, 1218},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToDots(tt.value, tt.unit, tt.dpi)
			if got != tt.want {
				t.Errorf("ToDots(%v, %v, %v) = %d, want %d", tt.value, tt.unit, tt.dpi, got, tt.want)
			}
		})
	}
}

func TestOrientationConstants(t *testing.T) {
	tests := []struct {
		orientation Orientation
		want        rune
	}{
		{OrientationNormal, 'N'},
		{OrientationRotated90, 'R'},
		{OrientationRotated180, 'I'},
		{OrientationRotated270, 'B'},
	}

	for _, tt := range tests {
		if rune(tt.orientation) != tt.want {
			t.Errorf("Orientation %v = %c, want %c", tt.orientation, tt.orientation, tt.want)
		}
	}
}

func TestJustificationConstants(t *testing.T) {
	tests := []struct {
		justification Justification
		want          rune
	}{
		{JustifyLeft, 'L'},
		{JustifyCenter, 'C'},
		{JustifyRight, 'R'},
		{JustifyJustified, 'J'},
	}

	for _, tt := range tests {
		if rune(tt.justification) != tt.want {
			t.Errorf("Justification %v = %c, want %c", tt.justification, tt.justification, tt.want)
		}
	}
}

func TestLineColorConstants(t *testing.T) {
	tests := []struct {
		color LineColor
		want  rune
	}{
		{LineColorBlack, 'B'},
		{LineColorWhite, 'W'},
	}

	for _, tt := range tests {
		if rune(tt.color) != tt.want {
			t.Errorf("LineColor %v = %c, want %c", tt.color, tt.color, tt.want)
		}
	}
}

func TestDPIConstants(t *testing.T) {
	tests := []struct {
		dpi  DPI
		want int
	}{
		{DPI203, 203},
		{DPI300, 300},
		{DPI600, 600},
	}

	for _, tt := range tests {
		if int(tt.dpi) != tt.want {
			t.Errorf("DPI %v = %d, want %d", tt.dpi, tt.dpi, tt.want)
		}
	}
}
