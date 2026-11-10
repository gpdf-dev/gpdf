package layout

import (
	"errors"
	"testing"
	"time"

	"github.com/gpdf-dev/gpdf/document"
)

// paginateWithTimeout runs PaginateWithError and fails the test instead of
// hanging when pagination does not terminate.
func paginateWithTimeout(t *testing.T, p *Paginator, doc *document.Document) ([]PageLayout, error) {
	t.Helper()
	type out struct {
		pages []PageLayout
		err   error
	}
	ch := make(chan out, 1)
	go func() {
		pages, err := p.PaginateWithError(doc)
		ch <- out{pages, err}
	}()
	select {
	case o := <-ch:
		return o.pages, o.err
	case <-time.After(5 * time.Second):
		t.Fatal("pagination did not terminate")
		return nil, nil
	}
}

func nofitText() *document.Text {
	return &document.Text{
		Content:   "hello world",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
}

func singlePageDoc(nodes ...document.DocumentNode) *document.Document {
	return &document.Document{Pages: []*document.Page{{Content: nodes}}}
}

func TestPaginateWithErrorContentDoesNotFit(t *testing.T) {
	tests := []struct {
		name    string
		size    document.Size
		margins document.Edges
		nodes   []document.DocumentNode
	}{
		{
			name:    "zero page size",
			size:    document.Size{},
			margins: document.UniformEdges(document.Mm(5)),
			nodes:   []document.DocumentNode{nofitText()},
		},
		{
			name:    "margins larger than page",
			size:    document.Size{Width: 100, Height: 100},
			margins: document.UniformEdges(document.Mm(20)),
			nodes:   []document.DocumentNode{nofitText()},
		},
		{
			name:    "body shorter than one line",
			size:    document.Size{Width: 60, Height: 40},
			margins: document.UniformEdges(document.Mm(5)),
			nodes:   []document.DocumentNode{nofitText()},
		},
		{
			name:    "padded box around unfit text",
			size:    document.Size{Width: 200, Height: 30},
			margins: document.UniformEdges(document.Pt(5)),
			nodes: []document.DocumentNode{&document.Box{
				Content:  []document.DocumentNode{nofitText()},
				BoxStyle: document.BoxStyle{Padding: document.UniformEdges(document.Pt(6))},
			}},
		},
		{
			name:    "content fits on page one, rest never fits",
			size:    document.Size{Width: 200, Height: 40},
			margins: document.UniformEdges(document.Pt(5)),
			nodes: []document.DocumentNode{
				nofitText(),
				&document.Text{Content: "huge", TextStyle: document.Style{FontSize: 48, LineHeight: 1.2}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPaginator(tt.size, tt.margins, &mockFontResolver{})
			_, err := paginateWithTimeout(t, p, singlePageDoc(tt.nodes...))
			if !errors.Is(err, ErrContentDoesNotFit) {
				t.Fatalf("err = %v, want ErrContentDoesNotFit", err)
			}
		})
	}
}

func TestPaginateNoFitDoesNotHang(t *testing.T) {
	p := NewPaginator(document.Size{}, document.UniformEdges(document.Mm(5)), &mockFontResolver{})
	done := make(chan struct{})
	go func() {
		p.Paginate(singlePageDoc(nofitText()))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Paginate did not terminate")
	}
}

func TestPaginateWithErrorNormalOverflow(t *testing.T) {
	var nodes []document.DocumentNode
	for i := 0; i < 200; i++ {
		nodes = append(nodes, nofitText())
	}
	p := NewPaginator(document.A6, document.UniformEdges(document.Mm(10)), &mockFontResolver{})
	pages, err := paginateWithTimeout(t, p, singlePageDoc(nodes...))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) < 2 {
		t.Fatalf("pages = %d, want multiple", len(pages))
	}
}

// Intentionally blank pages produced by page breaks around zero-height
// content still make progress and must not be reported as stuck.
func TestPaginateWithErrorBlankPagesFromBreaks(t *testing.T) {
	breakAfter := func() *document.Box {
		return &document.Box{BreakPolicy: document.BreakPolicy{BreakAfter: document.BreakPage}}
	}
	p := NewPaginator(document.A6, document.UniformEdges(document.Mm(10)), &mockFontResolver{})
	pages, err := paginateWithTimeout(t, p, singlePageDoc(
		nofitText(), breakAfter(), breakAfter(), breakAfter(), nofitText(),
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) < 3 {
		t.Fatalf("pages = %d, want at least 3", len(pages))
	}
}

func TestPaginateWithErrorEmptyPageWithZeroArea(t *testing.T) {
	// A page with no content never needs any room.
	p := NewPaginator(document.Size{}, document.Edges{}, &mockFontResolver{})
	if _, err := paginateWithTimeout(t, p, singlePageDoc()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHasVisibleContent(t *testing.T) {
	leaf := func(h float64) PlacedNode { return PlacedNode{Size: document.Size{Height: h}} }
	tests := []struct {
		name  string
		nodes []PlacedNode
		want  bool
	}{
		{"none", nil, false},
		{"zero-height leaf", []PlacedNode{leaf(0)}, false},
		{"tall container with only empty leaves", []PlacedNode{{
			Size: document.Size{Height: 20}, Children: []PlacedNode{leaf(0)},
		}}, false},
		{"nested visible leaf", []PlacedNode{{
			Children: []PlacedNode{leaf(0), {Children: []PlacedNode{leaf(14)}}},
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasVisibleContent(tt.nodes); got != tt.want {
				t.Errorf("hasVisibleContent = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemainingText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		lines []string
		k     int
		want  string
	}{
		{"word wrap consumes one space", "alpha beta gamma", []string{"alpha", "beta", "gamma"}, 1, "beta gamma"},
		{"mid-word break adds nothing", "SKU-0042", []string{"SKU-", "0042"}, 1, "0042"},
		{"CJK break adds nothing", "日本語の文章", []string{"日本語", "の文章"}, 1, "の文章"},
		{"forced newline kept", "one\ntwo\nthree", []string{"one", "two", "three"}, 1, "two\nthree"},
		{"blank line kept", "a\n\nb", []string{"a", "", "b"}, 1, "\nb"},
		{"leading space after char break", "ab cd", []string{"ab", " cd"}, 1, " cd"},
		{"first line", "alpha beta", []string{"alpha", "beta"}, 0, "alpha beta"},
		{"unmatched lines fall back to join", "a   b", []string{"a", "x"}, 1, "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remainingText(tt.text, tt.lines, tt.k); got != tt.want {
				t.Errorf("remainingText = %q, want %q", got, tt.want)
			}
		})
	}
}
