# Elements

Elements are the content building blocks placed inside columns. This document covers every element type available in gpdf.

## Text

The most common element. Renders a text block with optional styling.

### Builder API

```go
c.Text("Hello, World!")
c.Text("Large bold title", template.FontSize(24), template.Bold())
c.Text("Right-aligned red text", template.AlignRight(), template.TextColor(pdf.Red))
```

### JSON Schema

```json
{"span": 12, "text": "Hello, World!", "style": {"size": 24, "bold": true}}
```

Or in an elements array:

```json
{"type": "text", "content": "Hello, World!", "style": {"size": 24, "bold": true}}
```

### Text Options

| Option | Description |
|---|---|
| `FontSize(size)` | Font size in points |
| `Bold()` | Bold weight (700) |
| `Italic()` | Italic style |
| `TextColor(c)` | Foreground color |
| `BgColor(c)` | Background color |
| `AlignLeft()` | Left alignment (default) |
| `AlignCenter()` | Center alignment |
| `AlignRight()` | Right alignment |
| `FontFamily(name)` | Set font family |
| `LetterSpacing(pts)` | Extra space between characters |
| `TextIndent(v)` | First-line indentation |
| `Underline()` | Underline decoration |
| `Strikethrough()` | Strikethrough decoration |

Options can be combined freely:

```go
c.Text("Important note",
    template.FontSize(14),
    template.Bold(),
    template.Underline(),
    template.TextColor(pdf.RGBHex(0xFF0000)),
    template.BgColor(pdf.RGBHex(0xFFF3E0)),
    template.AlignCenter(),
)
```

---

## Rich Text

Renders mixed-style text in a single paragraph. Different fragments can have different fonts, sizes, colors, and decorations on the same line.

### Builder API

```go
c.RichText(func(rt *template.RichTextBuilder) {
    rt.Span("This is ")
    rt.Span("bold", template.Bold())
    rt.Span(" and this is ")
    rt.Span("red italic", template.Italic(), template.TextColor(pdf.Red))
    rt.Span(" in one line.")
})
```

The `RichTextBuilder.Span(text, opts...)` method appends a text fragment. Each fragment inherits the default style and applies its own options on top.

You can also pass text-level options to `RichText` itself to set the paragraph-level style (alignment, etc.):

```go
c.RichText(func(rt *template.RichTextBuilder) {
    rt.Span("Centered ")
    rt.Span("mixed", template.Bold())
    rt.Span(" text.")
}, template.AlignCenter())
```

---

## Image

Embeds a raster image (PNG or JPEG) in the PDF.

### Builder API

```go
imgData, _ := os.ReadFile("logo.png")

c.Image(imgData)
c.Image(imgData, template.FitWidth(document.Mm(50)))
c.Image(imgData, template.FitHeight(document.Mm(30)))
c.Image(imgData,
    template.FitWidth(document.Mm(50)),
    template.WithFitMode(document.FitContain),
    template.WithAlign(document.AlignCenter),
)
```

### JSON Schema

```json
{
    "span": 12,
    "image": {
        "src": "data:image/png;base64,iVBOR...",
        "width": "50mm",
        "fit": "contain",
        "align": "center"
    }
}
```

### Image Options

| Option | Description |
|---|---|
| `FitWidth(v)` | Fit image within the specified width |
| `FitHeight(v)` | Fit image within the specified height |
| `WithFitMode(mode)` | Set fit behavior |
| `WithAlign(align)` | Horizontal alignment within the column |

### Fit Modes

| Mode | Description |
|---|---|
| `document.FitContain` | Scale to fit within bounds, preserving aspect ratio (default) |
| `document.FitCover` | Scale to cover bounds, preserving aspect ratio (may crop) |
| `document.FitStretch` | Stretch to fill bounds exactly (may distort) |
| `document.FitOriginal` | Use original image dimensions |

### Image Source (JSON)

The `src` field supports multiple formats:

| Format | Example |
|---|---|
| Data URI | `"data:image/png;base64,iVBOR..."` |
| Raw base64 | `"iVBOR..."` |
| File path | `"./images/logo.png"` |
| Absolute path | `"/usr/share/images/logo.png"` |
| `file://` URI | `"file:///path/to/logo.png"` |

---

## Table

Renders a data table with headers, rows, and optional styling.

### Builder API

```go
c.Table(
    []string{"Name", "Age", "City"},          // header
    [][]string{                                // rows
        {"Alice", "30", "Tokyo"},
        {"Bob", "25", "New York"},
        {"Charlie", "35", "London"},
    },
)
```

### Styled Table

```go
c.Table(
    []string{"Product", "Qty", "Price"},
    [][]string{
        {"Widget", "10", "$9.99"},
        {"Gadget", "5", "$24.99"},
    },
    template.ColumnWidths(50, 25, 25),
    template.TableHeaderStyle(
        template.TextColor(pdf.White),
        template.BgColor(pdf.RGBHex(0x1A237E)),
    ),
    template.TableStripe(pdf.RGBHex(0xF5F5F5)),
    template.TableCellVAlign(document.VAlignMiddle),
)
```

### JSON Schema

```json
{
    "span": 12,
    "table": {
        "header": ["Product", "Qty", "Price"],
        "rows": [
            ["Widget", "10", "$9.99"],
            ["Gadget", "5", "$24.99"]
        ],
        "columnWidths": [50, 25, 25],
        "headerStyle": {"color": "#FFFFFF", "background": "#1A237E"},
        "stripeColor": "#F5F5F5"
    }
}
```

### Table Options

| Option | Description |
|---|---|
| `ColumnWidths(widths...)` | Column width percentages (should sum to 100) |
| `TableHeaderStyle(opts...)` | Header text/background styling (takes `TextOption`s) |
| `TableStripe(color)` | Alternating row background color |
| `TableCellVAlign(align)` | Vertical alignment for body cells (`VAlignTop`, `VAlignMiddle`, `VAlignBottom`) |

---

## List

Renders bulleted (unordered) or numbered (ordered) lists.

### Builder API

```go
// Unordered (bullet) list
c.List([]string{"First item", "Second item", "Third item"})

// Ordered (numbered) list
c.OrderedList([]string{"Step one", "Step two", "Step three"})

// With custom indent
c.List(items, template.ListIndent(document.Mm(10)))
```

### JSON Schema

```json
{
    "span": 12,
    "list": {
        "type": "unordered",
        "items": ["First item", "Second item", "Third item"]
    }
}
```

```json
{
    "span": 12,
    "list": {
        "type": "ordered",
        "items": ["Step one", "Step two", "Step three"]
    }
}
```

---

## Line

Renders a horizontal rule (separator line).

### Builder API

```go
c.Line()
c.Line(template.LineColor(pdf.RGBHex(0x1565C0)))
c.Line(
    template.LineColor(pdf.Gray(0.7)),
    template.LineThickness(document.Pt(2)),
)
```

### JSON Schema

```json
{"span": 12, "line": {"color": "#1565C0", "thickness": "2pt"}}
```

### Line Options

| Option | Description |
|---|---|
| `LineColor(c)` | Line color |
| `LineThickness(v)` | Line thickness (dimension value) |

---

## Spacer

Adds vertical empty space between elements.

### Builder API

```go
c.Spacer(document.Mm(10))
c.Spacer(document.Pt(24))
```

### JSON Schema

```json
{"span": 12, "spacer": "10mm"}
```

Or in elements array:

```json
{"type": "spacer", "height": "10mm"}
```

---

## QR Code

Generates and embeds a QR code image.

### Builder API

```go
c.QRCode("https://gpdf.dev")
c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(30)))
c.QRCode("HELLO",
    template.QRSize(document.Mm(25)),
    template.QRErrorCorrection(qrcode.LevelH),
    template.QRScale(8),
)
```

### JSON Schema

```json
{
    "span": 6,
    "qrcode": {
        "data": "https://gpdf.dev",
        "size": "30mm",
        "errorCorrection": "M"
    }
}
```

### QR Code Options

| Option | Description |
|---|---|
| `QRSize(v)` | Display size (width = height) |
| `QRErrorCorrection(level)` | Error correction: `LevelL`, `LevelM` (default), `LevelQ`, `LevelH` |
| `QRScale(s)` | Pixels per QR module (affects image resolution) |

### Error Correction Levels

| Level | Recovery | Best For |
|---|---|---|
| L | ~7% | Maximum data capacity |
| M | ~15% | General use (default) |
| Q | ~25% | Industrial / outdoor use |
| H | ~30% | Maximum reliability |

QR codes support Unicode content including CJK characters:

```go
c.QRCode("こんにちは世界", template.QRSize(document.Mm(30)))
```

---

## Barcode

Generates and embeds a Code128 barcode image.

### Builder API

```go
c.Barcode("INV-2026-0001")
c.Barcode("INV-2026-0001",
    template.BarcodeWidth(document.Mm(80)),
    template.BarcodeHeight(document.Mm(15)),
)
```

### JSON Schema

```json
{
    "span": 12,
    "barcode": {
        "data": "INV-2026-0001",
        "width": "80mm",
        "height": "15mm",
        "format": "code128"
    }
}
```

### Barcode Options

| Option | Description |
|---|---|
| `BarcodeWidth(v)` | Display width |
| `BarcodeHeight(v)` | Display height |
| `BarcodeFormat(f)` | Symbology (`barcode.Code128`) |

---

## Page Number / Total Pages

Insert the current page number or total page count. These are typically used in headers or footers.

### Builder API

```go
doc.Footer(func(p *template.PageBuilder) {
    p.AutoRow(func(r *template.RowBuilder) {
        r.Col(6, func(c *template.ColBuilder) {
            c.PageNumber(template.FontSize(8))
        })
        r.Col(6, func(c *template.ColBuilder) {
            c.TotalPages(template.AlignRight(), template.FontSize(8))
        })
    })
})
```

### JSON Schema

```json
{"type": "pageNumber", "style": {"size": 8}}
{"type": "totalPages", "style": {"size": 8, "align": "right"}}
```

Page numbers are resolved after pagination, so they correctly reflect the actual page numbers even when content overflows across multiple pages.

## See Also

- [Styling](06-styling.md) -- Colors, units, and decoration details
- [Builder API](02-builder-api.md) -- Full builder pattern reference
- [JSON Schema](03-json-schema.md) -- Declarative JSON format reference
