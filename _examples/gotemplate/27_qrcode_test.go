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
				{"span": 12, "text": "{{.Title}}", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.BasicLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "qrcode": {"data": "{{.URL}}", "size": "25mm"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.SizesLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}", "size": "20mm"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.Size20Label}}", "style": {"size": 9, "color": "#666666"}}
				]},
				{"span": 4, "elements": [
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}", "size": "30mm"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.Size30Label}}", "style": {"size": 9, "color": "#666666"}}
				]},
				{"span": 4, "elements": [
					{"type": "qrcode", "qrcode": {"data": "{{.URL}}", "size": "40mm"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "{{.Size40Label}}", "style": {"size": 9, "color": "#666666"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "{{.JapaneseLabel}}"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "qrcode": {"data": "{{.JapaneseText}}", "size": "30mm"}}
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
		"JapaneseLabel": "Japanese content:",
		"JapaneseText":  "こんにちは世界",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "27_qrcode.pdf", doc)
}
