package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_43_Opacity(t *testing.T) {
	photo := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Opacity", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "6mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "box": {
					"background": "#1A237E", "padding": "4mm", "opacity": 0.3,
					"elements": [{"type": "text", "content": "Box at 30%%", "style": {"color": "white", "bold": true}}]
				}},
				{"span": 6, "elements": [
					{"type": "text", "content": "Text at 40%%", "style": {"size": 24, "opacity": 0.4}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "6mm"},
					{"type": "image", "image": {"src": "%s", "width": "50mm", "opacity": 0.5}}
				]}
			]}}
		],
		"absolute": [
			{"x": "20mm", "y": "120mm", "elements": [
				{"type": "text", "content": "DRAFT", "style": {"size": 96, "bold": true, "color": "#FF0000", "opacity": 0.15}}
			]}
		]
	}`, base64.StdEncoding.EncodeToString(photo)))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "43_opacity.pdf", doc)
}
