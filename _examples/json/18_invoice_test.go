package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_18_Invoice(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "Invoice #INV-2026-001",
			"author": "ACME Corporation"
		},
		"body": [
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "ACME Corporation", "style": {"size": 24, "bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "123 Business Street"},
					{"type": "text", "content": "Suite 100"},
					{"type": "text", "content": "San Francisco, CA 94105"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "INVOICE", "style": {"size": 28, "bold": true, "align": "right", "color": "#1A237E"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "#INV-2026-001", "style": {"align": "right", "size": 12}},
					{"type": "text", "content": "Date: March 1, 2026", "style": {"align": "right"}},
					{"type": "text", "content": "Due: March 31, 2026", "style": {"align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "5mm"},
					{"type": "line", "line": {"color": "#1A237E", "thickness": "2pt"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Bill To:", "style": {"bold": true, "color": "gray(0.4)"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "John Smith", "style": {"bold": true}},
					{"type": "text", "content": "Tech Solutions Inc."},
					{"type": "text", "content": "456 Client Avenue"},
					{"type": "text", "content": "New York, NY 10001"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Payment Info:", "style": {"bold": true, "color": "gray(0.4)"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "Bank: First National Bank"},
					{"type": "text", "content": "Account: 1234-5678-9012"},
					{"type": "text", "content": "Routing: 021000021"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Description", "Qty", "Unit Price", "Amount"],
					"rows": [
						["Web Development - Frontend", "40 hrs", "$150.00", "$6,000.00"],
						["Web Development - Backend", "60 hrs", "$150.00", "$9,000.00"],
						["UI/UX Design", "20 hrs", "$120.00", "$2,400.00"],
						["Database Design", "15 hrs", "$130.00", "$1,950.00"],
						["QA Testing", "25 hrs", "$100.00", "$2,500.00"],
						["Project Management", "10 hrs", "$140.00", "$1,400.00"]
					],
					"columnWidths": [40, 15, 20, 25],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 8},
				{"span": 4, "elements": [
					{"type": "text", "content": "Subtotal:    $23,250.00", "style": {"align": "right"}},
					{"type": "text", "content": "Tax (10%):    $2,325.00", "style": {"align": "right"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "line", "line": {"thickness": "1pt"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "Total:       $25,575.00", "style": {"align": "right", "bold": true, "size": 14}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "15mm"},
					{"type": "line", "line": {"color": "gray(0.8)"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Thank you for your business!", "style": {"align": "center", "italic": true, "color": "gray(0.5)"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "18_invoice.pdf", doc)
}
