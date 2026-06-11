package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

const (
	workbookFileExtension  = ".xlsx"
	loadingWorkbookMessage = "Opening workbook..."
)

// OpenWorkbook prompts for a .xlsx workbook and opens the selected path.
func (a *App) OpenWorkbook() AppState {
	a.mu.RLock()
	ctx := a.ctx
	a.mu.RUnlock()
	// Open the native dialog without holding the app-state lock.

	if ctx == nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: statusKindError, Message: "file dialog is not available yet", Busy: false}

		return cloneAppState(a.state)
	}

	path, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "Open .xlsx workbook",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel workbooks (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    statusKindError,
			Message: fmt.Sprintf("could not open file dialog: %v", err),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if strings.TrimSpace(path) == "" {
		return a.State()
	}

	return a.OpenWorkbookPath(path)
}

// OpenWorkbookPath opens one .xlsx workbook from a filesystem path.
func (a *App) OpenWorkbookPath(path string) AppState {
	a.mu.Lock()
	a.state.Status = AppStatus{Kind: statusKindLoading, Message: loadingWorkbookMessage, Busy: true}
	a.mu.Unlock()

	// Workbook I/O can be slow, so don't block state reads while loading.
	workbook, view, err := loadWorkbookPath(path)
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: statusKindError, Message: err.Error(), Busy: false}

		return cloneAppState(a.state)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Replacing workbook/view should not reset the runtime appearance choice.
	a.state = AppState{
		Workbook:   workbook,
		View:       view,
		Status:     AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false},
		Appearance: normalizeAppearanceState(a.state.Appearance),
	}

	return cloneAppState(a.state)
}

// OpenDroppedFiles opens exactly one dropped .xlsx file path.
func (a *App) OpenDroppedFiles(paths []string) AppState {
	if len(paths) == 0 {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: statusKindError, Message: "drop one .xlsx workbook to open", Busy: false}

		return cloneAppState(a.state)
	}

	if len(paths) > 1 {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    statusKindError,
			Message: "only one .xlsx workbook can be opened at a time",
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	return a.OpenWorkbookPath(paths[0])
}

func loadWorkbookPath(path string) (WorkbookState, WorkbookViewState, error) {
	workbookPath, err := validateWorkbookPath(path)
	if err != nil {
		return WorkbookState{}, WorkbookViewState{}, err
	}

	file, err := excelize.OpenFile(workbookPath)
	if err != nil {
		return WorkbookState{}, WorkbookViewState{}, fmt.Errorf("could not read workbook: %w", err)
	}
	defer file.Close()

	return loadWorkbookFile(file, workbookPath)
}

func validateWorkbookPath(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", errors.New("choose a .xlsx workbook to open")
	}

	absolutePath, err := filepath.Abs(trimmedPath)
	if err == nil {
		// Store a stable file identity when the path can be resolved cleanly.
		trimmedPath = absolutePath
	}

	info, err := os.Stat(trimmedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("workbook path does not exist: %s", trimmedPath)
		}

		return "", fmt.Errorf("could not inspect workbook path %s: %w", trimmedPath, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("workbook path is a directory: %s", trimmedPath)
	}

	if !strings.EqualFold(filepath.Ext(trimmedPath), workbookFileExtension) {
		return "", errors.New("unsupported file type: only .xlsx workbooks can be opened")
	}

	return trimmedPath, nil
}

func loadWorkbookFile(file *excelize.File, path string) (WorkbookState, WorkbookViewState, error) {
	sheetNames := file.GetSheetList()
	if len(sheetNames) == 0 {
		return WorkbookState{}, WorkbookViewState{}, errors.New("workbook does not contain any sheets")
	}

	// Include the default style so renderers always have a fallback style entry.
	styleIDs := map[int]struct{}{0: {}}
	sheets := make([]WorkbookSheet, 0, len(sheetNames))
	for index, sheetName := range sheetNames {
		sheet, sheetStyleIDs, err := loadWorkbookSheet(file, sheetName, index)
		if err != nil {
			return WorkbookState{}, WorkbookViewState{}, fmt.Errorf("load sheet %q: %w", sheetName, err)
		}

		for styleID := range sheetStyleIDs {
			styleIDs[styleID] = struct{}{}
		}
		sheets = append(sheets, sheet)
	}

	styles, err := loadCellStyles(file, styleIDs)
	if err != nil {
		return WorkbookState{}, WorkbookViewState{}, err
	}

	fileName := filepath.Base(path)
	workbook := WorkbookState{
		HasWorkbook: true,
		Title:       fileName,
		FilePath:    path,
		FileName:    fileName,
		Sheets:      sheets,
		Styles:      styles,
	}

	return workbook, loadedWorkbookView(activeSheetName(file, sheetNames)), nil
}
