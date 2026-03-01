# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
![coverage](https://img.shields.io/badge/coverage-87.9%25-green)
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
- **Tables** — headers, column widths, striped rows
- **Headers & Footers** — consistent across all pages
- **Multiple units** — pt, mm, cm, in, em, %
- **Color spaces** — RGB, Grayscale, CMYK
- **Images** — JPEG and PNG embedding with fit options
- **Document metadata** — title, author, subject, creator

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

### Column Content

| Method | Description |
|---|---|
| `c.Text(text, opts...)` | Add text with styling options |
| `c.Table(header, rows, opts...)` | Add a table |
| `c.Image(data, opts...)` | Add an image (JPEG/PNG) |
| `c.Line(opts...)` | Add a horizontal line |
| `c.Spacer(height)` | Add vertical space |

### Text Options

| Option | Description |
|---|---|
| `template.FontSize(size)` | Set font size in points |
| `template.Bold()` | Bold weight |
| `template.Italic()` | Italic style |
| `template.FontFamily(name)` | Use a registered font |
| `template.TextColor(color)` | Set text color |
| `template.BgColor(color)` | Set background color |
| `template.AlignLeft()` | Left align (default) |
| `template.AlignCenter()` | Center align |
| `template.AlignRight()` | Right align |

### Table Options

| Option | Description |
|---|---|
| `template.ColumnWidths(w...)` | Set column width percentages |
| `template.TableHeaderStyle(opts...)` | Style the header row |
| `template.TableStripe(color)` | Set alternating row color |

### Image Options

| Option | Description |
|---|---|
| `template.FitWidth(value)` | Scale to fit width (keeps aspect ratio) |
| `template.FitHeight(value)` | Scale to fit height (keeps aspect ratio) |

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

## License

MIT
