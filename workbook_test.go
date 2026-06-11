package main

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

const (
	fixtureDataSheet    = "Data"
	fixtureSummarySheet = "Summary"
)

func TestOpenWorkbookPathLoadsWorkbookState(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()

	state := app.OpenWorkbookPath(path)
	assertReadyStatus(t, state.Status)
	assertDefaultAppearance(t, state.Appearance)
	assertLoadedWorkbookIdentity(t, state, path)
	dataSheet := assertLoadedSheetsAndView(t, state)
	styleID := assertLoadedCells(t, dataSheet)
	assertLoadedLayout(t, dataSheet)
	assertLoadedStyle(t, state.Workbook.Styles, styleID)
}

func TestOpenWorkbookPathPreservesAppearance(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	expectedAppearance := AppearanceState{
		Mode:           AppearanceModeDark,
		SystemTheme:    AppearanceThemeLight,
		EffectiveTheme: AppearanceThemeDark,
	}
	app := NewApp()
	app.state.Appearance = expectedAppearance

	state := app.OpenWorkbookPath(path)
	assertReadyStatus(t, state.Status)
	if state.Appearance != expectedAppearance {
		t.Fatalf("expected workbook open to preserve appearance, got %#v", state.Appearance)
	}
}

func TestOpenWorkbookRejectionsPreservePreviousWorkbook(t *testing.T) {
	t.Parallel()

	validPath := createWorkbookFixture(t)
	unsupportedPath := writeTestFile(t, filepath.Join(t.TempDir(), "notes.txt"), []byte("not a workbook"))
	badWorkbookPath := writeTestFile(t, filepath.Join(t.TempDir(), "bad.xlsx"), []byte("not a zip file"))
	missingPath := filepath.Join(t.TempDir(), "missing.xlsx")
	directoryPath := t.TempDir()

	tests := []struct {
		name        string
		open        func(*App) AppState
		wantMessage string
	}{
		{
			name: "empty path",
			open: func(app *App) AppState {
				return app.OpenWorkbookPath(" \t ")
			},
			wantMessage: "choose a .xlsx workbook to open",
		},
		{
			name: "missing path",
			open: func(app *App) AppState {
				return app.OpenWorkbookPath(missingPath)
			},
			wantMessage: "workbook path does not exist",
		},
		{
			name: "directory path",
			open: func(app *App) AppState {
				return app.OpenWorkbookPath(directoryPath)
			},
			wantMessage: "workbook path is a directory",
		},
		{
			name: "unsupported extension",
			open: func(app *App) AppState {
				return app.OpenWorkbookPath(unsupportedPath)
			},
			wantMessage: "unsupported file type",
		},
		{
			name: "invalid xlsx content",
			open: func(app *App) AppState {
				return app.OpenWorkbookPath(badWorkbookPath)
			},
			wantMessage: "could not read workbook",
		},
		{
			name: "multiple dropped files",
			open: func(app *App) AppState {
				return app.OpenDroppedFiles([]string{validPath, validPath})
			},
			wantMessage: "only one .xlsx workbook can be opened at a time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := NewApp()
			loaded := app.OpenWorkbookPath(validPath)
			if loaded.Status.Kind != statusKindReady {
				t.Fatalf("expected fixture to load before rejection case, got %#v", loaded.Status)
			}

			before := app.State()
			result := tt.open(app)
			if result.Status.Kind != statusKindError || result.Status.Busy {
				t.Fatalf("expected error status without busy flag, got %#v", result.Status)
			}
			if !strings.Contains(result.Status.Message, tt.wantMessage) {
				t.Fatalf("expected status message containing %q, got %q", tt.wantMessage, result.Status.Message)
			}
			if !reflect.DeepEqual(result.Workbook, before.Workbook) {
				t.Fatalf("expected failed open to preserve workbook\nbefore: %#v\nafter:  %#v", before.Workbook, result.Workbook)
			}
			if !reflect.DeepEqual(result.View, before.View) {
				t.Fatalf("expected failed open to preserve view\nbefore: %#v\nafter:  %#v", before.View, result.View)
			}
		})
	}
}

func TestOpenWorkbookWithoutContextSetsError(t *testing.T) {
	t.Parallel()

	app := NewApp()
	before := app.State()

	state := app.OpenWorkbook()
	if state.Status.Kind != statusKindError || state.Status.Busy {
		t.Fatalf("expected unavailable dialog to set error status, got %#v", state.Status)
	}
	if !strings.Contains(state.Status.Message, "file dialog is not available yet") {
		t.Fatalf("expected file-dialog error message, got %q", state.Status.Message)
	}
	if !reflect.DeepEqual(state.Workbook, before.Workbook) || state.View != before.View {
		t.Fatalf("expected unavailable dialog to preserve state, got %#v", state)
	}
}

func assertReadyStatus(t *testing.T, status AppStatus) {
	t.Helper()

	expected := AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false}
	if status != expected {
		t.Fatalf("expected ready status after loading workbook, got %#v", status)
	}
}

func assertLoadedWorkbookIdentity(t *testing.T, state AppState, path string) {
	t.Helper()

	fileName := filepath.Base(path)
	if state.Workbook.Title != fileName || state.Workbook.FileName != fileName || state.Workbook.FilePath != path {
		t.Fatalf("expected workbook identity for %q, got %#v", path, state.Workbook)
	}
	if state.Workbook.Dirty {
		t.Fatalf("expected loaded workbook to be clean, got %#v", state.Workbook)
	}
}

func assertLoadedSheetsAndView(t *testing.T, state AppState) WorkbookSheet {
	t.Helper()

	if len(state.Workbook.Sheets) != 2 {
		t.Fatalf("expected two sheets, got %#v", state.Workbook.Sheets)
	}
	if state.Workbook.Sheets[0].Name != fixtureDataSheet || state.Workbook.Sheets[1].Name != fixtureSummarySheet {
		t.Fatalf("expected sheet list Data, Summary, got %#v", state.Workbook.Sheets)
	}
	if state.View.ActiveSheetName != fixtureSummarySheet {
		t.Fatalf("expected active sheet %q, got %#v", fixtureSummarySheet, state.View)
	}
	if state.View.ActiveCell != (CellAddress{Ref: "A1", Row: 1, Column: 1}) || state.View.Selection.Ref != "A1" {
		t.Fatalf("expected loaded view to reset to A1, got %#v", state.View)
	}

	dataSheet := findSheet(t, state.Workbook, fixtureDataSheet)
	if dataSheet.Index != 0 || !dataSheet.Visible || dataSheet.State != sheetStateVisible {
		t.Fatalf("expected visible Data sheet at index 0, got %#v", dataSheet)
	}
	if dataSheet.Bounds.Ref != "A1:E5" {
		t.Fatalf("expected bounds to include cells and merge, got %#v", dataSheet.Bounds)
	}

	return dataSheet
}

func assertLoadedCells(t *testing.T, dataSheet WorkbookSheet) int {
	t.Helper()

	header := findCell(t, dataSheet, "A1")
	if header.Value != "Name" || header.RawValue != "Name" {
		t.Fatalf("expected A1 string value and raw value, got %#v", header)
	}

	number := findCell(t, dataSheet, "B2")
	if number.Value != "12.50" || number.RawValue != "12.5" {
		t.Fatalf("expected B2 displayed/raw numeric values, got %#v", number)
	}
	if number.StyleID <= 0 {
		t.Fatalf("expected B2 to have a custom style, got %#v", number)
	}

	formula := findCell(t, dataSheet, "C2")
	if !formula.HasFormula || formula.Formula != "SUM(B2,7.5)" || formula.Kind != "formula" {
		t.Fatalf("expected C2 formula metadata, got %#v", formula)
	}

	merged := findMergedCell(t, dataSheet, "D4:E5")
	if merged.Value != "Merged" {
		t.Fatalf("expected merged range value, got %#v", merged)
	}

	return number.StyleID
}

func assertLoadedLayout(t *testing.T, dataSheet WorkbookSheet) {
	t.Helper()

	column := findColumn(t, dataSheet, "B")
	assertClose(t, "column B width", column.Width, 18.25)

	row := findRow(t, dataSheet, 3)
	assertClose(t, "row 3 height", row.Height, 28.5)
}

func assertLoadedStyle(t *testing.T, styles []CellStyle, styleID int) {
	t.Helper()

	findStyle(t, styles, 0)
	style := findStyle(t, styles, styleID)
	if style.NumberFormatID != 2 || style.NumberFormat != "0.00" {
		t.Fatalf("expected number format mapping, got %#v", style)
	}
	if style.Font.Family != "Arial" || !style.Font.Bold || !style.Font.Italic || style.Font.Underline != "single" {
		t.Fatalf("expected font style mapping, got %#v", style.Font)
	}
	if !style.Font.Strikethrough || style.Font.Color != "#336699" {
		t.Fatalf("expected font color and strikethrough mapping, got %#v", style.Font)
	}
	if style.Fill.Type != "pattern" || style.Fill.Pattern != 1 || style.Fill.Color != "#FFEEAA" {
		t.Fatalf("expected fill style mapping, got %#v", style.Fill)
	}
	if style.Alignment.Horizontal != "center" || style.Alignment.Vertical != "center" || !style.Alignment.WrapText {
		t.Fatalf("expected alignment mapping, got %#v", style.Alignment)
	}

	border := findBorder(t, style, "left")
	if border.Style != 1 || border.Color != "#CC5500" {
		t.Fatalf("expected left border mapping, got %#v", border)
	}
}

func createWorkbookFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "budget.xlsx")
	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close workbook fixture: %v", err)
		}
	}()

	firstSheet := file.GetSheetName(file.GetActiveSheetIndex())
	must(t, file.SetSheetName(firstSheet, fixtureDataSheet))
	summaryIndex, err := file.NewSheet(fixtureSummarySheet)
	must(t, err)
	file.SetActiveSheet(summaryIndex)

	must(t, file.SetCellValue(fixtureSummarySheet, "A1", "Summary"))
	must(t, file.SetCellValue(fixtureDataSheet, "A1", "Name"))
	must(t, file.SetCellValue(fixtureDataSheet, "A2", "Alpha"))
	must(t, file.SetCellFloat(fixtureDataSheet, "B2", 12.5, -1, 64))
	must(t, file.SetCellFormula(fixtureDataSheet, "C2", "SUM(B2,7.5)"))
	must(t, file.SetCellValue(fixtureDataSheet, "D2", "After formula"))
	must(t, file.MergeCell(fixtureDataSheet, "D4", "E5"))
	must(t, file.SetCellValue(fixtureDataSheet, "D4", "Merged"))
	must(t, file.SetColWidth(fixtureDataSheet, "B", "B", 18.25))
	must(t, file.SetRowHeight(fixtureDataSheet, 3, 28.5))

	styleID, err := file.NewStyle(&excelize.Style{
		NumFmt: 2,
		Font: &excelize.Font{
			Family:    "Arial",
			Size:      11,
			Bold:      true,
			Italic:    true,
			Underline: "single",
			Strike:    true,
			Color:     "336699",
		},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFEEAA"}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{{Type: "left", Style: 1, Color: "CC5500"}},
	})
	must(t, err)
	must(t, file.SetCellStyle(fixtureDataSheet, "B2", "B2", styleID))
	must(t, file.SaveAs(path))

	return path
}

func writeTestFile(t *testing.T, path string, data []byte) string {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test file %q: %v", path, err)
	}

	return path
}

func must(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func findSheet(t *testing.T, workbook WorkbookState, name string) WorkbookSheet {
	t.Helper()

	for _, sheet := range workbook.Sheets {
		if sheet.Name == name {
			return sheet
		}
	}

	t.Fatalf("sheet %q not found in %#v", name, workbook.Sheets)
	return WorkbookSheet{}
}

func findCell(t *testing.T, sheet WorkbookSheet, ref string) CellData {
	t.Helper()

	for _, cell := range sheet.Cells {
		if cell.Ref == ref {
			return cell
		}
	}

	t.Fatalf("cell %q not found in %#v", ref, sheet.Cells)
	return CellData{}
}

func findMergedCell(t *testing.T, sheet WorkbookSheet, ref string) MergedCellRange {
	t.Helper()

	for _, mergedCell := range sheet.MergedCells {
		if mergedCell.Range.Ref == ref {
			return mergedCell
		}
	}

	t.Fatalf("merged range %q not found in %#v", ref, sheet.MergedCells)
	return MergedCellRange{}
}

func findColumn(t *testing.T, sheet WorkbookSheet, name string) ColumnLayout {
	t.Helper()

	for _, column := range sheet.Columns {
		if column.Name == name {
			return column
		}
	}

	t.Fatalf("column %q not found in %#v", name, sheet.Columns)
	return ColumnLayout{}
}

func findRow(t *testing.T, sheet WorkbookSheet, index int) RowLayout {
	t.Helper()

	for _, row := range sheet.Rows {
		if row.Index == index {
			return row
		}
	}

	t.Fatalf("row %d not found in %#v", index, sheet.Rows)
	return RowLayout{}
}

func findStyle(t *testing.T, styles []CellStyle, id int) CellStyle {
	t.Helper()

	for _, style := range styles {
		if style.ID == id {
			return style
		}
	}

	t.Fatalf("style %d not found in %#v", id, styles)
	return CellStyle{}
}

func findBorder(t *testing.T, style CellStyle, side string) CellBorderStyle {
	t.Helper()

	for _, border := range style.Borders {
		if border.Side == side {
			return border
		}
	}

	t.Fatalf("border %q not found in %#v", side, style.Borders)
	return CellBorderStyle{}
}

func assertClose(t *testing.T, label string, got float64, want float64) {
	t.Helper()

	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("expected %s %.6f, got %.6f", label, want, got)
	}
}

func TestNormalizeSavePath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "budget")
	path, err := normalizeSavePath("  " + basePath + "  ")
	must(t, err)
	assertPathEqual(t, path, basePath+workbookFileExtension)

	upperPath := filepath.Join(tempDir, "Budget.XLSX")
	path, err = normalizeSavePath(upperPath)
	must(t, err)
	assertPathEqual(t, path, upperPath)

	_, err = normalizeSavePath(filepath.Join(tempDir, "budget.xlsm"))
	if err == nil || !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("expected unsupported extension error, got %v", err)
	}

	directoryPath := filepath.Join(tempDir, "folder.xlsx")
	must(t, os.Mkdir(directoryPath, 0o700))
	_, err = normalizeSavePath(directoryPath)
	if err == nil || !strings.Contains(err.Error(), "save path is a directory") {
		t.Fatalf("expected directory error, got %v", err)
	}
}

func TestSaveOpenedWorkbookAppliesPendingEditsAndPreservesContent(t *testing.T) {
	t.Parallel()

	sourcePath := createWorkbookFixture(t)
	targetPath := filepath.Join(t.TempDir(), "copy.xlsx")
	pendingEdits := map[string]map[string]string{
		fixtureDataSheet: {
			"A3": "=A1+B1",
			"B2": "Budget",
			"D2": "",
		},
	}

	savedPath, err := saveOpenedWorkbook(sourcePath, pendingEdits, targetPath)
	must(t, err)
	if savedPath != targetPath {
		t.Fatalf("expected result path %q, got %q", targetPath, savedPath)
	}

	file := openExcelFile(t, targetPath)
	assertExcelCellValue(t, file, fixtureDataSheet, "B2", "Budget")
	if styleID, err := file.GetCellStyle(fixtureDataSheet, "B2"); err != nil || styleID == 0 {
		t.Fatalf("expected edited B2 to keep style, styleID=%d err=%v", styleID, err)
	}
	assertExcelCellValue(t, file, fixtureDataSheet, "A3", "=A1+B1")
	assertExcelCellFormula(t, file, fixtureDataSheet, "A3", "")
	assertExcelCellValue(t, file, fixtureDataSheet, "D2", "")
	assertExcelCellFormula(t, file, fixtureDataSheet, "C2", "SUM(B2,7.5)")
	assertExcelCellValue(t, file, fixtureSummarySheet, "A1", "Summary")
	assertExcelMergeExists(t, file, fixtureDataSheet, "D4:E5")
}

func TestSaveWorkbookAsUntitledCreatesXLSXAndClearsDirty(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.ctx = context.Background()
	targetBase := filepath.Join(t.TempDir(), "literal-output")
	var gotOptions runtime.SaveDialogOptions
	app.saveFileDialog = func(_ context.Context, options runtime.SaveDialogOptions) (string, error) {
		gotOptions = options

		return targetBase, nil
	}

	app.SetCellValue(defaultSheetName, "A1", "123")
	app.SetCellValue(defaultSheetName, "A2", "=A1+B1")
	app.SetCellValue(defaultSheetName, "B2", "clear me")
	app.SetCellValue(defaultSheetName, "B2", "")

	state := app.SaveWorkbookAs()
	wantPath := targetBase + workbookFileExtension
	wantFileName := filepath.Base(wantPath)
	if state.Workbook.FilePath != wantPath ||
		state.Workbook.FileName != wantFileName ||
		state.Workbook.Title != wantFileName {
		t.Fatalf("expected Save As identity for %q, got %#v", wantPath, state.Workbook)
	}
	if state.Workbook.Dirty {
		t.Fatalf("expected successful Save As to clear dirty state, got %#v", state.Workbook)
	}
	if len(app.pendingCellEdits) != 0 {
		t.Fatalf("expected successful Save As to clear pending edits, got %#v", app.pendingCellEdits)
	}
	if state.Status != (AppStatus{Kind: statusKindReady, Message: savedStatusMessage, Busy: false}) {
		t.Fatalf("expected saved status, got %#v", state.Status)
	}
	if gotOptions.DefaultFilename != defaultWorkbookTitle+workbookFileExtension {
		t.Fatalf("expected default Save As filename for untitled workbook, got %#v", gotOptions)
	}
	if len(gotOptions.Filters) != 1 || gotOptions.Filters[0].Pattern != "*.xlsx" {
		t.Fatalf("expected .xlsx Save As filter, got %#v", gotOptions.Filters)
	}

	file := openExcelFile(t, wantPath)
	if sheetName := file.GetSheetName(file.GetActiveSheetIndex()); sheetName != defaultSheetName {
		t.Fatalf("expected default sheet %q, got %q", defaultSheetName, sheetName)
	}
	assertExcelCellValue(t, file, defaultSheetName, "A1", "123")
	assertExcelCellValue(t, file, defaultSheetName, "A2", "=A1+B1")
	assertExcelCellFormula(t, file, defaultSheetName, "A2", "")
	assertExcelCellValue(t, file, defaultSheetName, "B2", "")
}

func TestSaveWorkbookInPlacePreservesContentAndClearsDirty(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	app.OpenWorkbookPath(path)
	app.SetCellValue(fixtureDataSheet, "B2", "Budget")
	app.SetCellValue(fixtureDataSheet, "A3", "=A1+B1")

	state := app.SaveWorkbook()
	fileName := filepath.Base(path)
	if state.Workbook.FilePath != path ||
		state.Workbook.FileName != fileName ||
		state.Workbook.Title != fileName {
		t.Fatalf("expected in-place save to preserve identity, got %#v", state.Workbook)
	}
	if state.Workbook.Dirty {
		t.Fatalf("expected in-place save to clear dirty state, got %#v", state.Workbook)
	}
	if len(app.pendingCellEdits) != 0 {
		t.Fatalf("expected in-place save to clear pending edits, got %#v", app.pendingCellEdits)
	}
	if state.Status != (AppStatus{Kind: statusKindReady, Message: savedStatusMessage, Busy: false}) {
		t.Fatalf("expected saved status, got %#v", state.Status)
	}

	file := openExcelFile(t, path)
	assertExcelCellValue(t, file, fixtureDataSheet, "B2", "Budget")
	assertExcelCellValue(t, file, fixtureDataSheet, "A3", "=A1+B1")
	assertExcelCellFormula(t, file, fixtureDataSheet, "A3", "")
	assertExcelCellFormula(t, file, fixtureDataSheet, "C2", "SUM(B2,7.5)")
	assertExcelMergeExists(t, file, fixtureDataSheet, "D4:E5")
}

func TestSaveWorkbookWhenCleanReportsSavedWithoutChangingState(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	app.OpenWorkbookPath(path)
	before := app.State()

	state := app.SaveWorkbook()
	if state.Status != (AppStatus{Kind: statusKindReady, Message: savedStatusMessage, Busy: false}) {
		t.Fatalf("expected saved ready status, got %#v", state.Status)
	}
	before.Status = state.Status
	if !reflect.DeepEqual(state, before) {
		t.Fatalf("expected clean save to preserve state except status\nbefore: %#v\nafter:  %#v", before, state)
	}
}

func TestSaveWorkbookFailurePreservesDirtyStateAndPendingEdits(t *testing.T) {
	t.Parallel()

	path := createWorkbookFixture(t)
	app := NewApp()
	app.OpenWorkbookPath(path)
	app.SetCellValue(fixtureDataSheet, "B2", "Budget")
	before := app.State()
	must(t, os.Remove(path))

	state := app.SaveWorkbook()
	if state.Status.Kind != statusKindError || state.Status.Busy {
		t.Fatalf("expected save error status without busy flag, got %#v", state.Status)
	}
	if !strings.Contains(state.Status.Message, "could not save workbook") {
		t.Fatalf("expected save error message, got %q", state.Status.Message)
	}
	if !reflect.DeepEqual(state.Workbook, before.Workbook) || !reflect.DeepEqual(state.View, before.View) {
		t.Fatalf("expected failed save to preserve state\nbefore: %#v\nafter:  %#v", before, state)
	}
	assertPendingEdit(t, app, fixtureDataSheet, "B2", "Budget")
}

func TestSaveWorkbookAsCancelIsNoOp(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.ctx = context.Background()
	app.SetCellValue(defaultSheetName, "A1", "local")
	app.saveFileDialog = func(context.Context, runtime.SaveDialogOptions) (string, error) {
		return "", nil
	}
	before := app.State()

	state := app.SaveWorkbookAs()
	if !reflect.DeepEqual(state, before) {
		t.Fatalf("expected Save As cancel to preserve state\nbefore: %#v\nafter:  %#v", before, state)
	}
	assertPendingEdit(t, app, defaultSheetName, "A1", "local")
}

func TestSaveWorkbookAsDialogErrorPreservesDirtyState(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.ctx = context.Background()
	app.SetCellValue(defaultSheetName, "A1", "local")
	app.saveFileDialog = func(context.Context, runtime.SaveDialogOptions) (string, error) {
		return "", errors.New("dialog failed")
	}
	before := app.State()

	state := app.SaveWorkbookAs()
	if state.Status.Kind != statusKindError || !strings.Contains(state.Status.Message, "could not open save dialog") {
		t.Fatalf("expected dialog error status, got %#v", state.Status)
	}
	before.Status = state.Status
	if !reflect.DeepEqual(state, before) {
		t.Fatalf("expected Save As dialog error to preserve state except status\nbefore: %#v\nafter:  %#v", before, state)
	}
	assertPendingEdit(t, app, defaultSheetName, "A1", "local")
}

func TestOpenWorkbookDirtyPromptCancelSkipsFilePicker(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.ctx = context.Background()
	app.SetCellValue(defaultSheetName, "A1", "local")
	openDialogCalls := 0
	app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
		return dirtyPromptCancel, nil
	}
	app.openFileDialog = func(context.Context, runtime.OpenDialogOptions) (string, error) {
		openDialogCalls++

		return createWorkbookFixture(t), nil
	}
	before := app.State()

	state := app.OpenWorkbook()
	if openDialogCalls != 0 {
		t.Fatalf("expected canceling dirty prompt to skip file dialog, got %d calls", openDialogCalls)
	}
	if !reflect.DeepEqual(state, before) {
		t.Fatalf("expected dirty prompt cancel to preserve state\nbefore: %#v\nafter:  %#v", before, state)
	}
	assertPendingEdit(t, app, defaultSheetName, "A1", "local")
}

func TestOpenDroppedFilesDirtyPromptSaveThenOpen(t *testing.T) {
	t.Parallel()

	currentPath := createWorkbookFixture(t)
	nextPath := createWorkbookFixture(t)
	app := NewApp()
	app.ctx = context.Background()
	app.OpenWorkbookPath(currentPath)
	app.SetCellValue(fixtureDataSheet, "A2", "Saved Before Open")
	app.messageDialog = func(_ context.Context, options runtime.MessageDialogOptions) (string, error) {
		assertDirtyPromptOptions(t, options)

		return dirtyPromptSave, nil
	}

	state := app.OpenDroppedFiles([]string{nextPath})
	assertReadyStatus(t, state.Status)
	if state.Workbook.FilePath != nextPath || state.Workbook.Dirty {
		t.Fatalf("expected dropped workbook to replace current cleanly, got %#v", state.Workbook)
	}
	if len(app.pendingCellEdits) != 0 {
		t.Fatalf("expected pending edits to clear after replacement, got %#v", app.pendingCellEdits)
	}

	currentFile := openExcelFile(t, currentPath)
	assertExcelCellValue(t, currentFile, fixtureDataSheet, "A2", "Saved Before Open")
	nextFile := openExcelFile(t, nextPath)
	assertExcelCellValue(t, nextFile, fixtureDataSheet, "A2", "Alpha")
}

func TestOpenDroppedFilesDirtyPromptSaveFailureCancelsReplacement(t *testing.T) {
	t.Parallel()

	currentPath := createWorkbookFixture(t)
	nextPath := createWorkbookFixture(t)
	app := NewApp()
	app.ctx = context.Background()
	app.OpenWorkbookPath(currentPath)
	app.SetCellValue(fixtureDataSheet, "A2", "Unsaved")
	before := app.State()
	must(t, os.Remove(currentPath))
	app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
		return dirtyPromptSave, nil
	}

	state := app.OpenDroppedFiles([]string{nextPath})
	if state.Status.Kind != statusKindError || !strings.Contains(state.Status.Message, "could not save workbook") {
		t.Fatalf("expected save failure status, got %#v", state.Status)
	}
	before.Status = state.Status
	if !reflect.DeepEqual(state, before) {
		t.Fatalf("expected failed prompt save to preserve workbook\nbefore: %#v\nafter:  %#v", before, state)
	}
	assertPendingEdit(t, app, fixtureDataSheet, "A2", "Unsaved")
}

func TestOpenDroppedFilesDirtyPromptDontSaveReplacesWorkbook(t *testing.T) {
	t.Parallel()

	nextPath := createWorkbookFixture(t)
	app := NewApp()
	app.ctx = context.Background()
	app.SetCellValue(defaultSheetName, "A1", "discard me")
	app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
		return dirtyPromptDontSave, nil
	}

	state := app.OpenDroppedFiles([]string{nextPath})
	assertReadyStatus(t, state.Status)
	if state.Workbook.FilePath != nextPath || state.Workbook.Dirty {
		t.Fatalf("expected replacement workbook to be clean, got %#v", state.Workbook)
	}
	if len(app.pendingCellEdits) != 0 {
		t.Fatalf("expected replacement to clear pending edits, got %#v", app.pendingCellEdits)
	}
}

func TestOpenDroppedFilesDirtyPromptNativeYesNoChoices(t *testing.T) {
	t.Parallel()

	t.Run("yes saves then opens", func(t *testing.T) {
		t.Parallel()

		currentPath := createWorkbookFixture(t)
		nextPath := createWorkbookFixture(t)
		app := NewApp()
		app.ctx = context.Background()
		app.OpenWorkbookPath(currentPath)
		app.SetCellValue(fixtureDataSheet, "A2", "Native Yes Saved")
		app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
			return "Yes", nil
		}

		state := app.OpenDroppedFiles([]string{nextPath})
		assertReadyStatus(t, state.Status)
		if state.Workbook.FilePath != nextPath || state.Workbook.Dirty {
			t.Fatalf("expected native Yes prompt to save then open next workbook, got %#v", state.Workbook)
		}

		currentFile := openExcelFile(t, currentPath)
		assertExcelCellValue(t, currentFile, fixtureDataSheet, "A2", "Native Yes Saved")
	})

	t.Run("no discards and opens", func(t *testing.T) {
		t.Parallel()

		nextPath := createWorkbookFixture(t)
		app := NewApp()
		app.ctx = context.Background()
		app.SetCellValue(defaultSheetName, "A1", "discard me")
		app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
			return "No", nil
		}

		state := app.OpenDroppedFiles([]string{nextPath})
		assertReadyStatus(t, state.Status)
		if state.Workbook.FilePath != nextPath || state.Workbook.Dirty {
			t.Fatalf("expected native No prompt to discard then open next workbook, got %#v", state.Workbook)
		}
		if len(app.pendingCellEdits) != 0 {
			t.Fatalf("expected replacement to clear pending edits, got %#v", app.pendingCellEdits)
		}
	})
}

func TestBeforeCloseDirtyPromptChoices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		choice      string
		wantPrevent bool
		wantDirty   bool
	}{
		{name: "cancel", choice: dirtyPromptCancel, wantPrevent: true, wantDirty: true},
		{name: "dont save", choice: dirtyPromptDontSave, wantPrevent: false, wantDirty: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := NewApp()
			app.SetCellValue(defaultSheetName, "A1", "local")
			app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
				return tt.choice, nil
			}

			prevent := app.beforeClose(context.Background())
			if prevent != tt.wantPrevent {
				t.Fatalf("expected prevent=%t, got %t", tt.wantPrevent, prevent)
			}
			state := app.State()
			if state.Workbook.Dirty != tt.wantDirty {
				t.Fatalf("expected dirty=%t, got %#v", tt.wantDirty, state.Workbook)
			}
			if tt.wantDirty {
				assertPendingEdit(t, app, defaultSheetName, "A1", "local")
			} else if len(app.pendingCellEdits) != 0 {
				t.Fatalf("expected Don't Save close to clear pending edits, got %#v", app.pendingCellEdits)
			}
		})
	}
}

func TestBeforeCloseSaveAsCancelPreventsClose(t *testing.T) {
	t.Parallel()

	app := NewApp()
	app.SetCellValue(defaultSheetName, "A1", "local")
	app.messageDialog = func(context.Context, runtime.MessageDialogOptions) (string, error) {
		return dirtyPromptSave, nil
	}
	app.saveFileDialog = func(context.Context, runtime.SaveDialogOptions) (string, error) {
		return "", nil
	}

	prevent := app.beforeClose(context.Background())
	if !prevent {
		t.Fatalf("expected Save As cancellation to prevent close")
	}
	state := app.State()
	if !state.Workbook.Dirty {
		t.Fatalf("expected Save As cancellation to preserve dirty state, got %#v", state.Workbook)
	}
	assertPendingEdit(t, app, defaultSheetName, "A1", "local")
}

func openExcelFile(t *testing.T, path string) *excelize.File {
	t.Helper()

	file, err := excelize.OpenFile(path)
	must(t, err)
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close workbook %q: %v", path, err)
		}
	})

	return file
}

func assertExcelCellValue(t *testing.T, file *excelize.File, sheetName string, cellRef string, want string) {
	t.Helper()

	got, err := file.GetCellValue(sheetName, cellRef)
	must(t, err)
	if got != want {
		t.Fatalf("expected %s!%s value %q, got %q", sheetName, cellRef, want, got)
	}
}

func assertExcelCellFormula(t *testing.T, file *excelize.File, sheetName string, cellRef string, want string) {
	t.Helper()

	got, err := file.GetCellFormula(sheetName, cellRef)
	must(t, err)
	if got != want {
		t.Fatalf("expected %s!%s formula %q, got %q", sheetName, cellRef, want, got)
	}
}

func assertExcelMergeExists(t *testing.T, file *excelize.File, sheetName string, ref string) {
	t.Helper()

	mergeCells, err := file.GetMergeCells(sheetName)
	must(t, err)
	for _, mergeCell := range mergeCells {
		if len(mergeCell) > 0 && mergeCell[0] == ref {
			return
		}
	}

	t.Fatalf("expected merge range %q in %s, got %#v", ref, sheetName, mergeCells)
}

func assertPathEqual(t *testing.T, got string, want string) {
	t.Helper()

	absoluteWant, err := filepath.Abs(want)
	must(t, err)
	if got != absoluteWant {
		t.Fatalf("expected path %q, got %q", absoluteWant, got)
	}
}

func assertDirtyPromptOptions(t *testing.T, options runtime.MessageDialogOptions) {
	t.Helper()

	if options.Type != runtime.QuestionDialog ||
		options.DefaultButton != dirtyPromptSave ||
		options.CancelButton != dirtyPromptCancel {
		t.Fatalf("expected question dirty prompt options, got %#v", options)
	}
	wantButtons := []string{dirtyPromptSave, dirtyPromptDontSave, dirtyPromptCancel}
	if !reflect.DeepEqual(options.Buttons, wantButtons) {
		t.Fatalf("expected dirty prompt buttons %#v, got %#v", wantButtons, options.Buttons)
	}
	if !strings.Contains(strings.ToLower(options.Message), "save") {
		t.Fatalf("expected dirty prompt save message, got %#v", options)
	}
}
