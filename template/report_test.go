package template

import (
	"testing"
)

func TestReport_Generate(t *testing.T) {
	doc := Report(ReportData{
		Title:    "Test Report",
		Subtitle: "Q1 2026",
		Author:   "Test Author",
		Date:     "2026-04-01",
		Sections: []ReportSection{
			{
				Title:   "Summary",
				Content: "This is a summary paragraph.",
			},
			{
				Title: "Data",
				Table: &ReportTable{
					Header: []string{"A", "B", "C"},
					Rows: [][]string{
						{"1", "2", "3"},
						{"4", "5", "6"},
					},
				},
			},
			{
				Title: "Metrics",
				Metrics: []ReportMetric{
					{Label: "Revenue", Value: "$1M", ColorHex: 0x2E7D32},
					{Label: "Growth", Value: "+10%"},
				},
			},
		},
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Report.Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Report generated empty PDF")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

func TestReport_MinimalData(t *testing.T) {
	doc := Report(ReportData{
		Title: "Minimal Report",
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Report.Generate failed: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}
