package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_05_Line(t *testing.T) {
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
				{"span": 12, "text": "{{.DefaultLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.RedLabel}}", "style": {"color": "#FF0000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#FF0000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.BlueLabel}}", "style": {"color": "#0000FF"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#0000FF"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.GreenLabel}}", "style": {"color": "#008000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#008000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.ThinLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"thickness": "0.5pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.MediumLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"thickness": "2pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.ThickLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"thickness": "5pt"}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":       "Line / Horizontal Rule Examples",
		"DefaultLabel": "Default line:",
		"RedLabel":    "Red line:",
		"BlueLabel":   "Blue line:",
		"GreenLabel":  "Green line:",
		"ThinLabel":   "Thickness 0.5pt:",
		"MediumLabel": "Thickness 2pt:",
		"ThickLabel":  "Thickness 5pt:",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "05_line.pdf", doc)
}
