package json_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

// TestJSON_34_ImageMinSize is the JSON-schema counterpart of
// TestExample_34_ImageMinSize. It shares the same golden file.
func TestJSON_34_ImageMinSize(t *testing.T) {
	imgData := testutil.TestImagePNG(t, 300, 500, color.RGBA{R: 100, G: 149, B: 237, A: 255})
	imgB64 := base64.StdEncoding.EncodeToString(imgData)

	// Build 23 filler rows identical to the Builder/GoTemplate cases.
	var filler strings.Builder
	for i := 0; i < 23; i++ {
		filler.WriteString(`{"row": {"cols": [
			{"span": 12, "elements": [
				{"type": "text", "content": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."}
			]}
		]}},`)
	}

	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Image Minimum Display Size", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			%s
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "image", "image": {"src": "%s", "width": "80mm", "minHeight": "100mm"}}
				]}
			]}}
		]
	}`, filler.String(), imgB64))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "34_image_min_size.pdf", doc)
}
