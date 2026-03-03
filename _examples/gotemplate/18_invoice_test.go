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
			"title": "{{.InvoiceTitle}}",
			"author": "{{.Company}}"
		},
		"body": [
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.Company}}", "style": {"size": 24, "bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "{{.CompanyAddress1}}"},
					{"type": "text", "content": "{{.CompanyAddress2}}"},
					{"type": "text", "content": "{{.CompanyAddress3}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.InvoiceTitle}}", "style": {"size": 28, "bold": true, "align": "right", "color": "#1A237E"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.InvoiceNumber}}", "style": {"align": "right", "size": 12}},
					{"type": "text", "content": "Date: {{.Date}}", "style": {"align": "right"}},
					{"type": "text", "content": "Due: {{.DueDate}}", "style": {"align": "right"}}
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
					{"type": "text", "content": "Bill To:", "style": {"bold": true, "color": "#666666"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.ClientName}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.ClientCompany}}"},
					{"type": "text", "content": "{{.ClientAddress}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Payment Info:", "style": {"bold": true, "color": "#666666"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.PaymentBank}}"},
					{"type": "text", "content": "{{.PaymentAccount}}"},
					{"type": "text", "content": "{{.PaymentRouting}}"}
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
					{"type": "text", "content": "Subtotal: {{.Subtotal}}", "style": {"align": "right"}},
					{"type": "text", "content": "Tax (10%): {{.Tax}}", "style": {"align": "right"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "line", "line": {"thickness": "1pt"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "Total: {{.Total}}", "style": {"size": 14, "bold": true, "align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "15mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "line", "line": {"color": "#CCCCCC"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.FooterNote}}", "style": {"align": "center", "italic": true, "color": "#999999"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Company":         "ACME Corporation",
		"CompanyAddress1": "123 Business Street",
		"CompanyAddress2": "Suite 100",
		"CompanyAddress3": "San Francisco, CA 94105",
		"InvoiceTitle":    "INVOICE",
		"InvoiceNumber":   "#INV-2026-001",
		"Date":            "March 1, 2026",
		"DueDate":         "March 31, 2026",
		"ClientName":      "John Smith",
		"ClientCompany":   "Tech Solutions Inc.",
		"ClientAddress":   "456 Client Avenue, New York, NY 10001",
		"PaymentBank":     "Bank: First National Bank",
		"PaymentAccount":  "Account: 1234-5678-9012",
		"PaymentRouting":  "Routing: 021000021",
		"Items": [][]string{
			{"Web Development - Frontend", "40 hrs", "$150.00", "$6,000.00"},
			{"Web Development - Backend", "60 hrs", "$150.00", "$9,000.00"},
			{"UI/UX Design", "20 hrs", "$120.00", "$2,400.00"},
			{"Database Design", "15 hrs", "$130.00", "$1,950.00"},
			{"QA Testing", "25 hrs", "$100.00", "$2,500.00"},
			{"Project Management", "10 hrs", "$140.00", "$1,400.00"},
		},
		"Subtotal":   "$23,250.00",
		"Tax":        "$2,325.00",
		"Total":      "$25,575.00",
		"FooterNote": "Thank you for your business! Payment due within 30 days.",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "18_invoice.pdf", doc)
}
