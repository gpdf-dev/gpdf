package merge_test

import (
	"fmt"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_Merge_04_MergeAndOverlay(t *testing.T) {
	// Generate two PDFs and merge them.
	part1 := generateDoc(t, 2, "Part 1")
	part2 := generateDoc(t, 2, "Part 2")

	merged, err := gpdf.Merge([]gpdf.Source{
		{Data: part1},
		{Data: part2},
	})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	// Open the merged PDF and add page numbers on every page.
	doc, err := gpdf.Open(merged)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	pageCount, err := doc.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	err = doc.EachPage(func(pageIndex int, p *template.PageBuilder) {
		p.Absolute(document.Mm(95), document.Mm(285), func(c *template.ColBuilder) {
			c.Text(fmt.Sprintf("Page %d / %d", pageIndex+1, pageCount),
				template.FontSize(9),
				template.TextColor(pdf.Gray(0.4)),
			)
		})
	})
	if err != nil {
		t.Fatalf("EachPage: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "04_merge_and_overlay.pdf", result)
}
