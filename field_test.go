package zpl

import (
	"testing"
)

func TestFieldOrigin(t *testing.T) {
	tests := []struct {
		name    string
		field   *FieldOrigin
		wantZPL string
	}{
		{
			name:    "basic origin",
			field:   NewFieldOrigin(50, 100),
			wantZPL: "^FO50,100",
		},
		{
			name:    "origin with justification",
			field:   NewFieldOrigin(50, 100).WithJustification(JustifyRight),
			wantZPL: "^FO50,100,R",
		},
		{
			name:    "origin at zero",
			field:   NewFieldOrigin(0, 0),
			wantZPL: "^FO0,0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestFieldData(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantZPL string
	}{
		{"simple text", "Hello", "^FDHello^FS"},
		{"with numbers", "Order #12345", "^FDOrder #12345^FS"},
		{"empty", "", "^FD^FS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewFieldData(tt.data)
			if got := field.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestFieldTypeset(t *testing.T) {
	tests := []struct {
		name    string
		field   *FieldTypeset
		wantZPL string
	}{
		{
			name:    "basic typeset",
			field:   NewFieldTypeset(100, 200),
			wantZPL: "^FT100,200",
		},
		{
			name:    "typeset with justification",
			field:   NewFieldTypeset(100, 200).WithJustification(JustifyCenter),
			wantZPL: "^FT100,200,C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestFieldReverse(t *testing.T) {
	fr := NewFieldReverse()
	if got := fr.ZPL(); got != "^FR" {
		t.Errorf("ZPL() = %s, want ^FR", got)
	}
}

func TestFieldBlock(t *testing.T) {
	tests := []struct {
		name    string
		block   *FieldBlock
		wantZPL string
	}{
		{
			name:    "basic block",
			block:   NewFieldBlock(400),
			wantZPL: "^FB400,1,0,L,0",
		},
		{
			name:    "multi-line centered",
			block:   NewFieldBlock(300).WithMaxLines(5).WithJustification(JustifyCenter),
			wantZPL: "^FB300,5,0,C,0",
		},
		{
			name:    "with line spacing and indent",
			block:   NewFieldBlock(500).WithMaxLines(3).WithLineSpacing(5).WithHangingIndent(20),
			wantZPL: "^FB500,3,5,L,20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.block.ZPL(); got != tt.wantZPL {
				t.Errorf("ZPL() = %s, want %s", got, tt.wantZPL)
			}
		})
	}
}

func TestComment(t *testing.T) {
	comment := NewComment("This is a test comment")
	expected := "^FX This is a test comment"
	if got := comment.ZPL(); got != expected {
		t.Errorf("ZPL() = %s, want %s", got, expected)
	}
}
