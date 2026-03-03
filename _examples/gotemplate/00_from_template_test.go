package gotemplate_test

import (
	"testing"
	gotemplate "text/template"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_00_FromTemplate(t *testing.T) {
	// Using FromTemplate with a pre-parsed Go template for more control.
	tmplStr := `{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "{{.Title}}"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{{- range $i, $section := .Sections}}
			{{- if $i}},{{end}}
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{$section.Heading}}", "style": {"size": 16, "bold": true, "color": "#1A237E"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{$section.Body}}"},
					{"type": "spacer", "height": "8mm"}
				]}
			]}}
			{{- end}}
		]
	}`

	tmpl, err := gotemplate.New("report").Funcs(template.TemplateFuncMap()).Parse(tmplStr)
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	type section struct {
		Heading string
		Body    string
	}

	data := map[string]any{
		"Title": "Quarterly Report - Q1 2026",
		"Sections": []section{
			{
				Heading: "Executive Summary",
				Body:    "This quarter showed strong growth across all product lines. Revenue increased 25% year-over-year, driven primarily by our new enterprise offerings.",
			},
			{
				Heading: "Product Development",
				Body:    "The gpdf library reached v0.8 with Go template integration and JSON schema support. Community adoption continues to accelerate with over 500 GitHub stars.",
			},
			{
				Heading: "Market Analysis",
				Body:    "The PDF generation market continues to grow, with increasing demand for programmatic document creation. Our zero-dependency approach resonates well with Go developers.",
			},
			{
				Heading: "Next Steps",
				Body:    "Focus areas for Q2 include reusable components (invoice, report, letter templates), fuzz testing, and preparation for the v1.0 release.",
			},
		},
	}

	doc, err := template.FromTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("FromTemplate error: %v", err)
	}
	testutil.GeneratePDF(t, "00_from_template.pdf", doc)
}
