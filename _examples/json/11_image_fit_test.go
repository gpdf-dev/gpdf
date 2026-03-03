package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_11_ImageFit(t *testing.T) {
	imgData := testutil.TestImagePNG(t, 300, 200, color.RGBA{R: 100, G: 149, B: 237, A: 255})
	imgB64 := base64.StdEncoding.EncodeToString(imgData)

	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Image Fit Options", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "FitWidth(80mm):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "%s", "width": "80mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "FitHeight(30mm):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "%s", "height": "30mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Default (no fit options):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "%s"}}
				]}
			]}}
		]
	}`, imgB64, imgB64, imgB64))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "11_image_fit.pdf", doc)
}
