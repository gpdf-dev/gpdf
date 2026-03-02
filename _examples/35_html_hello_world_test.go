package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/htmlpdf"
)

func TestExample_35_HTML_HelloWorld(t *testing.T) {
	html := `<h1>Hello, World!</h1><p>Generated from HTML using gpdf.</p>`

	result, err := htmlpdf.FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	data, err := result.Bytes()
	if err != nil {
		t.Fatalf("Bytes failed: %v", err)
	}

	assertValidPDF(t, data)
	writePDF(t, "35_html_hello_world.pdf", data)
	assertMatchesGolden(t, "35_html_hello_world.pdf", data)
}
