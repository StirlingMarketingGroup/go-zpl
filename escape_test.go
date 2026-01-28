package zpl

import "testing"

func TestEscapeFieldData(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no escaping needed", "Hello World", "Hello World"},
		{"caret escape", "Test^Value", "Test_5EValue"},
		{"tilde escape", "Test~Value", "Test_7EValue"},
		{"both characters", "^XA~HS", "_5EXA_7EHS"},
		{"multiple carets", "A^B^C", "A_5EB_5EC"},
		{"empty string", "", ""},
		{"only special chars", "^~^~", "_5E_7E_5E_7E"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeFieldData(tt.input)
			if got != tt.want {
				t.Errorf("EscapeFieldData(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMustEscapeFieldData(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"no special chars", "Hello World", false},
		{"has caret", "Test^Value", true},
		{"has tilde", "Test~Value", true},
		{"has both", "^~", true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustEscapeFieldData(tt.input)
			if got != tt.want {
				t.Errorf("MustEscapeFieldData(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
