package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_09_TableInColumns(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.TeamA}}", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "table", "table": {
						"header": ["Player", "Score"],
						"rows": {{toJSON .TeamARows}},
						"columnWidths": [60, 40]
					}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.TeamB}}", "style": {"bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "table", "table": {
						"header": ["Player", "Score"],
						"rows": {{toJSON .TeamBRows}},
						"columnWidths": [60, 40]
					}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":     "Tables in Grid Columns",
		"TeamA":     "Team A",
		"TeamB":     "Team B",
		"TeamARows": [][]string{{"Alice", "95"}, {"Bob", "87"}, {"Charlie", "92"}},
		"TeamBRows": [][]string{{"Diana", "91"}, {"Eve", "88"}, {"Frank", "85"}},
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "09_table_in_columns.pdf", doc)
}
