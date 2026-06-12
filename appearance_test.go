package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestInitializeAppearanceNormalizesAndRendersStyles(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Styles = []CellStyle{{ID: 1, Font: CellFontStyle{Color: "#000000"}}}

	state := app.InitializeAppearance(AppearanceModeDark, AppearanceThemeLight)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeDark, AppearanceThemeLight, AppearanceThemeDark)
	assertReadableDarkRender(t, findStyle(t, state.Workbook.Styles, 1), darkGridSurfaceRGB.cssColor())

	state = app.InitializeAppearance(AppearanceMode("sepia"), AppearanceTheme("sepia"))
	assertReadyStatus(t, state.Status)
	assertDefaultAppearance(t, state.Appearance)
	if style := findStyle(t, state.Workbook.Styles, 1); style.Render != (CellRenderStyle{}) {
		t.Fatalf("expected invalid initialization fallback to clear light-theme render metadata, got %#v", style.Render)
	}
}

func TestAppearanceCommandsResolveThemes(t *testing.T) {
	t.Parallel()

	app := NewApp()

	state := app.InitializeAppearance(AppearanceModeSystem, AppearanceThemeDark)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeSystem, AppearanceThemeDark, AppearanceThemeDark)

	state = app.SetAppearanceMode(AppearanceModeLight)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeLight, AppearanceThemeDark, AppearanceThemeLight)

	state = app.SetAppearanceMode(AppearanceModeDark)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeDark, AppearanceThemeDark, AppearanceThemeDark)

	state = app.SetSystemTheme(AppearanceThemeLight)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeDark, AppearanceThemeLight, AppearanceThemeDark)

	state = app.SetAppearanceMode(AppearanceModeSystem)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeSystem, AppearanceThemeLight, AppearanceThemeLight)

	state = app.SetSystemTheme(AppearanceThemeDark)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeSystem, AppearanceThemeDark, AppearanceThemeDark)
}

func TestInvalidAppearanceCommandsSetErrorAndPreserveState(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Sheets[0].Cells = []CellData{{Ref: "B2", Row: 2, Column: 2, Value: "Original"}}
	app.InitializeAppearance(AppearanceModeDark, AppearanceThemeLight)

	before := app.State()
	invalidMode := app.SetAppearanceMode(AppearanceMode("sepia"))
	assertAppearanceErrorPreservesState(t, invalidMode, before, "Appearance mode")

	before = app.State()
	invalidTheme := app.SetSystemTheme(AppearanceTheme("sepia"))
	assertAppearanceErrorPreservesState(t, invalidTheme, before, "System theme")
}

func TestAppearanceCommandsRecomputeWorkbookRenderStyles(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Styles = []CellStyle{{ID: 1, Font: CellFontStyle{Color: "#000000"}}}

	dark := app.SetAppearanceMode(AppearanceModeDark)
	assertReadyStatus(t, dark.Status)
	assertReadableDarkRender(t, findStyle(t, dark.Workbook.Styles, 1), darkGridSurfaceRGB.cssColor())

	light := app.SetAppearanceMode(AppearanceModeLight)
	assertReadyStatus(t, light.Status)
	if style := findStyle(t, light.Workbook.Styles, 1); style.Render != (CellRenderStyle{}) {
		t.Fatalf("expected light mode to clear render metadata, got %#v", style.Render)
	}

	systemLight := app.SetAppearanceMode(AppearanceModeSystem)
	assertReadyStatus(t, systemLight.Status)
	if style := findStyle(t, systemLight.Workbook.Styles, 1); style.Render != (CellRenderStyle{}) {
		t.Fatalf("expected system light theme to leave render metadata empty, got %#v", style.Render)
	}

	systemDark := app.SetSystemTheme(AppearanceThemeDark)
	assertReadyStatus(t, systemDark.Status)
	assertReadableDarkRender(t, findStyle(t, systemDark.Workbook.Styles, 1), darkGridSurfaceRGB.cssColor())
}

func TestOpenWorkbookPathAppliesCurrentAppearanceRenderStyles(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	app.InitializeAppearance(AppearanceModeDark, AppearanceThemeLight)

	state := app.OpenWorkbookPath(path)
	assertReadyStatus(t, state.Status)
	assertAppearance(t, state.Appearance, AppearanceModeDark, AppearanceThemeLight, AppearanceThemeDark)
	dataSheet := assertLoadedSheetsAndView(t, state)
	styleID := assertLoadedCells(t, dataSheet)

	assertStyleMissing(t, state.Workbook.Styles, 0)

	loadedStyle := findStyle(t, state.Workbook.Styles, styleID)
	assertReadableDarkRender(t, loadedStyle, loadedStyle.Fill.Color)

	light := app.SetAppearanceMode(AppearanceModeLight)
	assertReadyStatus(t, light.Status)
	if style := findStyle(t, light.Workbook.Styles, styleID); style.Render != (CellRenderStyle{}) {
		t.Fatalf("expected light mode to clear loaded workbook render metadata, got %#v", style.Render)
	}
}

func assertAppearance(
	t *testing.T,
	appearance AppearanceState,
	mode AppearanceMode,
	systemTheme AppearanceTheme,
	effectiveTheme AppearanceTheme,
) {
	t.Helper()

	expected := AppearanceState{Mode: mode, SystemTheme: systemTheme, EffectiveTheme: effectiveTheme}
	if appearance != expected {
		t.Fatalf("expected appearance %#v, got %#v", expected, appearance)
	}
}

func assertAppearanceErrorPreservesState(t *testing.T, state AppState, before AppState, messagePart string) {
	t.Helper()

	if state.Status.Kind != AppStatusKindError || state.Status.Busy {
		t.Fatalf("expected appearance command error status without busy flag, got %#v", state.Status)
	}
	if !strings.Contains(state.Status.Message, messagePart) {
		t.Fatalf("expected status message containing %q, got %q", messagePart, state.Status.Message)
	}
	if state.Appearance != before.Appearance {
		t.Fatalf(
			"expected invalid appearance command to preserve appearance, before %#v after %#v",
			before.Appearance,
			state.Appearance,
		)
	}
	if !reflect.DeepEqual(state.Workbook, before.Workbook) {
		t.Fatalf(
			"expected invalid appearance command to preserve workbook\nbefore: %#v\nafter:  %#v",
			before.Workbook,
			state.Workbook,
		)
	}
	if state.View != before.View {
		t.Fatalf("expected invalid appearance command to preserve view, before %#v after %#v", before.View, state.View)
	}
}

func assertReadableDarkRender(t *testing.T, style CellStyle, backgroundColor string) {
	t.Helper()

	if style.Render.TextColor == "" {
		t.Fatalf("expected style %d to have dark-theme render text color", style.ID)
	}

	textColor := mustParseCSSHexColor(t, style.Render.TextColor)
	background := mustParseCSSHexColor(t, backgroundColor)
	assertContrastAtLeast(t, textColor, background, minimumCellTextContrast)
}
