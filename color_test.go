package main

import (
	"math"
	"testing"
)

func TestParseCSSHexColor(t *testing.T) {
	t.Parallel()

	color, err := parseCSSHexColor(" #369 ")
	if err != nil {
		t.Fatalf("expected shorthand color to parse: %v", err)
	}
	if formatted := color.cssColor(); formatted != "#336699" {
		t.Fatalf("expected shorthand color to expand, got %q", formatted)
	}

	color, err = parseCSSHexColor("#E0E0E0")
	if err != nil {
		t.Fatalf("expected full hex color to parse: %v", err)
	}
	if formatted := color.cssColor(); formatted != "#E0E0E0" {
		t.Fatalf("expected full hex color to round-trip, got %q", formatted)
	}

	if _, err = parseCSSHexColor("#GGGGGG"); err == nil {
		t.Fatalf("expected invalid hex color to fail")
	}
	if _, err = parseCSSHexColor("#FFFFFFFF"); err == nil {
		t.Fatalf("expected alpha hex color to fail")
	}
}

func TestContrastHelpers(t *testing.T) {
	t.Parallel()

	black := mustParseCSSHexColor(t, "#000000")
	white := mustParseCSSHexColor(t, "#FFFFFF")
	if math.Abs(contrastRatio(black, white)-21) > 0.000001 {
		t.Fatalf("expected black/white contrast to be 21, got %.6f", contrastRatio(black, white))
	}

	mixed := mixRGBColor(black, white, 0.5)
	if formatted := mixed.cssColor(); formatted != "#808080" {
		t.Fatalf("expected midpoint mix to be #808080, got %q", formatted)
	}
}

func mustParseCSSHexColor(t *testing.T, value string) rgbColor {
	t.Helper()

	color, err := parseCSSHexColor(value)
	if err != nil {
		t.Fatalf("expected %q to parse as a CSS hex color: %v", value, err)
	}

	return color
}

func assertContrastAtLeast(t *testing.T, textColor rgbColor, background rgbColor, minimum float64) {
	t.Helper()

	contrast := contrastRatio(textColor, background)
	if contrast < minimum {
		t.Fatalf("expected contrast >= %.2f, got %.6f", minimum, contrast)
	}
}
