package template

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
)

// paginateForTest runs the same layout pipeline as Render and returns the
// resolved page layouts without producing PDF bytes.
func paginateForTest(d *Document) []layout.PageLayout {
	doc := d.buildDocument()
	paginator := layout.NewPaginator(d.config.PageSize, d.config.Margins, d.fontResolver)
	paginator.SetHeaderFooter(d.buildSection(d.headerFn), d.buildSection(d.footerFn))
	pages := paginator.Paginate(doc)
	layout.ResolvePageNumbersWithFonts(pages, d.fontResolver)
	return pages
}

// hasText reports whether a Text node with the given content is placed.
func hasText(nodes []layout.PlacedNode, want string) bool {
	for _, n := range nodes {
		if tx, ok := n.Node.(*document.Text); ok && tx.Content == want {
			return true
		}
		if hasText(n.Children, want) {
			return true
		}
	}
	return false
}

func tableRowsJSON(n int) string {
	rows := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, fmt.Sprintf(`["%d","Item %03d"]`, i, i))
	}
	return strings.Join(rows, ",")
}

// TestFooterPageNumbersOnAutoPaginatedPages is an end-to-end regression
// test for #44: the repeated footer must show the current page number on
// every automatically generated overflow page.
func TestFooterPageNumbersOnAutoPaginatedPages(t *testing.T) {
	js := `{"page":{"size":"A4","margins":"15mm"},
	 "footer":[{"row":{"cols":[{"span":12,"elements":[
	   {"type":"text","content":"Page"},{"type":"pageNumber"},{"type":"totalPages"}]}]}}],
	 "body":[{"row":{"cols":[{"span":12,"table":{"header":["ID","Description"],"rows":[` + tableRowsJSON(150) + `]}}]}}]}`
	doc, err := FromJSON([]byte(js), nil)
	if err != nil {
		t.Fatal(err)
	}
	pages := paginateForTest(doc)
	if len(pages) < 3 {
		t.Fatalf("expected at least 3 pages, got %d", len(pages))
	}
	total := fmt.Sprint(len(pages))
	for i, p := range pages {
		if !hasText(p.Children, fmt.Sprint(i+1)) {
			t.Errorf("page %d: current page number missing from footer", i+1)
		}
		if !hasText(p.Children, total) {
			t.Errorf("page %d: total pages %s missing from footer", i+1, total)
		}
	}
}

// collectTexts returns the content of every leaf text line placed.
func collectTexts(nodes []layout.PlacedNode, out []string) []string {
	for _, n := range nodes {
		if tx, ok := n.Node.(*document.Text); ok && len(n.Children) == 0 {
			out = append(out, tx.Content)
		}
		out = collectTexts(n.Children, out)
	}
	return out
}

// TestPageNumberInNarrowColumn is a regression test: in a column narrower
// than the placeholder sentinel, the placeholder used to be wrapped into
// fragments that were printed verbatim (e.g. "?GP", "DF", "_P"...).
func TestPageNumberInNarrowColumn(t *testing.T) {
	js := `{"page":{"size":"A4","margins":"15mm"},
	 "footer":[{"row":{"cols":[
	   {"span":2,"text":"Page","style":{"size":24}},
	   {"span":1,"elements":[{"type":"pageNumber","style":{"size":24}}]},
	   {"span":1,"text":"of","style":{"size":24}},
	   {"span":1,"elements":[{"type":"totalPages","style":{"size":24}}]}]}}],
	 "body":[{"row":{"cols":[{"span":12,"table":{"header":["ID","Description"],"rows":[` + tableRowsJSON(110) + `]}}]}}]}`
	doc, err := FromJSON([]byte(js), nil)
	if err != nil {
		t.Fatal(err)
	}
	pages := paginateForTest(doc)
	if len(pages) < 2 {
		t.Fatalf("expected multiple pages, got %d", len(pages))
	}
	for i, p := range pages {
		for _, s := range collectTexts(p.Children, nil) {
			if strings.Contains(s, "GPDF") || strings.ContainsRune(s, 0) {
				t.Fatalf("page %d: unresolved placeholder fragment %q", i+1, s)
			}
		}
		if !hasText(p.Children, fmt.Sprint(i+1)) {
			t.Errorf("page %d: page number missing", i+1)
		}
	}
}
