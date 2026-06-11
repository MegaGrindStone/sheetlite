package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
	if excelStyle == nil {
		return CellStyle{ID: styleID}
	}

	return CellStyle{
		ID:             styleID,
		NumberFormatID: excelStyle.NumFmt,
		NumberFormat:   numberFormatCode(excelStyle),
		Font:           fontStyleFromExcelStyle(file, excelStyle.Font),
		Fill:           fillStyleFromExcelStyle(excelStyle.Fill),
		Alignment:      alignmentStyleFromExcelStyle(excelStyle.Alignment),
		Borders:        borderStylesFromExcelStyle(excelStyle.Border),
	}
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
		return CellFontStyle{}
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
		Underline:     font.Underline,
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
		Type:    fill.Type,
		Pattern: fill.Pattern,
		Color:   color,
		Colors:  colors,
	}
}

func alignmentStyleFromExcelStyle(alignment *excelize.Alignment) CellAlignmentStyle {
	if alignment == nil {
		return CellAlignmentStyle{}
	}

	return CellAlignmentStyle{
		Horizontal:   alignment.Horizontal,
		Vertical:     alignment.Vertical,
		WrapText:     alignment.WrapText,
		TextRotation: alignment.TextRotation,
	}
}

func borderStylesFromExcelStyle(borders []excelize.Border) []CellBorderStyle {
	cellBorders := make([]CellBorderStyle, 0, len(borders))
	for _, border := range borders {
		cellBorders = append(cellBorders, CellBorderStyle{
			Side:  border.Type,
			Style: border.Style,
			Color: cssColor(border.Color),
		})
	}

	return cellBorders
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
