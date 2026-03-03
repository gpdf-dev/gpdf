package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_17_Colors(t *testing.T) {
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
					{"type": "text", "content": "{{.PredefinedLabel}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.Red}}", "style": {"color": "red"}},
					{"type": "text", "content": "{{.Green}}", "style": {"color": "green"}},
					{"type": "text", "content": "{{.Blue}}", "style": {"color": "blue"}},
					{"type": "text", "content": "{{.Yellow}}", "style": {"color": "yellow"}},
					{"type": "text", "content": "{{.Cyan}}", "style": {"color": "cyan"}},
					{"type": "text", "content": "{{.Magenta}}", "style": {"color": "magenta"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.RGBLabel}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.Orange}}", "style": {"color": "rgb(1.0, 0.5, 0.0)"}},
					{"type": "text", "content": "{{.Purple}}", "style": {"color": "rgb(0.5, 0.0, 0.5)"}},
					{"type": "text", "content": "{{.Teal}}", "style": {"color": "rgb(0.0, 0.5, 0.5)"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.HexLabel}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.Coral}}", "style": {"color": "#FF6B6B"}},
					{"type": "text", "content": "{{.Turquoise}}", "style": {"color": "#4ECDC4"}},
					{"type": "text", "content": "{{.SkyBlue}}", "style": {"color": "#45B7D1"}},
					{"type": "text", "content": "{{.Sage}}", "style": {"color": "#96CEB4"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.GrayLabel}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.Gray0}}", "style": {"color": "gray(0.0)"}},
					{"type": "text", "content": "{{.Gray3}}", "style": {"color": "gray(0.3)"}},
					{"type": "text", "content": "{{.Gray5}}", "style": {"color": "gray(0.5)"}},
					{"type": "text", "content": "{{.Gray7}}", "style": {"color": "gray(0.7)"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.BgLabel}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "{{.RedBg}}", "style": {"color": "white", "background": "red"}},
				{"span": 3, "text": "{{.GreenBg}}", "style": {"color": "white", "background": "green"}},
				{"span": 3, "text": "{{.BlueBg}}", "style": {"color": "white", "background": "blue"}},
				{"span": 3, "text": "{{.YellowBg}}", "style": {"background": "yellow"}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":           "Color System Examples",
		"PredefinedLabel": "Predefined Colors:",
		"Red":             "Red",
		"Green":           "Green",
		"Blue":            "Blue",
		"Yellow":          "Yellow",
		"Cyan":            "Cyan",
		"Magenta":         "Magenta",
		"RGBLabel":        "RGB Colors (float):",
		"Orange":          "RGB(1.0, 0.5, 0.0) - Orange",
		"Purple":          "RGB(0.5, 0.0, 0.5) - Purple",
		"Teal":            "RGB(0.0, 0.5, 0.5) - Teal",
		"HexLabel":        "Hex Colors:",
		"Coral":           "#FF6B6B - Coral",
		"Turquoise":       "#4ECDC4 - Turquoise",
		"SkyBlue":         "#45B7D1 - Sky Blue",
		"Sage":            "#96CEB4 - Sage",
		"GrayLabel":       "Grayscale:",
		"Gray0":           "Gray(0.0) - Black",
		"Gray3":           "Gray(0.3) - Dark gray",
		"Gray5":           "Gray(0.5) - Medium gray",
		"Gray7":           "Gray(0.7) - Light gray",
		"BgLabel":         "Background Color Swatches:",
		"RedBg":           " Red ",
		"GreenBg":         " Green ",
		"BlueBg":          " Blue ",
		"YellowBg":        " Yellow ",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "17_colors.pdf", doc)
}
