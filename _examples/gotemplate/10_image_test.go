package gotemplate_test

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_10_Image(t *testing.T) {
	pngData := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	jpegData := testutil.TestImageJPEG(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	greenImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	yellowImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 251, G: 188, B: 4, A: 255})

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Image Examples", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "PNG image (blue):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "{{.PngB64}}"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "JPEG image (red):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "{{.JpegB64}}"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Images side by side in grid columns:"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Green PNG"},
					{"type": "image", "image": {"src": "{{.GreenB64}}"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Yellow PNG"},
					{"type": "image", "image": {"src": "{{.YellowB64}}"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"PngB64":    base64.StdEncoding.EncodeToString(pngData),
		"JpegB64":   base64.StdEncoding.EncodeToString(jpegData),
		"GreenB64":  base64.StdEncoding.EncodeToString(greenImg),
		"YellowB64": base64.StdEncoding.EncodeToString(yellowImg),
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "10_image.pdf", doc)
}
