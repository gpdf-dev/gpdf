package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_29_QRBarcodeInvoice(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "Invoice with QR/Barcode", "author": "{{.Company}}"},
		"body": [
			{"row": {"cols": [
				{"span": 8, "elements": [
					{"type": "text", "content": "{{.Company}}", "style": {"size": 22, "bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "{{.InvoiceNumber}}", "style": {"size": 12, "color": "#666666"}}
				]},
				{"span": 4, "qrcode": {"data": "{{.PaymentURL}}", "size": "30mm"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "3mm"},
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
					{"type": "text", "content": "{{.ClientAddress1}}"},
					{"type": "text", "content": "{{.ClientAddress2}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.Date}}", "style": {"align": "right"}},
					{"type": "text", "content": "{{.DueDate}}", "style": {"align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "8mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Description", "Qty", "Price", "Amount"],
					"rows": {{toJSON .Items}},
					"columnWidths": [40, 15, 20, 25],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 8, "text": ""},
				{"span": 4, "elements": [
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.Total}}", "style": {"align": "right", "bold": true, "size": 14}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "10mm"},
					{"type": "line", "line": {"color": "#CCCCCC"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Order Reference:", "style": {"size": 9, "color": "#666666"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.BarcodeData}}", "width": "100mm", "height": "15mm", "format": "code128"}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "{{.ScanNote}}", "style": {"align": "center", "italic": true, "size": 9, "color": "#808080"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Company":        "ACME Corporation",
		"InvoiceNumber":  "Invoice #INV-2026-042",
		"PaymentURL":     "https://pay.acme.com/inv/2026-042",
		"ClientName":     "Jane Doe",
		"ClientCompany":  "Tech Solutions Inc.",
		"ClientAddress1": "456 Client Avenue",
		"ClientAddress2": "New York, NY 10001",
		"Date":           "Date: March 1, 2026",
		"DueDate":        "Due: March 31, 2026",
		"Items": [][]string{
			{"Web Development", "40 hrs", "$150.00", "$6,000.00"},
			{"UI/UX Design", "20 hrs", "$120.00", "$2,400.00"},
			{"QA Testing", "15 hrs", "$100.00", "$1,500.00"},
		},
		"Total":       "Total:  $9,900.00",
		"BarcodeData": "INV-2026-042",
		"ScanNote":    "Scan QR code to pay online",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "29_qr_barcode_invoice.pdf", doc)
}
