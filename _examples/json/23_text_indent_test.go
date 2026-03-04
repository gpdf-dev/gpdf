package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_23_TextIndent(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Text Indent Demo", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "This paragraph has a 24pt first-line indent. The first line starts further to the right, while subsequent lines wrap at the normal left margin. This is commonly used in book typography to indicate new paragraphs.", "style": {"textIndent": "24pt"}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "This paragraph uses a larger 48pt indent for a more dramatic effect. The indentation makes it easy to distinguish where a new paragraph begins without adding extra vertical space.", "style": {"textIndent": "48pt"}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "No indent on this paragraph for comparison. Standard left-aligned text without any first-line indentation starts flush with the left margin."}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "23_text_indent.pdf", doc)
}
