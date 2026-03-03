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
				{"span": 12, "text": "Text Styling Examples", "style": {"size": 20, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Size 8", "style": {"size": 8}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Size 12", "style": {"size": 12}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Size 18", "style": {"size": 18}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Size 24", "style": {"size": 24}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Size 36", "style": {"size": 36}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "Normal"},
				{"span": 3, "text": "Bold", "style": {"bold": true}},
				{"span": 3, "text": "Italic", "style": {"italic": true}},
				{"span": 3, "text": "Bold+Italic", "style": {"bold": true, "italic": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 2, "text": "Red", "style": {"color": "#FF0000"}},
				{"span": 2, "text": "Green", "style": {"color": "#008000"}},
				{"span": 2, "text": "Blue", "style": {"color": "#0000FF"}},
				{"span": 3, "text": "Orange", "style": {"color": "#FF8000"}},
				{"span": 3, "text": "Custom", "style": {"color": "#336699"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "text": "Yellow background", "style": {"background": "#FFFF00"}},
				{"span": 4, "text": "Cyan background", "style": {"background": "#00FFFF"}},
				{"span": 4, "text": "White on dark", "style": {"color": "#FFFFFF", "background": "#333333"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "text": "Left aligned", "style": {"align": "left"}},
				{"span": 4, "text": "Center aligned", "style": {"align": "center"}},
				{"span": 4, "text": "Right aligned", "style": {"align": "right"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "02_text_styling.pdf", doc)
}
