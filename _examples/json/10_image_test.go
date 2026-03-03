package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_10_Image(t *testing.T) {
	pngData := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	pngB64 := base64.StdEncoding.EncodeToString(pngData)

	jpegData := testutil.TestImageJPEG(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})
	jpegB64 := base64.StdEncoding.EncodeToString(jpegData)

	greenImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	greenB64 := base64.StdEncoding.EncodeToString(greenImg)

	yellowImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 251, G: 188, B: 4, A: 255})
	yellowB64 := base64.StdEncoding.EncodeToString(yellowImg)

	schema := []byte(fmt.Sprintf(`{
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
					{"type": "image", "image": {"src": "%s"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "JPEG image (red):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {"src": "%s"}},
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
					{"type": "image", "image": {"src": "%s"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Yellow PNG"},
					{"type": "image", "image": {"src": "%s"}}
				]}
			]}}
		]
	}`, pngB64, jpegB64, greenB64, yellowB64))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "10_image.pdf", doc)
}
