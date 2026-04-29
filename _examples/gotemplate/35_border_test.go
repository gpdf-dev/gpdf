package gotemplate_test

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

// TestTmpl_35_Border mirrors TestExample_35_Border / TestJSON_35_Border so the
// shared golden 35_border.pdf compares byte-identically across all three
// entry points.
func TestTmpl_35_Border(t *testing.T) {
	pngData := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},

			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "A. Outer border only", "style": {"size": 11, "bold": true}},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "table", "table": {
						"header": ["Name", "Age", "City"],
						"rows": {{toJSON .Rows}},
						"columnWidths": [40, 20, 40],
						"border": {"width": "1pt", "color": "{{.OuterColor}}"},
						"background": "#F5F5F5"
					}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},

			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "B. Cell grid only", "style": {"size": 11, "bold": true}},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "table", "table": {
						"header": ["Name", "Age", "City"],
						"rows": {{toJSON .Rows}},
						"columnWidths": [40, 20, 40],
						"cellBorder": {"width": "0.5pt", "color": "gray(0.5)"}
					}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},

			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "C. Outer frame + cell grid", "style": {"size": 11, "bold": true}},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "table", "table": {
						"header": ["Name", "Age", "City"],
						"rows": {{toJSON .Rows}},
						"columnWidths": [40, 20, 40],
						"border": {"width": "1pt", "color": "{{.OuterColor}}"},
						"cellBorder": {"width": "0.5pt", "color": "gray(0.5)"},
						"background": "#FAFAFA"
					}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},

			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "D. Dashed cell grid with stripes", "style": {"size": 11, "bold": true}},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "table", "table": {
						"header": ["Name", "Age", "City"],
						"rows": {{toJSON .Rows}},
						"columnWidths": [40, 20, 40],
						"cellBorder": {"width": "0.75pt", "color": "#0D47A1", "style": "dashed"},
						"stripeColor": "#E3F2FD"
					}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},

			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "E. Image with border + background", "style": {"size": 11, "bold": true}},
					{"type": "spacer", "height": "2mm"},
					{"type": "image", "image": {
						"src": "{{.ImageB64}}",
						"width": "60mm",
						"border": {"widths": ["2pt", "2pt", "2pt", "2pt"], "color": "#E53935", "style": "solid"},
						"background": "#FFF8E1"
					}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":      "Borders & backgrounds",
		"OuterColor": "#1A237E",
		"ImageB64":   base64.StdEncoding.EncodeToString(pngData),
		"Rows": [][]string{
			{"Alice", "30", "Tokyo"},
			{"Bob", "25", "New York"},
			{"Charlie", "35", "London"},
		},
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "35_border.pdf", doc)
}
