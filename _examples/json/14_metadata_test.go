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
				{"span": 12, "text": "Document with Metadata", "style": {"size": 20, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "This PDF document has the following metadata properties set:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Title: Annual Report 2026", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Author: gpdf Library", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Subject: Example of document metadata", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Creator: gpdf example_test.go", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "You can verify these metadata fields by opening the PDF in a viewer and checking the document properties (File > Properties or Ctrl+D in most PDF viewers).", "style": {"italic": true}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "14_metadata.pdf", doc)
}
