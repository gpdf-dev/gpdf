package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_18_Invoice(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "{{.InvoiceID}}",
			"author": "{{.Company}}"
		},
		"body": [
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.Company}}", "style": {"size": 24, "bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "{{.Addr1}}"},
					{"type": "text", "content": "{{.Addr2}}"},
					{"type": "text", "content": "{{.Addr3}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.InvoiceTitle}}", "style": {"size": 28, "bold": true, "align": "right", "color": "#1A237E"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.InvoiceNum}}", "style": {"align": "right", "size": 12}},
					{"type": "text", "content": "{{.DateLine}}", "style": {"align": "right"}},
					{"type": "text", "content": "{{.DueLine}}", "style": {"align": "right"}}
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
					{"type": "text", "content": "{{.BillToLabel}}", "style": {"bold": true, "color": "gray(0.4)"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.ClientName}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.ClientCompany}}"},
					{"type": "text", "content": "{{.ClientAddr1}}"},
					{"type": "text", "content": "{{.ClientAddr2}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.PayInfoLabel}}", "style": {"bold": true, "color": "gray(0.4)"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.Bank}}"},
					{"type": "text", "content": "{{.Account}}"},
					{"type": "text", "content": "{{.Routing}}"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Description", "Qty", "Unit Price", "Amount"],
					"rows": {{toJSON .Items}},
					"columnWidths": [40, 15, 20, 25],
					"headerStyle": {"color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 8},
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.Subtotal}}", "style": {"align": "right"}},
					{"type": "text", "content": "{{.Tax}}", "style": {"align": "right"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "line", "line": {"thickness": "1pt"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.Total}}", "style": {"align": "right", "bold": true, "size": 14}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "15mm"},
					{"type": "line", "line": {"color": "gray(0.8)"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.FooterNote}}", "style": {"align": "center", "italic": true, "color": "gray(0.5)"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Company":       "ACME Corporation",
		"Addr1":         "123 Business Street",
		"Addr2":         "Suite 100",
		"Addr3":         "San Francisco, CA 94105",
		"InvoiceTitle":  "INVOICE",
		"InvoiceID":     "Invoice #INV-2026-001",
		"InvoiceNum":    "#INV-2026-001",
		"DateLine":      "Date: March 1, 2026",
		"DueLine":       "Due: March 31, 2026",
		"BillToLabel":   "Bill To:",
		"ClientName":    "John Smith",
		"ClientCompany": "Tech Solutions Inc.",
		"ClientAddr1":   "456 Client Avenue",
		"ClientAddr2":   "New York, NY 10001",
		"PayInfoLabel":  "Payment Info:",
		"Bank":          "Bank: First National Bank",
		"Account":       "Account: 1234-5678-9012",
		"Routing":       "Routing: 021000021",
		"Items": [][]string{
			{"Web Development - Frontend", "40 hrs", "$150.00", "$6,000.00"},
			{"Web Development - Backend", "60 hrs", "$150.00", "$9,000.00"},
			{"UI/UX Design", "20 hrs", "$120.00", "$2,400.00"},
			{"Database Design", "15 hrs", "$130.00", "$1,950.00"},
			{"QA Testing", "25 hrs", "$100.00", "$2,500.00"},
			{"Project Management", "10 hrs", "$140.00", "$1,400.00"},
		},
		"Subtotal":   "Subtotal:    $23,250.00",
		"Tax":        "Tax (10%):    $2,325.00",
		"Total":      "Total:       $25,575.00",
		"FooterNote": "Thank you for your business!",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "18_invoice.pdf", doc)
}
