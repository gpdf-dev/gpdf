package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_25_TableVerticalAlign(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Table Vertical Align Demo", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "Default (Top) Alignment:", "style": {"bold": true}},
					{"type": "spacer", "height": "3mm"},
					{"type": "table", "table": {
						"header": ["Short", "Tall Cell"],
						"rows": [
							["A", "This cell has\nmuch more content\nthat spans\nmultiple lines"],
							["B", "Another tall\ncell with\nlong text"]
						],
						"headerStyle": {"bold": true, "color": "white", "background": "#1565C0"}
					}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "Middle Alignment:", "style": {"bold": true}},
					{"type": "spacer", "height": "3mm"},
					{"type": "table", "table": {
						"header": ["Short", "Tall Cell"],
						"rows": [
							["A", "This cell has\nmuch more content\nthat spans\nmultiple lines"],
							["B", "Another tall\ncell with\nlong text"]
						],
						"headerStyle": {"bold": true, "color": "white", "background": "#2E7D32"},
						"cellVAlign": "middle"
					}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "Bottom Alignment:", "style": {"bold": true}},
					{"type": "spacer", "height": "3mm"},
					{"type": "table", "table": {
						"header": ["Short", "Tall Cell"],
						"rows": [
							["A", "This cell has\nmuch more content\nthat spans\nmultiple lines"],
							["B", "Another tall\ncell with\nlong text"]
						],
						"headerStyle": {"bold": true, "color": "white", "background": "#E65100"},
						"cellVAlign": "bottom"
					}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "25_table_vertical_align.pdf", doc)
}
