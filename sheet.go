package main

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	excelizeDefaultColumnWidth = 9.140625
	layoutDimensionTolerance   = 0.000001
)

// SheetState describes workbook sheet visibility state.
type SheetState string

const (
	// SheetStateVisible marks a visible worksheet.
	SheetStateVisible SheetState = "visible"
	// SheetStateHidden marks a hidden worksheet.
	SheetStateHidden SheetState = "hidden"
	// SheetStateVeryHidden marks a worksheet hidden from Excel's normal UI.
	SheetStateVeryHidden SheetState = "veryHidden"
)

// AllSheetStates lists SheetState values for Wails enum binding.
var AllSheetStates = []struct {
	Value  SheetState
	TSName string
}{
	{SheetStateVisible, "Visible"},
	{SheetStateHidden, "Hidden"},
	{SheetStateVeryHidden, "VeryHidden"},
}

// WorkbookSheet describes one worksheet and the loaded data for it.
type WorkbookSheet struct {
	Index              int               `json:"index"`
	Name               string            `json:"name"`
	State              SheetState        `json:"state"`
	Visible            bool              `json:"visible"`
	Bounds             CellRange         `json:"bounds"`
	DefaultColumnWidth float64           `json:"defaultColumnWidth"`
	DefaultRowHeight   float64           `json:"defaultRowHeight"`
	Cells              []CellData        `json:"cells"`
	MergedCells        []MergedCellRange `json:"mergedCells"`
	Columns            []ColumnLayout    `json:"columns"`
	Rows               []RowLayout       `json:"rows"`
}

func (w *WorkbookSheet) setColumnWidth(index int, width float64) (bool, error) {
	if index < minExcelColumn || index > maxExcelColumn {
		return false, fmt.Errorf("column index must be between %d and %d", minExcelColumn, maxExcelColumn)
	}
	if !validLayoutDimension(width) {
		return false, errors.New("column width must be a positive finite number")
	}

	layoutIndex := slices.IndexFunc(w.Columns, func(column ColumnLayout) bool {
		return column.Index == index
	})
	if layoutIndex >= 0 {
		if sameLayoutDimension(w.Columns[layoutIndex].Width, width) {
			return false, nil
		}

		w.Columns[layoutIndex].Width = width
		slices.SortFunc(w.Columns, func(left ColumnLayout, right ColumnLayout) int {
			return left.Index - right.Index
		})

		return true, nil
	}

	if sameLayoutDimension(effectiveDefaultColumnWidth(w.DefaultColumnWidth), width) {
		return false, nil
	}

	name, err := excelize.ColumnNumberToName(index)
	if err != nil {
		return false, err
	}
	w.Columns = append(w.Columns, ColumnLayout{Index: index, Name: name, Width: width})
	slices.SortFunc(w.Columns, func(left ColumnLayout, right ColumnLayout) int {
		return left.Index - right.Index
	})

	return true, nil
}

func (w *WorkbookSheet) setRowHeight(index int, height float64) (bool, error) {
	if index < minExcelRow || index > maxExcelRow {
		return false, fmt.Errorf("row index must be between %d and %d", minExcelRow, maxExcelRow)
	}
	if !validLayoutDimension(height) {
		return false, errors.New("row height must be a positive finite number")
	}

	layoutIndex := slices.IndexFunc(w.Rows, func(row RowLayout) bool {
		return row.Index == index
	})
	if layoutIndex >= 0 {
		if sameLayoutDimension(w.Rows[layoutIndex].Height, height) {
			return false, nil
		}

		w.Rows[layoutIndex].Height = height
		slices.SortFunc(w.Rows, func(left RowLayout, right RowLayout) int {
			return left.Index - right.Index
		})

		return true, nil
	}

	if sameLayoutDimension(effectiveDefaultRowHeight(w.DefaultRowHeight), height) {
		return false, nil
	}

	w.Rows = append(w.Rows, RowLayout{Index: index, Height: height})
	slices.SortFunc(w.Rows, func(left RowLayout, right RowLayout) int {
		return left.Index - right.Index
	})

	return true, nil
}

func validLayoutDimension(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func sameLayoutDimension(left float64, right float64) bool {
	return math.Abs(left-right) < layoutDimensionTolerance
}

func effectiveDefaultColumnWidth(width float64) float64 {
	if validLayoutDimension(width) {
		return width
	}

	return defaultColumnWidth
}

func effectiveDefaultRowHeight(height float64) float64 {
	if validLayoutDimension(height) {
		return height
	}

	return defaultRowHeight
}

func (w *WorkbookSheet) setCellValue(address CellAddress, value string) (bool, error) {
	if value == "" {
		return w.clearCellValue(address), nil
	}

	expandedBounds, err := expandedSheetBounds(w.Bounds, address)
	if err != nil {
		return false, err
	}

	index := slices.IndexFunc(w.Cells, func(cell CellData) bool {
		return cell.Ref == address.Ref
	})
	if index >= 0 {
		oldCell := w.Cells[index]
		nextCell := oldCell
		nextCell.Ref = address.Ref
		nextCell.Row = address.Row
		nextCell.Column = address.Column
		nextCell.Value = value
		nextCell.RawValue = value
		nextCell.Formula = ""
		nextCell.HasFormula = false
		nextCell.Kind = CellKindString

		// Formula metadata and bounds can make an edit meaningful even when text matches.
		changed := nextCell != oldCell || expandedBounds != w.Bounds
		if !changed {
			return false, nil
		}

		w.Cells[index] = nextCell
		w.Bounds = expandedBounds
		slices.SortFunc(w.Cells, func(left CellData, right CellData) int {
			if left.Row != right.Row {
				return left.Row - right.Row
			}

			return left.Column - right.Column
		})

		return true, nil
	}

	w.Cells = append(w.Cells, CellData{
		Ref:      address.Ref,
		Row:      address.Row,
		Column:   address.Column,
		Value:    value,
		RawValue: value,
		Kind:     CellKindString,
	})
	w.Bounds = expandedBounds
	slices.SortFunc(w.Cells, func(left CellData, right CellData) int {
		if left.Row != right.Row {
			return left.Row - right.Row
		}

		return left.Column - right.Column
	})

	return true, nil
}

func (w *WorkbookSheet) clearCellValue(address CellAddress) bool {
	index := slices.IndexFunc(w.Cells, func(cell CellData) bool {
		return cell.Ref == address.Ref
	})
	if index < 0 {
		return false
	}

	oldCell := w.Cells[index]
	nextCell := oldCell
	nextCell.Ref = address.Ref
	nextCell.Row = address.Row
	nextCell.Column = address.Column
	nextCell.Value = ""
	nextCell.RawValue = ""
	nextCell.Formula = ""
	nextCell.HasFormula = false
	nextCell.Kind = CellKindUnset

	// Styled blanks stay in the sparse model so clears do not drop formatting.
	if nextCell.StyleID == 0 {
		w.Cells = append(w.Cells[:index], w.Cells[index+1:]...)

		return true
	}

	if nextCell == oldCell {
		return false
	}

	w.Cells[index] = nextCell
	slices.SortFunc(w.Cells, func(left CellData, right CellData) int {
		if left.Row != right.Row {
			return left.Row - right.Row
		}

		return left.Column - right.Column
	})

	return true
}

func loadWorkbookSheet(file *excelize.File, sheetName string, index int) (WorkbookSheet, map[int]struct{}, error) {
	styleIDs := map[int]struct{}{}
	bounds := cellBounds{}
	// A1 keeps empty sheets renderable even when Excelize reports no used cells.
	bounds.addRange(a1Range())

	dimensionRange, err := sheetDimensionRange(file, sheetName)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}
	bounds.addRange(dimensionRange)

	formattedRows, err := file.GetRows(sheetName)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}

	// Keep raw values beside Excelize's formatted/display values.
	rawRows, err := file.GetRows(sheetName, excelize.Options{RawCellValue: true})
	if err != nil {
		return WorkbookSheet{}, nil, err
	}

	cells, cellStyleIDs, loadedCellBounds, err := loadSheetCells(file, sheetName, formattedRows, rawRows)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}
	if loadedCellBounds.hasValues {
		bounds.addBounds(loadedCellBounds)
	}
	for styleID := range cellStyleIDs {
		styleIDs[styleID] = struct{}{}
	}

	mergedCells, mergedBounds, err := loadMergedCells(file, sheetName)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}
	if mergedBounds.hasValues {
		bounds.addBounds(mergedBounds)
	}

	finalBounds, err := bounds.rangeValue()
	if err != nil {
		return WorkbookSheet{}, nil, err
	}

	defaultColumnWidth, defaultRowHeight := sheetDefaults(file, sheetName)
	// Public Excelize layout APIs are per-row/per-column, so scan only useful bounds.
	columns, columnStyleIDs, err := loadColumnLayouts(
		file,
		sheetName,
		finalBounds.End.Column,
		defaultColumnWidth,
	)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}
	for styleID := range columnStyleIDs {
		styleIDs[styleID] = struct{}{}
	}

	rows, err := loadRowLayouts(file, sheetName, finalBounds.End.Row, defaultRowHeight)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}

	cells, cellStyleIDs, err = loadStyleOnlyCells(file, sheetName, cells, finalBounds)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}
	for styleID := range cellStyleIDs {
		styleIDs[styleID] = struct{}{}
	}
	sort.Slice(cells, func(i int, j int) bool {
		return cells[i].Row < cells[j].Row ||
			cells[i].Row == cells[j].Row && cells[i].Column < cells[j].Column
	})

	state, visible, err := sheetVisibility(file, sheetName)
	if err != nil {
		return WorkbookSheet{}, nil, err
	}

	return WorkbookSheet{
		Index:              index,
		Name:               sheetName,
		State:              state,
		Visible:            visible,
		Bounds:             finalBounds,
		DefaultColumnWidth: defaultColumnWidth,
		DefaultRowHeight:   defaultRowHeight,
		Cells:              cells,
		MergedCells:        mergedCells,
		Columns:            columns,
		Rows:               rows,
	}, styleIDs, nil
}

func loadMergedCells(file *excelize.File, sheetName string) ([]MergedCellRange, cellBounds, error) {
	bounds := cellBounds{}
	excelMergedCells, err := file.GetMergeCells(sheetName)
	if err != nil {
		return nil, bounds, err
	}

	mergedCells := make([]MergedCellRange, 0, len(excelMergedCells))
	for _, excelMergedCell := range excelMergedCells {
		startRef := excelMergedCell.GetStartAxis()
		endRef := excelMergedCell.GetEndAxis()
		if strings.TrimSpace(endRef) == "" {
			endRef = startRef
		}

		cellRange, err := parseCellRangeRef(startRef + ":" + endRef)
		if err != nil {
			return nil, bounds, err
		}

		bounds.addRange(cellRange)
		mergedCells = append(mergedCells, MergedCellRange{Range: cellRange, Value: excelMergedCell.GetCellValue()})
	}

	return mergedCells, bounds, nil
}

func loadColumnLayouts(
	file *excelize.File,
	sheetName string,
	maxColumn int,
	defaultWidth float64,
) ([]ColumnLayout, map[int]struct{}, error) {
	styleIDs := map[int]struct{}{}
	layouts := make([]ColumnLayout, 0)

	for column := minExcelColumn; column <= maxColumn; column++ {
		name, err := excelize.ColumnNumberToName(column)
		if err != nil {
			return nil, nil, err
		}

		width, err := file.GetColWidth(sheetName, name)
		if err != nil {
			return nil, nil, err
		}

		visible, err := file.GetColVisible(sheetName, name)
		if err != nil {
			return nil, nil, err
		}

		outlineLevel, err := file.GetColOutlineLevel(sheetName, name)
		if err != nil {
			return nil, nil, err
		}

		styleID, err := file.GetColStyle(sheetName, name)
		if err != nil {
			return nil, nil, err
		}
		if styleID > 0 {
			styleIDs[styleID] = struct{}{}
		}

		hidden := !visible
		// Skip default columns instead of serializing the whole sheet width.
		if sameLayoutDimension(width, defaultWidth) && !hidden && outlineLevel == 0 && styleID == 0 {
			continue
		}

		layouts = append(layouts, ColumnLayout{
			Index:        column,
			Name:         name,
			Width:        width,
			Hidden:       hidden,
			OutlineLevel: int(outlineLevel),
			StyleID:      styleID,
		})
	}

	return layouts, styleIDs, nil
}

func loadRowLayouts(
	file *excelize.File,
	sheetName string,
	maxRow int,
	defaultHeight float64,
) ([]RowLayout, error) {
	layouts := make([]RowLayout, 0)

	for row := minExcelRow; row <= maxRow; row++ {
		height, err := file.GetRowHeight(sheetName, row)
		if err != nil {
			return nil, err
		}

		visible, err := file.GetRowVisible(sheetName, row)
		if err != nil {
			return nil, err
		}

		outlineLevel, err := file.GetRowOutlineLevel(sheetName, row)
		if err != nil {
			return nil, err
		}

		hidden := !visible
		// Skip default rows instead of serializing every row in the useful bounds.
		if sameLayoutDimension(height, defaultHeight) && !hidden && outlineLevel == 0 {
			continue
		}

		layouts = append(layouts, RowLayout{
			Index:        row,
			Height:       height,
			Hidden:       hidden,
			OutlineLevel: int(outlineLevel),
		})
	}

	return layouts, nil
}

func sheetDimensionRange(file *excelize.File, sheetName string) (CellRange, error) {
	dimension, err := file.GetSheetDimension(sheetName)
	if err != nil {
		return CellRange{}, err
	}
	if strings.TrimSpace(dimension) == "" {
		return a1Range(), nil
	}

	return parseCellRangeRef(dimension)
}

func activeSheetName(file *excelize.File, sheetNames []string) string {
	activeSheetIndex := file.GetActiveSheetIndex()
	if activeSheetIndex >= 0 {
		if sheetName := file.GetSheetName(activeSheetIndex); sheetName != "" {
			return sheetName
		}
	}

	return sheetNames[0]
}

func loadedWorkbookView(activeSheetName string) WorkbookViewState {
	selection := a1Range()

	return WorkbookViewState{
		ActiveSheetName: activeSheetName,
		ActiveCell:      selection.Start,
		Selection:       selection,
		ZoomPercent:     defaultZoomPercent,
		Scroll: ScrollPosition{
			TopRow:     minExcelRow,
			LeftColumn: minExcelColumn,
		},
	}
}

func sheetDefaults(file *excelize.File, sheetName string) (float64, float64) {
	defaultColWidth := excelizeDefaultColumnWidth
	defaultRowHeightValue := defaultRowHeight

	props, err := file.GetSheetProps(sheetName)
	if err != nil {
		return defaultColWidth, defaultRowHeightValue
	}

	if props.DefaultColWidth != nil && *props.DefaultColWidth > 0 {
		defaultColWidth = *props.DefaultColWidth
	}
	if props.DefaultRowHeight != nil && *props.DefaultRowHeight > 0 {
		defaultRowHeightValue = *props.DefaultRowHeight
	}

	return defaultColWidth, defaultRowHeightValue
}

func sheetVisibility(file *excelize.File, sheetName string) (SheetState, bool, error) {
	visible, err := file.GetSheetVisible(sheetName)
	if err != nil {
		return "", false, err
	}

	state := normalizeSheetState(workbookSheetState(file, sheetName), visible)

	return state, visible, nil
}

func normalizeSheetState(state string, visible bool) SheetState {
	switch SheetState(state) {
	case SheetStateVisible:
		return SheetStateVisible
	case SheetStateHidden:
		return SheetStateHidden
	case SheetStateVeryHidden:
		return SheetStateVeryHidden
	default:
		if visible {
			return SheetStateVisible
		}

		return SheetStateHidden
	}
}

func workbookSheetState(file *excelize.File, sheetName string) string {
	if file.WorkBook == nil {
		return ""
	}

	for _, sheet := range file.WorkBook.Sheets.Sheet {
		if sheet.Name == sheetName {
			return sheet.State
		}
	}

	return ""
}
