package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_07_TableBasic(t *testing.T) {
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
				{"span": 12, "table": {
					"header": ["Name", "Age", "City"],
					"rows": {{toJSON .Rows}}
				}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title": "Basic Table",
		"Rows": [][]string{
			{"Alice", "30", "Tokyo"},
			{"Bob", "25", "New York"},
			{"Charlie", "35", "London"},
			{"Diana", "28", "Paris"},
		},
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "07_table_basic.pdf", doc)
}
