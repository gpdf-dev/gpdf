package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_16_Margins(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Custom Margins", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "This page has asymmetric margins: 10mm top/bottom, 40mm left/right. The wide side margins create a narrower text area, similar to a book layout."},
					{"type": "spacer", "height": "5mm"},
					{"type": "line"},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris."}
				]}
			]}}
		]
	}`)

	// The builder uses asymmetric margins: 10mm top/bottom, 40mm left/right.
	// JSON schema only supports uniform margins, so we override with options.
	doc, err := template.FromJSON(schema, nil, template.WithMargins(document.Edges{
		Top:    document.Mm(10),
		Right:  document.Mm(40),
		Bottom: document.Mm(10),
		Left:   document.Mm(40),
	}))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "16_margins.pdf", doc)
}
