package main

import (
	"context"
	"reflect"
	"testing"
)

func TestNewAppStartsWithNeutralState(t *testing.T) {
	t.Parallel()

	app := NewApp()
	assertNeutralAppState(t, app.State())

	var startupApp App
	startupApp.startup(context.Background())
	assertNeutralAppState(t, startupApp.State())
}

func TestStateReturnsSnapshotCopy(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Appearance = AppearanceState{
		Mode:           AppearanceModeDark,
		SystemTheme:    AppearanceThemeLight,
		EffectiveTheme: AppearanceThemeDark,
	}
	app.state.Workbook.Sheets[0].Cells = []CellData{{Ref: "B2", Row: 2, Column: 2, Value: "Original"}}
	app.state.Workbook.Sheets[0].MergedCells = []MergedCellRange{{Range: a1Range(), Value: "Merged"}}
	app.state.Workbook.Sheets[0].Columns = []ColumnLayout{{Index: 2, Name: "B", Width: 12.5}}
	app.state.Workbook.Sheets[0].Rows = []RowLayout{{Index: 2, Height: 20}}
	app.state.Workbook.Styles = []CellStyle{
		{
			ID:     1,
			Fill:   CellFillStyle{Colors: []string{"#FFFFFF"}},
			Render: CellRenderStyle{TextColor: "#E0E0E0", TextAdjusted: true},
			Borders: []CellBorderStyle{
				{Side: "left", Style: 1, Color: "#111111"},
			},
		},
	}

	state := app.State()
	state.Appearance.Mode = AppearanceModeLight
	state.Workbook.Sheets[0].Name = "Mutated"
	state.Workbook.Sheets[0].Cells[0].Value = "Mutated"
	state.Workbook.Sheets[0].MergedCells[0].Value = "Mutated"
	state.Workbook.Sheets[0].Columns[0].Width = 1
	state.Workbook.Sheets[0].Rows[0].Height = 1
	state.Workbook.Styles[0].Fill.Colors[0] = "#000000"
	state.Workbook.Styles[0].Borders[0].Color = "#000000"
	state.Workbook.Styles[0].Render.TextColor = "#000000"

	fresh := app.State()
	if fresh.Appearance.Mode != AppearanceModeDark {
		t.Fatalf("appearance snapshot mutation leaked into app state: %#v", fresh.Appearance)
	}
	if fresh.Workbook.Sheets[0].Name != defaultSheetName {
		t.Fatalf("state snapshot mutation leaked into app state: %#v", fresh.Workbook.Sheets[0])
	}
	if fresh.Workbook.Sheets[0].Cells[0].Value != "Original" {
		t.Fatalf("cell snapshot mutation leaked into app state: %#v", fresh.Workbook.Sheets[0].Cells[0])
	}
	if fresh.Workbook.Sheets[0].MergedCells[0].Value != "Merged" {
		t.Fatalf("merged-cell snapshot mutation leaked into app state: %#v", fresh.Workbook.Sheets[0].MergedCells[0])
	}
	if fresh.Workbook.Sheets[0].Columns[0].Width != 12.5 {
		t.Fatalf("column snapshot mutation leaked into app state: %#v", fresh.Workbook.Sheets[0].Columns[0])
	}
	if fresh.Workbook.Sheets[0].Rows[0].Height != 20 {
		t.Fatalf("row snapshot mutation leaked into app state: %#v", fresh.Workbook.Sheets[0].Rows[0])
	}
	if fresh.Workbook.Styles[0].Fill.Colors[0] != "#FFFFFF" {
		t.Fatalf("fill-color snapshot mutation leaked into app state: %#v", fresh.Workbook.Styles[0].Fill)
	}
	if fresh.Workbook.Styles[0].Borders[0].Color != "#111111" {
		t.Fatalf("border snapshot mutation leaked into app state: %#v", fresh.Workbook.Styles[0].Borders[0])
	}
	if fresh.Workbook.Styles[0].Render.TextColor != "#E0E0E0" {
		t.Fatalf("render snapshot mutation leaked into app state: %#v", fresh.Workbook.Styles[0].Render)
	}
}

func TestAppearanceHelpers(t *testing.T) {
	t.Parallel()

	if !AppearanceModeSystem.valid() || !AppearanceModeLight.valid() || !AppearanceModeDark.valid() {
		t.Fatalf("expected built-in appearance modes to be valid")
	}
	if AppearanceMode("sepia").valid() {
		t.Fatalf("expected unknown appearance mode to be invalid")
	}
	if !AppearanceThemeLight.valid() || !AppearanceThemeDark.valid() {
		t.Fatalf("expected built-in appearance themes to be valid")
	}
	if AppearanceTheme("sepia").valid() {
		t.Fatalf("expected unknown appearance theme to be invalid")
	}

	if resolveEffectiveTheme(AppearanceModeSystem, AppearanceThemeDark) != AppearanceThemeDark {
		t.Fatalf("expected system mode to follow dark system theme")
	}
	if resolveEffectiveTheme(AppearanceModeLight, AppearanceThemeDark) != AppearanceThemeLight {
		t.Fatalf("expected light mode to override dark system theme")
	}
	if resolveEffectiveTheme(AppearanceModeDark, AppearanceThemeLight) != AppearanceThemeDark {
		t.Fatalf("expected dark mode to override light system theme")
	}

	normalized := normalizeAppearanceState(AppearanceState{
		Mode:        AppearanceMode("invalid"),
		SystemTheme: AppearanceTheme("invalid"),
	})
	assertDefaultAppearance(t, normalized)
}

func TestViewCommandsUpdateState(t *testing.T) {
	t.Parallel()

	app := NewApp()

	activeSheet := app.SetActiveSheet(defaultSheetName)
	if activeSheet.View.ActiveSheetName != defaultSheetName || activeSheet.Status.Kind != statusKindReady {
		t.Fatalf("expected active Sheet 1 and ready status, got %#v", activeSheet)
	}

	selected := app.SelectCell("b2")
	expectedB2 := CellAddress{Ref: "B2", Row: 2, Column: 2}
	if selected.View.ActiveCell != expectedB2 {
		t.Fatalf("expected active cell B2, got %#v", selected.View.ActiveCell)
	}
	expectedSelection := CellRange{Ref: "B2", Start: expectedB2, End: expectedB2}
	if selected.View.Selection != expectedSelection {
		t.Fatalf("expected selection range B2, got %#v", selected.View.Selection)
	}

	scrolled := app.SetScrollPosition(4, 3)
	if scrolled.View.Scroll != (ScrollPosition{TopRow: 4, LeftColumn: 3}) {
		t.Fatalf("expected scroll position 4,3, got %#v", scrolled.View.Scroll)
	}

	zoomedHigh := app.SetZoom(maxZoomPercent + 1)
	if zoomedHigh.View.ZoomPercent != maxZoomPercent {
		t.Fatalf("expected zoom to clamp to %d, got %d", maxZoomPercent, zoomedHigh.View.ZoomPercent)
	}

	zoomedLow := app.SetZoom(minZoomPercent - 1)
	if zoomedLow.View.ZoomPercent != minZoomPercent {
		t.Fatalf("expected zoom to clamp to %d, got %d", minZoomPercent, zoomedLow.View.ZoomPercent)
	}
}

func TestInvalidViewCommandsSetErrorAndKeepView(t *testing.T) {
	t.Parallel()

	app := NewApp()

	beforeCell := app.State()
	invalidCell := app.SelectCell("XFE1")
	if invalidCell.Status.Kind != statusKindError {
		t.Fatalf("expected invalid cell to set error status, got %#v", invalidCell.Status)
	}
	if invalidCell.View != beforeCell.View {
		t.Fatalf("expected invalid cell to keep view unchanged, got %#v", invalidCell.View)
	}

	beforeSheet := app.State()
	invalidSheet := app.SetActiveSheet("Missing")
	if invalidSheet.Status.Kind != statusKindError {
		t.Fatalf("expected invalid sheet to set error status, got %#v", invalidSheet.Status)
	}
	if invalidSheet.View != beforeSheet.View {
		t.Fatalf("expected invalid sheet to keep view unchanged, got %#v", invalidSheet.View)
	}

	beforeScroll := app.State()
	invalidScroll := app.SetScrollPosition(0, 1)
	if invalidScroll.Status.Kind != statusKindError {
		t.Fatalf("expected invalid scroll to set error status, got %#v", invalidScroll.Status)
	}
	if invalidScroll.View != beforeScroll.View {
		t.Fatalf("expected invalid scroll to keep view unchanged, got %#v", invalidScroll.View)
	}
}

func assertNeutralAppState(t *testing.T, state AppState) {
	t.Helper()

	expectedSheet := WorkbookSheet{
		Index:              defaultSheetIndex,
		Name:               defaultSheetName,
		State:              sheetStateVisible,
		Visible:            true,
		Bounds:             a1Range(),
		DefaultColumnWidth: defaultColumnWidth,
		DefaultRowHeight:   defaultRowHeight,
		Cells:              []CellData{},
		MergedCells:        []MergedCellRange{},
		Columns:            []ColumnLayout{},
		Rows:               []RowLayout{},
	}
	if state.Workbook.Title != defaultWorkbookTitle || state.Workbook.FilePath != "" || state.Workbook.FileName != "" {
		t.Fatalf("expected neutral workbook identity, got %#v", state.Workbook)
	}
	if state.Workbook.Dirty {
		t.Fatalf("expected neutral workbook to be clean, got %#v", state.Workbook)
	}
	if !reflect.DeepEqual(state.Workbook.Sheets, []WorkbookSheet{expectedSheet}) {
		t.Fatalf("expected neutral sheet list, got %#v", state.Workbook.Sheets)
	}
	if len(state.Workbook.Styles) != 0 {
		t.Fatalf("expected neutral styles to be empty, got %#v", state.Workbook.Styles)
	}
	assertDefaultAppearance(t, state.Appearance)

	expectedView := WorkbookViewState{
		ActiveSheetName: defaultSheetName,
		ActiveCell:      CellAddress{Ref: "A1", Row: minExcelRow, Column: minExcelColumn},
		Selection:       a1Range(),
		ZoomPercent:     defaultZoomPercent,
		Scroll:          ScrollPosition{TopRow: minExcelRow, LeftColumn: minExcelColumn},
	}
	if state.View != expectedView {
		t.Fatalf("expected neutral view, got %#v", state.View)
	}

	expectedStatus := AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false}
	if state.Status != expectedStatus {
		t.Fatalf("expected ready neutral status, got %#v", state.Status)
	}
}

func assertDefaultAppearance(t *testing.T, appearance AppearanceState) {
	t.Helper()

	if appearance != defaultAppearanceState() {
		t.Fatalf("expected default appearance, got %#v", appearance)
	}
}
