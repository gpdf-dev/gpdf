package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_17_Colors(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "Color System Examples", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Predefined Colors:", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 2, "text": "Red", "style": {"color": "#FF0000"}},
				{"span": 2, "text": "Green", "style": {"color": "#008000"}},
				{"span": 2, "text": "Blue", "style": {"color": "#0000FF"}},
				{"span": 2, "text": "Yellow", "style": {"color": "#FFD700"}},
				{"span": 2, "text": "Cyan", "style": {"color": "#00FFFF"}},
				{"span": 2, "text": "Magenta", "style": {"color": "#FF00FF"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Hex Colors:", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "#FF6B6B Coral", "style": {"color": "#FF6B6B"}},
				{"span": 3, "text": "#4ECDC4 Turquoise", "style": {"color": "#4ECDC4"}},
				{"span": 3, "text": "#45B7D1 Sky Blue", "style": {"color": "#45B7D1"}},
				{"span": 3, "text": "#96CEB4 Sage", "style": {"color": "#96CEB4"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Background Color Swatches:", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "Red Background", "style": {"color": "#FFFFFF", "background": "#FF0000"}},
				{"span": 3, "text": "Green Background", "style": {"color": "#FFFFFF", "background": "#008000"}},
				{"span": 3, "text": "Blue Background", "style": {"color": "#FFFFFF", "background": "#0000FF"}},
				{"span": 3, "text": "Yellow Background", "style": {"color": "#000000", "background": "#FFD700"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "17_colors.pdf", doc)
}
