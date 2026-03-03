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
					{"type": "text", "content": "123 Business Avenue"},
					{"type": "text", "content": "Suite 456"},
					{"type": "text", "content": "San Francisco, CA 94102"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "INVOICE", "style": {"size": 28, "bold": true, "align": "right", "color": "#1A237E"}},
					{"type": "text", "content": "Invoice #: INV-2026-001", "style": {"align": "right"}},
					{"type": "text", "content": "Date: March 1, 2026", "style": {"align": "right"}},
					{"type": "text", "content": "Due Date: March 31, 2026", "style": {"align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#1A237E", "thickness": "2pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Bill To:", "style": {"bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "John Smith"},
					{"type": "text", "content": "Tech Solutions Inc."},
					{"type": "text", "content": "789 Client Street"},
					{"type": "text", "content": "New York, NY 10001"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Payment Info:", "style": {"bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "Bank: First National Bank"},
					{"type": "text", "content": "Account: 1234-5678-9012"},
					{"type": "text", "content": "Routing: 021000021"},
					{"type": "text", "content": "Terms: Net 30"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Description", "Qty", "Unit Price", "Amount"],
					"rows": [
						["Web Development - Frontend", "40", "$150.00", "$6,000.00"],
						["Web Development - Backend", "60", "$175.00", "$10,500.00"],
						["UI/UX Design", "20", "$125.00", "$2,500.00"],
						["Database Architecture", "15", "$200.00", "$3,000.00"],
						["Quality Assurance", "25", "$100.00", "$2,500.00"],
						["Project Management", "10", "$150.00", "$1,500.00"]
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
				{"span": 8, "text": ""},
				{"span": 4, "elements": [
					{"type": "text", "content": "Subtotal: $26,000.00", "style": {"align": "right"}},
					{"type": "text", "content": "Tax (10%): $2,600.00", "style": {"align": "right"}},
					{"type": "line"},
					{"type": "text", "content": "Total: $28,600.00", "style": {"align": "right", "bold": true, "size": 14}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#CCCCCC"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Thank you for your business!", "style": {"align": "center", "italic": true, "color": "#808080"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "18_invoice.pdf", doc)
}
