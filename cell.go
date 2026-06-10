package main

import (
	"strconv"
	"strings"
)

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
