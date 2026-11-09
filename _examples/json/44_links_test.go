package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_44_Links(t *testing.T) {
	logo := testutil.TestImagePNG(t, 120, 40, color.RGBA{R: 26, G: 35, B: 126, A: 255})
	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Links", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "6mm"},
					{"type": "link", "content": "Visit gpdf.dev", "style": {"link": "https://gpdf.dev"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "richText", "richText": {"fragments": [
						{"text": "Read the "},
						{"text": "documentation", "style": {"link": "https://gpdf.dev/docs", "color": "#0645AD", "underline": true}},
						{"text": " for every option."}
					]}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Plain-looking text can link too.", "style": {"link": "https://github.com/gpdf-dev/gpdf"}},
					{"type": "spacer", "height": "6mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "image": {"src": "%s", "width": "40mm", "link": "https://gpdf.dev"}},
				{"span": 6, "box": {
					"background": "#E8EAF6", "padding": "4mm", "link": "https://gpdf.dev/blog",
					"elements": [{"type": "text", "content": "The whole card is clickable"}]
				}}
			]}}
		]
	}`, base64.StdEncoding.EncodeToString(logo)))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "44_links.pdf", doc)
}
