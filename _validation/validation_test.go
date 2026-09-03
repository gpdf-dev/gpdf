// Package validation_test uses external tools (pdfcpu, poppler) to verify
// that PDF output from gpdf is structurally valid and readable.
//
// This test suite lives in its own module to avoid adding dependencies
// to the core gpdf library. It mirrors the _benchmark/ pattern.
//
// Run:
//
//	cd _validation && go test -v ./...
//
// Poppler tests (pdftotext, pdfinfo) are automatically skipped if the
// CLI tools are not installed.
package validation_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
)

// ---------------------------------------------------------------------------
// Test PDF generators — each returns a named PDF for validation
// ---------------------------------------------------------------------------

type testPDF struct {
	name string
	gen  func() ([]byte, error)
}

func testPDFs() []testPDF {
	return []testPDF{
		{"hello_world", genHelloWorld},
		{"styled_text", genStyledText},
		{"table", genTable},
		{"multi_page", genMultiPage},
		{"header_footer", genHeaderFooter},
		{"grid_layout", genGridLayout},
		{"invoice", genInvoice},
		{"report", genReport},
		{"letter", genLetter},
		{"rich_text", genRichText},
		{"list", genList},
		{"qrcode", genQRCode},
		{"barcode", genBarcode},
	}
}

func genHelloWorld() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Hello, World!", template.FontSize(24))
		})
	})
	return doc.Generate()
}

func genStyledText() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Bold Text", template.Bold(), template.FontSize(16))
			c.Text("Italic Text", template.Italic())
			c.Text("Colored Text", template.TextColor(pdf.Red))
			c.Text("Underlined Text", template.Underline())
			c.Text("Right Aligned", template.AlignRight())
			c.Text("Centered", template.AlignCenter())
		})
	})
	return doc.Generate()
}

func genTable() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Name", "Age", "City"},
				[][]string{
					{"Alice", "30", "Tokyo"},
					{"Bob", "25", "New York"},
					{"Charlie", "35", "London"},
				},
				template.ColumnWidths(40, 20, 40),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(pdf.RGBHex(0x333333)),
				),
				template.TableStripe(pdf.RGBHex(0xF0F0F0)),
			)
		})
	})
	return doc.Generate()
}

func genMultiPage() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	for i := range 5 {
		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text(fmt.Sprintf("Page %d of 5", i+1), template.FontSize(20))
				c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
			})
		})
	}
	return doc.Generate()
}

func genHeaderFooter() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	doc.Header(func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("Document Header", template.Bold(), template.FontSize(10))
				c.Line()
			})
		})
	})
	doc.Footer(func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Line()
				c.PageNumber(template.AlignCenter(), template.FontSize(8))
			})
		})
	})
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Body content with header and footer.")
		})
	})
	return doc.Generate()
}

func genGridLayout() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Column 1 (4/12)")
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Column 2 (4/12)")
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Column 3 (4/12)")
		})
	})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Half 1")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Half 2")
		})
	})
	return doc.Generate()
}

func genInvoice() ([]byte, error) {
	doc := template.Invoice(template.InvoiceData{
		Number:  "#INV-001",
		Date:    "2026-03-01",
		DueDate: "2026-03-31",
		From: template.InvoiceParty{
			Name:    "ACME Corp",
			Address: []string{"123 Main St", "City, ST 00000"},
		},
		To: template.InvoiceParty{
			Name:    "Client Inc.",
			Address: []string{"456 Side St"},
		},
		Items: []template.InvoiceItem{
			{Description: "Service A", Quantity: "10 hrs", UnitPrice: 100, Amount: 1000},
			{Description: "Service B", Quantity: "5 hrs", UnitPrice: 200, Amount: 1000},
		},
		TaxRate: 10,
		Notes:   "Thank you for your business!",
	})
	return doc.Generate()
}

func genReport() ([]byte, error) {
	doc := template.Report(template.ReportData{
		Title:    "Quarterly Report",
		Subtitle: "Q1 2026",
		Author:   "Test Author",
		Sections: []template.ReportSection{
			{
				Title:   "Summary",
				Content: "This is the executive summary section.",
				Metrics: []template.ReportMetric{
					{Label: "Revenue", Value: "$1M", ColorHex: 0x2E7D32},
					{Label: "Growth", Value: "+15%", ColorHex: 0x1565C0},
				},
			},
			{
				Title: "Data",
				Table: &template.ReportTable{
					Header: []string{"Item", "Value"},
					Rows:   [][]string{{"A", "100"}, {"B", "200"}},
				},
			},
		},
	})
	return doc.Generate()
}

func genLetter() ([]byte, error) {
	doc := template.Letter(template.LetterData{
		From: template.LetterParty{
			Name:    "ACME Corp",
			Address: []string{"123 Main St", "City, ST 00000"},
		},
		To: template.LetterParty{
			Name:    "Mr. John Smith",
			Address: []string{"456 Side St"},
		},
		Date:     "March 1, 2026",
		Subject:  "Test Letter",
		Greeting: "Dear Mr. Smith,",
		Body: []string{
			"This is the first paragraph.",
			"This is the second paragraph.",
		},
		Closing:   "Sincerely,",
		Signature: "Jane Doe",
	})
	return doc.Generate()
}

func genRichText() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.RichText(func(rt *template.RichTextBuilder) {
				rt.Span("Normal text ")
				rt.Span("bold", template.Bold())
				rt.Span(" and ")
				rt.Span("italic", template.Italic())
				rt.Span(" text.")
			})
		})
	})
	return doc.Generate()
}

func genList() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Unordered List:", template.Bold())
			c.List([]string{"First item", "Second item", "Third item"})
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Ordered List:", template.Bold())
			c.OrderedList([]string{"Step one", "Step two", "Step three"})
		})
	})
	return doc.Generate()
}

func genQRCode() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("QR Code:", template.Bold())
			c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(30)))
		})
	})
	return doc.Generate()
}

func genBarcode() ([]byte, error) {
	doc := template.New(template.WithPageSize(document.A4))
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Barcode:", template.Bold())
			c.Barcode("GPDF-2026-001",
				template.BarcodeWidth(document.Mm(60)),
				template.BarcodeHeight(document.Mm(15)))
		})
	})
	return doc.Generate()
}

// ---------------------------------------------------------------------------
// pdfcpu validation
// ---------------------------------------------------------------------------

func TestPdfcpu_Validate(t *testing.T) {
	for _, tp := range testPDFs() {
		t.Run(tp.name, func(t *testing.T) {
			data, err := tp.gen()
			if err != nil {
				t.Fatalf("generate %s: %v", tp.name, err)
			}
			if err := pdfcpuapi.Validate(bytes.NewReader(data), nil); err != nil {
				t.Errorf("pdfcpu validate %s: %v", tp.name, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// poppler validation (pdfinfo + pdftotext)
// ---------------------------------------------------------------------------

func TestPoppler_PdfInfo(t *testing.T) {
	if _, err := exec.LookPath("pdfinfo"); err != nil {
		t.Skip("pdfinfo not found; install poppler-utils to enable this test")
	}

	tmpDir := t.TempDir()

	for _, tp := range testPDFs() {
		t.Run(tp.name, func(t *testing.T) {
			data, err := tp.gen()
			if err != nil {
				t.Fatalf("generate %s: %v", tp.name, err)
			}

			path := filepath.Join(tmpDir, tp.name+".pdf")
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}

			out, err := exec.Command("pdfinfo", path).CombinedOutput()
			if err != nil {
				t.Errorf("pdfinfo %s failed: %v\n%s", tp.name, err, string(out))
			} else {
				t.Logf("pdfinfo %s:\n%s", tp.name, string(out))
			}
		})
	}
}

func TestPoppler_PdfToText(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not found; install poppler-utils to enable this test")
	}

	tmpDir := t.TempDir()

	for _, tp := range testPDFs() {
		t.Run(tp.name, func(t *testing.T) {
			data, err := tp.gen()
			if err != nil {
				t.Fatalf("generate %s: %v", tp.name, err)
			}

			path := filepath.Join(tmpDir, tp.name+".pdf")
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}

			out, err := exec.Command("pdftotext", path, "-").CombinedOutput()
			if err != nil {
				t.Errorf("pdftotext %s failed: %v\n%s", tp.name, err, string(out))
			} else {
				t.Logf("pdftotext %s: extracted %d bytes", tp.name, len(out))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Overlay with an embedded font — see gpdf#37
// ---------------------------------------------------------------------------

// TestOverlayEmbeddedFont checks that overlay text drawn in a font registered
// via WithFont produces a valid PDF. The font is embedded as a Type0 composite
// font through an incremental update; before the fix the overlay wrote raw
// UTF-8 bytes against a bare /TrueType dict and every non-ASCII character
// rendered as "?".
func TestOverlayEmbeddedFont(t *testing.T) {
	fontPath := filepath.Join("..", "..", "NotoSansJP-Regular.ttf")
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		t.Skipf("font fixture not found: %s", fontPath)
	}

	plain, err := genHelloWorld()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	doc, err := gpdf.Open(plain, template.WithFont("NotoSansJP", fontData))
	if err != nil {
		t.Fatalf("open PDF: %v", err)
	}
	err = doc.Overlay(0, func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("日本語オーバーレイ", template.FontFamily("NotoSansJP"))
			})
		})
	})
	if err != nil {
		t.Fatalf("overlay: %v", err)
	}
	out, err := doc.Save()
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := pdfcpuapi.Validate(bytes.NewReader(out), nil); err != nil {
		t.Errorf("pdfcpu validate overlay with embedded font: %v", err)
	}
	if !bytes.Contains(out, []byte("/Identity-H")) {
		t.Error("overlay font should be embedded as a Type0/Identity-H font")
	}
}

// ---------------------------------------------------------------------------
// Object stream (/ObjStm) round-trip — see gpdf#35
// ---------------------------------------------------------------------------

// TestQPDF_ObjectStreamRoundTrip converts a generated PDF into the compressed
// object stream layout that most real-world producers emit, then reads it back
// through gpdf's merge and overlay paths and validates the result with pdfcpu.
func TestQPDF_ObjectStreamRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("qpdf"); err != nil {
		t.Skip("qpdf not found; install qpdf to enable this test")
	}

	plain, err := genHelloWorld()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	tmpDir := t.TempDir()
	plainPath := filepath.Join(tmpDir, "plain.pdf")
	objStmPath := filepath.Join(tmpDir, "objstm.pdf")
	if err := os.WriteFile(plainPath, plain, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if out, err := exec.Command("qpdf", "--object-streams=generate", plainPath, objStmPath).CombinedOutput(); err != nil {
		t.Fatalf("qpdf --object-streams=generate: %v\n%s", err, out)
	}
	objStm, err := os.ReadFile(objStmPath)
	if err != nil {
		t.Fatalf("read objstm.pdf: %v", err)
	}

	t.Run("merge", func(t *testing.T) {
		merged, err := gpdf.Merge([]gpdf.Source{{Data: objStm}, {Data: objStm}})
		if err != nil {
			t.Fatalf("merge object stream PDF: %v", err)
		}
		if err := pdfcpuapi.Validate(bytes.NewReader(merged), nil); err != nil {
			t.Errorf("pdfcpu validate merged: %v", err)
		}
	})

	t.Run("overlay", func(t *testing.T) {
		doc, err := gpdf.Open(objStm)
		if err != nil {
			t.Fatalf("open object stream PDF: %v", err)
		}
		err = doc.Overlay(0, func(p *template.PageBuilder) {
			p.AutoRow(func(r *template.RowBuilder) {
				r.Col(12, func(c *template.ColBuilder) {
					c.Text("overlay")
				})
			})
		})
		if err != nil {
			t.Fatalf("overlay: %v", err)
		}
		out, err := doc.Save()
		if err != nil {
			t.Fatalf("save: %v", err)
		}
		if err := pdfcpuapi.Validate(bytes.NewReader(out), nil); err != nil {
			t.Errorf("pdfcpu validate overlay: %v", err)
		}
	})
}
