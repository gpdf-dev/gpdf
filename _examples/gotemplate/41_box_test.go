package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_41_Box(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "6mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 3, "box": {
					"background": "#E8EAF6", "height": "20mm", "padding": "3mm",
					"elements": [{"type": "text", "content": "LOGO", "style": {"bold": true, "align": "center"}}]
				}},
				{"span": 5, "box": {
					"padding": "2mm",
					"elements": [
						{"type": "text", "content": "line 1"},
						{"type": "text", "content": "line 2"},
						{"type": "text", "content": "line 3"}
					]
				}},
				{"span": 4, "box": {
					"border": {"width": "1pt", "color": "#1A237E"},
					"paddings": ["2mm", "3mm", "2mm", "3mm"],
					"elements": [
						{"type": "text", "content": "line 1", "style": {"align": "right"}},
						{"type": "text", "content": "line 2", "style": {"align": "right"}},
						{"type": "text", "content": "line 3", "style": {"align": "right"}}
					]
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "6mm"},
					{"type": "box", "box": {
						"border": {"width": "1pt", "color": "#1A237E"},
						"padding": "4mm",
						"elements": [
							{"type": "text", "content": "Outer box", "style": {"bold": true}},
							{"type": "box", "box": {
								"background": "#FFF3E0", "padding": "3mm",
								"elements": [{"type": "text", "content": "Nested box with its own background"}]
							}}
						]
					}}
				]}
			]}}
		]
	}`)

	data := map[string]any{"Title": "Box Demo"}
	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "41_box.pdf", doc)
}
