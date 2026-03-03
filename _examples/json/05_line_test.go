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
				{"span": 12, "text": "Line / Horizontal Rule Examples", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Default line:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Red line:", "style": {"color": "#FF0000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#FF0000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Blue line:", "style": {"color": "#0000FF"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#0000FF"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Green line:", "style": {"color": "#008000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#008000"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Thickness 0.5pt:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"thickness": "0.5pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Thickness 2pt:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"thickness": "2pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Thickness 5pt:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"thickness": "5pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Red 3pt:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#FF0000", "thickness": "3pt"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Blue 4pt:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#0000FF", "thickness": "4pt"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "05_line.pdf", doc)
}
