package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const minimumCellTextContrast = 4.5

type rgbColor struct {
	red   uint8
	green uint8
	blue  uint8
}

var (
	blackRGB              = rgbColor{red: 0x00, green: 0x00, blue: 0x00}
	whiteRGB              = rgbColor{red: 0xFF, green: 0xFF, blue: 0xFF}
	darkGridSurfaceRGB    = rgbColor{red: 0x16, green: 0x16, blue: 0x16}
	darkThemeGridTextRGB  = rgbColor{red: 0xE0, green: 0xE0, blue: 0xE0}
	lightThemeGridTextRGB = rgbColor{red: 0x20, green: 0x21, blue: 0x24}
)

func (r rgbColor) cssColor() string {
	return fmt.Sprintf("#%02X%02X%02X", r.red, r.green, r.blue)
}

func parseCSSHexColor(color string) (rgbColor, error) {
	hexColor := strings.TrimSpace(color)
	if hexColor == "" {
		return rgbColor{}, fmt.Errorf("empty color")
	}

	hexColor = strings.TrimPrefix(hexColor, "#")
	// Accept short CSS hex syntax from authored styles.
	if len(hexColor) == 3 {
		hexColor = strings.Repeat(string(hexColor[0]), 2) +
			strings.Repeat(string(hexColor[1]), 2) +
			strings.Repeat(string(hexColor[2]), 2)
	}
	if len(hexColor) != 6 {
		return rgbColor{}, fmt.Errorf("invalid color length %d", len(hexColor))
	}

	vals, err := strconv.ParseUint(hexColor, 16, 32)
	if err != nil {
		return rgbColor{}, fmt.Errorf("invalid color %q: %w", hexColor, err)
	}
	//nolint:gosec // G115: len(hexColor)==6 bounds parsed RGB channels to uint8.
	red := uint8(vals >> 16)
	green := uint8((vals >> 8) & 0xFF)
	blue := uint8(vals & 0xFF)

	return rgbColor{red: red, green: green, blue: blue}, nil
}

func relativeLuminance(color rgbColor) float64 {
	red := linearizedColorChannel(color.red)
	green := linearizedColorChannel(color.green)
	blue := linearizedColorChannel(color.blue)

	return 0.2126*red + 0.7152*green + 0.0722*blue
}

func linearizedColorChannel(value uint8) float64 {
	channel := float64(value) / 255.0
	if channel <= 0.03928 {
		return channel / 12.92
	}

	return math.Pow((channel+0.055)/1.055, 2.4)
}

func contrastRatio(firstColor rgbColor, secondColor rgbColor) float64 {
	firstLuminance := relativeLuminance(firstColor)
	secondLuminance := relativeLuminance(secondColor)
	lighter := math.Max(firstLuminance, secondLuminance)
	darker := math.Min(firstLuminance, secondLuminance)

	return (lighter + 0.05) / (darker + 0.05)
}

func mixRGBColor(source rgbColor, target rgbColor, amount float64) rgbColor {
	amount = min(max(amount, 0), 1)

	return rgbColor{
		red:   mixColorChannel(source.red, target.red, amount),
		green: mixColorChannel(source.green, target.green, amount),
		blue:  mixColorChannel(source.blue, target.blue, amount),
	}
}

func mixColorChannel(source uint8, target uint8, amount float64) uint8 {
	mixed := float64(source) + (float64(target)-float64(source))*amount

	return uint8(math.Round(mixed))
}

func mixColorToContrast(source rgbColor, target rgbColor, background rgbColor) (rgbColor, bool) {
	if contrastRatio(target, background) < minimumCellTextContrast {
		return rgbColor{}, false
	}

	// Binary search for the smallest visible hue shift.
	low := 0.0
	high := 1.0
	for range 24 {
		midpoint := (low + high) / 2
		mixedColor := mixRGBColor(source, target, midpoint)
		if contrastRatio(mixedColor, background) >= minimumCellTextContrast {
			high = midpoint
			continue
		}

		low = midpoint
	}

	mixedColor := mixRGBColor(source, target, high)
	if contrastRatio(mixedColor, background) < minimumCellTextContrast {
		return target, true
	}

	return mixedColor, true
}
