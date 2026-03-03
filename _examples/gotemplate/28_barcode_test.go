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
				{"span": 12, "text": "{{.Title}}", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Code128Label}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "{{.Code1}}", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.WidthLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "{{.Code2}}", "width": "80mm", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.HeightLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "{{.Code3}}", "height": "10mm", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.NumericLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "{{.NumericCode}}", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.ColumnsLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ItemALabel}}", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.ItemA}}", "width": "60mm", "format": "code128"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ItemBLabel}}", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "{{.ItemB}}", "width": "60mm", "format": "code128"}}
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
	testutil.GeneratePDF(t, "28_barcode.pdf", doc)
}
