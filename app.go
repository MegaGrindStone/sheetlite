package main

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	dirtyPromptSave     = "Save"
	dirtyPromptDontSave = "Don't Save"
	dirtyPromptCancel   = "Cancel"
)

// App owns Wails runtime context and the current application state snapshot.
type App struct {
	mu                 sync.RWMutex
	ctx                context.Context
	state              AppState
	pendingCellEdits   map[string]map[string]string
	pendingLayoutEdits pendingLayoutEdits
	openFileDialog     func(context.Context, runtime.OpenDialogOptions) (string, error)
	saveFileDialog     func(context.Context, runtime.SaveDialogOptions) (string, error)
	messageDialog      func(context.Context, runtime.MessageDialogOptions) (string, error)
}

// NewApp creates a new App application struct with neutral startup state.
func NewApp() *App {
	return &App{
		state:            initialAppState(),
		pendingCellEdits: map[string]map[string]string{},
		pendingLayoutEdits: pendingLayoutEdits{
			ColumnWidths: map[string]map[int]float64{},
			RowHeights:   map[string]map[int]float64{},
		},
	}
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
	if a.pendingCellEdits == nil {
		a.pendingCellEdits = map[string]map[string]string{}
	}
	a.pendingLayoutEdits = pendingLayoutEdits{
		ColumnWidths: map[string]map[int]float64{},
		RowHeights:   map[string]map[int]float64{},
	}
}

// State returns the current app state snapshot without mutating it.
func (a *App) State() AppState {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return cloneAppState(a.state)
}

// InitializeAppearance seeds backend appearance from frontend persistence and system theme.
func (a *App) InitializeAppearance(mode AppearanceMode, systemTheme AppearanceTheme) AppState {
	if !mode.valid() {
		mode = AppearanceModeSystem
	}
	if !systemTheme.valid() {
		systemTheme = AppearanceThemeLight
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.state.Appearance = AppearanceState{
		Mode:           mode,
		SystemTheme:    systemTheme,
		EffectiveTheme: resolveEffectiveTheme(mode, systemTheme),
	}
	for i := range a.state.Workbook.Styles {
		a.state.Workbook.Styles[i].render(a.state.Appearance.EffectiveTheme)
	}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetAppearanceMode changes the selected appearance mode.
func (a *App) SetAppearanceMode(mode AppearanceMode) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !mode.valid() {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: "Appearance mode must be system, light, or dark.",
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	appearance := normalizeAppearanceState(a.state.Appearance)
	appearance.Mode = mode
	appearance.EffectiveTheme = resolveEffectiveTheme(appearance.Mode, appearance.SystemTheme)
	a.state.Appearance = appearance
	for i := range a.state.Workbook.Styles {
		a.state.Workbook.Styles[i].render(appearance.EffectiveTheme)
	}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetSystemTheme changes the latest frontend-reported system theme.
func (a *App) SetSystemTheme(theme AppearanceTheme) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !theme.valid() {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: "System theme must be light or dark.",
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	appearance := normalizeAppearanceState(a.state.Appearance)
	appearance.SystemTheme = theme
	appearance.EffectiveTheme = resolveEffectiveTheme(appearance.Mode, appearance.SystemTheme)
	a.state.Appearance = appearance
	for i := range a.state.Workbook.Styles {
		a.state.Workbook.Styles[i].render(appearance.EffectiveTheme)
	}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetActiveSheet changes the active sheet when the sheet exists in the state.
func (a *App) SetActiveSheet(name string) AppState {
	sheetName := strings.TrimSpace(name)

	a.mu.Lock()
	defer a.mu.Unlock()

	if sheetName == "" {
		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: "Sheet name is required.", Busy: false}

		return cloneAppState(a.state)
	}

	if !slices.ContainsFunc(a.state.Workbook.Sheets, func(sheet WorkbookSheet) bool {
		return sheet.Name == sheetName
	}) {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Sheet %q was not found.", sheetName),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	a.state.View.ActiveSheetName = sheetName
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SelectCell changes the active cell for a valid Excel cell reference.
func (a *App) SelectCell(cellRef string) AppState {
	address, ok := parseCellAddress(cellRef)

	a.mu.Lock()
	defer a.mu.Unlock()

	if !ok {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Cell reference %q is invalid.", cellRef),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	// Active cell and selection are both updated to the single-cell range,
	// keeping selection separate for future multi-cell range expansion.
	a.state.View.ActiveCell = address
	a.state.View.Selection = CellRange{Ref: address.Ref, Start: address, End: address}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetCellValue applies one literal text edit to a cell in the current workbook.
func (a *App) SetCellValue(sheetName string, cellRef string, value string) AppState {
	trimmedSheetName := strings.TrimSpace(sheetName)
	address, validAddress := parseCellAddress(cellRef)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Keep zero-value App structs usable in focused backend tests and command calls.
	if a.state.Status.Kind == "" {
		a.state = initialAppState()
	}

	if trimmedSheetName == "" {
		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: "Sheet name is required.", Busy: false}

		return cloneAppState(a.state)
	}

	if !validAddress {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Cell reference %q is invalid.", cellRef),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	sheetIndex := slices.IndexFunc(a.state.Workbook.Sheets, func(sheet WorkbookSheet) bool {
		return sheet.Name == trimmedSheetName
	})
	if sheetIndex < 0 {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Sheet %q was not found.", trimmedSheetName),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	changed, err := a.state.Workbook.Sheets[sheetIndex].setCellValue(address, value)
	if err != nil {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Could not edit %s on sheet %q: %v", address.Ref, trimmedSheetName, err),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if changed {
		a.state.Workbook.Dirty = true
		// Keep the exact edit for later save logic, including empty clears.
		if a.pendingCellEdits == nil {
			a.pendingCellEdits = map[string]map[string]string{}
		}
		if a.pendingCellEdits[trimmedSheetName] == nil {
			a.pendingCellEdits[trimmedSheetName] = map[string]string{}
		}
		a.pendingCellEdits[trimmedSheetName][address.Ref] = value
		a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: unsavedChangesStatusMessage, Busy: false}

		return cloneAppState(a.state)
	}

	message := defaultStatusMessage
	if a.state.Workbook.Dirty {
		message = unsavedChangesStatusMessage
	}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: message, Busy: false}

	return cloneAppState(a.state)
}

// SetColumnWidth applies one committed width change to a worksheet column.
func (a *App) SetColumnWidth(sheetName string, columnIndex int, width float64) AppState {
	return a.setLayoutDimension(sheetName, columnIndex, width, layoutDimensionCommand{
		indexName: "Column index",
		sizeName:  "Column width",
		target:    "column",
		minIndex:  minExcelColumn,
		maxIndex:  maxExcelColumn,
		mutate: func(sheet *WorkbookSheet, index int, size float64) (bool, error) {
			return sheet.setColumnWidth(index, size)
		},
		record: func(app *App, name string, index int, size float64) {
			if app.pendingLayoutEdits.ColumnWidths[name] == nil {
				app.pendingLayoutEdits.ColumnWidths[name] = map[int]float64{}
			}
			app.pendingLayoutEdits.ColumnWidths[name][index] = size
		},
	})
}

// SetRowHeight applies one committed height change to a worksheet row.
func (a *App) SetRowHeight(sheetName string, rowIndex int, height float64) AppState {
	return a.setLayoutDimension(sheetName, rowIndex, height, layoutDimensionCommand{
		indexName: "Row index",
		sizeName:  "Row height",
		target:    "row",
		minIndex:  minExcelRow,
		maxIndex:  maxExcelRow,
		mutate: func(sheet *WorkbookSheet, index int, size float64) (bool, error) {
			return sheet.setRowHeight(index, size)
		},
		record: func(app *App, name string, index int, size float64) {
			if app.pendingLayoutEdits.RowHeights[name] == nil {
				app.pendingLayoutEdits.RowHeights[name] = map[int]float64{}
			}
			app.pendingLayoutEdits.RowHeights[name][index] = size
		},
	})
}

type layoutDimensionCommand struct {
	indexName string
	sizeName  string
	target    string
	minIndex  int
	maxIndex  int
	mutate    func(*WorkbookSheet, int, float64) (bool, error)
	record    func(*App, string, int, float64)
}

func (a *App) setLayoutDimension(
	sheetName string,
	index int,
	size float64,
	command layoutDimensionCommand,
) AppState {
	trimmedSheetName := strings.TrimSpace(sheetName)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Keep zero-value App structs usable in focused backend tests and command calls.
	if a.state.Status.Kind == "" {
		a.state = initialAppState()
		a.pendingLayoutEdits = pendingLayoutEdits{
			ColumnWidths: map[string]map[int]float64{},
			RowHeights:   map[string]map[int]float64{},
		}
	}

	if trimmedSheetName == "" {
		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: "Sheet name is required.", Busy: false}

		return cloneAppState(a.state)
	}

	if index < command.minIndex || index > command.maxIndex {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("%s must be between %d and %d.", command.indexName, command.minIndex, command.maxIndex),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if !validLayoutDimension(size) {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: command.sizeName + " must be a positive finite number.",
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	sheetIndex := slices.IndexFunc(a.state.Workbook.Sheets, func(sheet WorkbookSheet) bool {
		return sheet.Name == trimmedSheetName
	})
	if sheetIndex < 0 {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Sheet %q was not found.", trimmedSheetName),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	changed, err := command.mutate(&a.state.Workbook.Sheets[sheetIndex], index, size)
	if err != nil {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Could not resize %s %d on sheet %q: %v", command.target, index, trimmedSheetName, err),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if changed {
		a.state.Workbook.Dirty = true
		command.record(a, trimmedSheetName, index, size)
		a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: unsavedChangesStatusMessage, Busy: false}

		return cloneAppState(a.state)
	}

	message := defaultStatusMessage
	if a.state.Workbook.Dirty {
		message = unsavedChangesStatusMessage
	}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: message, Busy: false}

	return cloneAppState(a.state)
}

// SetScrollPosition changes the top-left visible cell coordinates.
func (a *App) SetScrollPosition(topRow int, leftColumn int) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	if topRow < minExcelRow || topRow > maxExcelRow {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Top row must be between %d and %d.", minExcelRow, maxExcelRow),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if leftColumn < minExcelColumn || leftColumn > maxExcelColumn {
		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("Left column must be between %d and %d.", minExcelColumn, maxExcelColumn),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	a.state.View.Scroll = ScrollPosition{TopRow: topRow, LeftColumn: leftColumn}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

// SetZoom changes the zoom percentage, clamped to the supported range.
func (a *App) SetZoom(percent int) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Match common spreadsheet zoom bounds instead of rejecting toolbar-style oversteps.
	a.state.View.ZoomPercent = clampZoomPercent(percent)
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

func (a *App) guardDestructiveTransition(discardOnDontSave bool) bool {
	a.mu.RLock()
	ctx := a.ctx
	dirty := a.state.Workbook.Dirty
	a.mu.RUnlock()

	if !dirty {
		return true
	}

	if ctx == nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: "unsaved-changes dialog is not available yet",
			Busy:    false,
		}

		return false
	}

	choice, err := a.runMessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Unsaved changes",
		Message:       "Save changes before continuing?",
		Buttons:       []string{dirtyPromptSave, dirtyPromptDontSave, dirtyPromptCancel},
		DefaultButton: dirtyPromptSave,
		CancelButton:  dirtyPromptCancel,
	})
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("could not show unsaved-changes dialog: %v", err),
			Busy:    false,
		}

		return false
	}

	switch normalizeDirtyPromptChoice(choice) {
	case dirtyPromptSave:
		state := a.SaveWorkbook()

		return !state.Workbook.Dirty && state.Status.Kind != AppStatusKindError
	case dirtyPromptDontSave:
		// Open/drop clear edits only after replacement succeeds; close has no later cleanup point.
		if discardOnDontSave {
			a.discardPendingEdits()
		}

		return true
	default:
		return false
	}
}

func normalizeDirtyPromptChoice(choice string) string {
	// Wails QuestionDialog returns native Yes/No on Linux and Windows even when
	// custom button labels are supplied; macOS returns the configured labels.
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case strings.ToLower(dirtyPromptSave), "yes":
		return dirtyPromptSave
	case strings.ToLower(dirtyPromptDontSave), "no":
		return dirtyPromptDontSave
	case strings.ToLower(dirtyPromptCancel), "":
		return dirtyPromptCancel
	default:
		return dirtyPromptCancel
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	if ctx != nil {
		a.mu.Lock()
		a.ctx = ctx
		a.mu.Unlock()
	}

	// Wails expects true to prevent close; guard returns true to continue.
	return !a.guardDestructiveTransition(true)
}

func (a *App) discardPendingEdits() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.pendingCellEdits = map[string]map[string]string{}
	a.pendingLayoutEdits = pendingLayoutEdits{
		ColumnWidths: map[string]map[int]float64{},
		RowHeights:   map[string]map[int]float64{},
	}
	a.state.Workbook.Dirty = false
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false}
}

func (a *App) runOpenFileDialog(ctx context.Context, options runtime.OpenDialogOptions) (string, error) {
	if a.openFileDialog != nil {
		return a.openFileDialog(ctx, options)
	}

	return runtime.OpenFileDialog(ctx, options)
}

func (a *App) runSaveFileDialog(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
	if a.saveFileDialog != nil {
		return a.saveFileDialog(ctx, options)
	}

	return runtime.SaveFileDialog(ctx, options)
}

func (a *App) runMessageDialog(ctx context.Context, options runtime.MessageDialogOptions) (string, error) {
	if a.messageDialog != nil {
		return a.messageDialog(ctx, options)
	}

	return runtime.MessageDialog(ctx, options)
}

type pendingLayoutEdits struct {
	ColumnWidths map[string]map[int]float64
	RowHeights   map[string]map[int]float64
}

func (p pendingLayoutEdits) clone() pendingLayoutEdits {
	clone := pendingLayoutEdits{
		ColumnWidths: map[string]map[int]float64{},
		RowHeights:   map[string]map[int]float64{},
	}
	for sheetName, sheetEdits := range p.ColumnWidths {
		clone.ColumnWidths[sheetName] = maps.Clone(sheetEdits)
	}
	for sheetName, sheetEdits := range p.RowHeights {
		clone.RowHeights[sheetName] = maps.Clone(sheetEdits)
	}

	return clone
}
