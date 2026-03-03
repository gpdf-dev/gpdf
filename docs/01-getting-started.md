# Getting Started

## Installation

```bash
go get github.com/gpdf-dev/gpdf
```

gpdf requires Go 1.22 or later. It has **zero external dependencies** -- only the Go standard library is used.

## Your First PDF

```go
package main

import (
    "os"

    gpdf "github.com/gpdf-dev/gpdf"
    "github.com/gpdf-dev/gpdf/document"
    "github.com/gpdf-dev/gpdf/template"
)

func main() {
    // 1. Create a document with options
    doc := gpdf.NewDocument(
        gpdf.WithPageSize(document.A4),
        gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
    )

    // 2. Add a page
    page := doc.AddPage()

    // 3. Add content using the grid system
    page.AutoRow(func(r *template.RowBuilder) {
        r.Col(12, func(c *template.ColBuilder) {
            c.Text("Hello, World!", template.FontSize(24), template.Bold())
        })
    })

    // 4. Generate PDF bytes
    data, err := doc.Generate()
    if err != nil {
        panic(err)
    }

    // 5. Write to file
    os.WriteFile("hello.pdf", data, 0644)
}
```

## Core Concepts

### Three Ways to Create PDFs

gpdf offers three approaches to PDF creation. Choose the one that best fits your use case:

| Approach | Best For | Import |
|---|---|---|
| **Builder API** | Full programmatic control in Go code | `template` package |
| **JSON Schema** | Declarative definitions, stored as data | `template.FromJSON` |
| **Go Templates** | Data-driven documents with template logic | `template.FromTemplate` |

All three approaches produce the same output -- they build the same internal document model.

### Document Structure

Every gpdf document follows this hierarchy:

```
Document
  -> Page(s)
       -> Row(s)
            -> Col(s)      (12-column grid)
                 -> Element(s)  (Text, Image, Table, etc.)
```

### Grid System

gpdf uses a **12-column grid** (similar to Bootstrap). Each row is divided into columns whose spans must sum to 12:

```go
page.AutoRow(func(r *template.RowBuilder) {
    r.Col(6, func(c *template.ColBuilder) {
        c.Text("Left half")
    })
    r.Col(6, func(c *template.ColBuilder) {
        c.Text("Right half")
    })
})
```

Common column layouts:
- `12` -- full width (single column)
- `6 + 6` -- two equal columns
- `4 + 4 + 4` -- three equal columns
- `3 + 3 + 3 + 3` -- four equal columns
- `8 + 4` -- wide left, narrow right

### Output Methods

```go
// Option A: Get PDF as byte slice
data, err := doc.Generate()

// Option B: Write PDF to any io.Writer (file, HTTP response, buffer, etc.)
err := doc.Render(writer)
```

## Using the Facade (`gpdf` package)

The top-level `gpdf` package re-exports the most commonly used functions for convenience:

```go
import gpdf "github.com/gpdf-dev/gpdf"

// These are equivalent:
doc := gpdf.NewDocument(gpdf.WithPageSize(gpdf.A4))
doc := template.New(template.WithPageSize(document.A4))
```

Available re-exports:

| Facade | Original |
|---|---|
| `gpdf.NewDocument` | `template.New` |
| `gpdf.WithPageSize` | `template.WithPageSize` |
| `gpdf.WithMargins` | `template.WithMargins` |
| `gpdf.WithFont` | `template.WithFont` |
| `gpdf.WithDefaultFont` | `template.WithDefaultFont` |
| `gpdf.WithMetadata` | `template.WithMetadata` |
| `gpdf.A4`, `A3`, `Letter`, `Legal` | `document.A4`, etc. |
| `gpdf.FromJSON` | `template.FromJSON` |
| `gpdf.FromTemplate` | `template.FromTemplate` |
| `gpdf.TemplateFuncMap` | `template.TemplateFuncMap` |
| `gpdf.NewInvoice` | `template.Invoice` |
| `gpdf.NewReport` | `template.Report` |
| `gpdf.NewLetter` | `template.Letter` |
| `gpdf.QRSize`, `QRErrorCorrection`, `QRScale` | `template.QR*` |
| `gpdf.BarcodeWidth`, `BarcodeHeight`, `BarcodeFormat` | `template.Barcode*` |

## Running Tests

```bash
cd gpdf

# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run example tests only
go test -run TestExample ./...

# Run benchmarks
cd _benchmark && go test -bench=. -benchmem
```

## Next Steps

- [Builder API](02-builder-api.md) -- Learn the full programmatic API
- [JSON Schema](03-json-schema.md) -- Define PDFs declaratively as JSON
- [Elements](05-elements.md) -- All available content elements
- [Styling](06-styling.md) -- Colors, fonts, and text decoration
