package handler

import (
	"fmt"
	"net/http"
	"time"

	"clothes-store/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type AdminReportHandler struct{ repo *repository.ReportsRepo }

func NewAdminReportHandler(repo *repository.ReportsRepo) *AdminReportHandler {
	return &AdminReportHandler{repo: repo}
}

func (h *AdminReportHandler) Excel(c *gin.Context) {
	sheets, err := h.repo.AllSheets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Color:  "FFFFFF",
			Family: "Manrope",
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E11900"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "B0140A", Style: 1},
		},
	})
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Color:  "14141A",
			Family: "Manrope",
		},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Italic: true,
			Size:   10,
			Color:  "6C6C78",
			Family: "Manrope",
		},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11, Color: "14141A", Family: "Manrope"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "E5E5E5", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	// Excelize creates "Sheet1" by default — we'll rename it for the first report.
	defaultSheet := f.GetSheetName(0)
	generatedAt := time.Now().Format("02.01.2006 15:04")

	for i, sheet := range sheets {
		name := truncSheetName(sheet.Name)
		var sheetIdx int
		if i == 0 {
			f.SetSheetName(defaultSheet, name)
			sheetIdx, _ = f.GetSheetIndex(name)
		} else {
			sheetIdx, _ = f.NewSheet(name)
		}

		// Title + subtitle
		f.SetCellValue(name, "A1", sheet.Name)
		f.SetCellStyle(name, "A1", "A1", titleStyle)
		f.SetCellValue(name, "A2", "Сформировано: "+generatedAt)
		f.SetCellStyle(name, "A2", "A2", subtitleStyle)

		// Header row at row 4
		headerRow := 4
		for col, h := range sheet.Headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, headerRow)
			f.SetCellValue(name, cell, h)
		}
		startCell, _ := excelize.CoordinatesToCellName(1, headerRow)
		endCell, _ := excelize.CoordinatesToCellName(len(sheet.Headers), headerRow)
		f.SetCellStyle(name, startCell, endCell, headerStyle)
		f.SetRowHeight(name, headerRow, 26)

		// Data rows
		for r, row := range sheet.Rows {
			for col, v := range row {
				cell, _ := excelize.CoordinatesToCellName(col+1, headerRow+1+r)
				f.SetCellValue(name, cell, v)
			}
		}
		if len(sheet.Rows) > 0 {
			firstData, _ := excelize.CoordinatesToCellName(1, headerRow+1)
			lastData, _ := excelize.CoordinatesToCellName(len(sheet.Headers), headerRow+len(sheet.Rows))
			f.SetCellStyle(name, firstData, lastData, cellStyle)
		}

		// Column widths
		for col := range sheet.Headers {
			colName, _ := excelize.ColumnNumberToName(col + 1)
			width := maxColWidth(sheet, col)
			f.SetColWidth(name, colName, colName, width)
		}

		// Freeze header row
		f.SetPanes(name, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      headerRow,
			TopLeftCell: "A" + fmt.Sprintf("%d", headerRow+1),
			ActivePane:  "bottomLeft",
		})

		_ = sheetIdx
	}

	f.SetActiveSheet(0)

	fname := fmt.Sprintf("report_%s.xlsx", time.Now().Format("2006-01-02_1504"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fname)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Status(http.StatusOK)

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// truncSheetName keeps sheet titles inside Excel's 31-char limit.
func truncSheetName(name string) string {
	runes := []rune(name)
	if len(runes) <= 31 {
		return name
	}
	return string(runes[:31])
}

// maxColWidth picks a column width based on header + sample row lengths.
func maxColWidth(s ReportSheet, col int) float64 {
	w := float64(len([]rune(s.Headers[col])) + 4)
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
	if w < 14 {
		w = 14
	}
	if w > 48 {
		w = 48
	}
	return w
}

// ReportSheet re-exported so handler's maxColWidth can use the type without importing repository here.
type ReportSheet = repository.ReportSheet
