package template

import (
	"bytes"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// generateTestPDF creates a simple PDF with the given number of pages.
func generateTestPDF(t *testing.T, numPages int) []byte {
	t.Helper()
	doc := New(
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(20))),
	)

	for i := 0; i < numPages; i++ {
		page := doc.AddPage()
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("Original content", FontSize(14))
			})
		})
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("generate test PDF: %v", err)
	}
	return data
}

func TestOpenExisting(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	count, err := doc.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 1 {
		t.Errorf("PageCount = %d, want 1", count)
	}
}

func TestOpenExisting_MultiplePages(t *testing.T) {
	data := generateTestPDF(t, 5)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	count, err := doc.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 5 {
		t.Errorf("PageCount = %d, want 5", count)
	}
}

func TestOverlay_SinglePage(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	err = doc.Overlay(0, func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("OVERLAY TEXT", FontSize(24))
			})
		})
	})
	if err != nil {
		t.Fatalf("Overlay: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Result should be larger than original.
	if len(result) <= len(data) {
		t.Errorf("result (%d bytes) should be larger than original (%d bytes)", len(result), len(data))
	}

	// Result should contain the overlay text.
	if !bytes.Contains(result, []byte("OVERLAY TEXT")) {
		t.Error("result should contain overlay text")
	}

	// Result should still be a valid PDF.
	r, err := pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	count, _ := r.PageCount()
	if count != 1 {
		t.Errorf("page count after overlay = %d, want 1", count)
	}
}

func TestOverlay_MultiplePages(t *testing.T) {
	data := generateTestPDF(t, 3)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	// Overlay on specific pages.
	for _, i := range []int{0, 2} {
		err = doc.Overlay(i, func(p *PageBuilder) {
			p.AutoRow(func(r *RowBuilder) {
				r.Col(12, func(c *ColBuilder) {
					c.Text("STAMP", FontSize(36))
				})
			})
		})
		if err != nil {
			t.Fatalf("Overlay(%d): %v", i, err)
		}
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	r, err := pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	count, _ := r.PageCount()
	if count != 3 {
		t.Errorf("page count = %d, want 3", count)
	}
}

func TestEachPage(t *testing.T) {
	data := generateTestPDF(t, 3)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	visited := 0
	err = doc.EachPage(func(pageIndex int, p *PageBuilder) {
		visited++
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("Footer", FontSize(10))
			})
		})
	})
	if err != nil {
		t.Fatalf("EachPage: %v", err)
	}
	if visited != 3 {
		t.Errorf("EachPage visited %d pages, want 3", visited)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	r, err := pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	count, _ := r.PageCount()
	if count != 3 {
		t.Errorf("page count = %d, want 3", count)
	}
}

func TestOverlay_NoContent(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	// Overlay with no content should be a no-op.
	err = doc.Overlay(0, func(p *PageBuilder) {
		// empty
	})
	if err != nil {
		t.Fatalf("Overlay: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Should still be valid.
	_, err = pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
}

func TestOverlay_AbsolutePosition(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	err = doc.Overlay(0, func(p *PageBuilder) {
		p.Absolute(document.Mm(100), document.Mm(200), func(c *ColBuilder) {
			c.Text("ABSOLUTE", FontSize(18))
		})
	})
	if err != nil {
		t.Fatalf("Overlay: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if !bytes.Contains(result, []byte("ABSOLUTE")) {
		t.Error("result should contain absolute-positioned text")
	}

	_, err = pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
}

func TestOverlay_PageOutOfRange(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	err = doc.Overlay(5, func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("Should fail", FontSize(12))
			})
		})
	})
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

func TestSave_NoModifications(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatalf("OpenExisting: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// No modifications should produce identical output.
	if !bytes.Equal(result, data) {
		t.Error("no-modification save should produce identical output")
	}
}
