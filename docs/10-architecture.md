# Architecture

gpdf uses a three-layer architecture with strict one-directional dependencies. This design separates concerns and allows each layer to be used independently.

## Layer Overview

```
┌─────────────────────────────────────────────────────┐
│  Layer 3: template                                  │
│  Declarative Builder API, 12-col Grid, Components   │
│  JSON Schema, Go Template integration               │
├─────────────────────────────────────────────────────┤
│  Layer 2: document                                  │
│  Node tree, Box Model, Style, Layout Engine,        │
│  Pagination, Renderer                               │
├─────────────────────────────────────────────────────┤
│  Layer 1: pdf                                       │
│  PDF Writer, Objects, Streams, TrueType Fonts,      │
│  Images, Cross-Reference Table                      │
└─────────────────────────────────────────────────────┘
```

**Dependency direction:** Layer 3 -> Layer 2 -> Layer 1 (never upward)

Layer 3 does **not** directly reference Layer 1 for content operations -- it builds Layer 2 node trees which are then rendered through the Layer 2 renderer that targets Layer 1.

## Layer 1: `pdf/` -- PDF Primitives

The lowest layer handles raw PDF file generation. It knows nothing about documents, layouts, or styles -- only PDF objects, streams, and the file format.

### Key Components

| File | Component | Description |
|---|---|---|
| `pdf/writer.go` | `Writer` | Assembles PDF files: objects, pages, fonts, images, xref, trailer |
| `pdf/object.go` | Object types | PDF primitives: `Dict`, `Array`, `Stream`, `Name`, `Integer`, `Real`, etc. |
| `pdf/stream.go` | `CompressFlate` | zlib/deflate compression for streams |
| `pdf/colorspace.go` | `Color` | RGB, Grayscale, CMYK color representations |
| `pdf/xref.go` | `XRefTable` | Cross-reference table tracking object byte offsets |
| `pdf/font/truetype.go` | `TrueTypeFont` | Full TrueType font parser (cmap, hmtx, metrics) |
| `pdf/font/subset.go` | `SubsetTrueType` | Font subsetting (extract only used glyphs) |
| `pdf/font/metrics.go` | `MeasureString`, `LineBreak` | Text measurement and line breaking with CJK/kinsoku support |
| `pdf/font/cmap.go` | `GenerateToUnicodeCMap` | Unicode mapping for text extraction |

### Writer API

```go
w := pdf.NewWriter(output)
w.SetCompression(true)
fontRef, fontName, _ := w.RegisterFont("MyFont", ttfData)
imgRef, imgName, _ := w.RegisterImage("logo", pngData)
w.AddPage(pageObject)
w.Close()
```

### PDF Object Model

```go
// All PDF objects implement the Object interface
type Object interface {
    WriteTo(w io.Writer) (int64, error)
}

// Examples: Dict, Array, Stream, Name, LiteralString, Integer, Real, etc.
```

## Layer 2: `document/` -- Document Model

The middle layer defines an abstract document tree independent of the PDF format. It handles layout computation, pagination, and rendering.

### Node System

All content is represented as a tree of `DocumentNode` values:

```go
type DocumentNode interface {
    NodeType() NodeType
    Children() []DocumentNode
    Style() Style
}
```

| Node Type | Description |
|---|---|
| `Document` | Root node containing pages |
| `Page` | Single page with size, margins, content |
| `Box` | Container with CSS-like box model (margin, padding, border) |
| `Text` | Text leaf node with content and style |
| `Image` | Image leaf node with source data and fit mode |
| `Table` | Table with header, body, footer rows |
| `List` | Ordered or unordered list |
| `RichText` | Mixed-style inline text fragments |
| `Path` | Vector path drawing (MoveTo, LineTo, CurveTo, Close) |

### Style System

```go
type Style struct {
    FontFamily    string
    FontSize      float64
    FontWeight    FontWeight    // 400 (Normal), 700 (Bold)
    FontStyle     FontStyle     // Normal, Italic
    Color         Color
    Background    *Color
    TextAlign     TextAlign     // Left, Center, Right, Justify
    LineHeight    float64
    LetterSpacing float64
    TextIndent    Value
    TextDecoration uint8        // Underline, Strikethrough, Overline
    VerticalAlign  VerticalAlign // Top, Middle, Bottom
    Margin, Padding Edges
    Border         BorderEdges
}
```

### Layout Engine (`document/layout/`)

The layout engine computes positions and sizes for all nodes:

```go
type Engine interface {
    Layout(node DocumentNode, constraints Constraints) Result
}

type Constraints struct {
    AvailableWidth, AvailableHeight float64
    FontResolver FontResolver
}

type Result struct {
    Bounds   Rectangle
    Children []PlacedNode
    Overflow DocumentNode  // content that didn't fit
}
```

| File | Component | Description |
|---|---|---|
| `layout/engine.go` | Engine interface | Layout abstraction and constraints |
| `layout/block.go` | Block layout | Box and block element positioning |
| `layout/flow.go` | Flow layout | Text wrapping and line breaking |
| `layout/table.go` | Table layout | Cell sizing and row layout |
| `layout/list.go` | List layout | List item positioning |
| `layout/paging.go` | Paginator | Pagination with header/footer injection |

### Renderer (`document/render/`)

The renderer traverses placed nodes and outputs to a target:

```go
type Renderer interface {
    BeginDocument(info DocumentMetadata) error
    BeginPage(size Size) error
    EndPage() error
    RenderText(text string, pos Point, style Style) error
    RenderRect(rect Rectangle, style RectStyle) error
    RenderImage(src ImageSource, pos Point, size Size) error
    RenderPath(path Path, style PathStyle) error
    RenderLine(from, to Point, style LineStyle) error
    EndDocument() error
}
```

`render/pdftarget.go` implements this interface targeting the Layer 1 `pdf.Writer`.

## Layer 3: `template/` -- Template API

The highest layer provides the user-facing API. It builds Layer 2 document trees from high-level constructs.

### Builder Pattern

```
Document → PageBuilder → RowBuilder → ColBuilder → Elements
```

| File | Component | Description |
|---|---|---|
| `template/builder.go` | `Document`, `PageBuilder`, `RowBuilder`, `ColBuilder` | Builder hierarchy |
| `template/grid.go` | Grid system | 12-column layout calculations |
| `template/component.go` | Option types | `TextOption`, `ImageOption`, `TableOption`, etc. |
| `template/schema.go` | JSON Schema | Schema types and JSON-to-builder conversion |
| `template/gotemplate.go` | Go Templates | `FromJSON`, `FromTemplate`, `TemplateFuncMap` |
| `template/fontresolver.go` | Font resolver | Font resolution for layout engine |
| `template/richtext.go` | `RichTextBuilder` | Mixed-style inline text builder |
| `template/invoice.go` | Invoice component | Pre-built invoice document |
| `template/report.go` | Report component | Pre-built report document |
| `template/letter.go` | Letter component | Pre-built business letter |

### Supporting Packages

| Package | Description |
|---|---|
| `barcode/` | Code128 barcode encoding and PNG rendering |
| `qrcode/` | QR code generation with error correction levels |

## Data Flow

The complete flow from user API to PDF output:

```
User Code (Builder / JSON / Template)
    │
    ▼
template.Document          ← Layer 3: builds node tree
    │  .Generate() / .Render()
    ▼
document.Document          ← Layer 2: abstract document tree
    │
    ▼
layout.Paginator           ← Layer 2: pagination + header/footer
    │  .Paginate()
    ▼
[]layout.PageResult        ← Layer 2: positioned nodes per page
    │
    ▼
layout.ResolvePageNumbers  ← Layer 2: replace placeholders
    │
    ▼
render.PDFRenderer         ← Layer 2→1: traverse and render
    │  .RenderDocument()
    ▼
pdf.Writer                 ← Layer 1: assemble PDF bytes
    │  .Close()
    ▼
io.Writer                  ← Final PDF output
```

## Font Pipeline

```
TTF file bytes
    │
    ▼
font.ParseTrueType()      ← Parse tables: head, cmap, hmtx, name, etc.
    │
    ▼
TrueTypeFont              ← Metrics, glyph mapping, encoding
    │
    ├── MeasureString()    ← Layout: text width calculation
    ├── LineBreak()        ← Layout: word/CJK wrapping + kinsoku
    │
    ▼
font.SubsetTrueType()     ← Extract used glyphs only
    │
    ▼
pdf.Writer.RegisterFont() ← Embed subset in PDF
    │
    ▼
font.GenerateToUnicodeCMap() ← Enable text selection/search
```

## Design Principles

1. **Layered isolation**: Each layer has a clear responsibility. Users can work at any level.
2. **Zero dependencies**: The entire library uses only Go's standard library.
3. **Interface-based**: Key abstractions (`DocumentNode`, `Engine`, `Renderer`, `FontResolver`) are interfaces, enabling extensibility.
4. **Functional options**: Configuration uses the `WithXxx` pattern for clean, extensible APIs.
5. **CJK-first**: Text measurement, line breaking, and font handling are designed with CJK support from the start.
6. **Performance**: Font subsetting, stream compression, and efficient PDF assembly keep output small and generation fast.

## See Also

- [Builder API](02-builder-api.md) -- Layer 3 user-facing API
- [Fonts](09-fonts.md) -- Font pipeline details
- [Layout](07-layout.md) -- Grid system and pagination
