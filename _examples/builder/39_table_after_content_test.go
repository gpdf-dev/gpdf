package builder_test

import (
	"fmt"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

// TestExample_39_TableAfterContent is a regression test: a multi-page table
// preceded by other rows must start on the first page and paginate from
// there (header repeated on each page), instead of the whole table being
// pushed to page 2 and leaving page 1 mostly blank. Rows were made atomic
// (BreakAvoid) for issue #24; rows containing a table must stay splittable.
func TestExample_39_TableAfterContent(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	rows := make([][]string, 60)
	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = fmt.Sprintf("Item-%03d", i+1)
		rows[i] = []string{ids[i], fmt.Sprintf("%d", (i+1)*100)}
	}

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Table After Content", template.FontSize(18), template.Bold())
			c.Text("Introductory paragraph before the table.")
			c.Spacer(document.Mm(5))
		})
	})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table([]string{"Item", "Amount"}, rows)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "39_table_after_content.pdf", doc)

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	testutil.AssertTableSplitsFromFirstPage(t, data, "Table After Content", "Item", ids)
}
