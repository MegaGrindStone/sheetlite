package main

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
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
	if activeSheet.View.ActiveSheetName != defaultSheetName || activeSheet.Status.Kind != AppStatusKindReady {
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
	if invalidCell.Status.Kind != AppStatusKindError {
		t.Fatalf("expected invalid cell to set error status, got %#v", invalidCell.Status)
	}
	if invalidCell.View != beforeCell.View {
		t.Fatalf("expected invalid cell to keep view unchanged, got %#v", invalidCell.View)
	}

	beforeSheet := app.State()
	invalidSheet := app.SetActiveSheet("Missing")
	if invalidSheet.Status.Kind != AppStatusKindError {
		t.Fatalf("expected invalid sheet to set error status, got %#v", invalidSheet.Status)
	}
	if invalidSheet.View != beforeSheet.View {
		t.Fatalf("expected invalid sheet to keep view unchanged, got %#v", invalidSheet.View)
	}

	beforeScroll := app.State()
	invalidScroll := app.SetScrollPosition(0, 1)
	if invalidScroll.Status.Kind != AppStatusKindError {
		t.Fatalf("expected invalid scroll to set error status, got %#v", invalidScroll.Status)
	}
	if invalidScroll.View != beforeScroll.View {
		t.Fatalf("expected invalid scroll to keep view unchanged, got %#v", invalidScroll.View)
	}
}

func TestPendingEditsInitialized(t *testing.T) {
	t.Parallel()

	app := NewApp()
	if app.pendingCellEdits == nil {
		t.Fatalf("expected NewApp to initialize pending cell edits")
	}
	assertNoPendingLayoutEdits(t, app)

	var startupApp App
	startupApp.startup(context.Background())
	if startupApp.pendingCellEdits == nil {
		t.Fatalf("expected startup to initialize pending cell edits")
	}
	assertNoPendingLayoutEdits(t, &startupApp)
}

func TestSetSheetCellValueInsertsSortedAndExpandsBounds(t *testing.T) {
	t.Parallel()

	sheet := WorkbookSheet{
		Name:   defaultSheetName,
		Bounds: a1Range(),
		Cells: []CellData{
			{Ref: "A1", Row: 1, Column: 1, Value: "start", RawValue: "start", Kind: CellKindString},
		},
	}

	changed, err := sheet.setCellValue(mustParseCellAddress(t, "c3"), "tail")
	if err != nil || !changed {
		t.Fatalf("expected C3 insertion to change sheet, changed=%t err=%v", changed, err)
	}
	changed, err = sheet.setCellValue(mustParseCellAddress(t, "b2"), "middle")
	if err != nil || !changed {
		t.Fatalf("expected B2 insertion to change sheet, changed=%t err=%v", changed, err)
	}

	if sheet.Bounds.Ref != "A1:C3" {
		t.Fatalf("expected bounds A1:C3 after insertions, got %#v", sheet.Bounds)
	}

	wantRefs := []string{"A1", "B2", "C3"}
	gotRefs := make([]string, 0, len(sheet.Cells))
	for _, cell := range sheet.Cells {
		gotRefs = append(gotRefs, cell.Ref)
	}
	if !reflect.DeepEqual(gotRefs, wantRefs) {
		t.Fatalf("expected sorted cells %v, got %v", wantRefs, gotRefs)
	}

	cell := findCell(t, sheet, "B2")
	if cell.Value != "middle" ||
		cell.RawValue != "middle" ||
		cell.Kind != CellKindString ||
		cell.HasFormula ||
		cell.Formula != "" {
		t.Fatalf("expected literal B2 cell, got %#v", cell)
	}
}

func TestSetSheetCellValuePreservesStyleAndClearsFormula(t *testing.T) {
	t.Parallel()

	sheet := WorkbookSheet{
		Name:   defaultSheetName,
		Bounds: a1Range(),
		Cells: []CellData{
			{
				Ref:        "A1",
				Row:        1,
				Column:     1,
				Value:      "3",
				RawValue:   "3",
				Formula:    "SUM(B1:C1)",
				HasFormula: true,
				Kind:       "formula",
				StyleID:    7,
			},
		},
	}

	changed, err := sheet.setCellValue(mustParseCellAddress(t, "A1"), "=literal")
	if err != nil || !changed {
		t.Fatalf("expected formula overwrite to change sheet, changed=%t err=%v", changed, err)
	}

	cell := findCell(t, sheet, "A1")
	if cell.Value != "=literal" || cell.RawValue != "=literal" || cell.Kind != CellKindString {
		t.Fatalf("expected literal string value, got %#v", cell)
	}
	if cell.HasFormula || cell.Formula != "" {
		t.Fatalf("expected formula metadata to be cleared, got %#v", cell)
	}
	if cell.StyleID != 7 {
		t.Fatalf("expected style ID to be preserved, got %#v", cell)
	}
}

func TestSetSheetCellValueClearSemantics(t *testing.T) {
	t.Parallel()

	sheet := WorkbookSheet{
		Name:   defaultSheetName,
		Bounds: mustCellRange(t, 1, 1, 2, 2),
		Cells: []CellData{
			{Ref: "A1", Row: 1, Column: 1, Value: "plain", RawValue: "plain", Kind: CellKindString},
			{Ref: "B2", Row: 2, Column: 2, Value: "styled", RawValue: "styled", Kind: CellKindString, StyleID: 4},
		},
	}

	changed, err := sheet.setCellValue(mustParseCellAddress(t, "A1"), "")
	if err != nil || !changed {
		t.Fatalf("expected clearing A1 to remove unstyled cell, changed=%t err=%v", changed, err)
	}
	if _, ok := cellByRef(sheet, "A1"); ok {
		t.Fatalf("expected unstyled A1 to be removed, got %#v", sheet.Cells)
	}

	changed, err = sheet.setCellValue(mustParseCellAddress(t, "B2"), "")
	if err != nil || !changed {
		t.Fatalf("expected clearing B2 to keep style-only cell, changed=%t err=%v", changed, err)
	}
	cell, ok := cellByRef(sheet, "B2")
	if !ok {
		t.Fatalf("expected styled B2 to remain, got %#v", sheet.Cells)
	}
	if cell.StyleID != 4 ||
		cell.Value != "" ||
		cell.RawValue != "" ||
		cell.Kind != CellKindUnset ||
		cell.HasFormula ||
		cell.Formula != "" {
		t.Fatalf("expected style-only B2 after clear, got %#v", cell)
	}
	if sheet.Bounds.Ref != "A1:B2" {
		t.Fatalf("expected clear not to shrink bounds, got %#v", sheet.Bounds)
	}
}

func TestSetSheetColumnWidthUpdatesInsertsAndPreservesMetadata(t *testing.T) {
	t.Parallel()

	sheet := WorkbookSheet{
		DefaultColumnWidth: defaultColumnWidth,
		Columns: []ColumnLayout{
			{Index: 3, Name: "C", Width: 12.5, Hidden: true, OutlineLevel: 2, StyleID: 7},
		},
	}

	changed, err := sheet.setColumnWidth(3, 22.75)
	if err != nil || !changed {
		t.Fatalf("expected existing column resize to change sheet, changed=%t err=%v", changed, err)
	}
	changed, err = sheet.setColumnWidth(2, 18.25)
	if err != nil || !changed {
		t.Fatalf("expected default column resize to create layout, changed=%t err=%v", changed, err)
	}

	want := []ColumnLayout{
		{Index: 2, Name: "B", Width: 18.25},
		{Index: 3, Name: "C", Width: 22.75, Hidden: true, OutlineLevel: 2, StyleID: 7},
	}
	if !reflect.DeepEqual(sheet.Columns, want) {
		t.Fatalf("expected sorted column layouts with metadata preserved\nwant: %#v\ngot:  %#v", want, sheet.Columns)
	}
}

func TestSetSheetRowHeightUpdatesInsertsAndPreservesMetadata(t *testing.T) {
	t.Parallel()

	sheet := WorkbookSheet{
		DefaultRowHeight: defaultRowHeight,
		Rows: []RowLayout{
			{Index: 4, Height: 21.5, Hidden: true, OutlineLevel: 3},
		},
	}

	changed, err := sheet.setRowHeight(4, 33.5)
	if err != nil || !changed {
		t.Fatalf("expected existing row resize to change sheet, changed=%t err=%v", changed, err)
	}
	changed, err = sheet.setRowHeight(2, 24.25)
	if err != nil || !changed {
		t.Fatalf("expected default row resize to create layout, changed=%t err=%v", changed, err)
	}

	want := []RowLayout{
		{Index: 2, Height: 24.25},
		{Index: 4, Height: 33.5, Hidden: true, OutlineLevel: 3},
	}
	if !reflect.DeepEqual(sheet.Rows, want) {
		t.Fatalf("expected sorted row layouts with metadata preserved\nwant: %#v\ngot:  %#v", want, sheet.Rows)
	}
}

func TestSetSheetLayoutSizeNoOpDoesNotMutate(t *testing.T) {
	t.Parallel()

	sheet := WorkbookSheet{
		DefaultColumnWidth: defaultColumnWidth,
		DefaultRowHeight:   defaultRowHeight,
		Columns:            []ColumnLayout{{Index: 2, Name: "B", Width: 18.25}},
		Rows:               []RowLayout{{Index: 3, Height: 28.5}},
	}
	beforeColumns := append([]ColumnLayout(nil), sheet.Columns...)
	beforeRows := append([]RowLayout(nil), sheet.Rows...)

	changed, err := sheet.setColumnWidth(2, 18.25)
	if err != nil || changed {
		t.Fatalf("expected existing column no-op, changed=%t err=%v", changed, err)
	}
	changed, err = sheet.setColumnWidth(4, defaultColumnWidth)
	if err != nil || changed {
		t.Fatalf("expected default column no-op, changed=%t err=%v", changed, err)
	}
	changed, err = sheet.setRowHeight(3, 28.5)
	if err != nil || changed {
		t.Fatalf("expected existing row no-op, changed=%t err=%v", changed, err)
	}
	changed, err = sheet.setRowHeight(5, defaultRowHeight)
	if err != nil || changed {
		t.Fatalf("expected default row no-op, changed=%t err=%v", changed, err)
	}

	if !reflect.DeepEqual(sheet.Columns, beforeColumns) || !reflect.DeepEqual(sheet.Rows, beforeRows) {
		t.Fatalf("expected no-op layout changes to preserve slices, columns=%#v rows=%#v", sheet.Columns, sheet.Rows)
	}
}

func TestSetSheetLayoutSizeRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*WorkbookSheet) (bool, error)
	}{
		{
			name: "column below bounds",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setColumnWidth(0, 12)
			},
		},
		{
			name: "column above bounds",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setColumnWidth(maxExcelColumn+1, 12)
			},
		},
		{
			name: "column zero width",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setColumnWidth(1, 0)
			},
		},
		{
			name: "column negative width",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setColumnWidth(1, -1)
			},
		},
		{
			name: "column nan width",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setColumnWidth(1, math.NaN())
			},
		},
		{
			name: "column infinite width",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setColumnWidth(1, math.Inf(1))
			},
		},
		{
			name: "row below bounds",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setRowHeight(0, 20)
			},
		},
		{
			name: "row above bounds",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setRowHeight(maxExcelRow+1, 20)
			},
		},
		{
			name: "row zero height",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setRowHeight(1, 0)
			},
		},
		{
			name: "row negative height",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setRowHeight(1, -1)
			},
		},
		{
			name: "row nan height",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setRowHeight(1, math.NaN())
			},
		},
		{
			name: "row infinite height",
			mutate: func(sheet *WorkbookSheet) (bool, error) {
				return sheet.setRowHeight(1, math.Inf(1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sheet := WorkbookSheet{
				DefaultColumnWidth: defaultColumnWidth,
				DefaultRowHeight:   defaultRowHeight,
				Columns:            []ColumnLayout{{Index: 2, Name: "B", Width: 18.25}},
				Rows:               []RowLayout{{Index: 3, Height: 28.5}},
			}
			before := sheet
			before.Columns = append([]ColumnLayout(nil), sheet.Columns...)
			before.Rows = append([]RowLayout(nil), sheet.Rows...)

			changed, err := tt.mutate(&sheet)
			if err == nil || changed {
				t.Fatalf("expected invalid layout mutation to fail without change, changed=%t err=%v", changed, err)
			}
			if !reflect.DeepEqual(sheet, before) {
				t.Fatalf("expected invalid layout mutation to preserve sheet\nbefore: %#v\nafter:  %#v", before, sheet)
			}
		})
	}
}

func TestSetCellValueEditsNeutralWorkbook(t *testing.T) {
	t.Parallel()

	app := NewApp()
	state := app.SetCellValue(defaultSheetName, "b2", "hello")

	if !state.Workbook.Dirty {
		t.Fatalf("expected neutral workbook edit to mark dirty, got %#v", state.Workbook)
	}
	if state.Status != (AppStatus{Kind: AppStatusKindReady, Message: unsavedChangesStatusMessage, Busy: false}) {
		t.Fatalf("expected unsaved ready status, got %#v", state.Status)
	}

	sheet := findSheet(t, state.Workbook, defaultSheetName)
	cell := findCell(t, sheet, "B2")
	if cell.Value != "hello" ||
		cell.RawValue != "hello" ||
		cell.Kind != CellKindString ||
		cell.HasFormula ||
		cell.Formula != "" {
		t.Fatalf("expected literal B2 cell, got %#v", cell)
	}
	if sheet.Bounds.Ref != "A1:B2" {
		t.Fatalf("expected bounds to expand to A1:B2, got %#v", sheet.Bounds)
	}
	assertPendingEdit(t, app, defaultSheetName, "B2", "hello")
}

func TestSetCellValueKeepsLiteralStrings(t *testing.T) {
	t.Parallel()

	app := NewApp()
	values := []string{"123", "true", "2026-06-12", "=A1+B1"}
	var state AppState
	for i, value := range values {
		state = app.SetCellValue(defaultSheetName, fmt.Sprintf("A%d", i+1), value)
	}

	sheet := findSheet(t, state.Workbook, defaultSheetName)
	for i, value := range values {
		ref := fmt.Sprintf("A%d", i+1)
		cell := findCell(t, sheet, ref)
		if cell.Value != value || cell.RawValue != value ||
			cell.Kind != CellKindString || cell.HasFormula || cell.Formula != "" {
			t.Fatalf("expected %s to remain literal %q, got %#v", ref, value, cell)
		}
		assertPendingEdit(t, app, defaultSheetName, ref, value)
	}
}

func TestSetCellValueNoOpDoesNotDirtyCleanWorkbook(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Sheets[0].Cells = []CellData{
		{Ref: "A1", Row: 1, Column: 1, Value: "same", RawValue: "same", Kind: CellKindString},
	}

	state := app.SetCellValue(defaultSheetName, "A1", "same")
	if state.Workbook.Dirty {
		t.Fatalf("expected no-op edit to keep workbook clean, got %#v", state.Workbook)
	}
	if len(app.pendingCellEdits) != 0 {
		t.Fatalf("expected no pending edits for no-op, got %#v", app.pendingCellEdits)
	}
	if state.Status != (AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}) {
		t.Fatalf("expected default ready status for no-op, got %#v", state.Status)
	}
}

func TestSetCellValueLoadedWorkbookPreservesMetadataAndOverwritesFormula(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	loaded := app.OpenWorkbookPath(path)
	dataSheet := findSheet(t, loaded.Workbook, fixtureDataSheet)
	originalStyleID := findCell(t, dataSheet, "B2").StyleID

	state := app.SetCellValue(fixtureDataSheet, "b2", "Budget")
	if !state.Workbook.Dirty {
		t.Fatalf("expected loaded workbook edit to mark dirty")
	}
	dataSheet = findSheet(t, state.Workbook, fixtureDataSheet)
	edited := findCell(t, dataSheet, "B2")
	if edited.Value != "Budget" || edited.RawValue != "Budget" || edited.Kind != CellKindString {
		t.Fatalf("expected B2 literal edit, got %#v", edited)
	}
	if edited.StyleID != originalStyleID {
		t.Fatalf("expected B2 style ID %d to be preserved, got %#v", originalStyleID, edited)
	}
	findColumn(t, dataSheet, "B")
	findRow(t, dataSheet, 3)
	assertPendingEdit(t, app, fixtureDataSheet, "B2", "Budget")

	state = app.SetCellValue(fixtureDataSheet, "C2", "=A1+B1")
	dataSheet = findSheet(t, state.Workbook, fixtureDataSheet)
	formula := findCell(t, dataSheet, "C2")
	if formula.Value != "=A1+B1" || formula.RawValue != "=A1+B1" || formula.Kind != CellKindString {
		t.Fatalf("expected C2 literal formula-looking text, got %#v", formula)
	}
	if formula.HasFormula || formula.Formula != "" {
		t.Fatalf("expected C2 formula metadata to be cleared, got %#v", formula)
	}
	assertPendingEdit(t, app, fixtureDataSheet, "C2", "=A1+B1")
}

func TestSetCellValueClearBehavior(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Sheets[0].Bounds = mustCellRange(t, 1, 1, 2, 2)
	app.state.Workbook.Sheets[0].Cells = []CellData{
		{Ref: "A1", Row: 1, Column: 1, Value: "plain", RawValue: "plain", Kind: CellKindString},
		{Ref: "B2", Row: 2, Column: 2, Value: "styled", RawValue: "styled", Kind: CellKindString, StyleID: 12},
	}

	state := app.SetCellValue(defaultSheetName, "A1", "")
	sheet := findSheet(t, state.Workbook, defaultSheetName)
	if _, ok := cellByRef(sheet, "A1"); ok {
		t.Fatalf("expected unstyled A1 to be removed, got %#v", sheet.Cells)
	}
	assertPendingEdit(t, app, defaultSheetName, "A1", "")

	state = app.SetCellValue(defaultSheetName, "B2", "")
	sheet = findSheet(t, state.Workbook, defaultSheetName)
	cell, ok := cellByRef(sheet, "B2")
	if !ok {
		t.Fatalf("expected styled B2 to remain, got %#v", sheet.Cells)
	}
	if cell.StyleID != 12 ||
		cell.Value != "" ||
		cell.RawValue != "" ||
		cell.Kind != CellKindUnset ||
		cell.HasFormula ||
		cell.Formula != "" {
		t.Fatalf("expected style-only B2 after clear, got %#v", cell)
	}
	assertPendingEdit(t, app, defaultSheetName, "B2", "")
}

func TestSetCellValueRejectsInvalidInputAndPreservesState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sheetName   string
		cellRef     string
		wantMessage string
	}{
		{name: "empty sheet", sheetName: " ", cellRef: "A1", wantMessage: "Sheet name is required"},
		{name: "invalid ref", sheetName: defaultSheetName, cellRef: "XFE1", wantMessage: "Cell reference"},
		{name: "missing sheet", sheetName: "Missing", cellRef: "A1", wantMessage: "was not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := NewApp()
			before := app.State()
			state := app.SetCellValue(tt.sheetName, tt.cellRef, "value")
			if state.Status.Kind != AppStatusKindError || state.Status.Busy {
				t.Fatalf("expected error status without busy flag, got %#v", state.Status)
			}
			if !strings.Contains(state.Status.Message, tt.wantMessage) {
				t.Fatalf("expected error containing %q, got %q", tt.wantMessage, state.Status.Message)
			}
			if !reflect.DeepEqual(state.Workbook, before.Workbook) {
				t.Fatalf("expected invalid edit to preserve workbook\nbefore: %#v\nafter:  %#v", before.Workbook, state.Workbook)
			}
			if state.View != before.View {
				t.Fatalf("expected invalid edit to preserve view\nbefore: %#v\nafter:  %#v", before.View, state.View)
			}
			if len(app.pendingCellEdits) != 0 {
				t.Fatalf("expected invalid edit not to record pending edits, got %#v", app.pendingCellEdits)
			}
		})
	}
}

func TestSetCellValuePendingEditsClearOnOpen(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	state := app.SetCellValue(defaultSheetName, "A1", "local")
	if !state.Workbook.Dirty {
		t.Fatalf("expected local edit to mark workbook dirty")
	}
	assertPendingEdit(t, app, defaultSheetName, "A1", "local")

	opened := app.OpenWorkbookPath(path)
	assertReadyStatus(t, opened.Status)
	if opened.Workbook.Dirty {
		t.Fatalf("expected opened workbook to be clean, got %#v", opened.Workbook)
	}
	if len(app.pendingCellEdits) != 0 {
		t.Fatalf("expected pending edits to clear after open, got %#v", app.pendingCellEdits)
	}
}

func TestResizeCommandsCreateLayoutsAndRecordPendingEdits(t *testing.T) {
	t.Parallel()

	app := NewApp()
	state := app.SetColumnWidth(defaultSheetName, 2, 18.25)
	if !state.Workbook.Dirty {
		t.Fatalf("expected column resize to mark workbook dirty")
	}
	if state.Status != (AppStatus{Kind: AppStatusKindReady, Message: unsavedChangesStatusMessage, Busy: false}) {
		t.Fatalf("expected unsaved ready status, got %#v", state.Status)
	}
	sheet := findSheet(t, state.Workbook, defaultSheetName)
	column := findColumn(t, sheet, "B")
	assertClose(t, "column B width", column.Width, 18.25)
	assertPendingColumnWidth(t, app, defaultSheetName, 2, 18.25)

	state = app.SetRowHeight(defaultSheetName, 3, 28.5)
	if !state.Workbook.Dirty {
		t.Fatalf("expected row resize to keep workbook dirty")
	}
	sheet = findSheet(t, state.Workbook, defaultSheetName)
	row := findRow(t, sheet, 3)
	assertClose(t, "row 3 height", row.Height, 28.5)
	assertPendingRowHeight(t, app, defaultSheetName, 3, 28.5)

	state = app.SetColumnWidth(defaultSheetName, 4, 19.5)
	sheet = findSheet(t, state.Workbook, defaultSheetName)
	column = findColumn(t, sheet, "D")
	assertClose(t, "column D width", column.Width, 19.5)
	assertPendingColumnWidth(t, app, defaultSheetName, 4, 19.5)

	state = app.SetRowHeight(defaultSheetName, 4, 29.5)
	sheet = findSheet(t, state.Workbook, defaultSheetName)
	row = findRow(t, sheet, 4)
	assertClose(t, "row 4 height", row.Height, 29.5)
	assertPendingRowHeight(t, app, defaultSheetName, 4, 29.5)
}

func TestResizeCommandsUpdateExistingLayoutAndPreserveMetadata(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Sheets[0].Columns = []ColumnLayout{
		{Index: 2, Name: "B", Width: 18.25, Hidden: true, OutlineLevel: 1, StyleID: 8},
	}
	app.state.Workbook.Sheets[0].Rows = []RowLayout{
		{Index: 3, Height: 28.5, Hidden: true, OutlineLevel: 2},
	}

	app.SetColumnWidth(defaultSheetName, 2, 22.5)
	state := app.SetRowHeight(defaultSheetName, 3, 31.25)
	sheet := findSheet(t, state.Workbook, defaultSheetName)

	column := findColumn(t, sheet, "B")
	wantColumn := ColumnLayout{Index: 2, Name: "B", Width: 22.5, Hidden: true, OutlineLevel: 1, StyleID: 8}
	if column != wantColumn {
		t.Fatalf("expected column metadata to be preserved\nwant: %#v\ngot:  %#v", wantColumn, column)
	}
	row := findRow(t, sheet, 3)
	wantRow := RowLayout{Index: 3, Height: 31.25, Hidden: true, OutlineLevel: 2}
	if row != wantRow {
		t.Fatalf("expected row metadata to be preserved\nwant: %#v\ngot:  %#v", wantRow, row)
	}
	assertPendingColumnWidth(t, app, defaultSheetName, 2, 22.5)
	assertPendingRowHeight(t, app, defaultSheetName, 3, 31.25)
}

func TestResizeCommandsNoOpDoesNotDirtyCleanWorkbook(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.state.Workbook.Sheets[0].Columns = []ColumnLayout{{Index: 2, Name: "B", Width: 18.25}}
	app.state.Workbook.Sheets[0].Rows = []RowLayout{{Index: 3, Height: 28.5}}

	state := app.SetColumnWidth(defaultSheetName, 2, 18.25)
	if state.Workbook.Dirty {
		t.Fatalf("expected column no-op to keep workbook clean, got %#v", state.Workbook)
	}
	state = app.SetRowHeight(defaultSheetName, 3, 28.5)
	if state.Workbook.Dirty {
		t.Fatalf("expected row no-op to keep workbook clean, got %#v", state.Workbook)
	}
	if state.Status != (AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}) {
		t.Fatalf("expected default ready status for no-op, got %#v", state.Status)
	}
	assertNoPendingLayoutEdits(t, app)
}

func TestResizeCommandsRejectInvalidInputAndPreserveState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		resize      func(*App) AppState
		wantMessage string
	}{
		{
			name: "empty column sheet",
			resize: func(app *App) AppState {
				return app.SetColumnWidth(" ", 1, 12)
			},
			wantMessage: "Sheet name is required",
		},
		{
			name: "missing column sheet",
			resize: func(app *App) AppState {
				return app.SetColumnWidth("Missing", 1, 12)
			},
			wantMessage: "was not found",
		},
		{
			name: "column index below bounds",
			resize: func(app *App) AppState {
				return app.SetColumnWidth(defaultSheetName, 0, 12)
			},
			wantMessage: "Column index must be between",
		},
		{
			name: "column non-finite width",
			resize: func(app *App) AppState {
				return app.SetColumnWidth(defaultSheetName, 1, math.Inf(1))
			},
			wantMessage: "Column width must be a positive finite number",
		},
		{
			name: "empty row sheet",
			resize: func(app *App) AppState {
				return app.SetRowHeight(" ", 1, 20)
			},
			wantMessage: "Sheet name is required",
		},
		{
			name: "row index above bounds",
			resize: func(app *App) AppState {
				return app.SetRowHeight(defaultSheetName, maxExcelRow+1, 20)
			},
			wantMessage: "Row index must be between",
		},
		{
			name: "row invalid height",
			resize: func(app *App) AppState {
				return app.SetRowHeight(defaultSheetName, 1, 0)
			},
			wantMessage: "Row height must be a positive finite number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := NewApp()
			app.SetColumnWidth(defaultSheetName, 2, 18.25)
			app.SetRowHeight(defaultSheetName, 3, 28.5)
			before := app.State()
			state := tt.resize(app)
			if state.Status.Kind != AppStatusKindError || state.Status.Busy {
				t.Fatalf("expected error status without busy flag, got %#v", state.Status)
			}
			if !strings.Contains(state.Status.Message, tt.wantMessage) {
				t.Fatalf("expected error containing %q, got %q", tt.wantMessage, state.Status.Message)
			}
			if !reflect.DeepEqual(state.Workbook, before.Workbook) || state.View != before.View {
				t.Fatalf("expected invalid resize to preserve state\nbefore: %#v\nafter:  %#v", before, state)
			}
			assertPendingColumnWidth(t, app, defaultSheetName, 2, 18.25)
			assertPendingRowHeight(t, app, defaultSheetName, 3, 28.5)
		})
	}
}

func TestResizePendingLayoutEditsClearOnOpen(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	app.SetColumnWidth(defaultSheetName, 2, 18.25)
	state := app.SetRowHeight(defaultSheetName, 3, 28.5)
	if !state.Workbook.Dirty {
		t.Fatalf("expected resizes to mark workbook dirty")
	}
	assertPendingColumnWidth(t, app, defaultSheetName, 2, 18.25)
	assertPendingRowHeight(t, app, defaultSheetName, 3, 28.5)

	opened := app.OpenWorkbookPath(path)
	assertReadyStatus(t, opened.Status)
	if opened.Workbook.Dirty {
		t.Fatalf("expected opened workbook to be clean, got %#v", opened.Workbook)
	}
	assertNoPendingLayoutEdits(t, app)
}

func mustParseCellAddress(t *testing.T, ref string) CellAddress {
	t.Helper()

	address, ok := parseCellAddress(ref)
	if !ok {
		t.Fatalf("expected %q to parse as a cell address", ref)
	}

	return address
}

func mustCellRange(t *testing.T, startRow int, startColumn int, endRow int, endColumn int) CellRange {
	t.Helper()

	cellRange, err := cellRangeFromCoordinates(startRow, startColumn, endRow, endColumn)
	if err != nil {
		t.Fatalf("create cell range: %v", err)
	}

	return cellRange
}

func cellByRef(sheet WorkbookSheet, ref string) (CellData, bool) {
	for _, cell := range sheet.Cells {
		if cell.Ref == ref {
			return cell, true
		}
	}

	return CellData{}, false
}

func assertPendingEdit(t *testing.T, app *App, sheetName string, cellRef string, value string) {
	t.Helper()

	sheetEdits, ok := app.pendingCellEdits[sheetName]
	if !ok {
		t.Fatalf("expected pending edits for sheet %q, got %#v", sheetName, app.pendingCellEdits)
	}
	got, ok := sheetEdits[cellRef]
	if !ok {
		t.Fatalf("expected pending edit for %s!%s, got %#v", sheetName, cellRef, app.pendingCellEdits)
	}
	if got != value {
		t.Fatalf("expected pending edit %s!%s=%q, got %q", sheetName, cellRef, value, got)
	}
}

func assertPendingColumnWidth(t *testing.T, app *App, sheetName string, columnIndex int, width float64) {
	t.Helper()

	sheetEdits, ok := app.pendingLayoutEdits.ColumnWidths[sheetName]
	if !ok {
		t.Fatalf("expected pending column widths for sheet %q, got %#v", sheetName, app.pendingLayoutEdits.ColumnWidths)
	}
	got, ok := sheetEdits[columnIndex]
	if !ok {
		t.Fatalf(
			"expected pending column width for %s!%d, got %#v",
			sheetName,
			columnIndex,
			app.pendingLayoutEdits.ColumnWidths,
		)
	}
	assertClose(t, fmt.Sprintf("pending column width %s!%d", sheetName, columnIndex), got, width)
}

func assertPendingRowHeight(t *testing.T, app *App, sheetName string, rowIndex int, height float64) {
	t.Helper()

	sheetEdits, ok := app.pendingLayoutEdits.RowHeights[sheetName]
	if !ok {
		t.Fatalf("expected pending row heights for sheet %q, got %#v", sheetName, app.pendingLayoutEdits.RowHeights)
	}
	got, ok := sheetEdits[rowIndex]
	if !ok {
		t.Fatalf("expected pending row height for %s!%d, got %#v", sheetName, rowIndex, app.pendingLayoutEdits.RowHeights)
	}
	assertClose(t, fmt.Sprintf("pending row height %s!%d", sheetName, rowIndex), got, height)
}

func assertNoPendingLayoutEdits(t *testing.T, app *App) {
	t.Helper()

	if len(app.pendingLayoutEdits.ColumnWidths) != 0 || len(app.pendingLayoutEdits.RowHeights) != 0 {
		t.Fatalf("expected no pending layout edits, got %#v", app.pendingLayoutEdits)
	}
}

func assertNeutralAppState(t *testing.T, state AppState) {
	t.Helper()

	expectedSheet := WorkbookSheet{
		Index:              defaultSheetIndex,
		Name:               defaultSheetName,
		State:              SheetStateVisible,
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

	expectedStatus := AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}
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
