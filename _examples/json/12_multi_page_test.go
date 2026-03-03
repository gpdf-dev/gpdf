package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_12_MultiPage(t *testing.T) {
	// The builder creates 5 explicit pages, each with title + line + 10 paragraphs.
	// We use the "pages" array to create multiple explicit pages.
	loremRow := `{"row": {"cols": [{"span": 12, "text": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."}]}}`

	pageBody := `{"body": [
		{"row": {"cols": [{"span": 12, "elements": [
			{"type": "text", "content": "Multi-Page Document", "style": {"size": 20, "bold": true}},
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

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "12_multi_page.pdf", doc)
}
