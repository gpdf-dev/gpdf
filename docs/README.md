# gpdf Documentation

**gpdf** is a pure Go, zero-dependency PDF generation library. It provides a layered architecture for creating PDFs programmatically with first-class CJK (Japanese/Chinese/Korean) support.

## Key Features

- **Zero dependencies** -- stdlib only, no external packages
- **Three ways to create PDFs** -- Go Builder API, JSON Schema, or Go Templates
- **12-column grid system** -- Bootstrap-style responsive layouts
- **Pre-built components** -- Invoice, Report, and Letter templates
- **CJK first-class support** -- Japanese line-breaking rules (kinsoku), character-level wrapping
- **TrueType font embedding** -- with automatic subsetting
- **High performance** -- 10-30x faster than alternatives

## Documentation Index

| # | Document | Description |
|---|---|---|
| 01 | [Getting Started](01-getting-started.md) | Installation, first PDF, basic concepts |
| 02 | [Builder API](02-builder-api.md) | Go Builder pattern (Document / Page / Row / Col) |
| 03 | [JSON Schema](03-json-schema.md) | Declarative PDF definition with JSON |
| 04 | [Go Templates](04-go-templates.md) | Data-driven PDFs using `text/template` |
| 05 | [Elements](05-elements.md) | Text, Image, Table, List, QR Code, Barcode, Absolute Positioning, and more |
| 06 | [Styling](06-styling.md) | Colors, fonts, alignment, decoration, units |
| 07 | [Layout](07-layout.md) | Grid system, page sizes, margins, headers/footers, pagination |
| 08 | [Components](08-components.md) | Pre-built Invoice, Report, and Letter |
| 09 | [Fonts](09-fonts.md) | TrueType font registration, CJK support, built-in fonts |
| 10 | [Architecture](10-architecture.md) | Three-layer architecture and internal design |

## Quick Example

```go
package main

import (
    "os"

    gpdf "github.com/gpdf-dev/gpdf"
    "github.com/gpdf-dev/gpdf/document"
    "github.com/gpdf-dev/gpdf/template"
)

func main() {
    doc := gpdf.NewDocument(
        gpdf.WithPageSize(document.A4),
        gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
    )

    page := doc.AddPage()
    page.AutoRow(func(r *template.RowBuilder) {
        r.Col(12, func(c *template.ColBuilder) {
            c.Text("Hello, World!", template.FontSize(24), template.Bold())
        })
    })

    data, err := doc.Generate()
    if err != nil {
        panic(err)
    }
    os.WriteFile("hello.pdf", data, 0644)
}
```

## Module Path

```
github.com/gpdf-dev/gpdf
```

## License

MIT
