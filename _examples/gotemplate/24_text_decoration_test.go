package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_24_TextDecoration(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "{{.NormalText}}"},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "{{.UnderlineText}}", "style": {"underline": true}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "{{.StrikeText}}", "style": {"strikethrough": true}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "{{.CombinedText}}", "style": {"underline": true, "strikethrough": true}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "{{.ColoredUnderline}}", "style": {"underline": true, "color": "#1565C0", "size": 14}},
					{"type": "spacer", "height": "4mm"},
					{"type": "text", "content": "{{.BoldUnderline}}", "style": {"bold": true, "underline": true, "size": 16}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":            "Text Decoration Demo",
		"NormalText":       "Normal text without decoration",
		"UnderlineText":    "Underlined text for emphasis",
		"StrikeText":       "Strikethrough text for deletions",
		"CombinedText":     "Combined underline and strikethrough",
		"ColoredUnderline": "Colored underlined text",
		"BoldUnderline":    "Bold underlined heading",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "24_text_decoration.pdf", doc)
}
