package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_27_QRCode(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "{{.Title}}", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.BasicLabel}}"},
					{"type": "spacer", "height": "2mm"},
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.SizesLabel}}"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.Size20Label}}", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}", "size": "20mm"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.Size30Label}}", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}", "size": "30mm"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "{{.Size40Label}}", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}", "size": "40mm"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.ECLabel}}"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [
					{"type": "text", "content": "{{.LevelL}}", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "{{.ECData}}", "size": "25mm", "errorCorrection": "L"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "{{.LevelM}}", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "{{.ECData}}", "size": "25mm", "errorCorrection": "M"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "{{.LevelQ}}", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "{{.ECData}}", "size": "25mm", "errorCorrection": "Q"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "{{.LevelH}}", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "{{.ECData}}", "size": "25mm", "errorCorrection": "H"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.JapaneseLabel}}"},
					{"type": "spacer", "height": "2mm"},
					{"type": "qrcode", "qrcode": {"data": "{{.JapaneseText}}", "size": "30mm"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":         "QR Code Examples",
		"BasicLabel":    "Basic QR code (URL):",
		"URL":           "https://gpdf.dev",
		"SizesLabel":    "QR codes with different sizes:",
		"Size20Label":   "20mm",
		"Size30Label":   "30mm",
		"Size40Label":   "40mm",
		"ECLabel":       "Error correction levels (L / M / Q / H):",
		"ECData":        "HELLO",
		"LevelL":        "Level L",
		"LevelM":        "Level M",
		"LevelQ":        "Level Q",
		"LevelH":        "Level H",
		"JapaneseLabel": "Japanese content:",
		"JapaneseText":  "\u3053\u3093\u306b\u3061\u306f\u4e16\u754c",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "27_qrcode.pdf", doc)
}
