package main

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
)

// App owns Wails runtime context and the current application state snapshot.
type App struct {
	mu    sync.RWMutex
	ctx   context.Context
	state AppState
}

// NewApp creates a new App application struct with neutral startup state.
func NewApp() *App {
	return &App{state: initialAppState()}
}

// startup is called when the app starts. The context is saved for runtime calls.
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ctx = ctx
	// Keep tests/free-standing construction usable if startup is called on a zero App.
	if a.state.Status.Kind == "" {
		a.state = initialAppState()
	}
}

// State returns the current app state snapshot without mutating it.
func (a *App) State() AppState {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return cloneAppState(a.state)
}

// SetActiveSheet changes the active sheet when the sheet exists in the state.
func (a *App) SetActiveSheet(name string) AppState {
	sheetName := strings.TrimSpace(name)

	a.mu.Lock()
	defer a.mu.Unlock()

	if sheetName == "" {
		a.state.Status = AppStatus{Kind: statusKindError, Message: "Sheet name is required.", Busy: false}

		return cloneAppState(a.state)
	}

	if !slices.ContainsFunc(a.state.Workbook.Sheets, func(sheet WorkbookSheet) bool {
		return sheet.Name == sheetName
	}) {
		a.state.Status = AppStatus{
			Kind:    statusKindError,
			Message: fmt.Sprintf("Sheet %q was not found.", sheetName),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	a.state.View.ActiveSheetName = sheetName
	a.state.Status = AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SelectCell changes the active cell for a valid Excel cell reference.
func (a *App) SelectCell(cellRef string) AppState {
	address, ok := parseCellAddress(cellRef)

	a.mu.Lock()
	defer a.mu.Unlock()

	if !ok {
		a.state.Status = AppStatus{
			Kind:    statusKindError,
			Message: fmt.Sprintf("Cell reference %q is invalid.", cellRef),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	// Selection is kept separate so a future range selection can survive focus movement.
	a.state.View.ActiveCell = address
	a.state.Status = AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetScrollPosition changes the top-left visible cell coordinates.
func (a *App) SetScrollPosition(topRow int, leftColumn int) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	if topRow < minExcelRow || topRow > maxExcelRow {
		a.state.Status = AppStatus{
			Kind:    statusKindError,
			Message: fmt.Sprintf("Top row must be between %d and %d.", minExcelRow, maxExcelRow),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if leftColumn < minExcelColumn || leftColumn > maxExcelColumn {
		a.state.Status = AppStatus{
			Kind:    statusKindError,
			Message: fmt.Sprintf("Left column must be between %d and %d.", minExcelColumn, maxExcelColumn),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	a.state.View.Scroll = ScrollPosition{TopRow: topRow, LeftColumn: leftColumn}
	a.state.Status = AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetZoom changes the zoom percentage, clamped to the supported range.
func (a *App) SetZoom(percent int) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Match common spreadsheet zoom bounds instead of rejecting toolbar-style oversteps.
	a.state.View.ZoomPercent = clampZoomPercent(percent)
	a.state.Status = AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}
