package overlay_test

import (
	"fmt"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_Overlay_02_PageNumbers(t *testing.T) {
	source := generateSourcePDF(t, 5)

	doc, err := gpdf.Open(source)
	if err != nil {
		t.Fatalf("gpdf.Open: %v", err)
	}

	count, err := doc.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}

	// Add page numbers to every page (bottom-right).
	err = doc.EachPage(func(pageIndex int, p *template.PageBuilder) {
		p.Absolute(document.Mm(170), document.Mm(285), func(c *template.ColBuilder) {
			c.Text(fmt.Sprintf("%d / %d", pageIndex+1, count),
				template.FontSize(10),
				template.AlignRight(),
			)
		}, template.AbsoluteWidth(document.Mm(20)))
	})
	if err != nil {
		t.Fatalf("EachPage: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "02_page_numbers.pdf", result)
}
