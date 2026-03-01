package examples_test

import (
	"testing"
	gotemplate "text/template"

	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_31_GoTemplate(t *testing.T) {
	// Define a document template with Go template expressions.
	// The {{.Field}} placeholders are resolved with the data map.
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "{{.Title}}", "author": "{{.Author}}"},
		"header": [
			{"row": {"cols": [
				{"span": 6, "text": "{{.Company}}", "style": {"size": 18, "bold": true, "color": "#1A237E"}},
				{"span": 6, "text": "{{.DocType}}", "style": {"size": 22, "bold": true, "align": "right", "color": "#1A237E"}}
			]}},
			{"row": {"cols": [
				{"span": 12, "line": {"color": "#1A237E", "thickness": "2pt"}}
			]}}
		],
		"body": [
			{"row": {"height": "25mm", "cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "Bill To:", "style": {"bold": true, "color": "#666666"}},
					{"type": "text", "content": "{{.ClientName}}", "style": {"bold": true}},
					{"type": "text", "content": "{{.ClientAddress}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "Date: {{.Date}}", "style": {"align": "right"}},
					{"type": "text", "content": "Invoice: {{.InvoiceNumber}}", "style": {"align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "10mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "table": {
					"header": ["Description", "Qty", "Unit Price", "Amount"],
					"rows": {{toJSON .Items}},
					"columnWidths": [40, 15, 20, 25],
					"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{"row": {"cols": [
				{"span": 8, "text": ""},
				{"span": 4, "elements": [
					{"type": "text", "content": "Subtotal: {{.Subtotal}}", "style": {"align": "right"}},
					{"type": "text", "content": "Tax ({{.TaxRate}}): {{.TaxAmount}}", "style": {"align": "right"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "line", "line": {"thickness": "1pt"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Total: {{.Total}}", "style": {"size": 14, "bold": true, "align": "right"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "15mm"}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "line", "line": {"color": "#CCCCCC"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.FooterNote}}", "style": {"align": "center", "italic": true, "color": "#999999"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":         "Invoice #INV-2026-042",
		"Author":        "ACME Corporation",
		"Company":       "ACME Corporation",
		"DocType":       "INVOICE",
		"ClientName":    "Jane Doe",
		"ClientAddress": "789 Client Blvd, Tokyo, Japan",
		"Date":          "March 1, 2026",
		"InvoiceNumber": "#INV-2026-042",
		"Items": [][]string{
			{"Web Application Development", "80 hrs", "$150.00", "$12,000.00"},
			{"API Integration", "30 hrs", "$160.00", "$4,800.00"},
			{"UI/UX Design", "25 hrs", "$120.00", "$3,000.00"},
			{"Performance Optimization", "15 hrs", "$170.00", "$2,550.00"},
			{"Documentation", "10 hrs", "$100.00", "$1,000.00"},
		},
		"Subtotal":   "$23,350.00",
		"TaxRate":    "10%",
		"TaxAmount":  "$2,335.00",
		"Total":      "$25,685.00",
		"FooterNote": "Thank you for your business! Payment due within 30 days.",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	generatePDF(t, "31_go_template.pdf", doc)
}

func TestExample_31_GoTemplate_FromTemplate(t *testing.T) {
	// Using FromTemplate with a pre-parsed Go template for more control.
	tmplStr := `{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "{{.Title}}"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
			]}},
			{"row": {"cols": [
				{"span": 12, "spacer": "5mm"}
			]}},
			{{- range $i, $section := .Sections}}
			{{- if $i}},{{end}}
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{$section.Heading}}", "style": {"size": 16, "bold": true, "color": "#1A237E"}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{$section.Body}}"},
					{"type": "spacer", "height": "8mm"}
				]}
			]}}
			{{- end}}
		]
	}`

	tmpl, err := gotemplate.New("report").Funcs(template.TemplateFuncMap()).Parse(tmplStr)
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	type section struct {
		Heading string
		Body    string
	}

	data := map[string]any{
		"Title": "Quarterly Report - Q1 2026",
		"Sections": []section{
			{
				Heading: "Executive Summary",
				Body:    "This quarter showed strong growth across all product lines. Revenue increased 25% year-over-year, driven primarily by our new enterprise offerings.",
			},
			{
				Heading: "Product Development",
				Body:    "The gpdf library reached v0.8 with Go template integration and JSON schema support. Community adoption continues to accelerate with over 500 GitHub stars.",
			},
			{
				Heading: "Market Analysis",
				Body:    "The PDF generation market continues to grow, with increasing demand for programmatic document creation. Our zero-dependency approach resonates well with Go developers.",
			},
			{
				Heading: "Next Steps",
				Body:    "Focus areas for Q2 include reusable components (invoice, report, letter templates), fuzz testing, and preparation for the v1.0 release.",
			},
		},
	}

	doc, err := template.FromTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("FromTemplate error: %v", err)
	}
	generatePDF(t, "31_go_template_from_template.pdf", doc)
}
