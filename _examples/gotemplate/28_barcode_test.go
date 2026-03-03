package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_28_Barcode(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "{{.Title}}", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Code128Label}}"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.Code1}}"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.WidthLabel}}"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.Code2}}", "width": "80mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.HeightLabel}}"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.Code3}}", "height": "10mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.NumericLabel}}"},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.NumericCode}}"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.ColumnsLabel}}"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ItemALabel}}", "style": {"size": 9}},
					{"type": "barcode", "barcode": {"data": "{{.ItemA}}", "width": "60mm"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ItemBLabel}}", "style": {"size": 9}},
					{"type": "barcode", "barcode": {"data": "{{.ItemB}}", "width": "60mm"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":        "Barcode Examples",
		"Code128Label": "Code 128 barcode:",
		"Code1":        "INV-2026-0001",
		"WidthLabel":   "With display width (80mm):",
		"Code2":        "PRODUCT-A-12345",
		"HeightLabel":  "With display height (10mm):",
		"Code3":        "SMALL-BAR",
		"NumericLabel": "Numeric data (Code C optimization):",
		"NumericCode":  "1234567890",
		"ColumnsLabel": "Barcodes in columns:",
		"ItemALabel":   "Item A",
		"ItemA":        "ITEM-A-001",
		"ItemBLabel":   "Item B",
		"ItemB":        "ITEM-B-002",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "28_barcode.pdf", doc)
}
