package store

import (
	"strings"
	"testing"
)

func TestSlugifyBasicNormalization(t *testing.T) {
	got := slugify("  Hello, World!  ")
	if got != "hello-world" {
		t.Fatalf("slugify() = %q, want hello-world", got)
	}
}

func TestSlugifyFallbackToEmptyForSymbols(t *testing.T) {
	got := slugify("---")
	if got != "" {
		t.Fatalf("slugify() = %q, want empty string", got)
	}
}

func TestSlugifyRespectsMaxLength(t *testing.T) {
	input := strings.Repeat("abcdef", 30)
	got := slugify(input)
	if len(got) > maxSlugLength {
		t.Fatalf("slug length = %d, want <= %d", len(got), maxSlugLength)
	}
}
