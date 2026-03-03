package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_02_TextStyling(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Text Styling Examples", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Font Size 8pt", "style": {"size": 8}},
					{"type": "text", "content": "Font Size 12pt (default)", "style": {"size": 12}},
					{"type": "text", "content": "Font Size 18pt", "style": {"size": 18}},
					{"type": "text", "content": "Font Size 24pt", "style": {"size": 24}},
					{"type": "text", "content": "Font Size 36pt", "style": {"size": 36}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Normal text"},
					{"type": "text", "content": "Bold text", "style": {"bold": true}},
					{"type": "text", "content": "Italic text", "style": {"italic": true}},
					{"type": "text", "content": "Bold + Italic text", "style": {"bold": true, "italic": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Red text", "style": {"color": "red"}},
					{"type": "text", "content": "Green text", "style": {"color": "green"}},
					{"type": "text", "content": "Blue text", "style": {"color": "blue"}},
					{"type": "text", "content": "Custom color (orange)", "style": {"color": "rgb(1.0, 0.5, 0.0)"}},
					{"type": "text", "content": "Hex color (#336699)", "style": {"color": "#336699"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Yellow background", "style": {"background": "yellow"}},
					{"type": "text", "content": "Cyan background", "style": {"background": "cyan"}},
					{"type": "text", "content": "White text on dark background", "style": {"color": "white", "background": "#333333"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Left aligned (default)", "style": {"align": "left"}},
					{"type": "text", "content": "Center aligned", "style": {"align": "center"}},
					{"type": "text", "content": "Right aligned", "style": {"align": "right"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "02_text_styling.pdf", doc)
}
