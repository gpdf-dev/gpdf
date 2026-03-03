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
				{"span": 12, "text": "Text Decoration Demo", "style": {"size": 20, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "8mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Normal text without decoration"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Underlined text for emphasis", "style": {"underline": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Strikethrough text for deletions", "style": {"strikethrough": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Combined underline and strikethrough", "style": {"underline": true, "strikethrough": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Colored underlined text", "style": {"underline": true, "color": "#1565C0", "size": 14}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "4mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Bold underlined heading", "style": {"bold": true, "underline": true, "size": 16}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "24_text_decoration.pdf", doc)
}
