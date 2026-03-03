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
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.FullWidth}}", "style": {"background": "#E3F2FD"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "text": "{{.Left6}}", "style": {"background": "#E8F5E9"}},
				{"span": 6, "text": "{{.Right6}}", "style": {"background": "#FFF3E0"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "text": "{{.Col4a}}", "style": {"background": "#FCE4EC"}},
				{"span": 4, "text": "{{.Col4b}}", "style": {"background": "#F3E5F5"}},
				{"span": 4, "text": "{{.Col4c}}", "style": {"background": "#E8EAF6"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "{{.Col3a}}", "style": {"background": "#E0F7FA"}},
				{"span": 3, "text": "{{.Col3b}}", "style": {"background": "#E0F2F1"}},
				{"span": 3, "text": "{{.Col3c}}", "style": {"background": "#FFF9C4"}},
				{"span": 3, "text": "{{.Col3d}}", "style": {"background": "#FFECB3"}}
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
				{"span": 8, "text": "{{.Article}}", "style": {"background": "#E1F5FE"}},
				{"span": 4, "text": "{{.SidePanel}}", "style": {"background": "#FBE9E7"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.LeftLine1}}"},
					{"type": "text", "content": "{{.LeftLine2}}"},
					{"type": "text", "content": "{{.LeftLine3}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.RightLine1}}"},
					{"type": "text", "content": "{{.RightLine2}}"},
					{"type": "text", "content": "{{.RightLine3}}"}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":       "12-Column Grid Layout",
		"FullWidth":   "Col 12 (full width)",
		"Left6":       "Col 6 (left)",
		"Right6":      "Col 6 (right)",
		"Col4a":       "Col 4",
		"Col4b":       "Col 4",
		"Col4c":       "Col 4",
		"Col3a":       "Col 3",
		"Col3b":       "Col 3",
		"Col3c":       "Col 3",
		"Col3d":       "Col 3",
		"Sidebar":     "Sidebar (3)",
		"MainContent": "Main content (9)",
		"Article":     "Article area (8)",
		"SidePanel":   "Side panel (4)",
		"LeftLine1":   "Left column - line 1",
		"LeftLine2":   "Left column - line 2",
		"LeftLine3":   "Left column - line 3",
		"RightLine1":  "Right column - line 1",
		"RightLine2":  "Right column - line 2",
		"RightLine3":  "Right column - line 3",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "03_grid_layout.pdf", doc)
}
