package gpdf

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument(
		WithPageSize(A4),
		WithMargins(document.UniformEdges(document.Mm(15))),
		WithMetadata(document.DocumentMetadata{
			Title: "Test",
		}),
	)
	if doc == nil {
		t.Fatal("NewDocument returned nil")
	}
}

func TestNewDocumentGenerate(t *testing.T) {
	doc := NewDocument()
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Test")
		})
	})
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Generated PDF is empty")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", data[:5])
	}
}

func TestPageSizes(t *testing.T) {
	if A4.Width <= 0 || A4.Height <= 0 {
		t.Error("A4 has invalid dimensions")
	}
	if A3.Width <= 0 || A3.Height <= 0 {
		t.Error("A3 has invalid dimensions")
	}
	if Letter.Width <= 0 || Letter.Height <= 0 {
		t.Error("Letter has invalid dimensions")
	}
	if Legal.Width <= 0 || Legal.Height <= 0 {
		t.Error("Legal has invalid dimensions")
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version is empty")
	}
}

func TestReexportedOptions(t *testing.T) {
	// Verify all re-exported option functions are non-nil.
	if WithPageSize == nil {
		t.Error("WithPageSize is nil")
	}
	if WithMargins == nil {
		t.Error("WithMargins is nil")
	}
	if WithFont == nil {
		t.Error("WithFont is nil")
	}
	if WithDefaultFont == nil {
		t.Error("WithDefaultFont is nil")
	}
	if WithMetadata == nil {
		t.Error("WithMetadata is nil")
	}
}

func TestOpen(t *testing.T) {
	// Generate a valid PDF first.
	doc := NewDocument()
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Hello from original")
		})
	})
	pdfData, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Open the generated PDF.
	existing, err := Open(pdfData)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if existing == nil {
		t.Fatal("Open returned nil document")
	}

	// Save without modifications to verify round-trip.
	result, err := existing.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("Save returned empty data")
	}
	if string(result[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", result[:5])
	}
}

func TestOpenInvalidData(t *testing.T) {
	_, err := Open([]byte("not a pdf"))
	if err == nil {
		t.Fatal("Open should fail with invalid data")
	}
}

func TestOpenEmptyData(t *testing.T) {
	_, err := Open(nil)
	if err == nil {
		t.Fatal("Open should fail with nil data")
	}
}
