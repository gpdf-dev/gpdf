package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_31_AbsolutePosition(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "Absolute Positioning", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "Absolute Positioning", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "This is normal flow content. The absolute elements below are overlaid on top."}
			]}},
			{"row": {"height": "5mm", "cols": [{"span": 12, "spacer": "5mm"}]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Product", "Quantity", "Unit Price"],
					"rows": [
						["Alpha", "10", "$5.00"],
						["Beta", "5", "$12.00"],
						["Gamma", "20", "$3.50"]
					]
				}}
			]}}
		],
		"absolute": [
			{
				"x": "130mm",
				"y": "20mm",
				"width": "40mm",
				"elements": [
					{"type": "text", "content": "CONFIDENTIAL", "style": {"size": 12, "bold": true, "color": "#CC3333"}}
				]
			},
			{
				"x": "30mm",
				"y": "150mm",
				"origin": "page",
				"elements": [
					{"type": "text", "content": "DRAFT", "style": {"size": 60, "color": "#E5E5E5"}}
				]
			},
			{
				"x": "0mm",
				"y": "230mm",
				"elements": [
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "20mm"}}
				]
			}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "31_absolute_position.pdf", doc)
}
