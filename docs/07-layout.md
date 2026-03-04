# Layout

gpdf provides a 12-column grid system, automatic pagination, and header/footer support for consistent document layouts.

## Page Sizes

### Predefined Sizes

| Name | Dimensions | Points |
|---|---|---|
| `document.A4` | 210mm x 297mm | 595.28 x 841.89 |
| `document.A3` | 297mm x 420mm | 841.89 x 1190.55 |
| `document.Letter` | 8.5" x 11" | 612 x 792 |
| `document.Legal` | 8.5" x 14" | 612 x 1008 |

### Setting Page Size

```go
// Builder API
doc := template.New(template.WithPageSize(document.A4))
doc := template.New(template.WithPageSize(document.Letter))
```

```json
// JSON Schema
"page": {"size": "A4"}
"page": {"size": "Letter"}
```

### Custom Page Size

Create a custom `document.Size` value. Dimensions are in PDF points (1pt = 1/72 inch):

```go
// B5 (176mm x 250mm)
b5 := document.Size{Width: 498.90, Height: 708.66}
doc := template.New(template.WithPageSize(b5))
```

## Margins

Margins define the space between the page edge and the content area.

### Uniform Margins

```go
// Same margin on all four sides
template.WithMargins(document.UniformEdges(document.Mm(20)))
```

### Custom Margins

```go
// Different margins per side
template.WithMargins(document.Edges{
    Top:    document.Mm(25),
    Right:  document.Mm(20),
    Bottom: document.Mm(25),
    Left:   document.Mm(20),
})
```

### JSON Schema

```json
"page": {"size": "A4", "margins": "15mm"}
```

> Note: JSON schema only supports uniform margins via the `margins` string. For asymmetric margins, use Go options with `FromJSON`.

## 12-Column Grid

Every row is divided into a 12-column grid, similar to Bootstrap's grid system. Columns are defined by their `span` (1-12), where the span represents the fraction of the total row width.

### Basic Layouts

```go
// Full width
page.AutoRow(func(r *template.RowBuilder) {
    r.Col(12, func(c *template.ColBuilder) {
        c.Text("Full width content")
    })
})

// Two equal columns
page.AutoRow(func(r *template.RowBuilder) {
    r.Col(6, func(c *template.ColBuilder) {
        c.Text("Left half")
    })
    r.Col(6, func(c *template.ColBuilder) {
        c.Text("Right half")
    })
})

// Three equal columns
page.AutoRow(func(r *template.RowBuilder) {
    r.Col(4, func(c *template.ColBuilder) { c.Text("Col 1") })
    r.Col(4, func(c *template.ColBuilder) { c.Text("Col 2") })
    r.Col(4, func(c *template.ColBuilder) { c.Text("Col 3") })
})

// Sidebar layout
page.AutoRow(func(r *template.RowBuilder) {
    r.Col(8, func(c *template.ColBuilder) {
        c.Text("Main content area")
    })
    r.Col(4, func(c *template.ColBuilder) {
        c.Text("Sidebar")
    })
})
```

### Common Span Combinations

| Layout | Spans |
|---|---|
| Full width | `12` |
| Two equal | `6 + 6` |
| Three equal | `4 + 4 + 4` |
| Four equal | `3 + 3 + 3 + 3` |
| Wide + narrow | `8 + 4` or `9 + 3` |
| Three unequal | `2 + 8 + 2` |

## Row Types

### Auto-Height Rows

Height is determined by the content. Most rows use this:

```go
page.AutoRow(func(r *template.RowBuilder) {
    // Row height expands to fit the tallest column
})
```

### Fixed-Height Rows

Rows with an explicit height:

```go
page.Row(document.Mm(30), func(r *template.RowBuilder) {
    // Row is exactly 30mm tall
})
```

## Headers and Footers

Headers and footers are defined once and automatically repeated on every page, including pages created by pagination overflow.

### Builder API

```go
doc.Header(func(p *template.PageBuilder) {
    p.AutoRow(func(r *template.RowBuilder) {
        r.Col(6, func(c *template.ColBuilder) {
            c.Text("Company Name", template.Bold(), template.FontSize(10))
        })
        r.Col(6, func(c *template.ColBuilder) {
            c.Text("Report Title", template.AlignRight(), template.FontSize(10),
                template.TextColor(pdf.Gray(0.5)))
        })
    })
    // Separator line
    p.AutoRow(func(r *template.RowBuilder) {
        r.Col(12, func(c *template.ColBuilder) {
            c.Line(template.LineColor(pdf.RGBHex(0x1565C0)),
                template.LineThickness(document.Pt(2)))
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
        r.Col(6, func(c *template.ColBuilder) {
            c.Text("Confidential", template.FontSize(8),
                template.TextColor(pdf.Gray(0.5)))
        })
        r.Col(6, func(c *template.ColBuilder) {
            c.PageNumber(template.AlignRight(), template.FontSize(8),
                template.TextColor(pdf.Gray(0.5)))
        })
    })
})
```

### JSON Schema

```json
{
    "header": [
        {"row": {"cols": [
            {"span": 6, "text": "Company", "style": {"bold": true, "size": 10}},
            {"span": 6, "text": "Report", "style": {"align": "right", "size": 10}}
        ]}}
    ],
    "footer": [
        {"row": {"cols": [
            {"span": 6, "elements": [
                {"type": "pageNumber", "style": {"size": 8, "color": "gray(0.5)"}}
            ]},
            {"span": 6, "elements": [
                {"type": "totalPages", "style": {"size": 8, "align": "right", "color": "gray(0.5)"}}
            ]}
        ]}}
    ]
}
```

### Header/Footer Layout

The space available for body content is reduced by the height of headers and footers:

```
+-------------------------------------------+
|  margin                                   |
|  +---------------------------------------+|
|  | HEADER                                ||
|  +---------------------------------------+|
|  | BODY CONTENT                          ||
|  |                                       ||
|  |                                       ||
|  +---------------------------------------+|
|  | FOOTER                                ||
|  +---------------------------------------+|
|  margin                                   |
+-------------------------------------------+
```

## Pagination

gpdf automatically paginates content that doesn't fit on a single page. When body content overflows:

1. A new page is created automatically
2. Headers and footers are repeated on the new page
3. Content continues from where it was cut off

### Multi-Page Documents

You can explicitly create multiple pages:

```go
// Explicit pages
for i := 0; i < 5; i++ {
    page := doc.AddPage()
    page.AutoRow(func(r *template.RowBuilder) {
        r.Col(12, func(c *template.ColBuilder) {
            c.Text(fmt.Sprintf("Page %d content", i+1))
        })
    })
}
```

Or let content overflow naturally:

```go
// Single page with lots of content — auto-paginates
page := doc.AddPage()
for i := 0; i < 100; i++ {
    page.AutoRow(func(r *template.RowBuilder) {
        r.Col(12, func(c *template.ColBuilder) {
            c.Text(fmt.Sprintf("Row %d of many", i+1))
        })
    })
}
```

### JSON: Explicit Pages vs. Single Body

```json
// Single body (auto-paginated)
{
    "page": {"size": "A4"},
    "body": [...]
}

// Explicit pages
{
    "page": {"size": "A4"},
    "pages": [
        {"body": [...]},
        {"body": [...]}
    ]
}
```

## Page Numbers

Insert page numbers and total page counts into headers or footers:

```go
doc.Footer(func(p *template.PageBuilder) {
    p.AutoRow(func(r *template.RowBuilder) {
        r.Col(6, func(c *template.ColBuilder) {
            c.PageNumber(template.FontSize(8))  // "1", "2", "3", ...
        })
        r.Col(6, func(c *template.ColBuilder) {
            c.TotalPages(template.AlignRight(), template.FontSize(8))  // "5"
        })
    })
})
```

Page numbers are placeholder strings that are resolved after the entire document is paginated, ensuring accuracy.

## Document Metadata

PDF metadata is stored in the document's Info dictionary:

```go
doc := template.New(
    template.WithMetadata(document.DocumentMetadata{
        Title:   "Annual Report 2026",
        Author:  "ACME Corporation",
        Subject: "Financial Summary",
        Creator: "Report Generator v2",
    }),
)
```

```json
"metadata": {
    "title": "Annual Report 2026",
    "author": "ACME Corporation",
    "subject": "Financial Summary",
    "creator": "Report Generator v2"
}
```

The `Producer` field is automatically set to `"gpdf/{version}"` if not explicitly specified.

## See Also

- [Builder API](02-builder-api.md) -- Full programmatic API reference
- [Elements](05-elements.md) -- Content elements within columns
- [Styling](06-styling.md) -- Units, colors, and dimensions
