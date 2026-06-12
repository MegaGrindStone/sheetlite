package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const styleOnlyCellScanLimit = 100_000

// CellKind describes the semantic kind Excelize reports for a loaded cell.
type CellKind string

const (
	// CellKindUnset marks cells with no concrete Excel cell kind, such as styled blanks.
	CellKindUnset CellKind = "unset"
	// CellKindBool marks boolean cells.
	CellKindBool CellKind = "bool"
	// CellKindDate marks date cells.
	CellKindDate CellKind = "date"
	// CellKindError marks formula/error cells.
	CellKindError CellKind = "error"
	// CellKindFormula marks formula cells.
	CellKindFormula CellKind = "formula"
	// CellKindInlineString marks inline string cells.
	CellKindInlineString CellKind = "inlineString"
	// CellKindNumber marks numeric cells.
	CellKindNumber CellKind = "number"
	// CellKindString marks shared-string and literal edit cells.
	CellKindString CellKind = "string"
)

// AllCellKinds lists CellKind values for Wails enum binding.
var AllCellKinds = []struct {
	Value  CellKind
	TSName string
}{
	{CellKindUnset, "Unset"},
	{CellKindBool, "Bool"},
	{CellKindDate, "Date"},
	{CellKindError, "Error"},
	{CellKindFormula, "Formula"},
	{CellKindInlineString, "InlineString"},
	{CellKindNumber, "Number"},
	{CellKindString, "String"},
}

func parseCellAddress(ref string) (CellAddress, bool) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return CellAddress{}, false
	}

	// Excel refs are column letters followed by row digits, with no separator.
	split := strings.IndexFunc(trimmed, func(r rune) bool {
		return r >= '0' && r <= '9'
	})
	if split <= 0 || split == len(trimmed) {
		return CellAddress{}, false
	}

	letters := strings.ToUpper(trimmed[:split])
	digits := trimmed[split:]
	if !isColumnLetters(letters) || !isRowDigits(digits) {
		return CellAddress{}, false
	}

	column, ok := columnLettersToNumber(letters)
	if !ok {
		return CellAddress{}, false
	}

	row, err := strconv.Atoi(digits)
	if err != nil || row < minExcelRow || row > maxExcelRow {
		return CellAddress{}, false
	}

	return CellAddress{Ref: letters + strconv.Itoa(row), Row: row, Column: column}, true
}

func loadSheetCells(
	file *excelize.File,
	sheetName string,
	formattedRows [][]string,
	rawRows [][]string,
) ([]CellData, map[int]struct{}, cellBounds, error) {
	styleIDs := map[int]struct{}{}
	bounds := cellBounds{}
	rowCount := max(len(formattedRows), len(rawRows))
	cells := make([]CellData, 0)

	for rowIndex := range rowCount {
		columnCount := max(len(rowValues(formattedRows, rowIndex)), len(rowValues(rawRows, rowIndex)))
		for columnIndex := range columnCount {
			formattedValue := valueAt(formattedRows, rowIndex, columnIndex)
			rawValue := valueAt(rawRows, rowIndex, columnIndex)
			cell, include, err := loadCell(file, sheetName, rowIndex+1, columnIndex+1, formattedValue, rawValue)
			if err != nil {
				return nil, nil, cellBounds{}, err
			}
			if !include {
				continue
			}

			if cell.StyleID > 0 {
				styleIDs[cell.StyleID] = struct{}{}
			}
			bounds.addAddress(CellAddress{Ref: cell.Ref, Row: cell.Row, Column: cell.Column})
			cells = append(cells, cell)
		}
	}

	return cells, styleIDs, bounds, nil
}

func loadStyleOnlyCells(
	file *excelize.File,
	sheetName string,
	cells []CellData,
	bounds CellRange,
) ([]CellData, map[int]struct{}, error) {
	styleIDs := map[int]struct{}{}
	rowCount := bounds.End.Row - bounds.Start.Row + 1
	columnCount := bounds.End.Column - bounds.Start.Column + 1
	// Style-only cells require probing; cap this to avoid huge blank-sheet scans.
	if rowCount <= 0 || columnCount <= 0 || rowCount*columnCount > styleOnlyCellScanLimit {
		return cells, styleIDs, nil
	}

	seen := make(map[string]struct{}, len(cells))
	for _, cell := range cells {
		seen[cell.Ref] = struct{}{}
	}

	for row := bounds.Start.Row; row <= bounds.End.Row; row++ {
		for column := bounds.Start.Column; column <= bounds.End.Column; column++ {
			ref, err := excelize.CoordinatesToCellName(column, row)
			if err != nil {
				return nil, nil, err
			}
			if _, ok := seen[ref]; ok {
				continue
			}

			styleID, err := file.GetCellStyle(sheetName, ref)
			if err != nil {
				return nil, nil, err
			}
			if styleID == 0 {
				continue
			}

			formattedValue, err := file.GetCellValue(sheetName, ref)
			if err != nil {
				return nil, nil, err
			}
			rawValue, err := file.GetCellValue(sheetName, ref, excelize.Options{RawCellValue: true})
			if err != nil {
				return nil, nil, err
			}

			cell, include, err := loadCell(file, sheetName, row, column, formattedValue, rawValue)
			if err != nil {
				return nil, nil, err
			}
			if !include {
				continue
			}

			styleIDs[cell.StyleID] = struct{}{}
			seen[ref] = struct{}{}
			cells = append(cells, cell)
		}
	}

	return cells, styleIDs, nil
}

func loadCell(
	file *excelize.File,
	sheetName string,
	row int,
	column int,
	value string,
	rawValue string,
) (CellData, bool, error) {
	ref, err := excelize.CoordinatesToCellName(column, row)
	if err != nil {
		return CellData{}, false, err
	}

	formula, err := file.GetCellFormula(sheetName, ref)
	if err != nil {
		return CellData{}, false, err
	}

	styleID, err := file.GetCellStyle(sheetName, ref)
	if err != nil {
		return CellData{}, false, err
	}

	cellType, err := file.GetCellType(sheetName, ref)
	if err != nil {
		return CellData{}, false, err
	}

	hasFormula := formula != ""
	kind, hasConcreteKind := cellKind(cellType, hasFormula)
	// Preserve styled and formula cells even when their displayed value is empty.
	include := value != "" || rawValue != "" || hasFormula || styleID != 0 || hasConcreteKind
	if !include {
		return CellData{}, false, nil
	}

	return CellData{
		Ref:        ref,
		Row:        row,
		Column:     column,
		Value:      value,
		RawValue:   rawValue,
		Formula:    formula,
		HasFormula: hasFormula,
		Kind:       kind,
		StyleID:    styleID,
	}, true, nil
}

func expandedSheetBounds(current CellRange, address CellAddress) (CellRange, error) {
	bounds := cellBounds{}
	if current.valid() {
		bounds.addRange(current)
	} else {
		bounds.addRange(a1Range())
	}
	bounds.addAddress(address)

	return bounds.rangeValue()
}

func (c CellRange) valid() bool {
	return c.Start.valid() &&
		c.End.valid() &&
		c.Start.Row <= c.End.Row &&
		c.Start.Column <= c.End.Column
}

func (c CellAddress) valid() bool {
	return c.Row >= minExcelRow && c.Row <= maxExcelRow &&
		c.Column >= minExcelColumn && c.Column <= maxExcelColumn
}

func parseCellRangeRef(ref string) (CellRange, error) {
	parts := strings.Split(strings.TrimSpace(ref), ":")
	if len(parts) == 0 || len(parts) > 2 {
		return CellRange{}, fmt.Errorf("invalid cell range %q", ref)
	}

	start, err := parseExcelCellAddress(parts[0])
	if err != nil {
		return CellRange{}, err
	}

	end := start
	if len(parts) == 2 {
		end, err = parseExcelCellAddress(parts[1])
		if err != nil {
			return CellRange{}, err
		}
	}

	if end.Row < start.Row {
		start.Row, end.Row = end.Row, start.Row
	}
	if end.Column < start.Column {
		start.Column, end.Column = end.Column, start.Column
	}

	return cellRangeFromCoordinates(start.Row, start.Column, end.Row, end.Column)
}

func parseExcelCellAddress(ref string) (CellAddress, error) {
	// Excelize coordinate helpers expect relative references, not $A$1 syntax.
	cleanRef := strings.ReplaceAll(strings.TrimSpace(ref), "$", "")
	column, row, err := excelize.CellNameToCoordinates(cleanRef)
	if err != nil {
		return CellAddress{}, err
	}

	if row < minExcelRow || row > maxExcelRow || column < minExcelColumn || column > maxExcelColumn {
		return CellAddress{}, fmt.Errorf("cell reference %q is outside supported Excel bounds", ref)
	}

	cellRef, err := excelize.CoordinatesToCellName(column, row)
	if err != nil {
		return CellAddress{}, err
	}

	return CellAddress{Ref: cellRef, Row: row, Column: column}, nil
}

func cellRangeFromCoordinates(startRow int, startColumn int, endRow int, endColumn int) (CellRange, error) {
	startRef, err := excelize.CoordinatesToCellName(startColumn, startRow)
	if err != nil {
		return CellRange{}, err
	}

	endRef, err := excelize.CoordinatesToCellName(endColumn, endRow)
	if err != nil {
		return CellRange{}, err
	}

	ref := startRef
	if startRef != endRef {
		ref = startRef + ":" + endRef
	}

	return CellRange{
		Ref:   ref,
		Start: CellAddress{Ref: startRef, Row: startRow, Column: startColumn},
		End:   CellAddress{Ref: endRef, Row: endRow, Column: endColumn},
	}, nil
}

func a1Range() CellRange {
	address := CellAddress{Ref: "A1", Row: minExcelRow, Column: minExcelColumn}

	return CellRange{Ref: address.Ref, Start: address, End: address}
}

func cellKind(cellType excelize.CellType, hasFormula bool) (CellKind, bool) {
	if hasFormula {
		return CellKindFormula, true
	}

	switch cellType {
	case excelize.CellTypeBool:
		return CellKindBool, true
	case excelize.CellTypeDate:
		return CellKindDate, true
	case excelize.CellTypeError:
		return CellKindError, true
	case excelize.CellTypeFormula:
		return CellKindFormula, true
	case excelize.CellTypeInlineString:
		return CellKindInlineString, true
	case excelize.CellTypeNumber:
		return CellKindNumber, true
	case excelize.CellTypeSharedString:
		return CellKindString, true
	case excelize.CellTypeUnset:
		return CellKindUnset, false
	default:
		return CellKindUnset, false
	}
}

func rowValues(rows [][]string, rowIndex int) []string {
	if rowIndex < 0 || rowIndex >= len(rows) {
		return nil
	}

	return rows[rowIndex]
}

func valueAt(rows [][]string, rowIndex int, columnIndex int) string {
	row := rowValues(rows, rowIndex)
	if columnIndex < 0 || columnIndex >= len(row) {
		return ""
	}

	return row[columnIndex]
}

func isColumnLetters(value string) bool {
	for i := range value {
		if value[i] < 'A' || value[i] > 'Z' {
			return false
		}
	}

	return true
}

func isRowDigits(value string) bool {
	if value[0] == '0' {
		return false
	}

	for i := range value {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}

	return true
}

func columnLettersToNumber(letters string) (int, bool) {
	column := 0
	// Excel columns are base-26 letters without a zero digit: A=1, Z=26, AA=27.
	for i := range letters {
		column = column*26 + int(letters[i]-'A'+1)
		if column > maxExcelColumn {
			return 0, false
		}
	}

	if column < minExcelColumn {
		return 0, false
	}

	return column, true
}

func clampZoomPercent(percent int) int {
	if percent < minZoomPercent {
		return minZoomPercent
	}

	if percent > maxZoomPercent {
		return maxZoomPercent
	}

	return percent
}

type cellBounds struct {
	minRow    int
	minColumn int
	maxRow    int
	maxColumn int
	hasValues bool
}

func (c *cellBounds) addAddress(address CellAddress) {
	if !c.hasValues {
		c.minRow = address.Row
		c.maxRow = address.Row
		c.minColumn = address.Column
		c.maxColumn = address.Column
		c.hasValues = true

		return
	}

	c.minRow = min(c.minRow, address.Row)
	c.maxRow = max(c.maxRow, address.Row)
	c.minColumn = min(c.minColumn, address.Column)
	c.maxColumn = max(c.maxColumn, address.Column)
}

func (c *cellBounds) addRange(cellRange CellRange) {
	c.addAddress(cellRange.Start)
	c.addAddress(cellRange.End)
}

func (c *cellBounds) addBounds(other cellBounds) {
	if !other.hasValues {
		return
	}

	// Bounds merging only needs coordinates; refs are rebuilt for final output.
	c.addAddress(CellAddress{Row: other.minRow, Column: other.minColumn})
	c.addAddress(CellAddress{Row: other.maxRow, Column: other.maxColumn})
}

func (c *cellBounds) rangeValue() (CellRange, error) {
	if !c.hasValues {
		return a1Range(), nil
	}

	return cellRangeFromCoordinates(c.minRow, c.minColumn, c.maxRow, c.maxColumn)
}
