package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_03_GridLayout(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.FullWidth}}", "style": {"background": "#E3F2FD"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "text": "{{.LeftCol}}", "style": {"background": "#E8F5E9"}},
				{"span": 6, "text": "{{.RightCol}}", "style": {"background": "#FFF3E0"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "{{.Sidebar}}", "style": {"background": "#D7CCC8"}},
				{"span": 9, "text": "{{.MainContent}}", "style": {"background": "#F5F5F5"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "text": "{{.Col1}}", "style": {"background": "#FCE4EC"}},
				{"span": 4, "text": "{{.Col2}}", "style": {"background": "#F3E5F5"}},
				{"span": 4, "text": "{{.Col3}}", "style": {"background": "#E8EAF6"}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":       "12-Column Grid Layout",
		"FullWidth":   "Full width (span 12)",
		"LeftCol":     "Half (span 6)",
		"RightCol":    "Half (span 6)",
		"Sidebar":     "Narrow (span 3)",
		"MainContent": "Wide (span 9)",
		"Col1":        "Third (span 4)",
		"Col2":        "Third (span 4)",
		"Col3":        "Third (span 4)",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "03_grid_layout.pdf", doc)
}
