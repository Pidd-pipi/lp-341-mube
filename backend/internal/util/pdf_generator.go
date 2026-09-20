package util

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

// PDFRow 报告表格行。
type PDFRow struct {
	Columns []string
}

// PDFReport 报告数据结构。
type PDFReport struct {
	Title    string
	SubTitle string
	Header   []string
	Rows     []PDFRow
	Footer   []string
}

// GeneratePDF 使用 gofpdf 生成报告 PDF。
func GeneratePDF(report PDFReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(report.Title, false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, report.Title)
	pdf.Ln(8)
	if report.SubTitle != "" {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(0, 7, report.SubTitle)
		pdf.Ln(8)
	}
	colW := 190.0 / float64(max(1, len(report.Header)))
	pdf.SetFont("Arial", "B", 10)
	for _, h := range report.Header {
		pdf.Cell(colW, 8, h)
	}
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 9)
	for _, row := range report.Rows {
		for _, col := range row.Columns {
			pdf.Cell(colW, 7, truncate(col, 24))
		}
		pdf.Ln(7)
	}
	for _, f := range report.Footer {
		pdf.Ln(4)
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(0, 6, f)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
