package main

import "slices"

const (
	statusKindReady   = "ready"
	statusKindLoading = "loading"
	statusKindError   = "error"

	defaultStatusMessage = "Ready"
	defaultWorkbookTitle = "Untitled"
	defaultSheetName     = "Sheet 1"
	defaultSheetIndex    = 0
	sheetStateVisible    = "visible"

	minExcelRow    = 1
	maxExcelRow    = 1048576
	minExcelColumn = 1
	maxExcelColumn = 16384

	minZoomPercent     = 25
	maxZoomPercent     = 400
	defaultZoomPercent = 100

	defaultColumnWidth = 8.43
	defaultRowHeight   = 15.0
)

// AppState is the complete Go-owned state snapshot rendered by the frontend.
type AppState struct {
	Workbook WorkbookState     `json:"workbook"`
	View     WorkbookViewState `json:"view"`
	Status   AppStatus         `json:"status"`
}

// AppStatus describes current user-facing backend status.
type AppStatus struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Busy    bool   `json:"busy"`
}

// WorkbookState describes the loaded workbook and its sheets.
type WorkbookState struct {
	HasWorkbook bool            `json:"hasWorkbook"`
	Title       string          `json:"title"`
	FilePath    string          `json:"filePath"`
	FileName    string          `json:"fileName"`
	Sheets      []WorkbookSheet `json:"sheets"`
	Styles      []CellStyle     `json:"styles"`
}

// WorkbookViewState describes the current workbook viewport and selection state.
type WorkbookViewState struct {
	ActiveSheetName string         `json:"activeSheetName"`
	ActiveCell      CellAddress    `json:"activeCell"`
	Selection       CellRange      `json:"selection"`
	ZoomPercent     int            `json:"zoomPercent"`
	Scroll          ScrollPosition `json:"scroll"`
}

// ScrollPosition describes the top-left visible cell using 1-based coordinates.
type ScrollPosition struct {
	TopRow     int `json:"topRow"`
	LeftColumn int `json:"leftColumn"`
}

// WorkbookSheet describes one worksheet and the loaded data for it.
type WorkbookSheet struct {
	Index              int               `json:"index"`
	Name               string            `json:"name"`
	State              string            `json:"state"`
	Visible            bool              `json:"visible"`
	Bounds             CellRange         `json:"bounds"`
	DefaultColumnWidth float64           `json:"defaultColumnWidth"`
	DefaultRowHeight   float64           `json:"defaultRowHeight"`
	Cells              []CellData        `json:"cells"`
	MergedCells        []MergedCellRange `json:"mergedCells"`
	Columns            []ColumnLayout    `json:"columns"`
	Rows               []RowLayout       `json:"rows"`
}

// CellData describes one loaded worksheet cell.
type CellData struct {
	Ref        string `json:"ref"`
	Row        int    `json:"row"`
	Column     int    `json:"column"`
	Value      string `json:"value"`
	RawValue   string `json:"rawValue"`
	Formula    string `json:"formula"`
	HasFormula bool   `json:"hasFormula"`
	Kind       string `json:"kind"`
	StyleID    int    `json:"styleId"`
}

// CellAddress describes one cell using an Excel reference and coordinates.
type CellAddress struct {
	Ref    string `json:"ref"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
}

// CellRange describes a rectangular cell range.
type CellRange struct {
	Ref   string      `json:"ref"`
	Start CellAddress `json:"start"`
	End   CellAddress `json:"end"`
}

// MergedCellRange describes one merged cell range and displayed value.
type MergedCellRange struct {
	Range CellRange `json:"range"`
	Value string    `json:"value"`
}

// ColumnLayout describes worksheet column display metadata.
type ColumnLayout struct {
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Width        float64 `json:"width"`
	Hidden       bool    `json:"hidden"`
	OutlineLevel int     `json:"outlineLevel"`
	StyleID      int     `json:"styleId"`
}

// RowLayout describes worksheet row display metadata.
type RowLayout struct {
	Index        int     `json:"index"`
	Height       float64 `json:"height"`
	Hidden       bool    `json:"hidden"`
	OutlineLevel int     `json:"outlineLevel"`
}

// CellStyle describes basic cell formatting metadata.
type CellStyle struct {
	ID             int                `json:"id"`
	NumberFormatID int                `json:"numberFormatId"`
	NumberFormat   string             `json:"numberFormat"`
	Font           CellFontStyle      `json:"font"`
	Fill           CellFillStyle      `json:"fill"`
	Alignment      CellAlignmentStyle `json:"alignment"`
	Borders        []CellBorderStyle  `json:"borders"`
}

// CellFontStyle describes cell font formatting metadata.
type CellFontStyle struct {
	Family        string  `json:"family"`
	Size          float64 `json:"size"`
	Bold          bool    `json:"bold"`
	Italic        bool    `json:"italic"`
	Underline     string  `json:"underline"`
	Strikethrough bool    `json:"strikethrough"`
	Color         string  `json:"color"`
}

// CellFillStyle describes cell fill formatting metadata.
type CellFillStyle struct {
	Type    string   `json:"type"`
	Pattern int      `json:"pattern"`
	Color   string   `json:"color"`
	Colors  []string `json:"colors"`
}

// CellAlignmentStyle describes cell text alignment metadata.
type CellAlignmentStyle struct {
	Horizontal   string `json:"horizontal"`
	Vertical     string `json:"vertical"`
	WrapText     bool   `json:"wrapText"`
	TextRotation int    `json:"textRotation"`
}

// CellBorderStyle describes one side of cell border formatting metadata.
type CellBorderStyle struct {
	Side  string `json:"side"`
	Style int    `json:"style"`
	Color string `json:"color"`
}

func initialAppState() AppState {
	a1 := CellAddress{Ref: "A1", Row: minExcelRow, Column: minExcelColumn}
	// Reuse the exact same single-cell range for sheet bounds and initial selection.
	a1Range := CellRange{Ref: "A1", Start: a1, End: a1}

	return AppState{
		Workbook: WorkbookState{
			HasWorkbook: false,
			Title:       defaultWorkbookTitle,
			FilePath:    "",
			FileName:    "",
			Sheets: []WorkbookSheet{
				{
					Index:              defaultSheetIndex,
					Name:               defaultSheetName,
					State:              sheetStateVisible,
					Visible:            true,
					Bounds:             a1Range,
					DefaultColumnWidth: defaultColumnWidth,
					DefaultRowHeight:   defaultRowHeight,
					Cells:              []CellData{},
					MergedCells:        []MergedCellRange{},
					Columns:            []ColumnLayout{},
					Rows:               []RowLayout{},
				},
			},
			Styles: []CellStyle{},
		},
		View: WorkbookViewState{
			ActiveSheetName: defaultSheetName,
			ActiveCell:      a1,
			Selection:       a1Range,
			ZoomPercent:     defaultZoomPercent,
			Scroll: ScrollPosition{
				TopRow:     minExcelRow,
				LeftColumn: minExcelColumn,
			},
		},
		Status: AppStatus{Kind: statusKindReady, Message: defaultStatusMessage, Busy: false},
	}
}

func cloneAppState(state AppState) AppState {
	// Returning AppState by value still shares slice backing arrays unless we copy them.
	state.Workbook.Sheets = slices.Clone(state.Workbook.Sheets)
	for i := range state.Workbook.Sheets {
		state.Workbook.Sheets[i].Cells = slices.Clone(state.Workbook.Sheets[i].Cells)
		state.Workbook.Sheets[i].MergedCells = slices.Clone(state.Workbook.Sheets[i].MergedCells)
		state.Workbook.Sheets[i].Columns = slices.Clone(state.Workbook.Sheets[i].Columns)
		state.Workbook.Sheets[i].Rows = slices.Clone(state.Workbook.Sheets[i].Rows)
	}

	// Style payloads have their own nested slices.
	state.Workbook.Styles = slices.Clone(state.Workbook.Styles)
	for i := range state.Workbook.Styles {
		state.Workbook.Styles[i].Fill.Colors = slices.Clone(state.Workbook.Styles[i].Fill.Colors)
		state.Workbook.Styles[i].Borders = slices.Clone(state.Workbook.Styles[i].Borders)
	}

	return state
}
