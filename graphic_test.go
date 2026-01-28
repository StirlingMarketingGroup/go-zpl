package zpl

import (
	"testing"
)

func TestGraphicBox(t *testing.T) {
	tests := []struct {
		name      string
		box       *GraphicBox
		wantZPL   string
	}{
		{
			name:    "basic box",
			box:     NewGraphicBox(100, 50, 2),
			wantZPL: "^GB100,50,2,B,0",
		},
		{
			name:    "box with white color",
			box:     NewGraphicBox(100, 50, 2).WithColor(LineColorWhite),
			wantZPL: "^GB100,50,2,W,0",
		},
		{
			name:    "box with rounded corners",
			box:     NewGraphicBox(100, 50, 2).WithCornerRadius(5),
			wantZPL: "^GB100,50,2,B,5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.box.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestGraphicCircle(t *testing.T) {
	tests := []struct {
		name      string
		circle    *GraphicCircle
		wantZPL   string
	}{
		{
			name:    "basic circle",
			circle:  NewGraphicCircle(100, 3),
			wantZPL: "^GC100,3,B",
		},
		{
			name:    "white circle",
			circle:  NewGraphicCircle(50, 1).WithColor(LineColorWhite),
			wantZPL: "^GC50,1,W",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.circle.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestGraphicDiagonalLine(t *testing.T) {
	tests := []struct {
		name    string
		line    *GraphicDiagonalLine
		wantZPL string
	}{
		{
			name:    "basic diagonal",
			line:    NewGraphicDiagonalLine(100, 100, 2),
			wantZPL: "^GD100,100,2,B,R",
		},
		{
			name:    "left leaning diagonal",
			line:    NewGraphicDiagonalLine(100, 100, 2).WithLeftLeaning(),
			wantZPL: "^GD100,100,2,B,L",
		},
		{
			name:    "white diagonal",
			line:    NewGraphicDiagonalLine(50, 75, 1).WithColor(LineColorWhite),
			wantZPL: "^GD50,75,1,W,R",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.line.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestGraphicEllipse(t *testing.T) {
	tests := []struct {
		name    string
		ellipse *GraphicEllipse
		wantZPL string
	}{
		{
			name:    "basic ellipse",
			ellipse: NewGraphicEllipse(100, 50, 2),
			wantZPL: "^GE100,50,2,B",
		},
		{
			name:    "white ellipse",
			ellipse: NewGraphicEllipse(80, 40, 3).WithColor(LineColorWhite),
			wantZPL: "^GE80,40,3,W",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ellipse.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestGraphicSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbol  *GraphicSymbol
		wantZPL string
	}{
		{
			name:    "registered trademark",
			symbol:  NewGraphicSymbol(SymbolRegisteredTrademark, 50, 50),
			wantZPL: "^GSN,50,50^FDA",
		},
		{
			name:    "copyright rotated",
			symbol:  NewGraphicSymbol(SymbolCopyright, 30, 30).WithOrientation(OrientationRotated90),
			wantZPL: "^GSR,30,30^FDB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.symbol.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}
