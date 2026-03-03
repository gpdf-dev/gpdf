package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_06_Spacer(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "Spacer Examples", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Text before 5mm spacer"},
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "Text after 5mm spacer"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Text before 15mm spacer"},
					{"type": "spacer", "height": "15mm"},
					{"type": "text", "content": "Text after 15mm spacer"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Text before 30mm spacer"},
					{"type": "spacer", "height": "30mm"},
					{"type": "text", "content": "Text after 30mm spacer"}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "06_spacer.pdf", doc)
}
