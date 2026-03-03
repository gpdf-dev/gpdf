package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_12_MultiPage(t *testing.T) {
	loremRow := `{"row": {"cols": [{"span": 12, "text": "{{.Lorem}}"}]}}`

	pageBody := `{"body": [
		{"row": {"cols": [{"span": 12, "elements": [
			{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
			{"type": "spacer", "height": "5mm"},
			{"type": "line"},
			{"type": "spacer", "height": "10mm"}
		]}]}},
		` + loremRow + `,` + loremRow + `,` + loremRow + `,` + loremRow + `,` + loremRow + `,` +
		loremRow + `,` + loremRow + `,` + loremRow + `,` + loremRow + `,` + loremRow + `
	]}`

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"pages": [
			` + pageBody + `,
			` + pageBody + `,
			` + pageBody + `,
			` + pageBody + `,
			` + pageBody + `
		]
	}`)

	data := map[string]any{
		"Title": "Multi-Page Document",
		"Lorem": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "12_multi_page.pdf", doc)
}
