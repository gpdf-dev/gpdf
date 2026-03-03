package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_02_TextStyling(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 20, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "{{.NormalText}}"},
				{"span": 3, "text": "{{.BoldText}}", "style": {"bold": true}},
				{"span": 3, "text": "{{.ItalicText}}", "style": {"italic": true}},
				{"span": 3, "text": "{{.BoldItalicText}}", "style": {"bold": true, "italic": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "text": "{{.RedText}}", "style": {"color": "#FF0000"}},
				{"span": 4, "text": "{{.GreenText}}", "style": {"color": "#008000"}},
				{"span": 4, "text": "{{.BlueText}}", "style": {"color": "#0000FF"}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":          "Text Styling Examples",
		"NormalText":     "Normal",
		"BoldText":       "Bold",
		"ItalicText":     "Italic",
		"BoldItalicText": "Bold+Italic",
		"RedText":        "Red",
		"GreenText":      "Green",
		"BlueText":       "Blue",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "02_text_styling.pdf", doc)
}
