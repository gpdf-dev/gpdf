package gotemplate_test

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_11_ImageFit(t *testing.T) {
	imgData := testutil.TestImagePNG(t, 300, 200, color.RGBA{R: 100, G: 149, B: 237, A: 255})

	schema := []byte(`{
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
					{"type": "image", "image": {"src": "{{.ImgB64}}", "width": "80mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "FitHeight(30mm):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "{{.ImgB64}}", "height": "30mm"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Default (no fit options):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "{{.ImgB64}}"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"ImgB64": base64.StdEncoding.EncodeToString(imgData),
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "11_image_fit.pdf", doc)
}
