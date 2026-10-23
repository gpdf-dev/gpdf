package json_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

// TestJSON_39_TableAfterContent mirrors the builder regression test: a
// multi-page table preceded by other rows must start on the first page.
func TestJSON_39_TableAfterContent(t *testing.T) {
	ids := make([]string, 60)
	rowJSON := make([]string, len(ids))
	for i := range ids {
		ids[i] = fmt.Sprintf("Item-%03d", i+1)
		rowJSON[i] = fmt.Sprintf(`["%s", "%d"]`, ids[i], (i+1)*100)
	}

	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Table After Content", "style": {"size": 18, "bold": true}},
					{"type": "text", "content": "Introductory paragraph before the table."},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Item", "Amount"],
					"rows": [%s]
				}}
			]}}
		]
	}`, strings.Join(rowJSON, ",")))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "39_table_after_content.pdf", doc)

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	testutil.AssertTableSplitsFromFirstPage(t, data, "Table After Content", "Item", ids)
}
