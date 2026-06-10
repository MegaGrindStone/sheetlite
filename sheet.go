package main

import (
	"math"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

const excelizeDefaultColumnWidth = 9.140625

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
		if math.Abs(width-defaultWidth) < 0.000001 && !hidden && outlineLevel == 0 && styleID == 0 {
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
		if math.Abs(height-defaultHeight) < 0.000001 && !hidden && outlineLevel == 0 {
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

func sheetVisibility(file *excelize.File, sheetName string) (string, bool, error) {
	visible, err := file.GetSheetVisible(sheetName)
	if err != nil {
		return "", false, err
	}

	// Excelize exposes visibility as bool; keep the workbook state string when available.
	state := workbookSheetState(file, sheetName)
	if state == "" {
		if visible {
			state = sheetStateVisible
		} else {
			state = "hidden"
		}
	}

	return state, visible, nil
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
