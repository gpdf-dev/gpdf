package template

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
)

func generateWithTimeout(t *testing.T, doc *Document) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() {
		_, err := doc.Generate()
		ch <- err
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Generate did not return")
		return nil
	}
}

func TestGenerateContentDoesNotFit(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{"zero page size", []Option{WithPageSize(document.Size{})}},
		// 50 x 30 mm label with the default 20 mm margins.
		{"default margins on small label", []Option{WithPageSize(document.Size{
			Width:  document.Mm(50).Resolve(0, 0),
			Height: document.Mm(30).Resolve(0, 0),
		})}},
		{"body shorter than one line", []Option{
			WithPageSize(document.Size{Width: 60, Height: 40}),
			WithMargins(document.UniformEdges(document.Mm(5))),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New(tt.opts...)
			doc.AddPage().AutoRow(func(r *RowBuilder) {
				r.Col(12, func(c *ColBuilder) {
					// A bold 14 pt word wider than the column forces a
					// mid-word break, which used to grow the overflow text
					// on every page instead of repeating it.
					c.Text("SKU-0042", Bold(), FontSize(14))
					c.Text("hello world")
				})
			})
			if err := generateWithTimeout(t, doc); !errors.Is(err, layout.ErrContentDoesNotFit) {
				t.Fatalf("err = %v, want layout.ErrContentDoesNotFit", err)
			}
		})
	}
}

func TestGenerateHeaderTallerThanPage(t *testing.T) {
	doc := New(
		WithPageSize(document.Size{Width: 200, Height: 80}),
		WithMargins(document.UniformEdges(document.Pt(5))),
	)
	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("line 1")
				c.Text("line 2")
				c.Text("line 3")
				c.Text("line 4")
			})
		})
	})
	doc.AddPage().AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("body") })
	})
	if err := generateWithTimeout(t, doc); !errors.Is(err, layout.ErrContentDoesNotFit) {
		t.Fatalf("err = %v, want layout.ErrContentDoesNotFit", err)
	}
}

func TestGenerateSmallLabelWithTightMargins(t *testing.T) {
	doc := New(
		WithPageSize(document.Size{
			Width:  document.Mm(50).Resolve(0, 0),
			Height: document.Mm(30).Resolve(0, 0),
		}),
		WithMargins(document.UniformEdges(document.Mm(3))),
	)
	doc.AddPage().AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("SKU-0042") })
	})
	if err := generateWithTimeout(t, doc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Text that overflows onto the next page keeps its forced line breaks
// instead of having them collapsed into spaces.
func TestOverflowTextKeepsNewlines(t *testing.T) {
	doc := New(
		WithPageSize(document.Size{Width: 200, Height: 90}),
		WithMargins(document.UniformEdges(document.Pt(10))),
	)
	doc.AddPage().AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("line one\nline two\nline three\nline four\nline five\nline six\nline seven")
		})
	})
	paginator := layout.NewPaginator(doc.config.PageSize, doc.config.Margins, doc.fontResolver)
	pages, err := paginator.PaginateWithError(doc.buildDocument())
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) < 2 {
		t.Fatalf("pages = %d, want at least 2", len(pages))
	}
	var lines []string
	var walk func([]layout.PlacedNode)
	walk = func(ns []layout.PlacedNode) {
		for _, n := range ns {
			if txt, ok := n.Node.(*document.Text); ok && len(n.Children) == 0 {
				lines = append(lines, txt.Content)
			}
			walk(n.Children)
		}
	}
	for _, p := range pages {
		walk(p.Children)
	}
	want := []string{"line one", "line two", "line three", "line four", "line five", "line six", "line seven"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("lines = %q, want %q", lines, want)
	}
}
