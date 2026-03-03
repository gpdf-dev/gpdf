package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_08_TableStyled(t *testing.T) {
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
				{"span": 12, "table": {
					"header": ["Product", "Category", "Qty", "Unit Price", "Total"],
					"rows": {{toJSON .Rows}},
					"columnWidths": [30, 20, 10, 20, 20],
					"headerStyle": {"color": "white", "background": "#1A237E"},
					"stripeColor": "#F5F5F5"
				}}
			]}}
		]
	}`)

	data := map[string]any{
		"Title": "Styled Table",
		"Rows": [][]string{
			{"Laptop Pro 15", "Electronics", "2", "$1,299.00", "$2,598.00"},
			{"Wireless Mouse", "Accessories", "10", "$29.99", "$299.90"},
			{"USB-C Hub", "Accessories", "5", "$49.99", "$249.95"},
			{"Monitor 27\"", "Electronics", "3", "$399.00", "$1,197.00"},
			{"Keyboard", "Accessories", "10", "$79.99", "$799.90"},
			{"Webcam HD", "Electronics", "4", "$89.99", "$359.96"},
		},
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "08_table_styled.pdf", doc)
}
