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
				{"span": 12, "elements": [
					{"type": "text", "content": "Color System Examples", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Predefined Colors:", "style": {"bold": true}},
					{"type": "text", "content": "Red", "style": {"color": "red"}},
					{"type": "text", "content": "Green", "style": {"color": "green"}},
					{"type": "text", "content": "Blue", "style": {"color": "blue"}},
					{"type": "text", "content": "Yellow", "style": {"color": "yellow"}},
					{"type": "text", "content": "Cyan", "style": {"color": "cyan"}},
					{"type": "text", "content": "Magenta", "style": {"color": "magenta"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "RGB Colors (float):", "style": {"bold": true}},
					{"type": "text", "content": "RGB(1.0, 0.5, 0.0) - Orange", "style": {"color": "rgb(1.0, 0.5, 0.0)"}},
					{"type": "text", "content": "RGB(0.5, 0.0, 0.5) - Purple", "style": {"color": "rgb(0.5, 0.0, 0.5)"}},
					{"type": "text", "content": "RGB(0.0, 0.5, 0.5) - Teal", "style": {"color": "rgb(0.0, 0.5, 0.5)"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Hex Colors:", "style": {"bold": true}},
					{"type": "text", "content": "#FF6B6B - Coral", "style": {"color": "#FF6B6B"}},
					{"type": "text", "content": "#4ECDC4 - Turquoise", "style": {"color": "#4ECDC4"}},
					{"type": "text", "content": "#45B7D1 - Sky Blue", "style": {"color": "#45B7D1"}},
					{"type": "text", "content": "#96CEB4 - Sage", "style": {"color": "#96CEB4"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Grayscale:", "style": {"bold": true}},
					{"type": "text", "content": "Gray(0.0) - Black", "style": {"color": "gray(0.0)"}},
					{"type": "text", "content": "Gray(0.3) - Dark gray", "style": {"color": "gray(0.3)"}},
					{"type": "text", "content": "Gray(0.5) - Medium gray", "style": {"color": "gray(0.5)"}},
					{"type": "text", "content": "Gray(0.7) - Light gray", "style": {"color": "gray(0.7)"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Background Color Swatches:", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": " Red ", "style": {"color": "white", "background": "red"}},
				{"span": 3, "text": " Green ", "style": {"color": "white", "background": "green"}},
				{"span": 3, "text": " Blue ", "style": {"color": "white", "background": "blue"}},
				{"span": 3, "text": " Yellow ", "style": {"background": "yellow"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "17_colors.pdf", doc)
}
