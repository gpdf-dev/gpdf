package gotemplate_test

import (
	"fmt"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

// TestTmpl_39_TableAfterContent mirrors the builder regression test: a
// multi-page table preceded by other rows must start on the first page.
func TestTmpl_39_TableAfterContent(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "text", "content": "{{.Intro}}"},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Item", "Amount"],
					"rows": {{toJSON .Rows}}
				}}
			]}}
		]
	}`)

	ids := make([]string, 60)
	rows := make([][]string, len(ids))
	for i := range ids {
		ids[i] = fmt.Sprintf("Item-%03d", i+1)
		rows[i] = []string{ids[i], fmt.Sprintf("%d", (i+1)*100)}
	}
	data := map[string]any{
		"Title": "Table After Content",
		"Intro": "Introductory paragraph before the table.",
		"Rows":  rows,
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "39_table_after_content.pdf", doc)

	out, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	testutil.AssertTableSplitsFromFirstPage(t, out, "Table After Content", "Item", ids)
}
