package main

import "testing"

func TestCellStyleRenderClearsLightThemeRender(t *testing.T) {
	t.Parallel()

	style := CellStyle{
		ID:     1,
		Font:   CellFontStyle{Color: "#000000"},
		Fill:   CellFillStyle{Color: darkGridSurfaceRGB.cssColor()},
		Render: CellRenderStyle{TextColor: darkThemeGridTextRGB.cssColor(), TextAdjusted: true},
	}

	style.render(AppearanceThemeLight)
	if style.Render != (CellRenderStyle{}) {
		t.Fatalf("expected light theme render metadata to be empty, got %#v", style.Render)
	}
}

func TestCellStyleRenderDerivesDarkThemeReadableText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		style        CellStyle
		background   string
		wantColor    string
		wantAdjusted bool
	}{
		{
			name:         "black on dark grid",
			style:        CellStyle{Font: CellFontStyle{Color: "#000000"}},
			background:   darkGridSurfaceRGB.cssColor(),
			wantAdjusted: true,
		},
		{
			name:         "dark blue on dark grid",
			style:        CellStyle{Font: CellFontStyle{Color: "#002060"}},
			background:   darkGridSurfaceRGB.cssColor(),
			wantAdjusted: true,
		},
		{
			name: "white on white fill",
			style: CellStyle{
				Font: CellFontStyle{Color: "#FFFFFF"},
				Fill: CellFillStyle{Color: "#FFFFFF"},
			},
			background:   "#FFFFFF",
			wantAdjusted: true,
		},
		{
			name:       "default text on no fill",
			style:      CellStyle{},
			background: darkGridSurfaceRGB.cssColor(),
			wantColor:  darkThemeGridTextRGB.cssColor(),
		},
		{
			name: "default text on light fill",
			style: CellStyle{
				Fill: CellFillStyle{Color: "#FFEEAA"},
			},
			background: "#FFEEAA",
			wantColor:  lightThemeGridTextRGB.cssColor(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			style := test.style
			style.render(AppearanceThemeDark)
			render := style.Render
			if render.TextColor == "" {
				t.Fatalf("expected dark theme render text color")
			}
			if render.TextAdjusted != test.wantAdjusted {
				t.Fatalf("expected textAdjusted %t, got %#v", test.wantAdjusted, render)
			}
			if test.wantColor != "" && render.TextColor != test.wantColor {
				t.Fatalf("expected render text color %q, got %#v", test.wantColor, render)
			}

			textColor := mustParseCSSHexColor(t, render.TextColor)
			background := mustParseCSSHexColor(t, test.background)
			assertContrastAtLeast(t, textColor, background, minimumCellTextContrast)
		})
	}
}
