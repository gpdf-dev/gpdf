package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_33_ComponentReport(t *testing.T) {
	doc := template.Report(template.ReportData{
		Title:    "Quarterly Report",
		Subtitle: "Q1 2026 - Financial Summary",
		Author:   "ACME Corporation",
		Date:     "April 1, 2026",
		Sections: []template.ReportSection{
			{
				Title: "Executive Summary",
				Content: "This report presents the financial performance of ACME Corporation " +
					"for the first quarter of 2026. Revenue increased by 15% compared to Q4 2025, " +
					"driven primarily by strong growth in the cloud services division.",
				Metrics: []template.ReportMetric{
					{Label: "Revenue", Value: "$12.5M", ColorHex: 0x2E7D32},
					{Label: "Growth", Value: "+15%", ColorHex: 0x2E7D32},
					{Label: "Customers", Value: "2,450", ColorHex: 0x1565C0},
					{Label: "Margin", Value: "22%", ColorHex: 0x1565C0},
				},
			},
			{
				Title: "Revenue Breakdown",
				Table: &template.ReportTable{
					Header:       []string{"Division", "Q1 2026", "Q4 2025", "Change"},
					ColumnWidths: []float64{35, 22, 22, 21},
					Rows: [][]string{
						{"Cloud Services", "$5,200,000", "$4,100,000", "+26.8%"},
						{"Enterprise Software", "$3,800,000", "$3,500,000", "+8.6%"},
						{"Consulting", "$2,100,000", "$1,900,000", "+10.5%"},
						{"Support & Maintenance", "$1,400,000", "$1,350,000", "+3.7%"},
					},
				},
			},
			{
				Title: "Expense Summary",
				Table: &template.ReportTable{
					Header:       []string{"Category", "Amount", "% of Revenue"},
					ColumnWidths: []float64{40, 30, 30},
					Rows: [][]string{
						{"Personnel", "$5,500,000", "44.0%"},
						{"Infrastructure", "$1,800,000", "14.4%"},
						{"Marketing", "$1,200,000", "9.6%"},
						{"R&D", "$950,000", "7.6%"},
						{"General & Admin", "$300,000", "2.4%"},
					},
				},
			},
		},
	})

	generatePDF(t, "33_component_report.pdf", doc)
}
