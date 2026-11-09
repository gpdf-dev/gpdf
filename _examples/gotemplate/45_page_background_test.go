package gotemplate_test

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_45_PageBackground(t *testing.T) {
	letterhead := testutil.TestImagePNG(t, 300, 100, color.RGBA{R: 255, G: 213, B: 79, A: 255})
	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm", "background": {
			"color": "#FFFDF5", "image": "%s", "fit": "contain", "opacity": 0.25
		}},
		"pages": [
			{"body": [{"row": {"cols": [{"span": 12, "elements": [
				{"type": "text", "content": "{{.Title}} (page 1)", "style": {"size": 20, "bold": true}},
				{"type": "spacer", "height": "4mm"},
				{"type": "text", "content": "The background is drawn beneath the content of every page, ignoring margins."}
			]}]}}]},
			{"body": [{"row": {"cols": [{"span": 12, "elements": [
				{"type": "text", "content": "{{.Title}} (page 2)", "style": {"size": 20, "bold": true}},
				{"type": "spacer", "height": "4mm"},
				{"type": "text", "content": "The background is drawn beneath the content of every page, ignoring margins."}
			]}]}}]}
		]
	}`, base64.StdEncoding.EncodeToString(letterhead)))

	data := map[string]any{"Title": "Page Background"}
	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "45_page_background.pdf", doc)
}
