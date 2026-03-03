package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_09_TableInColumns(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Tables in Grid Columns", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Team A", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "table", "table": {
						"header": ["Player", "Score"],
						"rows": [
							["Alice", "95"],
							["Bob", "87"],
							["Charlie", "92"]
						],
						"columnWidths": [60, 40]
					}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Team B", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "table", "table": {
						"header": ["Player", "Score"],
						"rows": [
							["Diana", "91"],
							["Eve", "88"],
							["Frank", "85"]
						],
						"columnWidths": [60, 40]
					}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "09_table_in_columns.pdf", doc)
}
