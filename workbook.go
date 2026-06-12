package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
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
	if !a.guardDestructiveTransition(false) {
		return a.State()
	}

	a.mu.RLock()
	ctx := a.ctx
	a.mu.RUnlock()
	// Open the native dialog without holding the app-state lock.

	if ctx == nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: "file dialog is not available yet", Busy: false}

		return cloneAppState(a.state)
	}

	path, err := a.runOpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "Open .xlsx workbook",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel workbooks (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
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
	a.state.Status = AppStatus{Kind: AppStatusKindLoading, Message: loadingWorkbookMessage, Busy: true}
	a.mu.Unlock()

	// Workbook I/O can be slow, so don't block state reads while loading.
	workbook, view, err := loadWorkbookPath(path)
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: err.Error(), Busy: false}

		return cloneAppState(a.state)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	appearance := normalizeAppearanceState(a.state.Appearance)
	for i := range workbook.Styles {
		workbook.Styles[i].render(appearance.EffectiveTheme)
	}

	// Replacing workbook/view should not reset the runtime appearance choice.
	a.state = AppState{
		Workbook:   workbook,
		View:       view,
		Status:     AppStatus{Kind: AppStatusKindReady, Message: defaultStatusMessage, Busy: false},
		Appearance: appearance,
	}
	a.pendingCellEdits = map[string]map[string]string{}
	a.pendingLayoutEdits = pendingLayoutEdits{
		ColumnWidths: map[string]map[int]float64{},
		RowHeights:   map[string]map[int]float64{},
	}

	return cloneAppState(a.state)
}

// OpenDroppedFiles opens exactly one dropped .xlsx file path.
func (a *App) OpenDroppedFiles(paths []string) AppState {
	if len(paths) == 0 {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: "drop one .xlsx workbook to open", Busy: false}

		return cloneAppState(a.state)
	}

	if len(paths) > 1 {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: "only one .xlsx workbook can be opened at a time",
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if !a.guardDestructiveTransition(false) {
		return a.State()
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

	// Style 0 is the workbook default and intentionally falls through to app/grid CSS.
	styleIDs := map[int]struct{}{}
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
		Title:    fileName,
		FilePath: path,
		FileName: fileName,
		Dirty:    false,
		Sheets:   sheets,
		Styles:   styles,
	}

	return workbook, loadedWorkbookView(activeSheetName(file, sheetNames)), nil
}

// SaveWorkbook saves the current workbook. Untitled workbooks route through Save As.
func (a *App) SaveWorkbook() AppState {
	a.mu.RLock()
	filePath := a.state.Workbook.FilePath
	dirty := a.state.Workbook.Dirty
	a.mu.RUnlock()

	if strings.TrimSpace(filePath) == "" {
		return a.SaveWorkbookAs()
	}

	if !dirty {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: savedStatusMessage, Busy: false}

		return cloneAppState(a.state)
	}

	return a.saveWorkbookToPath(filePath, false)
}

// SaveWorkbookAs prompts for a target path and saves the current workbook there.
func (a *App) SaveWorkbookAs() AppState {
	a.mu.RLock()
	ctx := a.ctx
	workbook := a.state.Workbook
	a.mu.RUnlock()

	if ctx == nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: "save dialog is not available yet", Busy: false}

		return cloneAppState(a.state)
	}

	path, err := a.runSaveFileDialog(ctx, saveDialogOptions(workbook))
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("could not open save dialog: %v", err),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	if strings.TrimSpace(path) == "" {
		return a.State()
	}

	return a.saveWorkbookToPath(path, true)
}

func (a *App) saveWorkbookToPath(path string, updateIdentity bool) AppState {
	targetPath, err := normalizeSavePath(path)
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{Kind: AppStatusKindError, Message: err.Error(), Busy: false}

		return cloneAppState(a.state)
	}

	a.mu.Lock()
	if a.state.Status.Kind == "" {
		a.state = initialAppState()
	}
	// Save I/O runs without the app lock, so snapshot mutable state first.
	workbook := cloneAppState(a.state).Workbook
	pendingEdits := make(map[string]map[string]string, len(a.pendingCellEdits))
	for sheetName, sheetEdits := range a.pendingCellEdits {
		pendingEdits[sheetName] = maps.Clone(sheetEdits)
	}
	layoutEdits := a.pendingLayoutEdits.clone()
	a.state.Status = AppStatus{Kind: AppStatusKindLoading, Message: savingWorkbookMessage, Busy: true}
	a.mu.Unlock()

	savedPath, err := saveWorkbookFile(workbook, pendingEdits, layoutEdits, targetPath)
	if err != nil {
		a.mu.Lock()
		defer a.mu.Unlock()

		a.state.Status = AppStatus{
			Kind:    AppStatusKindError,
			Message: fmt.Sprintf("could not save workbook: %v", err),
			Busy:    false,
		}

		return cloneAppState(a.state)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if updateIdentity {
		a.state.Workbook.FilePath = savedPath
		a.state.Workbook.FileName = filepath.Base(savedPath)
		a.state.Workbook.Title = a.state.Workbook.FileName
	}
	a.state.Workbook.Dirty = false
	a.pendingCellEdits = map[string]map[string]string{}
	a.pendingLayoutEdits = pendingLayoutEdits{
		ColumnWidths: map[string]map[int]float64{},
		RowHeights:   map[string]map[int]float64{},
	}
	a.state.Status = AppStatus{Kind: AppStatusKindReady, Message: savedStatusMessage, Busy: false}

	return cloneAppState(a.state)
}

func saveWorkbookFile(
	workbook WorkbookState,
	pendingEdits map[string]map[string]string,
	pendingLayoutEdits pendingLayoutEdits,
	targetPath string,
) (string, error) {
	if strings.TrimSpace(workbook.FilePath) != "" {
		return saveOpenedWorkbook(workbook.FilePath, pendingEdits, pendingLayoutEdits, targetPath)
	}

	return saveUntitledWorkbook(workbook, targetPath)
}

func saveOpenedWorkbook(
	sourcePath string,
	pendingEdits map[string]map[string]string,
	pendingLayoutEdits pendingLayoutEdits,
	targetPath string,
) (string, error) {
	workbookPath, err := validateWorkbookPath(sourcePath)
	if err != nil {
		return "", err
	}

	// Reopen the source file so untouched workbook XML survives the save.
	file, err := excelize.OpenFile(workbookPath)
	if err != nil {
		return "", fmt.Errorf("could not read workbook: %w", err)
	}
	defer file.Close()

	if err := applyPendingCellEdits(file, pendingEdits); err != nil {
		return "", err
	}
	if err := applyPendingColumnWidths(file, pendingLayoutEdits.ColumnWidths); err != nil {
		return "", err
	}
	if err := applyPendingRowHeights(file, pendingLayoutEdits.RowHeights); err != nil {
		return "", err
	}

	if sameWorkbookPath(workbookPath, targetPath) {
		if err := file.Save(); err != nil {
			return "", err
		}

		return targetPath, nil
	}

	if err := file.SaveAs(targetPath); err != nil {
		return "", err
	}

	return targetPath, nil
}

func saveUntitledWorkbook(workbook WorkbookState, targetPath string) (string, error) {
	file := excelize.NewFile()
	defer file.Close()

	// Untitled saves intentionally write only current non-empty text cells and dimensions.
	sheets := workbook.Sheets
	if len(sheets) == 0 {
		sheets = []WorkbookSheet{{Name: defaultSheetName}}
	}

	defaultSheet := file.GetSheetName(file.GetActiveSheetIndex())
	for index, sheet := range sheets {
		sheetName := strings.TrimSpace(sheet.Name)
		if sheetName == "" {
			sheetName = defaultSheetName
		}

		if index == 0 && defaultSheet != sheetName {
			if err := file.SetSheetName(defaultSheet, sheetName); err != nil {
				return "", err
			}
		}
		if index > 0 {
			if _, err := file.NewSheet(sheetName); err != nil {
				return "", err
			}
		}

		for _, cell := range sheet.Cells {
			value := cell.RawValue
			if value == "" {
				value = cell.Value
			}
			if value == "" {
				continue
			}
			if err := file.SetCellStr(sheetName, cell.Ref, value); err != nil {
				return "", err
			}
		}
		if err := applySheetLayouts(file, sheetName, sheet); err != nil {
			return "", err
		}
	}

	if err := file.SaveAs(targetPath); err != nil {
		return "", err
	}

	return targetPath, nil
}

func applyPendingCellEdits(file *excelize.File, pendingEdits map[string]map[string]string) error {
	sheetNames := make([]string, 0, len(pendingEdits))
	for sheetName := range pendingEdits {
		sheetNames = append(sheetNames, sheetName)
	}
	// Map iteration order is random; stable write order keeps errors deterministic.
	slices.Sort(sheetNames)

	for _, sheetName := range sheetNames {
		sheetIndex, err := file.GetSheetIndex(sheetName)
		if err != nil {
			return err
		}
		if sheetIndex < 0 {
			return fmt.Errorf("sheet %q was not found", sheetName)
		}

		cellRefs := make([]string, 0, len(pendingEdits[sheetName]))
		for cellRef := range pendingEdits[sheetName] {
			cellRefs = append(cellRefs, cellRef)
		}
		slices.Sort(cellRefs)

		for _, cellRef := range cellRefs {
			if err := file.SetCellStr(sheetName, cellRef, pendingEdits[sheetName][cellRef]); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyPendingColumnWidths(file *excelize.File, pendingWidths map[string]map[int]float64) error {
	sheetNames := make([]string, 0, len(pendingWidths))
	for sheetName := range pendingWidths {
		sheetNames = append(sheetNames, sheetName)
	}
	// Map iteration order is random; stable write order keeps errors deterministic.
	slices.Sort(sheetNames)

	for _, sheetName := range sheetNames {
		sheetIndex, err := file.GetSheetIndex(sheetName)
		if err != nil {
			return err
		}
		if sheetIndex < 0 {
			return fmt.Errorf("sheet %q was not found", sheetName)
		}

		columnIndexes := make([]int, 0, len(pendingWidths[sheetName]))
		for columnIndex := range pendingWidths[sheetName] {
			columnIndexes = append(columnIndexes, columnIndex)
		}
		// Keep column application order stable for deterministic save failures.
		slices.Sort(columnIndexes)

		for _, columnIndex := range columnIndexes {
			width := pendingWidths[sheetName][columnIndex]
			if err := setColumnWidth(file, sheetName, columnIndex, width); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyPendingRowHeights(file *excelize.File, pendingHeights map[string]map[int]float64) error {
	sheetNames := make([]string, 0, len(pendingHeights))
	for sheetName := range pendingHeights {
		sheetNames = append(sheetNames, sheetName)
	}
	// Map iteration order is random; stable write order keeps errors deterministic.
	slices.Sort(sheetNames)

	for _, sheetName := range sheetNames {
		sheetIndex, err := file.GetSheetIndex(sheetName)
		if err != nil {
			return err
		}
		if sheetIndex < 0 {
			return fmt.Errorf("sheet %q was not found", sheetName)
		}

		rowIndexes := make([]int, 0, len(pendingHeights[sheetName]))
		for rowIndex := range pendingHeights[sheetName] {
			rowIndexes = append(rowIndexes, rowIndex)
		}
		// Keep row application order stable for deterministic save failures.
		slices.Sort(rowIndexes)

		for _, rowIndex := range rowIndexes {
			height := pendingHeights[sheetName][rowIndex]
			if err := setRowHeight(file, sheetName, rowIndex, height); err != nil {
				return err
			}
		}
	}

	return nil
}

func applySheetLayouts(file *excelize.File, sheetName string, sheet WorkbookSheet) error {
	// Untitled workbooks save from in-memory layout slices, so sort copies before writing.
	columns := slices.Clone(sheet.Columns)
	slices.SortFunc(columns, func(left ColumnLayout, right ColumnLayout) int {
		return left.Index - right.Index
	})
	for _, column := range columns {
		if column.Hidden || !validLayoutDimension(column.Width) {
			continue
		}
		if err := setColumnWidth(file, sheetName, column.Index, column.Width); err != nil {
			return err
		}
	}

	rows := slices.Clone(sheet.Rows)
	slices.SortFunc(rows, func(left RowLayout, right RowLayout) int {
		return left.Index - right.Index
	})
	for _, row := range rows {
		if row.Hidden || !validLayoutDimension(row.Height) {
			continue
		}
		if err := setRowHeight(file, sheetName, row.Index, row.Height); err != nil {
			return err
		}
	}

	return nil
}

func setColumnWidth(file *excelize.File, sheetName string, columnIndex int, width float64) error {
	if columnIndex < minExcelColumn || columnIndex > maxExcelColumn {
		return fmt.Errorf("column index must be between %d and %d", minExcelColumn, maxExcelColumn)
	}
	if !validLayoutDimension(width) {
		return errors.New("column width must be a positive finite number")
	}

	columnName, err := excelize.ColumnNumberToName(columnIndex)
	if err != nil {
		return err
	}

	return file.SetColWidth(sheetName, columnName, columnName, width)
}

func setRowHeight(file *excelize.File, sheetName string, rowIndex int, height float64) error {
	if rowIndex < minExcelRow || rowIndex > maxExcelRow {
		return fmt.Errorf("row index must be between %d and %d", minExcelRow, maxExcelRow)
	}
	if !validLayoutDimension(height) {
		return errors.New("row height must be a positive finite number")
	}

	return file.SetRowHeight(sheetName, rowIndex, height)
}

func normalizeSavePath(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", errors.New("choose a .xlsx path to save")
	}

	extension := filepath.Ext(trimmedPath)
	if extension == "" {
		trimmedPath += workbookFileExtension
	} else if !strings.EqualFold(extension, workbookFileExtension) {
		return "", errors.New("unsupported file type: only .xlsx workbooks can be saved")
	}

	absolutePath, err := filepath.Abs(trimmedPath)
	if err == nil {
		trimmedPath = absolutePath
	}

	info, err := os.Stat(trimmedPath)
	if err == nil && info.IsDir() {
		return "", fmt.Errorf("save path is a directory: %s", trimmedPath)
	}
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("could not inspect save path %s: %w", trimmedPath, err)
	}

	return trimmedPath, nil
}

func saveDialogOptions(workbook WorkbookState) runtime.SaveDialogOptions {
	fileName := strings.TrimSpace(workbook.FileName)
	if fileName == "" {
		fileName = defaultWorkbookTitle + workbookFileExtension
	}

	options := runtime.SaveDialogOptions{
		Title:                      "Save .xlsx workbook",
		DefaultFilename:            fileName,
		CanCreateDirectories:       true,
		TreatPackagesAsDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel workbooks (*.xlsx)", Pattern: "*.xlsx"},
		},
	}
	if strings.TrimSpace(workbook.FilePath) != "" {
		options.DefaultDirectory = filepath.Dir(workbook.FilePath)
	}

	return options
}

func sameWorkbookPath(left string, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	if leftErr == nil {
		left = leftAbs
	}
	rightAbs, rightErr := filepath.Abs(right)
	if rightErr == nil {
		right = rightAbs
	}

	return filepath.Clean(left) == filepath.Clean(right)
}
