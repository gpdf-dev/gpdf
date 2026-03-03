package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_02_TextStyling(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Size8}}", "style": {"size": 8}},
					{"type": "text", "content": "{{.Size12}}", "style": {"size": 12}},
					{"type": "text", "content": "{{.Size18}}", "style": {"size": 18}},
					{"type": "text", "content": "{{.Size24}}", "style": {"size": 24}},
					{"type": "text", "content": "{{.Size36}}", "style": {"size": 36}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Normal}}"},
					{"type": "text", "content": "{{.BoldText}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.ItalicText}}", "style": {"italic": true}},
					{"type": "text", "content": "{{.BoldItalicText}}", "style": {"bold": true, "italic": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.RedText}}", "style": {"color": "red"}},
					{"type": "text", "content": "{{.GreenText}}", "style": {"color": "green"}},
					{"type": "text", "content": "{{.BlueText}}", "style": {"color": "blue"}},
					{"type": "text", "content": "{{.OrangeText}}", "style": {"color": "rgb(1.0, 0.5, 0.0)"}},
					{"type": "text", "content": "{{.HexText}}", "style": {"color": "#336699"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.YellowBg}}", "style": {"background": "yellow"}},
					{"type": "text", "content": "{{.CyanBg}}", "style": {"background": "cyan"}},
					{"type": "text", "content": "{{.WhiteOnDark}}", "style": {"color": "white", "background": "#333333"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.LeftAlign}}", "style": {"align": "left"}},
					{"type": "text", "content": "{{.CenterAlign}}", "style": {"align": "center"}},
					{"type": "text", "content": "{{.RightAlign}}", "style": {"align": "right"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":          "Text Styling Examples",
		"Size8":          "Font Size 8pt",
		"Size12":         "Font Size 12pt (default)",
		"Size18":         "Font Size 18pt",
		"Size24":         "Font Size 24pt",
		"Size36":         "Font Size 36pt",
		"Normal":         "Normal text",
		"BoldText":       "Bold text",
		"ItalicText":     "Italic text",
		"BoldItalicText": "Bold + Italic text",
		"RedText":        "Red text",
		"GreenText":      "Green text",
		"BlueText":       "Blue text",
		"OrangeText":     "Custom color (orange)",
		"HexText":        "Hex color (#336699)",
		"YellowBg":       "Yellow background",
		"CyanBg":         "Cyan background",
		"WhiteOnDark":    "White text on dark background",
		"LeftAlign":      "Left aligned (default)",
		"CenterAlign":    "Center aligned",
		"RightAlign":     "Right aligned",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "02_text_styling.pdf", doc)
}
