package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_14_Metadata(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "Annual Report 2026",
			"author": "gpdf Library",
			"subject": "Example of document metadata",
			"creator": "gpdf example_test.go"
		},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Document with Metadata", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "This PDF has the following metadata set:"},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Title: Annual Report 2026"},
					{"type": "text", "content": "Author: gpdf Library"},
					{"type": "text", "content": "Subject: Example of document metadata"},
					{"type": "text", "content": "Creator: gpdf example_test.go"},
					{"type": "text", "content": "Producer: gpdf (set automatically)"},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "Open the PDF properties in your viewer to verify.", "style": {"italic": true}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "14_metadata.pdf", doc)
}
