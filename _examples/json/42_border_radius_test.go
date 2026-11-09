package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_42_BorderRadius(t *testing.T) {
	banner := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	avatar := testutil.TestImagePNG(t, 100, 100, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Border Radius", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "6mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 4, "box": {
					"background": "#E8EAF6", "padding": "4mm", "borderRadius": "4mm",
					"elements": [{"type": "text", "content": "Filled card"}]
				}},
				{"span": 4, "box": {
					"border": {"width": "1pt", "color": "#1A237E"}, "padding": "4mm", "borderRadius": "3mm",
					"elements": [{"type": "text", "content": "Outlined card"}]
				}},
				{"span": 4, "elements": [
					{"type": "text", "content": "Pill label", "style": {
						"background": "#FFE0B2", "padding": "2mm", "borderRadius": "8pt", "align": "center"
					}}
				]}
			]}},
			{"row": {"cols": [{"span": 12, "elements": [{"type": "spacer", "height": "8mm"}]}]}},
			{"row": {"cols": [
				{"span": 6, "image": {"src": "%s", "width": "60mm", "borderRadius": "5mm"}},
				{"span": 6, "image": {"src": "%s", "width": "30mm", "borderRadius": "15mm"}}
			]}}
		]
	}`, base64.StdEncoding.EncodeToString(banner), base64.StdEncoding.EncodeToString(avatar)))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "42_border_radius.pdf", doc)
}
