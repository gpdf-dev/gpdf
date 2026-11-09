package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_40_RichText(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "6mm"},
					{"type": "richText", "richText": {
						"fragments": [
							{"text": "Welcome to ", "style": {"size": 12}},
							{"text": "{{.AppName}}", "style": {"size": 14, "bold": true, "color": "#007BFF"}},
							{"text": "! This paragraph mixes inline styles.", "style": {"size": 12, "italic": true}}
						]
					}},
					{"type": "spacer", "height": "4mm"},
					{"type": "richText", "richText": {
						"style": {"align": "center"},
						"fragments": [
							{"text": "Status: "},
							{"text": "{{.Status}}", "style": {"bold": true, "color": "#2E7D32"}},
							{"text": " — "},
							{"text": "old price", "style": {"strikethrough": true}},
							{"text": " "},
							{"text": "new price", "style": {"underline": true}}
						]
					}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":   "Rich Text Demo",
		"AppName": "gpdf Engine",
		"Status":  "PAID",
	}
	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "40_rich_text.pdf", doc)
}
