package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_00_Overview(t *testing.T) {
	// Define a PDF document entirely in JSON.
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "JSON Schema Example", "author": "gpdf"},
		"header": [
			{"row": {"cols": [
				{"span": 6, "text": "gpdf JSON Schema", "style": {"size": 16, "bold": true, "color": "#1A237E"}},
				{"span": 6, "text": "Document Header", "style": {"align": "right", "italic": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#1A237E", "thickness": "1pt"}}
			]}}
		],
		"footer": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "line"},
					{"type": "pageNumber", "style": {"align": "center"}}
				]}
			]}}
		],
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "JSON Schema Generation", "style": {"size": 24, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "text": "This PDF was generated entirely from a JSON schema definition. No Go builder code needed!"}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Features", "style": {"size": 16, "bold": true}},
					{"type": "list", "list": {"items": [
						"Declarative document definition",
						"All element types supported",
						"Style options (bold, italic, color, align)",
						"Header and footer support"
					]}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Supported Elements", "style": {"size": 16, "bold": true}},
					{"type": "list", "list": {"type": "ordered", "items": [
						"Text with styles",
						"Tables with headers",
						"Lists (ordered/unordered)",
						"Lines and spacers",
						"QR codes and barcodes",
						"Images (base64)"
					]}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Feature", "Format", "Status"],
					"rows": [
						["Text styling", "JSON style object", "Supported"],
						["Tables", "header + rows arrays", "Supported"],
						["Lists", "ordered/unordered", "Supported"],
						["Images", "base64 encoded", "Supported"],
						["QR codes", "data string", "Supported"],
						["Barcodes", "Code128", "Supported"]
					],
					"columnWidths": [35, 35, 30],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 6, "qrcode": {"data": "https://gpdf.dev", "size": "25mm"}},
				{"span": 6, "barcode": {"data": "GPDF-JSON-001", "format": "code128"}}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDF(t, "00_overview.pdf", doc)
}
