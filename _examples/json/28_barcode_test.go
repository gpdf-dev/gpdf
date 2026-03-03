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
				{"span": 12, "text": "Barcode Examples", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Code 128 barcode:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "INV-2026-0001", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "With display width (80mm):"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "PRODUCT-A-12345", "width": "80mm", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "With display height (10mm):"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "SMALL-BAR", "height": "10mm", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Numeric data (Code C optimization):"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "barcode": {"data": "1234567890", "format": "code128"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Barcodes in columns:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Item A", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "ITEM-A-001", "width": "60mm", "format": "code128"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Item B", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "barcode", "barcode": {"data": "ITEM-B-002", "width": "60mm", "format": "code128"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "28_barcode.pdf", doc)
}
