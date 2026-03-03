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
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Heading}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "{{.Description}}"},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.TitleLine}}"},
					{"type": "text", "content": "{{.AuthorLine}}"},
					{"type": "text", "content": "{{.SubjectLine}}"},
					{"type": "text", "content": "{{.CreatorLine}}"},
					{"type": "text", "content": "{{.ProducerLine}}"},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "{{.Note}}", "style": {"italic": true}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"MetaTitle":    "Annual Report 2026",
		"MetaAuthor":   "gpdf Library",
		"MetaSubject":  "Example of document metadata",
		"MetaCreator":  "gpdf example_test.go",
		"Heading":      "Document with Metadata",
		"Description":  "This PDF has the following metadata set:",
		"TitleLine":    "Title: Annual Report 2026",
		"AuthorLine":   "Author: gpdf Library",
		"SubjectLine":  "Subject: Example of document metadata",
		"CreatorLine":  "Creator: gpdf example_test.go",
		"ProducerLine": "Producer: gpdf (set automatically)",
		"Note":         "Open the PDF properties in your viewer to verify.",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "14_metadata.pdf", doc)
}
