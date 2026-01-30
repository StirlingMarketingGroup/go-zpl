package pdf417

import "testing"

func compareIntSlice(t *testing.T, expected, actual []int, testStr string) {
	if len(actual) != len(expected) {
		t.Errorf("Invalid slice size. Expected %d got %d while encoding %q", len(expected), len(actual), testStr)
		return
	}
	for i, a := range actual {
		if e := expected[i]; e != a {
			t.Errorf("Unexpected value at position %d. Expected %d got %d while encoding %q", i, e, a, testStr)
		}
	}
}

func TestHighlevelEncode(t *testing.T) {
	runTest := func(msg string, expected ...int) {
		if codes, err := highlevelEncode(msg); err != nil {
			t.Error(err)
		} else {
			compareIntSlice(t, expected, codes, msg)
		}
	}

	runTest("01234", 902, 112, 434)
	runTest("Super !", 567, 615, 137, 809, 329)
	runTest("Super ", 567, 615, 137, 809)
	runTest("ABC123", 1, 88, 32, 119)
	runTest("123ABC", 841, 63, 840, 32)
}

func TestBinaryEncoder(t *testing.T) {
	runTest := func(msg string, expected ...int) {
		codes := encodeBinary([]byte(msg), encText)
		compareIntSlice(t, expected, codes, msg)
	}

	runTest("alcool", 924, 163, 238, 432, 766, 244)
	runTest("alcoolique", 901, 163, 238, 432, 766, 244, 105, 113, 117, 101)
}

func TestControlCharacterEncoding(t *testing.T) {
	testCases := []struct {
		name string
		data string
	}{
		{"03_text_with_RS", "Hello\x1EWorld"},
		{"04_text_with_GS", "Hello\x1DWorld"},
		{"05_fedex_header", "[)>\x1E01"},
		{"06_fedex_header_gs", "[)>\x1E01\x1D"},
		{"07_fedex_first_segment", "[)>\x1E01\x1D0294105"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codewords, err := highlevelEncode(tc.data)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}
			t.Logf("Input: %q (%d bytes)", tc.data, len(tc.data))
			t.Logf("Codewords (%d): %v", len(codewords), codewords)

			// Decode mode switches
			for i, cw := range codewords {
				switch cw {
				case 900:
					t.Logf("  [%d] = 900 (latch to text)", i)
				case 901:
					t.Logf("  [%d] = 901 (latch to byte padded)", i)
				case 902:
					t.Logf("  [%d] = 902 (latch to numeric)", i)
				case 913:
					t.Logf("  [%d] = 913 (shift to byte)", i)
				case 924:
					t.Logf("  [%d] = 924 (latch to byte)", i)
				}
			}
		})
	}
}
