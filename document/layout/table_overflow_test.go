package layout

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
)

// issue56Cell returns a table cell holding one 8pt line of text.
func issue56Cell(s string) document.TableCell {
	return document.TableCell{Content: []document.DocumentNode{
		&document.Text{Content: s, TextStyle: document.Style{FontSize: 8, LineHeight: 1}},
	}}
}

// issue56Table builds a one-column table with the given header and body rows.
func issue56Table(header []string, body ...string) *document.Table {
	tbl := &document.Table{}
	for _, s := range header {
		tbl.Header = append(tbl.Header, document.TableRow{Cells: []document.TableCell{issue56Cell(s)}})
	}
	for _, s := range body {
		tbl.Body = append(tbl.Body, document.TableRow{Cells: []document.TableCell{issue56Cell(s)}})
	}
	return tbl
}

// issue56BreakAvoidRow wraps a node the way template rows do: a horizontal
// BreakAvoid box with a single column.
func issue56BreakAvoidRow(n document.DocumentNode) *document.Box {
	return &document.Box{
		BoxStyle:    document.BoxStyle{Direction: document.DirectionHorizontal},
		BreakPolicy: document.BreakPolicy{BreakInside: document.BreakAvoid},
		Content:     []document.DocumentNode{&document.Box{Content: []document.DocumentNode{n}}},
	}
}

// placedText is a text line with its absolute position on the page.
type placedText struct {
	content string
	y       float64
}

// collectTexts returns every placed text node with its absolute Y.
func collectTexts(nodes []PlacedNode, originY float64) []placedText {
	var out []placedText
	for _, n := range nodes {
		y := originY + n.Position.Y
		if t, ok := n.Node.(*document.Text); ok && len(n.Children) == 0 {
			out = append(out, placedText{content: t.Content, y: y})
		}
		out = append(out, collectTexts(n.Children, y)...)
	}
	return out
}

// assertLinesOnPage checks that the page shows exactly want, in order, at
// strictly increasing Y positions (i.e. no two lines overlap).
func assertLinesOnPage(t *testing.T, page PageLayout, want []string) {
	t.Helper()
	got := collectTexts(page.Children, 0)
	if len(got) != len(want) {
		t.Fatalf("page has %d text lines %v, want %v", len(got), got, want)
	}
	for i, w := range want {
		if got[i].content != w {
			t.Errorf("line %d = %q, want %q", i, got[i].content, w)
		}
		if i > 0 && got[i].y <= got[i-1].y {
			t.Errorf("line %q at y=%.2f overlaps %q at y=%.2f", got[i].content, got[i].y, got[i-1].content, got[i-1].y)
		}
	}
}

// TestIssue56_TableAtPageBottom covers a table that starts with less room
// than one row at the bottom of a page. Its rows used to be measured at ~0
// height and drawn on top of each other there, with no overflow reported.
func TestIssue56_TableAtPageBottom(t *testing.T) {
	const pageH, left = 400.0, 5.0 // only 5pt left for the table
	size := document.Size{Width: 500, Height: pageH}
	spacer := func() document.DocumentNode {
		return &document.Box{BoxStyle: document.BoxStyle{Height: document.Pt(pageH - left)}}
	}

	tests := []struct {
		name string
		node document.DocumentNode
		want []string // lines expected on page 2
	}{
		{
			name: "break-avoid row, no header",
			node: issue56BreakAvoidRow(issue56Table(nil, "row 1", "row 2", "row 3")),
			want: []string{"row 1", "row 2", "row 3"},
		},
		{
			name: "break-avoid row, with header",
			node: issue56BreakAvoidRow(issue56Table([]string{"head"}, "row 1", "row 2", "row 3")),
			want: []string{"head", "row 1", "row 2", "row 3"},
		},
		{
			name: "bare table, no header",
			node: issue56Table(nil, "row 1", "row 2", "row 3"),
			want: []string{"row 1", "row 2", "row 3"},
		},
		{
			name: "bare table, with header",
			node: issue56Table([]string{"head"}, "row 1", "row 2"),
			want: []string{"head", "row 1", "row 2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &document.Document{Pages: []*document.Page{{
				Size:    size,
				Content: []document.DocumentNode{spacer(), tt.node},
			}}}
			p := NewPaginator(size, document.Edges{}, &mockFontResolver{})
			pages, err := paginateWithTimeout(t, p, doc)
			if err != nil {
				t.Fatalf("PaginateWithError: %v", err)
			}
			if len(pages) != 2 {
				t.Fatalf("want the table on page 2, got %d page(s): rows overlap at the bottom of page 1", len(pages))
			}
			assertLinesOnPage(t, pages[0], nil)
			assertLinesOnPage(t, pages[1], tt.want)
		})
	}
}

// TestIssue56_TableSplitsAtOverflowingRow checks that when the first rows
// fit, the row that does not fit and the rows after it continue on the
// next page, after the repeated header.
func TestIssue56_TableSplitsAtOverflowingRow(t *testing.T) {
	// 8pt lines: header + 2 body rows (24pt) fit in 28pt, the third does not.
	size := document.Size{Width: 500, Height: 28}
	tbl := issue56Table([]string{"head"}, "row 1", "row 2", "row 3", "row 4")
	doc := &document.Document{Pages: []*document.Page{{Size: size, Content: []document.DocumentNode{tbl}}}}

	p := NewPaginator(size, document.Edges{}, &mockFontResolver{})
	pages, err := paginateWithTimeout(t, p, doc)
	if err != nil {
		t.Fatalf("PaginateWithError: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2", len(pages))
	}
	assertLinesOnPage(t, pages[0], []string{"head", "row 1", "row 2"})
	assertLinesOnPage(t, pages[1], []string{"head", "row 3", "row 4"})
}

// TestIssue56_TallRowAtPageTopIsStillPlaced makes sure a row taller than a
// whole page is still placed (clipped) when the table is at the top of the
// page, where moving it to the next page cannot help.
func TestIssue56_TallRowAtPageTopIsStillPlaced(t *testing.T) {
	size := document.Size{Width: 500, Height: 20}
	tall := document.TableCell{Content: []document.DocumentNode{
		&document.Text{Content: "a", TextStyle: document.Style{FontSize: 8, LineHeight: 1}},
		&document.Text{Content: "b", TextStyle: document.Style{FontSize: 8, LineHeight: 1}},
		&document.Text{Content: "c", TextStyle: document.Style{FontSize: 8, LineHeight: 1}},
	}}
	for _, withHeader := range []bool{false, true} {
		tbl := &document.Table{Body: []document.TableRow{{Cells: []document.TableCell{tall}}}}
		if withHeader {
			tbl.Header = []document.TableRow{{Cells: []document.TableCell{issue56Cell("head")}}}
		}
		doc := &document.Document{Pages: []*document.Page{{Size: size, Content: []document.DocumentNode{tbl}}}}

		p := NewPaginator(size, document.Edges{}, &mockFontResolver{})
		pages, err := paginateWithTimeout(t, p, doc)
		if err != nil {
			t.Fatalf("header=%v: PaginateWithError: %v", withHeader, err)
		}
		if len(pages) != 1 {
			t.Fatalf("header=%v: got %d pages, want 1", withHeader, len(pages))
		}
		if len(collectTexts(pages[0].Children, 0)) == 0 {
			t.Errorf("header=%v: tall row was dropped", withHeader)
		}
	}
}
