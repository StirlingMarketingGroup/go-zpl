package zpl

import (
	"strings"
	"testing"
)

func TestParseFieldJustifyAndTextBlock(t *testing.T) {
	for _, field := range []string{"^FO780,280,1", "^FT10,20,1", "^TBR,310,20"} {
		t.Run(field, func(t *testing.T) {
			label, err := Parse("^XA" + field + "^FDX^FS^XZ")
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := label.ZPL(); !strings.Contains(got, field) {
				t.Errorf("parsed label prints %q, want it to keep %s", got, field)
			}
		})
	}
}
