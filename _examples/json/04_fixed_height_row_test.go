package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_04_FixedHeightRow(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "Fixed-Height Row Examples", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"height": "30mm", "cols": [
				{"span": 12, "text": "This row is 30mm tall", "style": {"background": "#E3F2FD"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"height": "50mm", "cols": [
				{"span": 6, "text": "Left column in 50mm row", "style": {"background": "#E8F5E9"}},
				{"span": 6, "text": "Right column in 50mm row", "style": {"background": "#FFF3E0"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "This row has auto height", "style": {"background": "#FCE4EC"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "04_fixed_height_row.pdf", doc)
}
