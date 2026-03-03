package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_22_LetterSpacing(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 20, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "8mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Normal}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Spacing1}}", "style": {"letterSpacing": 1}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Spacing3}}", "style": {"letterSpacing": 3}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.WideHeader}}", "style": {"size": 16, "bold": true, "letterSpacing": 5}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Tight}}", "style": {"letterSpacing": -0.5}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":      "Letter Spacing Demo",
		"Normal":     "Normal spacing (0pt)",
		"Spacing1":   "Letter spacing 1pt",
		"Spacing3":   "Letter spacing 3pt",
		"WideHeader": "WIDE HEADER",
		"Tight":      "Tight spacing -0.5pt",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "22_letter_spacing.pdf", doc)
}
