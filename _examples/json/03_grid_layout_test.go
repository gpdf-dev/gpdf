package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_03_GridLayout(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "12-Column Grid Layout", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Col 12 (full width)", "style": {"background": "#E3F2FD"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "text": "Col 6 (left)", "style": {"background": "#E8F5E9"}},
				{"span": 6, "text": "Col 6 (right)", "style": {"background": "#FFF3E0"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "text": "Col 4", "style": {"background": "#FCE4EC"}},
				{"span": 4, "text": "Col 4", "style": {"background": "#F3E5F5"}},
				{"span": 4, "text": "Col 4", "style": {"background": "#E8EAF6"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "Col 3", "style": {"background": "#E0F7FA"}},
				{"span": 3, "text": "Col 3", "style": {"background": "#E0F2F1"}},
				{"span": 3, "text": "Col 3", "style": {"background": "#FFF9C4"}},
				{"span": 3, "text": "Col 3", "style": {"background": "#FFECB3"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "Sidebar (3)", "style": {"background": "#D7CCC8"}},
				{"span": 9, "text": "Main content (9)", "style": {"background": "#F5F5F5"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 8, "text": "Article area (8)", "style": {"background": "#E1F5FE"}},
				{"span": 4, "text": "Side panel (4)", "style": {"background": "#FBE9E7"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Left column - line 1"},
					{"type": "text", "content": "Left column - line 2"},
					{"type": "text", "content": "Left column - line 3"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Right column - line 1"},
					{"type": "text", "content": "Right column - line 2"},
					{"type": "text", "content": "Right column - line 3"}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "03_grid_layout.pdf", doc)
}
