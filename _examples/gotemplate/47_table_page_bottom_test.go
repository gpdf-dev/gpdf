package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_47_TablePageBottom(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "{{.Gap}}"}
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
		"Gap": "266mm",
		"Rows": [][]string{
			{"Alice", "30", "Tokyo"},
			{"Bob", "25", "New York"},
			{"Charlie", "35", "London"},
		},
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "47_table_page_bottom.pdf", doc)
}
