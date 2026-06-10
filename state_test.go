package main

import "testing"

func TestStateReturnsSnapshotCopy(t *testing.T) {
	t.Parallel()

	app := NewApp()
	state := app.State()
	state.Workbook.Sheets[0].Name = "Mutated"

	fresh := app.State()
	if fresh.Workbook.Sheets[0].Name != defaultSheetName {
		t.Fatalf("state snapshot mutation leaked into app state: %#v", fresh.Workbook.Sheets[0])
	}
}

func TestViewCommandsUpdateState(t *testing.T) {
	t.Parallel()

	app := NewApp()

	activeSheet := app.SetActiveSheet(defaultSheetName)
	if activeSheet.View.ActiveSheetName != defaultSheetName || activeSheet.Status.Kind != statusKindReady {
		t.Fatalf("expected active Sheet 1 and ready status, got %#v", activeSheet)
	}

	selected := app.SelectCell("b2")
	if selected.View.ActiveCell != (CellAddress{Ref: "B2", Row: 2, Column: 2}) {
		t.Fatalf("expected active cell B2, got %#v", selected.View.ActiveCell)
	}
	if selected.View.Selection.Ref != "A1" {
		t.Fatalf("expected selection to stay A1, got %#v", selected.View.Selection)
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
