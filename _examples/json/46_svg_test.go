package json_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func svgBase64(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("../testdata/svg/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func TestJSON_46_SVG(t *testing.T) {
	logo := svgBase64(t, "logo.svg")
	icons := svgBase64(t, "icons.svg")
	chart := svgBase64(t, "chart.svg")

	schema := []byte(fmt.Sprintf(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "SVG Images", "style": {"size": 20, "bold": true}},
					{"type": "text", "content": "Embedded as vector graphics: sharp at any zoom level."},
					{"type": "spacer", "height": "6mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 4, "image": {"src": "%[1]s", "width": "40mm"}},
				{"span": 8, "image": {"src": "%[2]s", "width": "100mm"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "8mm"},
					{"type": "image", "image": {"src": "%[3]s", "width": "60mm"}},
					{"type": "spacer", "height": "4mm"},
					{"type": "image", "image": {"src": "%[3]s", "width": "150mm"}}
				]}
			]}}
		]
	}`, logo, chart, icons))

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "46_svg.pdf", doc)
}
