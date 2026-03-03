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
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.DefaultLabel}}"},
					{"type": "line"},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.RedLabel}}"},
					{"type": "line", "line": {"color": "red"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.BlueLabel}}"},
					{"type": "line", "line": {"color": "blue"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.GreenLabel}}"},
					{"type": "line", "line": {"color": "green"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.ThinLabel}}"},
					{"type": "line", "line": {"thickness": "0.5pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.MediumLabel}}"},
					{"type": "line", "line": {"thickness": "2pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.ThickLabel}}"},
					{"type": "line", "line": {"thickness": "5pt"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.ThickRedLabel}}"},
					{"type": "line", "line": {"color": "red", "thickness": "3pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.ThickBlueLabel}}"},
					{"type": "line", "line": {"color": "blue", "thickness": "4pt"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":          "Line / Horizontal Rule Examples",
		"DefaultLabel":   "Default line (gray, 1pt):",
		"RedLabel":       "Red line:",
		"BlueLabel":      "Blue line:",
		"GreenLabel":     "Green line:",
		"ThinLabel":      "Thin line (0.5pt):",
		"MediumLabel":    "Medium line (2pt):",
		"ThickLabel":     "Thick line (5pt):",
		"ThickRedLabel":  "Thick red line (3pt):",
		"ThickBlueLabel": "Thick blue line (4pt):",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "05_line.pdf", doc)
}
