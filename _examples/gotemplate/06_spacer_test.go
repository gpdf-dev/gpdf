package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_06_Spacer(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Before5mm}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.After5mm}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Before15mm}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "15mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.After15mm}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.Before30mm}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "30mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.After30mm}}"}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":      "Spacer Examples",
		"Before5mm":  "Text before 5mm spacer",
		"After5mm":   "Text after 5mm spacer",
		"Before15mm": "Text before 15mm spacer",
		"After15mm":  "Text after 15mm spacer",
		"Before30mm": "Text before 30mm spacer",
		"After30mm":  "Text after 30mm spacer",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "06_spacer.pdf", doc)
}
