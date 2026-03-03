package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_19_Report(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "Quarterly Report Q1 2026",
			"author": "ACME Corporation",
			"subject": "Q1 2026 Financial Summary"
		},
		"header": [
			{"row": {"cols": [
				{"span": 6, "text": "ACME Corp", "style": {"bold": true, "size": 9, "color": "#1565C0"}},
				{"span": 6, "text": "Q1 2026 Report", "style": {"align": "right", "size": 9, "color": "#808080"}}
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
				{"span": 12, "text": "Confidential - For Internal Use Only", "style": {"align": "center", "size": 7, "color": "#808080"}}
			]}}
		],
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "Quarterly Report", "style": {"size": 28, "bold": true, "align": "center", "color": "#1A237E"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Q1 2026 - Financial Summary", "style": {"size": 16, "align": "center", "color": "#666666"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#1A237E", "thickness": "2pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Executive Summary", "style": {"size": 16, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "The first quarter of 2026 has shown strong performance across all business divisions. Revenue grew by 15% year-over-year, driven primarily by expansion in the enterprise segment and successful product launches. Our customer base expanded to 2,450 active accounts, reflecting a 12% increase from the previous quarter."}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [
					{"type": "text", "content": "Revenue", "style": {"bold": true, "align": "center"}},
					{"type": "text", "content": "$12.5M", "style": {"size": 20, "bold": true, "align": "center", "color": "#1A237E"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "Growth", "style": {"bold": true, "align": "center"}},
					{"type": "text", "content": "+15%", "style": {"size": 20, "bold": true, "align": "center", "color": "#2E7D32"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "Customers", "style": {"bold": true, "align": "center"}},
					{"type": "text", "content": "2,450", "style": {"size": 20, "bold": true, "align": "center", "color": "#1565C0"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "Margin", "style": {"bold": true, "align": "center"}},
					{"type": "text", "content": "22%", "style": {"size": 20, "bold": true, "align": "center", "color": "#F57F17"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Revenue Breakdown", "style": {"size": 16, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Division", "Q1 Revenue", "Q4 Revenue", "Change", "% of Total"],
					"rows": [
						["Enterprise", "$5,200,000", "$4,500,000", "+15.6%", "41.6%"],
						["Mid-Market", "$3,800,000", "$3,400,000", "+11.8%", "30.4%"],
						["SMB", "$2,100,000", "$1,900,000", "+10.5%", "16.8%"],
						["Partnerships", "$1,400,000", "$1,050,000", "+33.3%", "11.2%"]
					],
					"columnWidths": [20, 20, 20, 20, 20],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Highlights", "style": {"size": 14, "bold": true, "color": "#2E7D32"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "list", "list": {"items": [
						"Enterprise segment grew 15.6% QoQ",
						"Partnership revenue up 33.3%",
						"Customer acquisition cost down 8%",
						"Net promoter score improved to 72"
					]}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Challenges", "style": {"size": 14, "bold": true, "color": "#C62828"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "list", "list": {"items": [
						"SMB churn rate increased to 5.2%",
						"Infrastructure costs rose 12%",
						"Hiring targets missed by 15%",
						"APAC expansion delayed to Q2"
					]}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "19_report.pdf", doc)
}
