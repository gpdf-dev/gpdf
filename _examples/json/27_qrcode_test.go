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
				{"span": 12, "elements": [
					{"type": "text", "content": "QR Code Examples", "style": {"size": 18, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Basic QR code (URL):"},
					{"type": "spacer", "height": "2mm"},
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "QR codes with different sizes:"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 4, "elements": [
					{"type": "text", "content": "20mm", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "20mm"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "30mm", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "30mm"}}
				]},
				{"span": 4, "elements": [
					{"type": "text", "content": "40mm", "style": {"size": 9, "color": "gray(0.4)"}},
					{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "40mm"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Error correction levels (L / M / Q / H):"},
					{"type": "spacer", "height": "2mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [
					{"type": "text", "content": "Level L", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "HELLO", "size": "25mm", "errorCorrection": "L"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "Level M", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "HELLO", "size": "25mm", "errorCorrection": "M"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "Level Q", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "HELLO", "size": "25mm", "errorCorrection": "Q"}}
				]},
				{"span": 3, "elements": [
					{"type": "text", "content": "Level H", "style": {"size": 9}},
					{"type": "qrcode", "qrcode": {"data": "HELLO", "size": "25mm", "errorCorrection": "H"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Japanese content:"},
					{"type": "spacer", "height": "2mm"},
					{"type": "qrcode", "qrcode": {"data": "こんにちは世界", "size": "30mm"}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "27_qrcode.pdf", doc)
}
