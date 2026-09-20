package util

import (
	"bytes"
	"testing"
)

func TestGeneratePDF(t *testing.T) {
	report := PDFReport{
		Title:    "体检报告 GB202608160001",
		SubTitle: "体检人：王小明",
		Header:   []string{"项目", "结果", "参考值", "异常"},
		Rows: []PDFRow{
			{Columns: []string{"血常规", "5.2", "3.5-9.5", "否"}},
			{Columns: []string{"肝功能ALT", "72", "9-50", "是"}},
		},
		Footer: []string{"结论：发现 1 项异常"},
	}
	content, err := GeneratePDF(report)
	if err != nil {
		t.Fatalf("GeneratePDF() error = %v", err)
	}
	if len(content) < 100 {
		t.Fatalf("PDF too small: %d bytes", len(content))
	}
	if !bytes.HasPrefix(content, []byte("%PDF")) {
		t.Fatal("not a valid PDF header")
	}
}
