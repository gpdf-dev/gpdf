package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_24_TextDecoration(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Text Decoration Demo", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "Normal text without decoration"},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "Underlined text for emphasis", "style": {"underline": true}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "Strikethrough text for deletions", "style": {"strikethrough": true}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "Combined underline and strikethrough", "style": {"underline": true, "strikethrough": true}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "Colored underlined text", "style": {"underline": true, "color": "#1565C0", "size": 14}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "Bold underlined heading", "style": {"bold": true, "underline": true, "size": 16}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "24_text_decoration.pdf", doc)
}
