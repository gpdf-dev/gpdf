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
				{"span": 12, "text": "{{.Title}}", "style": {"size": 20, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "8mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.NormalText}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.UnderlineText}}", "style": {"underline": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.StrikeText}}", "style": {"strikethrough": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.CombinedText}}", "style": {"underline": true, "strikethrough": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.ColoredUnderline}}", "style": {"underline": true, "color": "#1565C0", "size": 14}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.BoldUnderline}}", "style": {"bold": true, "underline": true, "size": 16}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":           "Text Decoration Demo",
		"NormalText":      "Normal text without decoration",
		"UnderlineText":   "Underlined text for emphasis",
		"StrikeText":      "Strikethrough text for deletions",
		"CombinedText":    "Combined underline and strikethrough",
		"ColoredUnderline": "Colored underlined text",
		"BoldUnderline":   "Bold underlined heading",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "24_text_decoration.pdf", doc)
}
