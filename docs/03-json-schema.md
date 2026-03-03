# JSON Schema

gpdf supports defining PDFs declaratively using JSON. This is useful for storing document templates as data, loading them from files or databases, and separating layout from application logic.

## Basic Usage

```go
import "github.com/gpdf-dev/gpdf/template"

schema := []byte(`{
    "page": {"size": "A4", "margins": "20mm"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "Hello, World!", "style": {"size": 24, "bold": true}}
        ]}}
    ]
}`)

doc, err := template.FromJSON(schema, nil)
if err != nil {
    // handle error
}
data, err := doc.Generate()
```

## With Data Binding (Go Templates)

JSON schemas can include Go template expressions (`{{.Field}}`). Pass a data object as the second argument to resolve them:

```go
schema := []byte(`{
    "page": {"size": "A4", "margins": "20mm"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "Hello, {{.Name}}!", "style": {"size": 24, "bold": true}}
        ]}}
    ]
}`)

data := map[string]string{"Name": "World"}
doc, err := template.FromJSON(schema, data)
```

When `data` is `nil`, no template processing is performed and the JSON is parsed directly.

## Schema Structure

The top-level JSON object has the following fields:

```json
{
    "page": { ... },
    "metadata": { ... },
    "header": [ ... ],
    "footer": [ ... ],
    "body": [ ... ],
    "pages": [ ... ]
}
```

### `page` -- Page Settings

| Field | Type | Description | Example |
|---|---|---|---|
| `size` | string | Page size name | `"A4"`, `"A3"`, `"Letter"`, `"Legal"` |
| `margins` | string | Uniform margins | `"20mm"`, `"15pt"`, `"1in"` |

```json
"page": {"size": "A4", "margins": "15mm"}
```

### `metadata` -- Document Metadata (optional)

| Field | Type | Description |
|---|---|---|
| `title` | string | Document title |
| `author` | string | Document author |
| `subject` | string | Document subject |
| `creator` | string | Creator application name |

```json
"metadata": {"title": "Invoice #001", "author": "ACME Corp"}
```

### `header` / `footer` -- Repeated Sections (optional)

Array of row definitions. Headers appear at the top of every page; footers appear at the bottom.

```json
"header": [
    {"row": {"cols": [
        {"span": 6, "text": "Company Name", "style": {"bold": true}},
        {"span": 6, "text": "Confidential", "style": {"align": "right"}}
    ]}}
]
```

### `body` -- Page Content

Array of row definitions. All rows are placed on a single page (with automatic pagination on overflow).

```json
"body": [
    {"row": {"cols": [
        {"span": 12, "text": "Page content here"}
    ]}}
]
```

### `pages` -- Multiple Explicit Pages (optional)

When you need explicit page breaks, use `pages` instead of `body`:

```json
"pages": [
    {"body": [
        {"row": {"cols": [{"span": 12, "text": "Page 1 content"}]}}
    ]},
    {"body": [
        {"row": {"cols": [{"span": 12, "text": "Page 2 content"}]}}
    ]}
]
```

## Row Definition

```json
{"row": {"height": "auto", "cols": [...]}}
```

| Field | Type | Description | Default |
|---|---|---|---|
| `height` | string | Row height (`"auto"` or dimension) | `"auto"` |
| `cols` | array | Column definitions | required |

## Column Definition

### Shorthand (single element per column)

```json
{"span": 6, "text": "Hello", "style": {"size": 18, "bold": true}}
```

| Field | Type | Description |
|---|---|---|
| `span` | int | Column width (1-12) |
| `text` | string | Text content |
| `image` | object | Image definition |
| `table` | object | Table definition |
| `list` | object | List definition |
| `line` | object | Horizontal line |
| `spacer` | string | Vertical space (dimension string) |
| `qrcode` | object | QR code |
| `barcode` | object | Barcode |
| `style` | object | Style for `text` shorthand |

### Multiple Elements (elements array)

To put multiple elements in a single column, use the `elements` array:

```json
{
    "span": 12,
    "elements": [
        {"type": "text", "content": "Title", "style": {"size": 20, "bold": true}},
        {"type": "spacer", "height": "5mm"},
        {"type": "text", "content": "Subtitle"},
        {"type": "line"},
        {"type": "spacer", "height": "10mm"}
    ]
}
```

### Element Types

| Type | Fields | Description |
|---|---|---|
| `text` | `content`, `style` | Text block |
| `image` | `image: {src, width, height, fit, align}` | Image |
| `table` | `table: {header, rows, columnWidths, headerStyle, stripeColor}` | Data table |
| `list` | `list: {type, items}` | Bullet or numbered list |
| `line` | `line: {color, thickness}` | Horizontal rule |
| `spacer` | `height` | Vertical spacing |
| `qrcode` | `qrcode: {data, size, errorCorrection}` | QR code |
| `barcode` | `barcode: {data, width, height, format}` | Barcode |
| `pageNumber` | `style` | Current page number |
| `totalPages` | `style` | Total page count |

## Style Object

The `style` object can be applied to text elements:

```json
{
    "size": 16,
    "bold": true,
    "italic": false,
    "align": "center",
    "color": "#1A237E",
    "background": "#F5F5F5",
    "fontFamily": "NotoSansJP",
    "underline": true,
    "strikethrough": false,
    "letterSpacing": 0.5
}
```

| Field | Type | Values |
|---|---|---|
| `size` | number | Font size in points |
| `bold` | bool | Bold weight |
| `italic` | bool | Italic style |
| `align` | string | `"left"`, `"center"`, `"right"` |
| `color` | string | `"#RRGGBB"`, `"rgb(r,g,b)"`, `"gray(v)"`, or named color |
| `background` | string | Same as color |
| `fontFamily` | string | Registered font family name |
| `underline` | bool | Underline decoration |
| `strikethrough` | bool | Strikethrough decoration |
| `letterSpacing` | number | Extra space between characters (points) |

### Named Colors

`black`, `white`, `red`, `green`, `blue`, `yellow`, `cyan`, `magenta`

### Color Formats

- Hex: `"#1A237E"`
- RGB (0.0-1.0): `"rgb(0.1, 0.14, 0.49)"`
- Grayscale (0.0-1.0): `"gray(0.5)"`
- Named: `"red"`, `"blue"`, etc.

## Image Object

```json
"image": {
    "src": "data:image/png;base64,iVBOR...",
    "width": "50mm",
    "height": "30mm",
    "fit": "contain",
    "align": "center"
}
```

| Field | Type | Description |
|---|---|---|
| `src` | string | Base64, data URI, or file path |
| `width` | string | Display width (dimension) |
| `height` | string | Display height (dimension) |
| `fit` | string | `"contain"`, `"cover"`, `"stretch"`, `"original"` |
| `align` | string | `"left"`, `"center"`, `"right"` |

## Table Object

```json
"table": {
    "header": ["Name", "Age", "City"],
    "rows": [
        ["Alice", "30", "Tokyo"],
        ["Bob", "25", "New York"]
    ],
    "columnWidths": [40, 30, 30],
    "headerStyle": {"bold": true, "color": "#FFFFFF", "background": "#1A237E"},
    "stripeColor": "#F5F5F5"
}
```

## List Object

```json
"list": {
    "type": "ordered",
    "items": ["First item", "Second item", "Third item"]
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | `"ordered"` or `"unordered"` (default) |
| `items` | array | List item strings |

## QR Code Object

```json
"qrcode": {
    "data": "https://gpdf.dev",
    "size": "30mm",
    "errorCorrection": "M"
}
```

| Field | Type | Description |
|---|---|---|
| `data` | string | Data to encode |
| `size` | string | Display size (dimension) |
| `errorCorrection` | string | `"L"`, `"M"`, `"Q"`, `"H"` |

## Barcode Object

```json
"barcode": {
    "data": "INV-2026-001",
    "width": "80mm",
    "height": "15mm",
    "format": "code128"
}
```

## Dimension Strings

All dimension fields accept these formats:

| Format | Example | Description |
|---|---|---|
| `Nmm` | `"20mm"` | Millimeters |
| `Npt` | `"12pt"` | PDF points (1/72 inch) |
| `Ncm` | `"2.5cm"` | Centimeters |
| `Nin` | `"1in"` | Inches |
| `Nem` | `"1.5em"` | Relative to font size |
| `N%` | `"50%"` | Percentage of parent |
| `N` | `"12"` | Bare number (defaults to points) |
| `auto` | `"auto"` | Auto-calculated |

## Overriding with Go Options

Options passed to `FromJSON` override settings in the schema:

```go
doc, err := template.FromJSON(schema, data,
    template.WithFont("NotoSansJP", fontData),
    template.WithDefaultFont("NotoSansJP", 12),
)
```

## Complete Example

```json
{
    "page": {"size": "A4", "margins": "15mm"},
    "metadata": {"title": "Monthly Report", "author": "ACME"},
    "header": [
        {"row": {"cols": [
            {"span": 6, "text": "ACME Corp", "style": {"bold": true, "size": 10}},
            {"span": 6, "text": "Confidential", "style": {"align": "right", "size": 10, "color": "gray(0.5)"}}
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
    ],
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "Monthly Report", "style": {"size": 24, "bold": true, "color": "#1A237E"}}
        ]}},
        {"row": {"cols": [
            {"span": 12, "spacer": "10mm"}
        ]}},
        {"row": {"cols": [
            {"span": 12, "table": {
                "header": ["Metric", "This Month", "Last Month", "Change"],
                "rows": [
                    ["Revenue", "$125,000", "$112,000", "+11.6%"],
                    ["Users", "8,500", "7,200", "+18.1%"],
                    ["Orders", "3,200", "2,800", "+14.3%"]
                ],
                "columnWidths": [30, 25, 25, 20],
                "headerStyle": {"color": "#FFFFFF", "background": "#1A237E"},
                "stripeColor": "#F5F5F5"
            }}
        ]}}
    ]
}
```

## See Also

- [Go Templates](04-go-templates.md) -- More advanced data binding with loops and conditionals
- [Elements](05-elements.md) -- Detailed reference for each element type
