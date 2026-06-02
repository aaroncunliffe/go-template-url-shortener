package shortcode

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "default length", length: defaultLength},
		{name: "custom length", length: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				code string
				err  error
			)

			if tt.length == defaultLength {
				code, err = Generate()
			} else {
				code, err = GenerateLength(tt.length)
			}

			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if len(code) != tt.length {
				t.Fatalf("expected length %d, got %d", tt.length, len(code))
			}

			// Check only valid characters are used
			for _, r := range code {
				if !strings.ContainsRune(charset, r) {
					t.Fatalf("unexpected character %q in %q", r, code)
				}
			}
		})
	}
}
