package gotemplate_test

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

// TestTmpl_34_ImageMinSize is the Go-template counterpart of
// TestExample_34_ImageMinSize. It shares the same golden file.
func TestTmpl_34_ImageMinSize(t *testing.T) {
	imgData := testutil.TestImagePNG(t, 300, 500, color.RGBA{R: 100, G: 149, B: 237, A: 255})

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{{range .Fillers}}
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.}}"}
				]}
			]}},
			{{end}}
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "image", "image": {"src": "{{.ImgB64}}", "width": "80mm", "minHeight": "100mm"}}
				]}
			]}}
		]
	}`)

	fillers := make([]string, 23)
	for i := range fillers {
		fillers[i] = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
			"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	}

	data := map[string]any{
		"Title":   "Image Minimum Display Size",
		"Fillers": fillers,
		"ImgB64":  base64.StdEncoding.EncodeToString(imgData),
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "34_image_min_size.pdf", doc)
}
