package gotemplate_test

import (
	"testing"
	gotemplate "text/template"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_19_Report(t *testing.T) {
	tmplStr := `{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "{{.Title}}",
			"author": "{{.Company}}",
			"subject": "{{.Subtitle}}"
		},
		"header": [
			{"row": {"cols": [
				{"span": 6, "text": "{{.Company}}", "style": {"bold": true, "size": 9, "color": "#1565C0"}},
				{"span": 6, "text": "{{.HeaderRight}}", "style": {"align": "right", "size": 9, "color": "#808080"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#1565C0"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}}
		],
		"footer": [
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#CCCCCC"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.FooterText}}", "style": {"align": "center", "size": 7, "color": "#808080"}}
			]}}
		],
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "20mm"},
					{"type": "text", "content": "{{.Title}}", "style": {"size": 28, "bold": true, "align": "center", "color": "#1A237E"}},
					{"type": "text", "content": "{{.Subtitle}}", "style": {"size": 16, "align": "center", "color": "#666666"}},
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
				{{- range $i, $m := .Metrics}}
				{{- if $i}},{{end}}
				{"span": 3, "elements": [
					{"type": "text", "content": "{{$m.Label}}", "style": {"color": "#808080", "size": 9}},
					{"type": "text", "content": "{{$m.Value}}", "style": {"size": 18, "bold": true, "color": "{{$m.Color}}"}}
				]}
				{{- end}}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "10mm"},
					{"type": "text", "content": "{{.RevenueHeading}}", "style": {"size": 16, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Division", "Q1 2026", "Q4 2025", "Change"],
					"rows": {{toJSON .RevenueRows}},
					"columnWidths": [35, 22, 22, 21],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
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
		]
	}`

	tmpl, err := gotemplate.New("report").Funcs(template.TemplateFuncMap()).Parse(tmplStr)
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	type metric struct {
		Label string
		Value string
		Color string
	}

	data := map[string]any{
		"Title":       "Quarterly Report",
		"Subtitle":    "Q1 2026 - Financial Summary",
		"Company":     "ACME Corp",
		"HeaderRight": "Q1 2026 Report",
		"FooterText":  "Confidential - For Internal Use Only",
		"SummaryHeading": "Executive Summary",
		"SummaryText1":   "This report presents the financial performance of ACME Corporation for the first quarter of 2026. Revenue increased by 15% compared to Q4 2025, driven primarily by strong growth in the cloud services division. Operating margins improved to 22%, up from 19% in the previous quarter.",
		"SummaryText2":   "Key highlights include the successful launch of three new product lines, expansion into the European market, and a 20% reduction in customer churn rate. The company remains well-positioned for continued growth throughout 2026.",
		"MetricsHeading": "Key Metrics",
		"Metrics": []metric{
			{Label: "Revenue", Value: "$12.5M", Color: "#2E7D32"},
			{Label: "Growth", Value: "+15%", Color: "#2E7D32"},
			{Label: "Customers", Value: "2,450", Color: "#1565C0"},
			{Label: "Margin", Value: "22%", Color: "#1565C0"},
		},
		"RevenueHeading": "Revenue Breakdown",
		"RevenueRows": [][]string{
			{"Cloud Services", "$5,200,000", "$4,100,000", "+26.8%"},
			{"Enterprise Software", "$3,800,000", "$3,500,000", "+8.6%"},
			{"Consulting", "$2,100,000", "$1,900,000", "+10.5%"},
			{"Support & Maintenance", "$1,400,000", "$1,350,000", "+3.7%"},
		},
		"HighlightsHeading":  "Highlights",
		"HighlightsText":     "Cloud services revenue grew 26.8%, exceeding projections by 5%. New enterprise clients added: 47.",
		"ChallengesHeading":  "Challenges",
		"ChallengesText":     "Infrastructure costs rose 12% due to scaling needs. Two major client renewals deferred to Q2.",
	}

	doc, err := template.FromTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("FromTemplate error: %v", err)
	}
	testutil.GeneratePDF(t, "19_report.pdf", doc)
}
