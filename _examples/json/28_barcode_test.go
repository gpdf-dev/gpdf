package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_28_Barcode(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "Barcode Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Barcode Examples", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Code 128 barcode:"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "INV-2026-0001"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "With display width (80mm):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "PRODUCT-A-12345", "width": "80mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "With display height (10mm):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "SMALL-BAR", "height": "10mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Numeric data (Code C optimization):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "1234567890"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Barcodes in columns:"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Item A", "style": {"size": 9}},
					{"type": "barcode", "barcode": {"data": "ITEM-A-001", "width": "60mm"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Item B", "style": {"size": 9}},
					{"type": "barcode", "barcode": {"data": "ITEM-B-002", "width": "60mm"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "28_barcode.pdf", doc)
}
