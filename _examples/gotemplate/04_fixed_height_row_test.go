package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_04_FixedHeightRow(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"height": "30mm", "cols": [
				{"span": 12, "text": "{{.Row30mm}}", "style": {"background": "#E3F2FD"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"height": "50mm", "cols": [
				{"span": 6, "text": "{{.LeftRow50mm}}", "style": {"background": "#E8F5E9"}},
				{"span": 6, "text": "{{.RightRow50mm}}", "style": {"background": "#FFF3E0"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.AutoRow}}", "style": {"background": "#FCE4EC"}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":       "Fixed-Height Row Examples",
		"Row30mm":     "This row is 30mm tall",
		"LeftRow50mm": "Left column in 50mm row",
		"RightRow50mm": "Right column in 50mm row",
		"AutoRow":     "This row has auto height",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "04_fixed_height_row.pdf", doc)
}
