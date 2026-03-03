package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_19_Report(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "{{.MetaTitle}}",
			"author": "{{.MetaAuthor}}",
			"subject": "{{.MetaSubject}}"
		},
		"header": [
			{"row": {"cols": [
				{"span": 6, "text": "{{.HeaderLeft}}", "style": {"bold": true, "size": 9, "color": "#1565C0"}},
				{"span": 6, "text": "{{.HeaderRight}}", "style": {"align": "right", "size": 9, "color": "gray(0.5)"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "line", "line": {"color": "#1565C0"}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}}
		],
		"footer": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "3mm"},
					{"type": "line", "line": {"color": "gray(0.8)"}},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.FooterText}}", "style": {"align": "center", "size": 7, "color": "gray(0.5)"}}
			]}}
		],
		"pages": [
			{"body": [
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "spacer", "height": "20mm"},
						{"type": "text", "content": "{{.Title}}", "style": {"size": 28, "bold": true, "align": "center", "color": "#1A237E"}},
						{"type": "text", "content": "{{.Subtitle}}", "style": {"size": 16, "align": "center", "color": "gray(0.4)"}},
						{"type": "spacer", "height": "15mm"},
						{"type": "line", "line": {"color": "#1A237E", "thickness": "2pt"}},
						{"type": "spacer", "height": "10mm"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "text", "content": "{{.SummaryHeading}}", "style": {"size": 16, "bold": true}},
						{"type": "spacer", "height": "3mm"},
						{"type": "text", "content": "{{.SummaryText1}}"},
						{"type": "spacer", "height": "5mm"},
						{"type": "text", "content": "{{.SummaryText2}}"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "spacer", "height": "10mm"},
						{"type": "text", "content": "{{.MetricsHeading}}", "style": {"size": 14, "bold": true}},
						{"type": "spacer", "height": "3mm"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 3, "elements": [
						{"type": "text", "content": "{{.Metric1Label}}", "style": {"color": "gray(0.5)", "size": 9}},
						{"type": "text", "content": "{{.Metric1Value}}", "style": {"size": 18, "bold": true, "color": "#2E7D32"}}
					]},
					{"span": 3, "elements": [
						{"type": "text", "content": "{{.Metric2Label}}", "style": {"color": "gray(0.5)", "size": 9}},
						{"type": "text", "content": "{{.Metric2Value}}", "style": {"size": 18, "bold": true, "color": "#2E7D32"}}
					]},
					{"span": 3, "elements": [
						{"type": "text", "content": "{{.Metric3Label}}", "style": {"color": "gray(0.5)", "size": 9}},
						{"type": "text", "content": "{{.Metric3Value}}", "style": {"size": 18, "bold": true, "color": "#1565C0"}}
					]},
					{"span": 3, "elements": [
						{"type": "text", "content": "{{.Metric4Label}}", "style": {"color": "gray(0.5)", "size": 9}},
						{"type": "text", "content": "{{.Metric4Value}}", "style": {"size": 18, "bold": true, "color": "#1565C0"}}
					]}
				]}}
			]},
			{"body": [
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "text", "content": "{{.RevenueHeading}}", "style": {"size": 16, "bold": true}},
						{"type": "spacer", "height": "5mm"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "table", "table": {
							"header": ["Division", "Q1 2026", "Q4 2025", "Change"],
							"rows": [
								["Cloud Services", "$5,200,000", "$4,100,000", "+26.8%"],
								["Enterprise Software", "$3,800,000", "$3,500,000", "+8.6%"],
								["Consulting", "$2,100,000", "$1,900,000", "+10.5%"],
								["Support & Maintenance", "$1,400,000", "$1,350,000", "+3.7%"]
							],
							"columnWidths": [35, 22, 22, 21],
							"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
							"stripeColor": "#F5F5F5"
						}},
						{"type": "spacer", "height": "10mm"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "text", "content": "{{.ExpenseHeading}}", "style": {"size": 16, "bold": true}},
						{"type": "spacer", "height": "5mm"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 12, "elements": [
						{"type": "table", "table": {
							"header": ["Category", "Amount", "% of Revenue"],
							"rows": [
								["Personnel", "$5,500,000", "44.0%"],
								["Infrastructure", "$1,800,000", "14.4%"],
								["Marketing", "$1,200,000", "9.6%"],
								["R&D", "$950,000", "7.6%"],
								["General & Admin", "$300,000", "2.4%"]
							],
							"columnWidths": [40, 30, 30],
							"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
							"stripeColor": "#F5F5F5"
						}},
						{"type": "spacer", "height": "10mm"}
					]}
				]}},
				{"row": {"cols": [
					{"span": 6, "elements": [
						{"type": "text", "content": "{{.HighlightsHeading}}", "style": {"bold": true, "color": "#2E7D32"}},
						{"type": "spacer", "height": "2mm"},
						{"type": "text", "content": "{{.HighlightsText}}"}
					]},
					{"span": 6, "elements": [
						{"type": "text", "content": "{{.ChallengesHeading}}", "style": {"bold": true, "color": "#C62828"}},
						{"type": "spacer", "height": "2mm"},
						{"type": "text", "content": "{{.ChallengesText}}"}
					]}
				]}}
			]}
		]
	}`)

	data := map[string]any{
		"MetaTitle":          "Quarterly Report Q1 2026",
		"MetaAuthor":         "ACME Corporation",
		"MetaSubject":        "Q1 2026 Financial Summary",
		"HeaderLeft":         "ACME Corp",
		"HeaderRight":        "Q1 2026 Report",
		"FooterText":         "Confidential - For Internal Use Only",
		"Title":              "Quarterly Report",
		"Subtitle":           "Q1 2026 - Financial Summary",
		"SummaryHeading":     "Executive Summary",
		"SummaryText1":       "This report presents the financial performance of ACME Corporation for the first quarter of 2026. Revenue increased by 15% compared to Q4 2025, driven primarily by strong growth in the cloud services division. Operating margins improved to 22%, up from 19% in the previous quarter.",
		"SummaryText2":       "Key highlights include the successful launch of three new product lines, expansion into the European market, and a 20% reduction in customer churn rate. The company remains well-positioned for continued growth throughout 2026.",
		"MetricsHeading":     "Key Metrics",
		"Metric1Label":       "Revenue",
		"Metric1Value":       "$12.5M",
		"Metric2Label":       "Growth",
		"Metric2Value":       "+15%",
		"Metric3Label":       "Customers",
		"Metric3Value":       "2,450",
		"Metric4Label":       "Margin",
		"Metric4Value":       "22%",
		"RevenueHeading":     "Revenue Breakdown",
		"ExpenseHeading":     "Expense Summary",
		"HighlightsHeading":  "Highlights",
		"HighlightsText":     "Cloud services revenue grew 26.8%, exceeding projections by 5%. New enterprise clients added: 47.",
		"ChallengesHeading":  "Challenges",
		"ChallengesText":     "Infrastructure costs rose 12% due to scaling needs. Two major client renewals deferred to Q2.",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "19_report.pdf", doc)
}
