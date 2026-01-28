package zpl

import (
	"testing"
)

func TestScalableFont(t *testing.T) {
	tests := []struct {
		name    string
		font    *ScalableFont
		wantZPL string
	}{
		{
			name:    "font 0 normal",
			font:    NewScalableFont(Font0, 30, 30),
			wantZPL: "^A0N,30,30",
		},
		{
			name:    "font A rotated 90",
			font:    NewScalableFont(FontA, 50, 40).WithOrientation(OrientationRotated90),
			wantZPL: "^AAR,50,40",
		},
		{
			name:    "font E inverted",
			font:    NewScalableFont(FontE, 100, 80).WithOrientation(OrientationRotated180),
			wantZPL: "^AEI,100,80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.font.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestChangeFont(t *testing.T) {
	tests := []struct {
		name    string
		cf      *ChangeFont
		wantZPL string
	}{
		{
			name:    "change to font 0",
			cf:      NewChangeFont(Font0, 30, 30),
			wantZPL: "^CF0,30,30",
		},
		{
			name:    "change to font B",
			cf:      NewChangeFont(FontB, 20, 15),
			wantZPL: "^CFB,20,15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cf.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestCharacterSet(t *testing.T) {
	tests := []struct {
		name    string
		cs      *CharacterSet
		wantZPL string
	}{
		{
			name:    "USA charset",
			cs:      NewCharacterSet(CharSetUSA),
			wantZPL: "^CI0",
		},
		{
			name:    "UTF-8",
			cs:      NewCharacterSet(CharSetUTF8),
			wantZPL: "^CI28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestFontConstants(t *testing.T) {
	fonts := map[Font]rune{
		FontA: 'A',
		FontB: 'B',
		FontC: 'C',
		FontD: 'D',
		FontE: 'E',
		FontF: 'F',
		FontG: 'G',
		FontH: 'H',
		Font0: '0',
	}

	for font, expected := range fonts {
		if rune(font) != expected {
			t.Errorf("Font %c expected to be %c", font, expected)
		}
	}
}
