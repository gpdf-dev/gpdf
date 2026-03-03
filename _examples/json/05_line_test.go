package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_05_Line(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Line / Horizontal Rule Examples", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Default line (gray, 1pt):"},
					{"type": "line"},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Red line:"},
					{"type": "line", "line": {"color": "red"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Blue line:"},
					{"type": "line", "line": {"color": "blue"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Green line:"},
					{"type": "line", "line": {"color": "green"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Thin line (0.5pt):"},
					{"type": "line", "line": {"thickness": "0.5pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Medium line (2pt):"},
					{"type": "line", "line": {"thickness": "2pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Thick line (5pt):"},
					{"type": "line", "line": {"thickness": "5pt"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Thick red line (3pt):"},
					{"type": "line", "line": {"color": "red", "thickness": "3pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Thick blue line (4pt):"},
					{"type": "line", "line": {"color": "blue", "thickness": "4pt"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "05_line.pdf", doc)
}
