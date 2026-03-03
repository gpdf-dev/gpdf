package gotemplate_test

import (
	"testing"
	gotemplate "text/template"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_12_MultiPage(t *testing.T) {
	tmplStr := `{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "5mm"},
					{"type": "line"},
					{"type": "spacer", "height": "10mm"}
				]}
			]}}
			{{- range .Paragraphs}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.}}"}
			]}}
			{{- end}}
		]
	}`

	tmpl, err := gotemplate.New("test").Funcs(template.TemplateFuncMap()).Parse(tmplStr)
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	paragraphs := make([]string, 20)
	for i := range paragraphs {
		paragraphs[i] = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	}

	data := map[string]any{
		"Title":      "Multi-Page Document with Template",
		"Paragraphs": paragraphs,
	}

	doc, err := template.FromTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("FromTemplate error: %v", err)
	}
	testutil.GeneratePDF(t, "12_multi_page.pdf", doc)
}
