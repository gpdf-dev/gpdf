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
				{"span": 12, "text": "{{.Title}}", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.PredefinedLabel}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 2, "text": "{{.RedText}}", "style": {"color": "#FF0000"}},
				{"span": 2, "text": "{{.GreenText}}", "style": {"color": "#008000"}},
				{"span": 2, "text": "{{.BlueText}}", "style": {"color": "#0000FF"}},
				{"span": 2, "text": "{{.YellowText}}", "style": {"color": "#FFD700"}},
				{"span": 2, "text": "{{.CyanText}}", "style": {"color": "#00FFFF"}},
				{"span": 2, "text": "{{.MagentaText}}", "style": {"color": "#FF00FF"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.HexLabel}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "{{.CoralText}}", "style": {"color": "#FF6B6B"}},
				{"span": 3, "text": "{{.TurquoiseText}}", "style": {"color": "#4ECDC4"}},
				{"span": 3, "text": "{{.SkyBlueText}}", "style": {"color": "#45B7D1"}},
				{"span": 3, "text": "{{.SageText}}", "style": {"color": "#96CEB4"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.BackgroundLabel}}", "style": {"bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "3mm"}
			]}},
			{"row": {"cols": [
				{"span": 3, "text": "{{.RedBgText}}", "style": {"color": "#FFFFFF", "background": "#FF0000"}},
				{"span": 3, "text": "{{.GreenBgText}}", "style": {"color": "#FFFFFF", "background": "#008000"}},
				{"span": 3, "text": "{{.BlueBgText}}", "style": {"color": "#FFFFFF", "background": "#0000FF"}},
				{"span": 3, "text": "{{.YellowBgText}}", "style": {"color": "#000000", "background": "#FFD700"}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":           "Color System Examples",
		"PredefinedLabel": "Predefined Colors:",
		"RedText":         "Red",
		"GreenText":       "Green",
		"BlueText":        "Blue",
		"YellowText":      "Yellow",
		"CyanText":        "Cyan",
		"MagentaText":     "Magenta",
		"HexLabel":        "Hex Colors:",
		"CoralText":       "#FF6B6B Coral",
		"TurquoiseText":   "#4ECDC4 Turquoise",
		"SkyBlueText":     "#45B7D1 Sky Blue",
		"SageText":        "#96CEB4 Sage",
		"BackgroundLabel": "Background Color Swatches:",
		"RedBgText":       "Red Background",
		"GreenBgText":     "Green Background",
		"BlueBgText":      "Blue Background",
		"YellowBgText":    "Yellow Background",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "17_colors.pdf", doc)
}
