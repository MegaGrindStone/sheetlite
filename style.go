package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// FontUnderlineStyle describes supported font underline values.
type FontUnderlineStyle string

const (
	// FontUnderlineStyleNone marks fonts without underline.
	FontUnderlineStyleNone FontUnderlineStyle = "none"
	// FontUnderlineStyleSingle marks single-underlined fonts.
	FontUnderlineStyleSingle FontUnderlineStyle = "single"
	// FontUnderlineStyleDouble marks double-underlined fonts.
	FontUnderlineStyleDouble FontUnderlineStyle = "double"
)

// AllFontUnderlineStyles lists FontUnderlineStyle values for Wails enum binding.
var AllFontUnderlineStyles = []struct {
	Value  FontUnderlineStyle
	TSName string
}{
	{FontUnderlineStyleNone, "None"},
	{FontUnderlineStyleSingle, "Single"},
	{FontUnderlineStyleDouble, "Double"},
}

// FillType describes supported cell fill types.
type FillType string

const (
	// FillTypeNone marks cells without fill metadata.
	FillTypeNone FillType = "none"
	// FillTypePattern marks pattern fills.
	FillTypePattern FillType = "pattern"
	// FillTypeGradient marks gradient fills.
	FillTypeGradient FillType = "gradient"
)

// AllFillTypes lists FillType values for Wails enum binding.
var AllFillTypes = []struct {
	Value  FillType
	TSName string
}{
	{FillTypeNone, "None"},
	{FillTypePattern, "Pattern"},
	{FillTypeGradient, "Gradient"},
}

// HorizontalAlignment describes supported horizontal alignment values.
type HorizontalAlignment string

const (
	// HorizontalAlignmentGeneral marks default horizontal alignment.
	HorizontalAlignmentGeneral HorizontalAlignment = "general"
	// HorizontalAlignmentLeft aligns content left.
	HorizontalAlignmentLeft HorizontalAlignment = "left"
	// HorizontalAlignmentCenter aligns content in the center.
	HorizontalAlignmentCenter HorizontalAlignment = "center"
	// HorizontalAlignmentRight aligns content right.
	HorizontalAlignmentRight HorizontalAlignment = "right"
	// HorizontalAlignmentFill fills content across the cell.
	HorizontalAlignmentFill HorizontalAlignment = "fill"
	// HorizontalAlignmentJustify justifies content horizontally.
	HorizontalAlignmentJustify HorizontalAlignment = "justify"
	// HorizontalAlignmentCenterContinuous centers content across selected cells.
	HorizontalAlignmentCenterContinuous HorizontalAlignment = "centerContinuous"
	// HorizontalAlignmentDistributed distributes content horizontally.
	HorizontalAlignmentDistributed HorizontalAlignment = "distributed"
)

// AllHorizontalAlignments lists HorizontalAlignment values for Wails enum binding.
var AllHorizontalAlignments = []struct {
	Value  HorizontalAlignment
	TSName string
}{
	{HorizontalAlignmentGeneral, "General"},
	{HorizontalAlignmentLeft, "Left"},
	{HorizontalAlignmentCenter, "Center"},
	{HorizontalAlignmentRight, "Right"},
	{HorizontalAlignmentFill, "Fill"},
	{HorizontalAlignmentJustify, "Justify"},
	{HorizontalAlignmentCenterContinuous, "CenterContinuous"},
	{HorizontalAlignmentDistributed, "Distributed"},
}

// VerticalAlignment describes supported vertical alignment values.
type VerticalAlignment string

const (
	// VerticalAlignmentGeneral marks default vertical alignment.
	VerticalAlignmentGeneral VerticalAlignment = "general"
	// VerticalAlignmentTop aligns content to the top.
	VerticalAlignmentTop VerticalAlignment = "top"
	// VerticalAlignmentCenter aligns content vertically centered.
	VerticalAlignmentCenter VerticalAlignment = "center"
	// VerticalAlignmentBottom aligns content to the bottom.
	VerticalAlignmentBottom VerticalAlignment = "bottom"
	// VerticalAlignmentJustify justifies content vertically.
	VerticalAlignmentJustify VerticalAlignment = "justify"
	// VerticalAlignmentDistributed distributes content vertically.
	VerticalAlignmentDistributed VerticalAlignment = "distributed"
)

// AllVerticalAlignments lists VerticalAlignment values for Wails enum binding.
var AllVerticalAlignments = []struct {
	Value  VerticalAlignment
	TSName string
}{
	{VerticalAlignmentGeneral, "General"},
	{VerticalAlignmentTop, "Top"},
	{VerticalAlignmentCenter, "Center"},
	{VerticalAlignmentBottom, "Bottom"},
	{VerticalAlignmentJustify, "Justify"},
	{VerticalAlignmentDistributed, "Distributed"},
}

// BorderSide describes supported cell border sides.
type BorderSide string

const (
	// BorderSideLeft marks the left border.
	BorderSideLeft BorderSide = "left"
	// BorderSideRight marks the right border.
	BorderSideRight BorderSide = "right"
	// BorderSideTop marks the top border.
	BorderSideTop BorderSide = "top"
	// BorderSideBottom marks the bottom border.
	BorderSideBottom BorderSide = "bottom"
	// BorderSideDiagonalUp marks a diagonal-up border.
	BorderSideDiagonalUp BorderSide = "diagonalUp"
	// BorderSideDiagonalDown marks a diagonal-down border.
	BorderSideDiagonalDown BorderSide = "diagonalDown"
)

// AllBorderSides lists BorderSide values for Wails enum binding.
var AllBorderSides = []struct {
	Value  BorderSide
	TSName string
}{
	{BorderSideLeft, "Left"},
	{BorderSideRight, "Right"},
	{BorderSideTop, "Top"},
	{BorderSideBottom, "Bottom"},
	{BorderSideDiagonalUp, "DiagonalUp"},
	{BorderSideDiagonalDown, "DiagonalDown"},
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
	Render         CellRenderStyle    `json:"render"`
}

// CellFontStyle describes cell font formatting metadata.
type CellFontStyle struct {
	Family        string             `json:"family"`
	Size          float64            `json:"size"`
	Bold          bool               `json:"bold"`
	Italic        bool               `json:"italic"`
	Underline     FontUnderlineStyle `json:"underline"`
	Strikethrough bool               `json:"strikethrough"`
	Color         string             `json:"color"`
}

// CellFillStyle describes cell fill formatting metadata.
type CellFillStyle struct {
	Type    FillType `json:"type"`
	Pattern int      `json:"pattern"`
	Color   string   `json:"color"`
	Colors  []string `json:"colors"`
}

// CellAlignmentStyle describes cell text alignment metadata.
type CellAlignmentStyle struct {
	Horizontal   HorizontalAlignment `json:"horizontal"`
	Vertical     VerticalAlignment   `json:"vertical"`
	WrapText     bool                `json:"wrapText"`
	TextRotation int                 `json:"textRotation"`
}

// CellBorderStyle describes one side of cell border formatting metadata.
type CellBorderStyle struct {
	Side  BorderSide `json:"side"`
	Style int        `json:"style"`
	Color string     `json:"color"`
}

// CellRenderStyle describes display-only cell style metadata derived by the backend.
type CellRenderStyle struct {
	TextColor    string `json:"textColor"`
	TextAdjusted bool   `json:"textAdjusted"`
}

var englishNumberFormats = map[int]string{
	0:  "general",
	1:  "0",
	2:  "0.00",
	3:  "#,##0",
	4:  "#,##0.00",
	9:  "0%",
	10: "0.00%",
	11: "0.00E+00",
	12: "# ?/?",
	13: "# ??/??",
	14: "mm-dd-yy",
	15: "d-mmm-yy",
	16: "d-mmm",
	17: "mmm-yy",
	18: "h:mm AM/PM",
	19: "h:mm:ss AM/PM",
	20: "hh:mm",
	21: "hh:mm:ss",
	22: "m/d/yy hh:mm",
	37: "#,##0 ;(#,##0)",
	38: "#,##0 ;[red](#,##0)",
	39: "#,##0.00 ;(#,##0.00)",
	40: "#,##0.00 ;[red](#,##0.00)",
	41: "_(* #,##0_);_(* \\(#,##0\\);_(* \"-\"_);_(@_)",
	42: "_(\"$\"* #,##0_);_(\"$\"* \\(#,##0\\);_(\"$\"* \"-\"_);_(@_)",
	43: "_(* #,##0.00_);_(* \\(#,##0.00\\);_(* \"-\"??_);_(@_)",
	44: "_(\"$\"* #,##0.00_);_(\"$\"* \\(#,##0.00\\);_(\"$\"* \"-\"??_);_(@_)",
	45: "mm:ss",
	46: "[h]:mm:ss",
	47: "mm:ss.0",
	48: "##0.0E+0",
	49: "@",
}

func loadCellStyles(file *excelize.File, styleIDs map[int]struct{}) ([]CellStyle, error) {
	ids := make([]int, 0, len(styleIDs))
	for styleID := range styleIDs {
		if styleID < 0 {
			continue
		}
		ids = append(ids, styleID)
	}
	sort.Ints(ids)

	styles := make([]CellStyle, 0, len(ids))
	for _, styleID := range ids {
		excelStyle, err := file.GetStyle(styleID)
		if err != nil {
			return nil, fmt.Errorf("load style %d: %w", styleID, err)
		}

		styles = append(styles, cellStyleFromExcelStyle(file, styleID, excelStyle))
	}

	return styles, nil
}

func cellStyleFromExcelStyle(file *excelize.File, styleID int, excelStyle *excelize.Style) CellStyle {
	style := CellStyle{
		ID: styleID,
		Font: CellFontStyle{
			Underline: FontUnderlineStyleNone,
		},
		Fill: CellFillStyle{
			Type:   FillTypeNone,
			Colors: []string{},
		},
		Alignment: CellAlignmentStyle{
			Horizontal: HorizontalAlignmentGeneral,
			Vertical:   VerticalAlignmentGeneral,
		},
		Borders: []CellBorderStyle{},
	}
	if excelStyle == nil {
		return style
	}

	style.NumberFormatID = excelStyle.NumFmt
	style.NumberFormat = numberFormatCode(excelStyle)
	style.Font = fontStyleFromExcelStyle(file, excelStyle.Font)
	style.Fill = fillStyleFromExcelStyle(excelStyle.Fill)
	style.Alignment = alignmentStyleFromExcelStyle(excelStyle.Alignment)
	style.Borders = borderStylesFromExcelStyle(excelStyle.Border)

	return style
}

func numberFormatCode(excelStyle *excelize.Style) string {
	if excelStyle.CustomNumFmt != nil {
		return *excelStyle.CustomNumFmt
	}

	// Excelize doesn't expose built-in format strings, so unknown IDs use General.
	formatCode, ok := englishNumberFormats[excelStyle.NumFmt]
	if !ok {
		return "general"
	}

	return formatCode
}

func fontStyleFromExcelStyle(file *excelize.File, font *excelize.Font) CellFontStyle {
	if font == nil {
		return CellFontStyle{Underline: FontUnderlineStyleNone}
	}

	color := font.Color
	// Resolve indexed/theme colors when Excelize can map them to a base color.
	if baseColor := file.GetBaseColor(font.Color, font.ColorIndexed, font.ColorTheme); baseColor != "" {
		color = baseColor
	}

	return CellFontStyle{
		Family:        font.Family,
		Size:          font.Size,
		Bold:          font.Bold,
		Italic:        font.Italic,
		Underline:     normalizeFontUnderlineStyle(font.Underline),
		Strikethrough: font.Strike,
		Color:         cssColor(color),
	}
}

func fillStyleFromExcelStyle(fill excelize.Fill) CellFillStyle {
	colors := make([]string, 0, len(fill.Color))
	for _, color := range fill.Color {
		if formatted := cssColor(color); formatted != "" {
			colors = append(colors, formatted)
		}
	}

	color := ""
	if len(colors) > 0 {
		color = colors[0]
	}

	return CellFillStyle{
		Type:    normalizeFillType(fill.Type),
		Pattern: fill.Pattern,
		Color:   color,
		Colors:  colors,
	}
}

func alignmentStyleFromExcelStyle(alignment *excelize.Alignment) CellAlignmentStyle {
	if alignment == nil {
		return CellAlignmentStyle{
			Horizontal: HorizontalAlignmentGeneral,
			Vertical:   VerticalAlignmentGeneral,
		}
	}

	return CellAlignmentStyle{
		Horizontal:   normalizeHorizontalAlignment(alignment.Horizontal),
		Vertical:     normalizeVerticalAlignment(alignment.Vertical),
		WrapText:     alignment.WrapText,
		TextRotation: alignment.TextRotation,
	}
}

func borderStylesFromExcelStyle(borders []excelize.Border) []CellBorderStyle {
	cellBorders := make([]CellBorderStyle, 0, len(borders))
	for _, border := range borders {
		side, ok := normalizeBorderSide(border.Type)
		if !ok {
			continue
		}

		cellBorders = append(cellBorders, CellBorderStyle{
			Side:  side,
			Style: border.Style,
			Color: cssColor(border.Color),
		})
	}

	return cellBorders
}

func normalizeFontUnderlineStyle(value string) FontUnderlineStyle {
	switch FontUnderlineStyle(strings.TrimSpace(value)) {
	case FontUnderlineStyleNone:
		return FontUnderlineStyleNone
	case FontUnderlineStyleSingle:
		return FontUnderlineStyleSingle
	case FontUnderlineStyleDouble:
		return FontUnderlineStyleDouble
	default:
		return FontUnderlineStyleNone
	}
}

func normalizeFillType(value string) FillType {
	switch FillType(strings.TrimSpace(value)) {
	case FillTypeNone:
		return FillTypeNone
	case FillTypePattern:
		return FillTypePattern
	case FillTypeGradient:
		return FillTypeGradient
	default:
		return FillTypeNone
	}
}

func normalizeHorizontalAlignment(value string) HorizontalAlignment {
	switch HorizontalAlignment(strings.TrimSpace(value)) {
	case HorizontalAlignmentGeneral:
		return HorizontalAlignmentGeneral
	case HorizontalAlignmentLeft:
		return HorizontalAlignmentLeft
	case HorizontalAlignmentCenter:
		return HorizontalAlignmentCenter
	case HorizontalAlignmentRight:
		return HorizontalAlignmentRight
	case HorizontalAlignmentFill:
		return HorizontalAlignmentFill
	case HorizontalAlignmentJustify:
		return HorizontalAlignmentJustify
	case HorizontalAlignmentCenterContinuous:
		return HorizontalAlignmentCenterContinuous
	case HorizontalAlignmentDistributed:
		return HorizontalAlignmentDistributed
	default:
		return HorizontalAlignmentGeneral
	}
}

func normalizeVerticalAlignment(value string) VerticalAlignment {
	switch VerticalAlignment(strings.TrimSpace(value)) {
	case VerticalAlignmentGeneral:
		return VerticalAlignmentGeneral
	case VerticalAlignmentTop:
		return VerticalAlignmentTop
	case VerticalAlignmentCenter:
		return VerticalAlignmentCenter
	case VerticalAlignmentBottom:
		return VerticalAlignmentBottom
	case VerticalAlignmentJustify:
		return VerticalAlignmentJustify
	case VerticalAlignmentDistributed:
		return VerticalAlignmentDistributed
	default:
		return VerticalAlignmentGeneral
	}
}

func normalizeBorderSide(value string) (BorderSide, bool) {
	switch BorderSide(strings.TrimSpace(value)) {
	case BorderSideLeft:
		return BorderSideLeft, true
	case BorderSideRight:
		return BorderSideRight, true
	case BorderSideTop:
		return BorderSideTop, true
	case BorderSideBottom:
		return BorderSideBottom, true
	case BorderSideDiagonalUp:
		return BorderSideDiagonalUp, true
	case BorderSideDiagonalDown:
		return BorderSideDiagonalDown, true
	default:
		return "", false
	}
}

func cssColor(color string) string {
	trimmedColor := strings.TrimSpace(color)
	if trimmedColor == "" {
		return ""
	}

	trimmedColor = strings.TrimPrefix(trimmedColor, "#")
	// Excel colors are often opaque ARGB; frontend CSS wants RGB here.
	if len(trimmedColor) == 8 && strings.HasPrefix(strings.ToUpper(trimmedColor), "FF") {
		trimmedColor = trimmedColor[2:]
	}

	return "#" + strings.ToUpper(trimmedColor)
}

func darkTextColor(sourceText rgbColor, background rgbColor) (rgbColor, bool) {
	if contrastRatio(sourceText, background) >= minimumCellTextContrast {
		return sourceText, false
	}

	// Move toward the endpoint with stronger contrast against the fill.
	target := whiteRGB
	if contrastRatio(blackRGB, background) > contrastRatio(whiteRGB, background) {
		target = blackRGB
	}

	if mixedColor, ok := mixColorToContrast(sourceText, target, background); ok {
		return mixedColor, true
	}

	fallback := lightThemeGridTextRGB
	if contrastRatio(darkThemeGridTextRGB, background) >= contrastRatio(lightThemeGridTextRGB, background) {
		fallback = darkThemeGridTextRGB
	}

	if contrastRatio(fallback, background) >= minimumCellTextContrast {
		return fallback, true
	}
	if contrastRatio(blackRGB, background) > contrastRatio(whiteRGB, background) {
		return blackRGB, true
	}

	return whiteRGB, true
}

func (c *CellStyle) render(effectiveTheme AppearanceTheme) {
	if effectiveTheme != AppearanceThemeDark {
		// Light mode leaves workbook-authored font colors in charge.
		c.Render = CellRenderStyle{}
		return
	}

	background := c.renderBackground()
	textColor := c.renderTextColor(background)

	readableColor, adjusted := darkTextColor(textColor, background)
	c.Render.TextColor = readableColor.cssColor()
	c.Render.TextAdjusted = adjusted
}

func (c *CellStyle) renderBackground() rgbColor {
	if color, err := parseCSSHexColor(c.Fill.Color); err == nil {
		return color
	}

	for _, fillColor := range c.Fill.Colors {
		if color, err := parseCSSHexColor(fillColor); err == nil {
			return color
		}
	}

	return darkGridSurfaceRGB
}

func (c *CellStyle) renderTextColor(background rgbColor) rgbColor {
	if color, err := parseCSSHexColor(c.Font.Color); err == nil {
		return color
	}

	// Missing font colors behave like automatic text.
	if contrastRatio(darkThemeGridTextRGB, background) >= contrastRatio(lightThemeGridTextRGB, background) {
		return darkThemeGridTextRGB
	}

	return lightThemeGridTextRGB
}
