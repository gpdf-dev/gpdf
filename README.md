# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
![coverage](https://img.shields.io/badge/coverage-85.3%25-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)

**English** | [日本語](README_ja.md) | [中文](README_zh.md) | [한국어](README_ko.md) | [Español](README_es.md) | [Português](README_pt.md)

A pure Go, zero-dependency PDF generation library with a layered architecture and declarative builder API.

## Features

- **Zero dependencies** — only the Go standard library
- **Layered architecture** — low-level PDF primitives, document model, and high-level template API
- **12-column grid system** — Bootstrap-style responsive layout
- **TrueType font support** — embed custom fonts with subsetting
- **CJK ready** — full CJK text support from day one
- **Tables** — headers, column widths, striped rows, vertical alignment
- **Headers & Footers** — consistent across all pages with page numbers
- **Lists** — bulleted and numbered lists
- **QR codes** — pure Go QR code generation with error correction levels
- **Barcodes** — Code 128 barcode generation
- **Text decorations** — underline, strikethrough, letter spacing, text indent
- **Page numbers** — automatic page number and total page count
- **Go template integration** — generate PDFs from Go templates
- **Reusable components** — pre-built Invoice, Report, and Letter templates
- **JSON schema** — define documents entirely in JSON
- **Multiple units** — pt, mm, cm, in, em, %
- **Color spaces** — RGB, Grayscale, CMYK
- **Images** — JPEG and PNG embedding with fit options
- **Absolute positioning** — place elements at exact XY coordinates on the page
- **Existing PDF overlay** — open existing PDFs and add text, images, stamps on top
- **Document metadata** — title, author, subject, creator
- **Encryption** — AES-256 encryption (ISO 32000-2, Rev 6) with owner/user passwords and permissions
- **PDF/A** — PDF/A-1b and PDF/A-2b conformance with ICC profiles and XMP metadata
- **Digital signatures** — CMS/PKCS#7 signatures with RSA/ECDSA keys and optional RFC 3161 timestamping

## Benchmark

Comparison with [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [signintech/gopdf](https://github.com/signintech/gopdf), and [maroto v2](https://github.com/johnfercher/maroto).
Median of 5 runs, 100 iterations each. Apple M1, Go 1.25.

**Execution time** (lower is better):

| Benchmark | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| Single page | **13 µs** | 132 µs | 423 µs | 237 µs |
| Table (4x10) | **108 µs** | 241 µs | 835 µs | 8.6 ms |
| 100 pages | **683 µs** | 11.7 ms | 8.6 ms | 19.8 ms |
| Complex document | **133 µs** | 254 µs | 997 µs | 10.4 ms |

**Memory usage** (lower is better):

| Benchmark | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| Single page | **16 KB** | 1.2 MB | 1.8 MB | 61 KB |
| Table (4x10) | **209 KB** | 1.3 MB | 1.9 MB | 1.6 MB |
| 100 pages | **909 KB** | 121 MB | 83 MB | 4.0 MB |
| Complex document | **246 KB** | 1.3 MB | 2.0 MB | 2.0 MB |

### Why is gpdf fast?

- **Single page** — Single-pass build→layout→render pipeline with no intermediate data structures. Concrete struct types throughout (no `interface{}` boxing), so the document tree is built with minimal heap allocations.
- **Table** — Cell content is written directly as PDF content stream commands via a reusable `strings.Builder` buffer. No per-cell object wrapping or repeated font lookups; the font is resolved once per document.
- **100 pages** — Layout scales linearly O(n). Overflow pagination passes remaining nodes by slice reference (no deep copies). The font is parsed once and shared across all pages.
- **Complex document** — Single-pass layout without re-measurement combines all the above. Font subsetting embeds only the glyphs actually used, and Flate compression is applied by default, keeping both memory and output size small.

Run benchmarks yourself:

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## Architecture

```
┌─────────────────────────────────────┐
│  gpdf (entry point)                 │
├─────────────────────────────────────┤
│  template  — Builder API, Grid      │  Layer 3
├─────────────────────────────────────┤
│  document  — Nodes, Style, Layout   │  Layer 2
├─────────────────────────────────────┤
│  pdf       — Writer, Fonts, Streams │  Layer 1
└─────────────────────────────────────┘
```

## Requirements

- Go 1.22 or later

## Install

```bash
go get github.com/gpdf-dev/gpdf
```

## Quick Start

```go
package main

import (
	"os"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func main() {
	doc := gpdf.NewDocument(
		gpdf.WithPageSize(gpdf.A4),
		gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Hello, World!", template.FontSize(24), template.Bold())
		})
	})

	data, _ := doc.Generate()
	os.WriteFile("hello.pdf", data, 0644)
}
```

## Examples

### Text Styling

Font size, weight, style, color, background color, and alignment:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("Large bold title", template.FontSize(24), template.Bold())
		c.Text("Italic text", template.Italic())
		c.Text("Bold + Italic", template.Bold(), template.Italic())
		c.Text("Red text", template.TextColor(pdf.Red))
		c.Text("Custom color", template.TextColor(pdf.RGBHex(0x336699)))
		c.Text("With background", template.BgColor(pdf.Yellow))
		c.Text("Centered", template.AlignCenter())
		c.Text("Right aligned", template.AlignRight())
	})
})
```

### CJK Fonts (Japanese / Chinese / Korean)

Embed TrueType fonts for CJK text rendering. Each language needs its own Noto Sans font:

```go
fontData, _ := os.ReadFile("NotoSansJP-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithFont("NotoSansJP", fontData),
	gpdf.WithDefaultFont("NotoSansJP", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("こんにちは世界", template.FontSize(18))
	})
})
```

For multi-language documents, register multiple fonts and switch with `FontFamily()`:

```go
jpFont, _ := os.ReadFile("NotoSansJP-Regular.ttf")
scFont, _ := os.ReadFile("NotoSansSC-Regular.ttf")
krFont, _ := os.ReadFile("NotoSansKR-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithFont("NotoSansJP", jpFont),
	gpdf.WithFont("NotoSansSC", scFont),
	gpdf.WithFont("NotoSansKR", krFont),
	gpdf.WithDefaultFont("NotoSansJP", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(4, func(c *template.ColBuilder) {
		c.Text("日本語", template.FontFamily("NotoSansJP"))
	})
	r.Col(4, func(c *template.ColBuilder) {
		c.Text("中文", template.FontFamily("NotoSansSC"))
	})
	r.Col(4, func(c *template.ColBuilder) {
		c.Text("한국어", template.FontFamily("NotoSansKR"))
	})
})
```

Recommended fonts (all free, OFL license):

| Font | Language |
|---|---|
| [Noto Sans JP](https://fonts.google.com/noto/specimen/Noto+Sans+JP) | Japanese |
| [Noto Sans SC](https://fonts.google.com/noto/specimen/Noto+Sans+SC) | Simplified Chinese |
| [Noto Sans KR](https://fonts.google.com/noto/specimen/Noto+Sans+KR) | Korean |

### 12-Column Grid Layout

Build layouts using a Bootstrap-style 12-column grid:

```go
// Two equal columns
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Left half")
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Right half")
	})
})

// Sidebar + main content
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) {
		c.Text("Sidebar")
	})
	r.Col(9, func(c *template.ColBuilder) {
		c.Text("Main content")
	})
})

// Four equal columns
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) { c.Text("Col 1") })
	r.Col(3, func(c *template.ColBuilder) { c.Text("Col 2") })
	r.Col(3, func(c *template.ColBuilder) { c.Text("Col 3") })
	r.Col(3, func(c *template.ColBuilder) { c.Text("Col 4") })
})
```

### Fixed-Height Rows

Use `Row()` with a specific height, or `AutoRow()` for content-based height:

```go
// Fixed height: 30mm
page.Row(document.Mm(30), func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("This row is 30mm tall")
	})
})

// Auto height (fits content)
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("Height adjusts to content")
	})
})
```

### Tables

Basic table:

```go
c.Table(
	[]string{"Name", "Age", "City"},
	[][]string{
		{"Alice", "30", "Tokyo"},
		{"Bob", "25", "New York"},
		{"Charlie", "35", "London"},
	},
)
```

Styled table with header colors, column widths, and stripe rows:

```go
c.Table(
	[]string{"Product", "Category", "Qty", "Unit Price", "Total"},
	[][]string{
		{"Laptop Pro 15", "Electronics", "2", "$1,299.00", "$2,598.00"},
		{"Wireless Mouse", "Accessories", "10", "$29.99", "$299.90"},
		{"USB-C Hub", "Accessories", "5", "$49.99", "$249.95"},
	},
	template.ColumnWidths(30, 20, 10, 20, 20),
	template.TableHeaderStyle(
		template.TextColor(pdf.White),
		template.BgColor(pdf.RGBHex(0x1A237E)),
	),
	template.TableStripe(pdf.RGBHex(0xF5F5F5)),
)
```

### Images

Embed JPEG and PNG images with optional fit options:

```go
// Default size (original dimensions)
c.Image(imgData)

// Fit to specific width (maintains aspect ratio)
c.Image(imgData, template.FitWidth(document.Mm(80)))

// Fit to specific height (maintains aspect ratio)
c.Image(imgData, template.FitHeight(document.Mm(30)))
```

Images in grid columns:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Image 1")
		c.Image(pngData)
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Image 2")
		c.Image(jpegData)
	})
})
```

### Lines & Spacers

Horizontal rules with color and thickness:

```go
c.Line()                                           // Default (gray, 1pt)
c.Line(template.LineColor(pdf.Red))                 // Colored
c.Line(template.LineThickness(document.Pt(3)))      // Thick
c.Line(template.LineColor(pdf.Blue),                // Combined
	template.LineThickness(document.Pt(2)))
```

Vertical spacing:

```go
c.Spacer(document.Mm(5))   // 5mm gap
c.Spacer(document.Mm(15))  // 15mm gap
```

### Lists

Bulleted and numbered lists:

```go
// Bulleted list
c.List([]string{"First item", "Second item", "Third item"})

// Numbered list
c.OrderedList([]string{"Step one", "Step two", "Step three"})
```

### QR Codes

Generate QR codes with configurable size and error correction:

```go
// Basic QR code
c.QRCode("https://gpdf.dev")

// Custom size and error correction level
c.QRCode("https://gpdf.dev",
	template.QRSize(document.Mm(30)),
	template.QRErrorCorrection(qrcode.LevelH))
```

### Barcodes

Generate Code 128 barcodes:

```go
// Basic barcode
c.Barcode("INV-2026-0001")

// Custom width
c.Barcode("INV-2026-0001", template.BarcodeWidth(document.Mm(80)))
```

### Page Numbers

Automatic page numbers and total page count:

```go
doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Generated by gpdf", template.FontSize(8))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.PageNumber(template.AlignRight(), template.FontSize(8))
		})
	})
})

doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.TotalPages(template.AlignRight(), template.FontSize(9))
		})
	})
})
```

### Text Decorations

Underline, strikethrough, letter spacing, and text indent:

```go
c.Text("Underlined text", template.Underline())
c.Text("Strikethrough text", template.Strikethrough())
c.Text("Wide spacing", template.LetterSpacing(3))
c.Text("Indented paragraph...", template.TextIndent(document.Pt(24)))
```

### Headers & Footers

Define headers and footers that repeat on every page:

```go
doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ACME Corporation", template.Bold(), template.FontSize(10))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Confidential", template.AlignRight(), template.FontSize(10),
				template.TextColor(pdf.Gray(0.5)))
		})
	})
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Line(template.LineColor(pdf.RGBHex(0x1565C0)))
			c.Spacer(document.Mm(5))
		})
	})
})

doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
			c.Line(template.LineColor(pdf.Gray(0.7)))
			c.Spacer(document.Mm(2))
		})
	})
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Generated by gpdf", template.AlignCenter(),
				template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
		})
	})
})
```

### Multi-Page Documents

```go
for i := 1; i <= 5; i++ {
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Page content here")
		})
	})
}
```

### Existing PDF Overlay

Open an existing PDF and overlay content using the same builder API:

```go
// Open an existing PDF
doc, err := gpdf.Open(existingPDFBytes)

// Add a "DRAFT" watermark on page 1
doc.Overlay(0, func(p *template.PageBuilder) {
	p.Absolute(document.Mm(50), document.Mm(140), func(c *template.ColBuilder) {
		c.Text("DRAFT", template.FontSize(72),
			template.TextColor(pdf.Gray(0.85)))
	})
})

// Add page numbers to every page
count, _ := doc.PageCount()
doc.EachPage(func(i int, p *template.PageBuilder) {
	p.Absolute(document.Mm(170), document.Mm(285), func(c *template.ColBuilder) {
		c.Text(fmt.Sprintf("%d / %d", i+1, count), template.FontSize(10))
	}, template.AbsoluteWidth(document.Mm(20)))
})

result, _ := doc.Save()
```

### JSON Schema

Define documents entirely in JSON:

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "Report", "author": "gpdf"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "Hello from JSON", "style": {"size": 24, "bold": true}}
		]}},
		{"row": {"cols": [
			{"span": 12, "table": {
				"header": ["Name", "Value"],
				"rows": [["Alpha", "100"], ["Beta", "200"]],
				"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"}
			}}
		]}}
	]
}`)

doc, err := template.FromJSON(schema, nil)
data, _ := doc.Generate()
```

### Go Template Integration

Use Go templates with JSON schema for dynamic content:

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "{{.Title}}"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
		]}}
	]
}`)

data := map[string]any{"Title": "Dynamic Report"}
doc, err := template.FromJSON(schema, data)
```

For more control, use a pre-parsed Go template:

```go
tmpl, _ := gotemplate.New("doc").Funcs(template.TemplateFuncMap()).Parse(schemaStr)
doc, err := template.FromTemplate(tmpl, data)
```

### Reusable Components

Generate common document types with a single function call:

**Invoice:**

```go
doc := template.Invoice(template.InvoiceData{
	Number:  "#INV-2026-001",
	Date:    "March 1, 2026",
	DueDate: "March 31, 2026",
	From:    template.InvoiceParty{Name: "ACME Corp", Address: []string{"123 Main St"}},
	To:      template.InvoiceParty{Name: "Client Inc.", Address: []string{"456 Side St"}},
	Items: []template.InvoiceItem{
		{Description: "Web Development", Quantity: "40 hrs", UnitPrice: 150, Amount: 6000},
		{Description: "UI/UX Design", Quantity: "20 hrs", UnitPrice: 120, Amount: 2400},
	},
	TaxRate: 10,
	Notes:   "Thank you for your business!",
})
data, _ := doc.Generate()
```

**Report:**

```go
doc := template.Report(template.ReportData{
	Title:    "Quarterly Report",
	Subtitle: "Q1 2026",
	Author:   "ACME Corp",
	Sections: []template.ReportSection{
		{
			Title:   "Executive Summary",
			Content: "Revenue increased by 15% compared to Q4 2025.",
			Metrics: []template.ReportMetric{
				{Label: "Revenue", Value: "$12.5M", ColorHex: 0x2E7D32},
				{Label: "Growth", Value: "+15%", ColorHex: 0x2E7D32},
			},
		},
		{
			Title: "Revenue Breakdown",
			Table: &template.ReportTable{
				Header: []string{"Division", "Q1 2026", "Change"},
				Rows:   [][]string{{"Cloud", "$5.2M", "+26.8%"}, {"Enterprise", "$3.8M", "+8.6%"}},
			},
		},
	},
})
```

**Letter:**

```go
doc := template.Letter(template.LetterData{
	From:     template.LetterParty{Name: "ACME Corp", Address: []string{"123 Main St"}},
	To:       template.LetterParty{Name: "Mr. John Smith", Address: []string{"456 Side St"}},
	Date:     "March 1, 2026",
	Subject:  "Partnership Proposal",
	Greeting: "Dear Mr. Smith,",
	Body:     []string{"We are writing to propose a strategic partnership..."},
	Closing:  "Sincerely,",
	Signature: "Jane Doe",
})
```

### Encryption

AES-256 encryption with owner/user passwords and permission control:

```go
// Owner password only (PDF opens without password, editing restricted)
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithEncryption(
		encrypt.WithOwnerPassword("owner-secret"),
	),
)

// Both passwords with permission control
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithEncryption(
		encrypt.WithOwnerPassword("owner-pass"),
		encrypt.WithUserPassword("user-pass"),
		encrypt.WithPermissions(encrypt.PermPrint|encrypt.PermCopy),
	),
)
```

### PDF/A Conformance

Generate PDF/A-1b or PDF/A-2b compliant documents:

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithPDFA(
		pdfa.WithLevel(pdfa.LevelA2b),
		pdfa.WithMetadata(pdfa.MetadataInfo{
			Title:  "Archived Report",
			Author: "gpdf",
		}),
	),
)
```

### Digital Signatures

Sign PDFs with CMS/PKCS#7 using RSA or ECDSA keys:

```go
data, _ := doc.Generate()

signed, err := gpdf.SignDocument(data, signature.Signer{
	Certificate: cert,
	PrivateKey:  key,
	Chain:       intermediates,
},
	signature.WithReason("Approved"),
	signature.WithLocation("Tokyo"),
	signature.WithTimestamp("http://tsa.example.com"),
)
```

### Document Metadata

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithMetadata(document.DocumentMetadata{
		Title:   "Annual Report 2026",
		Author:  "gpdf Library",
		Subject: "Example of document metadata",
		Creator: "My Application",
	}),
)
```

### Page Sizes & Margins

```go
// Available page sizes
document.A4      // 210mm x 297mm
document.A3      // 297mm x 420mm
document.Letter  // 8.5in x 11in
document.Legal   // 8.5in x 14in

// Uniform margins
template.WithMargins(document.UniformEdges(document.Mm(20)))

// Asymmetric margins
template.WithMargins(document.Edges{
	Top:    document.Mm(10),
	Right:  document.Mm(40),
	Bottom: document.Mm(10),
	Left:   document.Mm(40),
})
```

### Output Options

```go
// Generate returns []byte
data, err := doc.Generate()

// Render writes to any io.Writer
var buf bytes.Buffer
err := doc.Render(&buf)

// Write directly to a file
f, _ := os.Create("output.pdf")
defer f.Close()
doc.Render(f)
```

## API Reference

### Document Options

| Function | Description |
|---|---|
| `WithPageSize(size)` | Set page size (A4, A3, Letter, Legal) |
| `WithMargins(edges)` | Set page margins |
| `WithFont(family, data)` | Register a TrueType font |
| `WithDefaultFont(family, size)` | Set the default font |
| `WithMetadata(meta)` | Set document metadata |
| `WithEncryption(opts...)` | Enable AES-256 encryption |
| `WithPDFA(opts...)` | Enable PDF/A conformance |

### Column Content

| Method | Description |
|---|---|
| `c.Text(text, opts...)` | Add text with styling options |
| `c.Table(header, rows, opts...)` | Add a table |
| `c.Image(data, opts...)` | Add an image (JPEG/PNG) |
| `c.QRCode(data, opts...)` | Add a QR code |
| `c.Barcode(data, opts...)` | Add a barcode (Code 128) |
| `c.List(items, opts...)` | Add a bulleted list |
| `c.OrderedList(items, opts...)` | Add a numbered list |
| `c.PageNumber(opts...)` | Add current page number |
| `c.TotalPages(opts...)` | Add total page count |
| `c.Line(opts...)` | Add a horizontal line |
| `c.Spacer(height)` | Add vertical space |

### Page-Level Content

| Method | Description |
|---|---|
| `page.AutoRow(fn)` | Add an auto-height row |
| `page.Row(height, fn)` | Add a fixed-height row |
| `page.Absolute(x, y, fn, opts...)` | Place content at exact XY coordinates |

#### Absolute Positioning Options

| Option | Description |
|---|---|
| `gpdf.AbsoluteWidth(value)` | Set explicit width (default: remaining space) |
| `gpdf.AbsoluteHeight(value)` | Set explicit height (default: remaining space) |
| `gpdf.AbsoluteOriginPage()` | Use page corner as origin instead of content area |

### Existing PDF Operations

| Function / Method | Description |
|---|---|
| `gpdf.Open(data, opts...)` | Open an existing PDF for overlay |
| `doc.PageCount()` | Get the number of pages |
| `doc.Overlay(page, fn)` | Add content on top of a specific page |
| `doc.EachPage(fn)` | Apply overlay to every page |
| `doc.Save()` | Save the modified PDF |

### Text Options

| Option | Description |
|---|---|
| `template.FontSize(size)` | Set font size in points |
| `template.Bold()` | Bold weight |
| `template.Italic()` | Italic style |
| `template.FontFamily(name)` | Use a registered font |
| `template.TextColor(color)` | Set text color |
| `template.BgColor(color)` | Set background color |
| `template.Underline()` | Underline decoration |
| `template.Strikethrough()` | Strikethrough decoration |
| `template.LetterSpacing(pts)` | Set letter spacing in points |
| `template.TextIndent(value)` | Set first-line indent |
| `template.AlignLeft()` | Left align (default) |
| `template.AlignCenter()` | Center align |
| `template.AlignRight()` | Right align |

### Table Options

| Option | Description |
|---|---|
| `template.ColumnWidths(w...)` | Set column width percentages |
| `template.TableHeaderStyle(opts...)` | Style the header row |
| `template.TableStripe(color)` | Set alternating row color |
| `template.TableCellVAlign(align)` | Set cell vertical alignment (Top/Middle/Bottom) |

### Image Options

| Option | Description |
|---|---|
| `template.FitWidth(value)` | Scale to fit width (keeps aspect ratio) |
| `template.FitHeight(value)` | Scale to fit height (keeps aspect ratio) |

### QR Code Options

| Option | Description |
|---|---|
| `template.QRSize(value)` | Set QR code size |
| `template.QRErrorCorrection(level)` | Set error correction (L/M/Q/H) |
| `template.QRScale(n)` | Set module scale factor |

### Barcode Options

| Option | Description |
|---|---|
| `template.BarcodeWidth(value)` | Set barcode width |
| `template.BarcodeHeight(value)` | Set barcode height |
| `template.BarcodeFormat(fmt)` | Set barcode format (Code 128) |

### Encryption Options

| Option | Description |
|---|---|
| `encrypt.WithOwnerPassword(pw)` | Set owner password |
| `encrypt.WithUserPassword(pw)` | Set user password |
| `encrypt.WithPermissions(perm)` | Set document permissions (PermPrint, PermCopy, PermModify, etc.) |

### PDF/A Options

| Option | Description |
|---|---|
| `pdfa.WithLevel(level)` | Set conformance level (LevelA1b, LevelA2b) |
| `pdfa.WithMetadata(info)` | Set XMP metadata (Title, Author, Subject, etc.) |

### Digital Signature

| Function / Option | Description |
|---|---|
| `gpdf.SignDocument(data, signer, opts...)` | Sign a PDF with a digital signature |
| `signature.WithReason(reason)` | Set signing reason |
| `signature.WithLocation(location)` | Set signing location |
| `signature.WithTimestamp(tsaURL)` | Enable RFC 3161 timestamping |
| `signature.WithSignTime(t)` | Set signing time |

### Template Generation

| Function | Description |
|---|---|
| `template.FromJSON(schema, data)` | Generate document from JSON schema |
| `template.FromTemplate(tmpl, data)` | Generate document from Go template |
| `template.TemplateFuncMap()` | Get template helper functions (includes `toJSON`) |

### Reusable Components

| Function | Description |
|---|---|
| `template.Invoice(data)` | Generate a professional invoice PDF |
| `template.Report(data)` | Generate a structured report PDF |
| `template.Letter(data)` | Generate a business letter PDF |

### Line Options

| Option | Description |
|---|---|
| `template.LineColor(color)` | Set line color |
| `template.LineThickness(value)` | Set line thickness |

### Units

```go
document.Pt(72)    // Points (1/72 inch)
document.Mm(10)    // Millimeters
document.Cm(2.5)   // Centimeters
document.In(1)     // Inches
document.Em(1.5)   // Relative to font size
document.Pct(50)   // Percentage
```

### Colors

```go
pdf.RGB(0.2, 0.4, 0.8)   // RGB (0.0–1.0)
pdf.RGBHex(0xFF5733)      // RGB from hex
pdf.Gray(0.5)             // Grayscale
pdf.CMYK(0, 0.5, 1, 0)   // CMYK

// Predefined
pdf.Black, pdf.White, pdf.Red, pdf.Green, pdf.Blue
pdf.Yellow, pdf.Cyan, pdf.Magenta
```

## License

MIT
