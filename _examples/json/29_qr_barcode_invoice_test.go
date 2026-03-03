package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_29_QRBarcodeInvoice(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "Invoice with QR/Barcode", "author": "ACME Corporation"},
		"body": [
			{"row": {"cols": [
				{"span": 8, "elements": [
					{"type": "text", "content": "ACME Corporation", "style": {"size": 22, "bold": true, "color": "#1A237E"}},
					{"type": "text", "content": "Invoice #INV-2026-042", "style": {"size": 12, "color": "gray(0.4)"}}
				]},
				{"span": 4, "qrcode": {"data": "https://pay.acme.com/inv/2026-042", "size": "30mm", "errorCorrection": "H"}}
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
					{"type": "text", "content": "Bill To:", "style": {"bold": true, "color": "gray(0.4)"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "Jane Doe", "style": {"bold": true}},
					{"type": "text", "content": "Tech Solutions Inc."},
					{"type": "text", "content": "456 Client Avenue"},
					{"type": "text", "content": "New York, NY 10001"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Date: March 1, 2026", "style": {"align": "right"}},
					{"type": "text", "content": "Due: March 31, 2026", "style": {"align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "8mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Description", "Qty", "Price", "Amount"],
					"rows": [
						["Web Development", "40 hrs", "$150.00", "$6,000.00"],
						["UI/UX Design", "20 hrs", "$120.00", "$2,400.00"],
						["QA Testing", "15 hrs", "$100.00", "$1,500.00"]
					],
					"columnWidths": [40, 15, 20, 25],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 8},
				{"span": 4, "elements": [
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Total:  $9,900.00", "style": {"align": "right", "bold": true, "size": 14}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "10mm"},
					{"type": "line", "line": {"color": "gray(0.8)"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Order Reference:", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "INV-2026-042", "width": "100mm", "height": "15mm"}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "Scan QR code to pay online", "style": {"align": "center", "italic": true, "size": 9, "color": "gray(0.5)"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "29_qr_barcode_invoice.pdf", doc)
}
