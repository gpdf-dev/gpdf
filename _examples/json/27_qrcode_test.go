package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_27_QRCode(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "QR Code Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "QR Code Examples", "style": {"size": 18, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Basic QR code (URL):"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "qrcode": {"data": "https://gpdf.dev", "size": "25mm"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "QR codes with different sizes:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "20mm"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "20mm", "style": {"size": 9, "color": "#666666"}}
				]},
				{"span": 4, "elements": [
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "30mm"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "30mm", "style": {"size": 9, "color": "#666666"}}
				]},
				{"span": 4, "elements": [
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "40mm"}},
					{"type": "spacer", "height": "2mm"},
					{"type": "text", "content": "40mm", "style": {"size": 9, "color": "#666666"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "Japanese content:"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "2mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "qrcode": {"data": "こんにちは世界", "size": "30mm"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "27_qrcode.pdf", doc)
}
