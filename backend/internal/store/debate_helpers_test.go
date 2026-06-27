package store

import (
	"regexp"
	"testing"
	"time"
)

func TestTurnWindowByMode(t *testing.T) {
	tests := []struct {
		mode string
		want time.Duration
	}{
		{mode: "marathon", want: 7 * 24 * time.Hour},
		{mode: "standard", want: 48 * time.Hour},
		{mode: "rapid", want: 12 * time.Hour},
		{mode: "blitz", want: 2 * time.Hour},
		{mode: "unknown", want: 48 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			got := turnWindow(tt.mode)
			if got != tt.want {
				t.Fatalf("turnWindow(%q) = %s, want %s", tt.mode, got, tt.want)
			}
		})
	}
}

func TestOtherSide(t *testing.T) {
	if got := otherSide("a"); got != "b" {
		t.Fatalf("otherSide(a) = %s, want b", got)
	}
	if got := otherSide("b"); got != "a" {
		t.Fatalf("otherSide(b) = %s, want a", got)
	}
}

func TestGenerateAnonymousIDFormat(t *testing.T) {
	pattern := regexp.MustCompile(`^A#[0-9]{4}$`)
	for i := 0; i < 10; i++ {
		id, err := generateAnonymousID("A")
		if err != nil {
			t.Fatalf("generateAnonymousID() error = %v", err)
		}
		if !pattern.MatchString(id) {
			t.Fatalf("anonymous id = %q, want pattern %s", id, pattern.String())
		}
	}
}
