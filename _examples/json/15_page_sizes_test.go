package json_test

import (
	"fmt"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_15_PageSizes(t *testing.T) {
	sizes := []struct {
		name string
		size string
		file string
	}{
		{"A4", "A4", "15a_pagesize_a4.pdf"},
		{"A3", "A3", "15b_pagesize_a3.pdf"},
		{"Letter", "Letter", "15c_pagesize_letter.pdf"},
		{"Legal", "Legal", "15d_pagesize_legal.pdf"},
	}
	for _, s := range sizes {
		t.Run(s.name, func(t *testing.T) {
			schema := []byte(fmt.Sprintf(`{
				"page": {"size": "%s", "margins": "20mm"},
				"body": [
					{"row": {"cols": [
						{"span": 12, "text": "Page Size: %s", "style": {"size": 20, "bold": true}}
					]}},
					{"row": {"cols": [
						{"span": 12, "spacer": "10mm"}
					]}},
					{"row": {"cols": [
						{"span": 12, "text": "This page demonstrates the %s page format."}
					]}}
				]
			}`, s.size, s.name, s.name))
			doc, err := template.FromJSON(schema, nil)
			if err != nil {
				t.Fatalf("FromJSON error: %v", err)
			}
			testutil.GeneratePDF(t, s.file, doc)
		})
	}
}
