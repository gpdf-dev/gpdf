package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_14_Metadata(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {
			"title": "{{.MetaTitle}}",
			"author": "{{.MetaAuthor}}",
			"subject": "{{.MetaSubject}}",
			"creator": "{{.MetaCreator}}"
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
				{"span": 12, "text": "Title: {{.MetaTitle}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Author: {{.MetaAuthor}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Subject: {{.MetaSubject}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Creator: {{.MetaCreator}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Open the PDF properties in your viewer to verify these metadata fields.", "style": {"italic": true}}
			]}}
		]
	}`)

	data := map[string]any{
		"MetaTitle":   "Annual Report 2026",
		"MetaAuthor":  "gpdf Library",
		"MetaSubject": "Example of document metadata via template",
		"MetaCreator": "gpdf template example",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "14_metadata.pdf", doc)
}
