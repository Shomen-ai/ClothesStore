// admin_report_handler.go builds and streams the branded Excel analytics report
// served at GET /api/admin/reports/excel (admin-only). The handler pulls report
// sheets and summary KPIs from the repository and renders them into a styled
// .xlsx workbook (contents page, one sheet per report, a category chart and a
// table of contents with hyperlinks). The bulk of this file is presentation:
// excelize style construction and per-sheet layout.
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"clothes-store/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// AdminReportHandler serves the Excel report export, backed by the reports repo.
type AdminReportHandler struct{ repo *repository.ReportsRepo }

// NewAdminReportHandler constructs an AdminReportHandler backed by the given repo.
func NewAdminReportHandler(repo *repository.ReportsRepo) *AdminReportHandler {
	return &AdminReportHandler{repo: repo}
}

// ── re-exports so this file can refer to repo types without a long path ──
type (
	ReportSheet = repository.ReportSheet
	ColType     = repository.ColType
)

const (
	colText    = repository.ColText
	colInt     = repository.ColInt
	colMoney   = repository.ColMoney
	colPercent = repository.ColPercent
	colDate    = repository.ColDate

	// brand palette
	cAccent     = "E11900"
	cAccentDark = "B0140A"
	cText       = "14141A"
	cTextSoft   = "6C6C78"
	cBorder     = "E5E5E5"
	cZebra      = "F8F4ED"
	cTotals     = "ECE7DF"
	cKpiBg      = "FAF6EE"
	cContentsBg = "F6F3EE"

	// format strings
	fmtMoney   = `# ##0" ₽"`
	fmtInt     = `# ##0`
	fmtPercent = `0.0"%"`

	contentsSheet = "Содержание"
	chartSheet    = "Выручка по категориям"
)

type reportStyles struct {
	// Contents
	pageTitle      int
	pageSubtitle   int
	kpiCardLabel   int
	kpiCardMoney   int
	kpiCardInt     int
	kpiCardPercent int
	tocHeader      int
	tocLink        int

	// Sheet chrome
	sheetTitle    int
	sheetSubtitle int
	header        int

	// Per-column body / zebra / totals styles
	body   map[ColType]int
	zebra  map[ColType]int
	totals map[ColType]int
}

// ─────────────────────────────────────────────────────────────────────────────

// Excel serves GET /api/admin/reports/excel and streams a generated .xlsx
// workbook as a file download. It fetches all report sheets and summary KPIs,
// builds the shared styles, renders the contents page and one sheet per report,
// then writes the file to the response body. A failure before the body is
// written yields 500 JSON; a failure during streaming (after headers are
// flushed) can only be logged via Gin.
func (h *AdminReportHandler) Excel(c *gin.Context) {
	sheets, err := h.repo.AllSheets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	kpis, err := h.repo.KPIs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	styles, err := buildStyles(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// First sheet → "Содержание".
	defaultSheet := f.GetSheetName(0)
	f.SetSheetName(defaultSheet, contentsSheet)

	renderContents(f, kpis, sheets, styles)

	// One sheet per report. Track names because excelize truncates sheet names
	// to 31 chars and we need the same key for hyperlinks.
	displayNames := make([]string, len(sheets))
	for i, s := range sheets {
		name := truncSheetName(s.Name)
		displayNames[i] = name
		if _, err := f.NewSheet(name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		renderSheet(f, name, s, styles)
		if strings.HasPrefix(s.Name, chartSheet) {
			addCategoryChart(f, name, s)
		}
	}

	// TOC links now that all sheets exist.
	renderTOCLinks(f, displayNames, styles)

	f.SetActiveSheet(0)

	fname := fmt.Sprintf("clothesstore-report_%s.xlsx", time.Now().Format("2006-01-02_1504"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fname)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Status(http.StatusOK)

	if err := f.Write(c.Writer); err != nil {
		// Headers are already flushed; just log via Gin.
		_ = c.Error(err)
	}
}

// ─── styles ─────────────────────────────────────────────────────────────────

// buildStyles registers every excelize cell style used by the report once and
// returns them grouped in reportStyles. Body/zebra/totals styles are keyed by
// ColType so number formats and alignment match each column's data type.
func buildStyles(f *excelize.File) (*reportStyles, error) {
	str := func(s string) *string { return &s }
	s := &reportStyles{
		body:   map[ColType]int{},
		zebra:  map[ColType]int{},
		totals: map[ColType]int{},
	}

	// Contents — page-level decoration
	id, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 26, Color: cText, Family: "Manrope"},
	})
	if err != nil {
		return nil, err
	}
	s.pageTitle = id

	if id, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Size: 11, Color: cTextSoft, Family: "Manrope"},
	}); err != nil {
		return nil, err
	}
	s.pageSubtitle = id

	if id, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 9, Color: cTextSoft, Family: "JetBrains Mono"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{cKpiBg}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: cBorder, Style: 1},
			{Type: "left", Color: cBorder, Style: 1},
			{Type: "right", Color: cBorder, Style: 1},
		},
	}); err != nil {
		return nil, err
	}
	s.kpiCardLabel = id

	kpiValueStyle := func(numFmt string) (int, error) {
		return f.NewStyle(&excelize.Style{
			Font:         &excelize.Font{Bold: true, Size: 22, Color: cText, Family: "Manrope"},
			Alignment:    &excelize.Alignment{Horizontal: "left", Vertical: "center"},
			Fill:         excelize.Fill{Type: "pattern", Color: []string{cKpiBg}, Pattern: 1},
			CustomNumFmt: str(numFmt),
			Border: []excelize.Border{
				{Type: "bottom", Color: cBorder, Style: 1},
				{Type: "left", Color: cBorder, Style: 1},
				{Type: "right", Color: cBorder, Style: 1},
			},
		})
	}
	if id, err = kpiValueStyle(fmtMoney); err != nil {
		return nil, err
	}
	s.kpiCardMoney = id
	if id, err = kpiValueStyle(fmtInt); err != nil {
		return nil, err
	}
	s.kpiCardInt = id
	if id, err = kpiValueStyle(fmtPercent); err != nil {
		return nil, err
	}
	s.kpiCardPercent = id

	if id, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: cTextSoft, Family: "JetBrains Mono"},
		Alignment: &excelize.Alignment{Horizontal: "left"},
		Border:    []excelize.Border{{Type: "bottom", Color: cBorder, Style: 1}},
	}); err != nil {
		return nil, err
	}
	s.tocHeader = id

	if id, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 12, Color: cAccent, Family: "Manrope", Underline: "single"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	}); err != nil {
		return nil, err
	}
	s.tocLink = id

	// Sheet chrome
	if id, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: cText, Family: "Manrope"},
	}); err != nil {
		return nil, err
	}
	s.sheetTitle = id

	if id, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Size: 10, Color: cTextSoft, Family: "Manrope"},
	}); err != nil {
		return nil, err
	}
	s.sheetSubtitle = id

	if id, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF", Family: "Manrope"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{cAccent}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    []excelize.Border{{Type: "bottom", Color: cAccentDark, Style: 1}},
	}); err != nil {
		return nil, err
	}
	s.header = id

	// Per-column body / zebra / totals
	for _, t := range []ColType{colText, colInt, colMoney, colPercent, colDate} {
		baseFont := &excelize.Font{Size: 11, Color: cText, Family: "Manrope"}
		mono := &excelize.Font{Size: 11, Color: cText, Family: "JetBrains Mono"}
		font := baseFont
		if t == colDate {
			font = mono
		}
		al := &excelize.Alignment{Vertical: "center", Horizontal: alignFor(t)}
		var numFmt *string
		switch t {
		case colMoney:
			numFmt = str(fmtMoney)
		case colInt:
			numFmt = str(fmtInt)
		case colPercent:
			numFmt = str(fmtPercent)
		}
		bottom := []excelize.Border{{Type: "bottom", Color: cBorder, Style: 1}}

		id1, err := f.NewStyle(&excelize.Style{Font: font, Alignment: al, CustomNumFmt: numFmt, Border: bottom})
		if err != nil {
			return nil, err
		}
		s.body[t] = id1

		id2, err := f.NewStyle(&excelize.Style{
			Font: font, Alignment: al, CustomNumFmt: numFmt, Border: bottom,
			Fill: excelize.Fill{Type: "pattern", Color: []string{cZebra}, Pattern: 1},
		})
		if err != nil {
			return nil, err
		}
		s.zebra[t] = id2

		totalsFont := &excelize.Font{Size: 11, Bold: true, Color: cText, Family: "Manrope"}
		if t == colDate {
			totalsFont = &excelize.Font{Size: 11, Bold: true, Color: cText, Family: "JetBrains Mono"}
		}
		id3, err := f.NewStyle(&excelize.Style{
			Font:         totalsFont,
			Alignment:    al,
			CustomNumFmt: numFmt,
			Fill:         excelize.Fill{Type: "pattern", Color: []string{cTotals}, Pattern: 1},
			Border:       []excelize.Border{{Type: "top", Color: cAccent, Style: 1}, {Type: "bottom", Color: cBorder, Style: 1}},
		})
		if err != nil {
			return nil, err
		}
		s.totals[t] = id3
	}

	return s, nil
}

func alignFor(t ColType) string {
	switch t {
	case colMoney, colInt, colPercent:
		return "right"
	case colDate:
		return "center"
	}
	return "left"
}

// ─── Contents sheet ─────────────────────────────────────────────────────────

// renderContents lays out the "Содержание" landing sheet: title, generation
// timestamp, a 3x2 grid of KPI cards and the table-of-contents skeleton. The
// TOC hyperlink targets are filled in later by renderTOCLinks, once every report
// sheet has been created.
func renderContents(f *excelize.File, k repository.SummaryKPIs, sheets []ReportSheet, s *reportStyles) {
	f.SetCellValue(contentsSheet, "A1", "Аналитический отчёт ClothesStore")
	f.SetCellStyle(contentsSheet, "A1", "A1", s.pageTitle)
	f.SetRowHeight(contentsSheet, 1, 38)

	f.SetCellValue(contentsSheet, "A2", "Сформировано: "+time.Now().Format("02.01.2006 15:04"))
	f.SetCellStyle(contentsSheet, "A2", "A2", s.pageSubtitle)

	// KPI grid: 3 columns × 2 rows
	type kpi struct {
		label string
		value any
		style int
	}
	cards := []kpi{
		{"Валовая выручка", k.TotalRevenue, s.kpiCardMoney},
		{"Заказов всего", k.TotalOrders, s.kpiCardInt},
		{"Средний чек (AOV)", k.AOV, s.kpiCardMoney},
		{"Клиентов", k.Customers, s.kpiCardInt},
		{"Активных товаров", k.ActiveProducts, s.kpiCardInt},
		{"Доля распродажи", k.SaleShare, s.kpiCardPercent},
	}

	// Cards span B..C (col 2..3), D..E (4..5), F..G (6..7), rows 4-5 and 7-8.
	cardCols := [][2]int{{2, 3}, {4, 5}, {6, 7}}
	for i, c := range cards {
		rowGroup := i / 3
		labelRow := 4 + rowGroup*3
		valueRow := labelRow + 1
		colStart := cardCols[i%3][0]
		colEnd := cardCols[i%3][1]

		startLbl, _ := excelize.CoordinatesToCellName(colStart, labelRow)
		endLbl, _ := excelize.CoordinatesToCellName(colEnd, labelRow)
		_ = f.MergeCell(contentsSheet, startLbl, endLbl)
		f.SetCellValue(contentsSheet, startLbl, "  "+strings.ToUpper(c.label))
		f.SetCellStyle(contentsSheet, startLbl, endLbl, s.kpiCardLabel)
		f.SetRowHeight(contentsSheet, labelRow, 20)

		startVal, _ := excelize.CoordinatesToCellName(colStart, valueRow)
		endVal, _ := excelize.CoordinatesToCellName(colEnd, valueRow)
		_ = f.MergeCell(contentsSheet, startVal, endVal)
		f.SetCellValue(contentsSheet, startVal, c.value)
		f.SetCellStyle(contentsSheet, startVal, endVal, c.style)
		f.SetRowHeight(contentsSheet, valueRow, 42)
	}

	// TOC header
	f.SetCellValue(contentsSheet, "B11", "Содержимое отчёта")
	f.SetCellStyle(contentsSheet, "B11", "G11", s.tocHeader)
	f.SetRowHeight(contentsSheet, 11, 26)

	// TOC rows are populated by renderTOCLinks (after sheets are created).
	for i := range sheets {
		row := 12 + i
		idxCell, _ := excelize.CoordinatesToCellName(2, row)
		f.SetCellValue(contentsSheet, idxCell, fmt.Sprintf("%02d", i+1))
		f.SetCellStyle(contentsSheet, idxCell, idxCell, s.pageSubtitle)
		f.SetRowHeight(contentsSheet, row, 22)
	}

	// Columns and visuals
	_ = f.SetColWidth(contentsSheet, "A", "A", 2)
	_ = f.SetColWidth(contentsSheet, "B", "B", 6)
	_ = f.SetColWidth(contentsSheet, "C", "C", 26)
	_ = f.SetColWidth(contentsSheet, "D", "D", 6)
	_ = f.SetColWidth(contentsSheet, "E", "E", 26)
	_ = f.SetColWidth(contentsSheet, "F", "F", 6)
	_ = f.SetColWidth(contentsSheet, "G", "G", 26)
	_ = f.SetSheetView(contentsSheet, 0, &excelize.ViewOptions{ShowGridLines: boolp(false)})
}

// renderTOCLinks writes the clickable table-of-contents rows linking to each
// report sheet. It runs after all sheets exist so the internal hyperlinks
// resolve.
func renderTOCLinks(f *excelize.File, displayNames []string, s *reportStyles) {
	for i, name := range displayNames {
		row := 12 + i
		linkCell, _ := excelize.CoordinatesToCellName(3, row)
		endCell, _ := excelize.CoordinatesToCellName(7, row)
		_ = f.MergeCell(contentsSheet, linkCell, endCell)
		f.SetCellValue(contentsSheet, linkCell, name)
		f.SetCellStyle(contentsSheet, linkCell, endCell, s.tocLink)
		display := name
		_ = f.SetCellHyperLink(contentsSheet, linkCell,
			fmt.Sprintf("'%s'!A1", name), "Location",
			excelize.HyperlinkOpts{Display: &display, Tooltip: &display},
		)
	}
}

// ─── Per-report sheet ───────────────────────────────────────────────────────

// renderSheet renders a single report's data onto its sheet: title/subtitle,
// header row, zebra-striped data rows, an optional SUM totals row for the
// columns listed in TotalsCols, an auto-filter, data bars on money/int columns
// and a frozen header pane.
func renderSheet(f *excelize.File, name string, s ReportSheet, st *reportStyles) {
	gen := time.Now().Format("02.01.2006 15:04")

	// Title + subtitle
	f.SetCellValue(name, "A1", s.Name)
	f.SetCellStyle(name, "A1", "A1", st.sheetTitle)
	f.SetRowHeight(name, 1, 30)
	f.SetCellValue(name, "A2", "Сформировано: "+gen)
	f.SetCellStyle(name, "A2", "A2", st.sheetSubtitle)

	headerRow := 4
	for col, h := range s.Headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, headerRow)
		f.SetCellValue(name, cell, h)
	}
	startCell, _ := excelize.CoordinatesToCellName(1, headerRow)
	endCell, _ := excelize.CoordinatesToCellName(len(s.Headers), headerRow)
	f.SetCellStyle(name, startCell, endCell, st.header)
	f.SetRowHeight(name, headerRow, 26)

	// Data rows with zebra striping
	for r, row := range s.Rows {
		rowNum := headerRow + 1 + r
		zebra := r%2 == 1
		for col, v := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
			t := colTypeAt(s, col)
			f.SetCellValue(name, cell, cellValue(v, t))
			if zebra {
				f.SetCellStyle(name, cell, cell, st.zebra[t])
			} else {
				f.SetCellStyle(name, cell, cell, st.body[t])
			}
		}
	}

	// Totals row (SUM formulas) for columns marked in TotalsCols
	if len(s.TotalsCols) > 0 && len(s.Rows) > 0 {
		totalsRow := headerRow + 1 + len(s.Rows)
		// Label "Итого" in the first column
		labelCell, _ := excelize.CoordinatesToCellName(1, totalsRow)
		f.SetCellValue(name, labelCell, "Итого")
		f.SetCellStyle(name, labelCell, labelCell, st.totals[colText])

		dataStart := headerRow + 1
		dataEnd := headerRow + len(s.Rows)

		totalsSet := map[int]bool{}
		for _, c := range s.TotalsCols {
			totalsSet[c] = true
		}

		for col := 1; col < len(s.Headers); col++ {
			cell, _ := excelize.CoordinatesToCellName(col+1, totalsRow)
			t := colTypeAt(s, col)
			if totalsSet[col] {
				colName, _ := excelize.ColumnNumberToName(col + 1)
				formula := fmt.Sprintf("SUM(%s%d:%s%d)", colName, dataStart, colName, dataEnd)
				_ = f.SetCellFormula(name, cell, formula)
			}
			f.SetCellStyle(name, cell, cell, st.totals[t])
		}
		f.SetRowHeight(name, totalsRow, 24)
	}

	// AutoFilter on header
	_ = f.AutoFilter(name, fmt.Sprintf("%s:%s", startCell, endCell), []excelize.AutoFilterOptions{})

	// Column widths
	for col := range s.Headers {
		colName, _ := excelize.ColumnNumberToName(col + 1)
		_ = f.SetColWidth(name, colName, colName, widthForCol(s, col))
	}

	// Data bars on Money/Int data columns (skip totals row)
	if len(s.Rows) > 1 {
		for col, t := range s.ColTypes {
			if t != colMoney && t != colInt {
				continue
			}
			colName, _ := excelize.ColumnNumberToName(col + 1)
			ref := fmt.Sprintf("%s%d:%s%d", colName, headerRow+1, colName, headerRow+len(s.Rows))
			zero := 0
			_ = f.SetConditionalFormat(name, ref, []excelize.ConditionalFormatOptions{{
				Type: "data_bar", Criteria: "=",
				MinType: "min", MaxType: "max",
				BarColor: "#" + cAccent,
				BarOnly:  false,
				Format:   &zero,
			}})
		}
	}

	// Freeze the title + header rows
	_ = f.SetPanes(name, &excelize.Panes{
		Freeze: true, Split: false,
		XSplit: 0, YSplit: headerRow,
		TopLeftCell: "A" + fmt.Sprintf("%d", headerRow+1),
		ActivePane:  "bottomLeft",
	})
}

// ─── chart ──────────────────────────────────────────────────────────────────

// addCategoryChart draws a bar chart of revenue-by-category next to the table,
// using the first column as categories and the last column as values.
func addCategoryChart(f *excelize.File, sheet string, s ReportSheet) {
	if len(s.Rows) == 0 {
		return
	}
	headerRow := 4
	firstData := headerRow + 1
	lastData := headerRow + len(s.Rows)
	revCol, _ := excelize.ColumnNumberToName(len(s.Headers)) // Revenue is last col
	catCol := "A"

	categories := fmt.Sprintf("'%s'!$%s$%d:$%s$%d", sheet, catCol, firstData, catCol, lastData)
	values := fmt.Sprintf("'%s'!$%s$%d:$%s$%d", sheet, revCol, firstData, revCol, lastData)
	nameRef := fmt.Sprintf("'%s'!$%s$%d", sheet, revCol, headerRow)

	anchor, _ := excelize.CoordinatesToCellName(len(s.Headers)+2, headerRow)
	_ = f.AddChart(sheet, anchor, &excelize.Chart{
		Type: excelize.Col,
		Series: []excelize.ChartSeries{{
			Name:       nameRef,
			Categories: categories,
			Values:     values,
			Fill:       excelize.Fill{Type: "pattern", Color: []string{"#" + cAccent}, Pattern: 1},
		}},
		Title:      []excelize.RichTextRun{{Text: "Выручка по категориям, ₽"}},
		Legend:     excelize.ChartLegend{Position: "none"},
		PlotArea:   excelize.ChartPlotArea{ShowVal: false, ShowCatName: false, ShowSerName: false},
		Dimension:  excelize.ChartDimension{Width: 520, Height: 320},
	})
}

// ─── helpers ────────────────────────────────────────────────────────────────

func colTypeAt(s ReportSheet, col int) ColType {
	if col < len(s.ColTypes) {
		return s.ColTypes[col]
	}
	return colText
}

// cellValue normalises pq's []byte/string scan results to the correct Go type
// so Excel number formats and SUM formulas actually apply.
func cellValue(v any, t ColType) any {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		// already a real Go type — pass through (KPI floats, manual labels, etc.)
		return v
	}
	switch t {
	case colInt:
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(f)
		}
	case colMoney, colPercent:
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return v
}

// widthForCol estimates a column width from the header label and a sample of up
// to the first 12 data rows, clamped per column type and capped at 52.
func widthForCol(s ReportSheet, col int) float64 {
	w := float64(len([]rune(s.Headers[col])) + 4)
	switch colTypeAt(s, col) {
	case colMoney:
		if w < 16 {
			w = 16
		}
	case colInt:
		if w < 12 {
			w = 12
		}
	case colPercent:
		if w < 12 {
			w = 12
		}
	case colDate:
		if w < 14 {
			w = 14
		}
	}
	sample := s.Rows
	if len(sample) > 12 {
		sample = sample[:12]
	}
	for _, row := range sample {
		if col >= len(row) {
			continue
		}
		v := fmt.Sprintf("%v", row[col])
		if l := float64(len([]rune(v)) + 2); l > w {
			w = l
		}
	}
	if w > 52 {
		w = 52
	}
	return w
}

// truncSheetName keeps sheet titles within Excel's 31-char limit.
func truncSheetName(name string) string {
	runes := []rune(name)
	if len(runes) <= 31 {
		return name
	}
	return string(runes[:31])
}

func boolp(b bool) *bool { return &b }
