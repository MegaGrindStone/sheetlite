package main

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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
	assertLoadedWorkbookIdentity(t, state, path)
	dataSheet := assertLoadedSheetsAndView(t, state)
	styleID := assertLoadedCells(t, dataSheet)
	assertLoadedLayout(t, dataSheet)
	assertLoadedStyle(t, state.Workbook.Styles, styleID)
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
	if !state.Workbook.HasWorkbook {
		t.Fatalf("expected workbook to be loaded, got %#v", state.Workbook)
	}
	if state.Workbook.Title != fileName || state.Workbook.FileName != fileName || state.Workbook.FilePath != path {
		t.Fatalf("expected workbook identity for %q, got %#v", path, state.Workbook)
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
