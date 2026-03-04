# Styling

gpdf provides a comprehensive styling system for text, colors, alignment, decoration, and dimensions. Styles are applied through functional options (Builder API) or style objects (JSON Schema).

## Text Options (Builder API)

All text styling is done via `TextOption` functions passed to `c.Text(...)`, `c.RichText(...)`, `c.PageNumber(...)`, and `c.TotalPages(...)`.

```go
c.Text("Styled text",
    template.FontSize(18),
    template.Bold(),
    template.Italic(),
    template.TextColor(pdf.RGBHex(0x1A237E)),
    template.BgColor(pdf.Yellow),
    template.AlignCenter(),
    template.Underline(),
    template.LetterSpacing(0.5),
    template.FontFamily("NotoSansJP"),
)
```

### Complete Option Reference

| Option | Type | Description |
|---|---|---|
| `FontSize(size)` | `float64` | Font size in points |
| `Bold()` | -- | Set font weight to bold (700) |
| `Italic()` | -- | Set font style to italic |
| `TextColor(c)` | `pdf.Color` | Text foreground color |
| `BgColor(c)` | `pdf.Color` | Background highlight color |
| `AlignLeft()` | -- | Left text alignment (default) |
| `AlignCenter()` | -- | Center text alignment |
| `AlignRight()` | -- | Right text alignment |
| `FontFamily(name)` | `string` | Set font family (must be registered) |
| `LetterSpacing(pts)` | `float64` | Extra inter-character spacing in points |
| `TextIndent(v)` | `Value` | First-line indentation |
| `Underline()` | -- | Add underline decoration |
| `Strikethrough()` | -- | Add strikethrough decoration |

Multiple decorations can be combined:

```go
c.Text("Bold underlined", template.Bold(), template.Underline())
```

## Colors

gpdf supports multiple color models and formats.

### Creating Colors (Go)

```go
import "github.com/gpdf-dev/gpdf/pdf"

// RGB (each component 0.0 to 1.0)
color := pdf.RGB(0.1, 0.14, 0.49)

// RGB from hex (0xRRGGBB)
color := pdf.RGBHex(0x1A237E)

// Grayscale (0.0 = black, 1.0 = white)
color := pdf.Gray(0.5)

// CMYK (each component 0.0 to 1.0)
color := pdf.CMYK(0.0, 0.0, 0.0, 1.0)
```

### Predefined Colors

| Constant | Value |
|---|---|
| `pdf.Black` | RGB(0, 0, 0) |
| `pdf.White` | RGB(1, 1, 1) |
| `pdf.Red` | RGB(1, 0, 0) |
| `pdf.Green` | RGB(0, 1, 0) |
| `pdf.Blue` | RGB(0, 0, 1) |
| `pdf.Yellow` | RGB(1, 1, 0) |
| `pdf.Cyan` | RGB(0, 1, 1) |
| `pdf.Magenta` | RGB(1, 0, 1) |

### Color Strings (JSON Schema)

In JSON schemas, colors are specified as strings:

| Format | Example | Description |
|---|---|---|
| Hex | `"#1A237E"` | `#RRGGBB` format |
| RGB | `"rgb(0.1, 0.14, 0.49)"` | Float components (0.0-1.0) |
| Gray | `"gray(0.5)"` | Grayscale (0.0-1.0) |
| Named | `"red"` | Predefined name |

**Named colors:** `black`, `white`, `red`, `green`, `blue`, `yellow`, `cyan`, `magenta`

### Usage Examples

```go
// Text color
c.Text("Error message", template.TextColor(pdf.Red))

// Background color
c.Text("Highlighted", template.BgColor(pdf.RGBHex(0xFFF3E0)))

// Line color
c.Line(template.LineColor(pdf.RGBHex(0x1565C0)))

// Table header styling
template.TableHeaderStyle(
    template.TextColor(pdf.White),
    template.BgColor(pdf.RGBHex(0x1A237E)),
)

// Table stripe color
template.TableStripe(pdf.RGBHex(0xF5F5F5))
```

## Units and Dimensions

gpdf supports multiple measurement units. All values are ultimately resolved to PDF points (1/72 inch) internally.

### Creating Values (Go)

```go
import "github.com/gpdf-dev/gpdf/document"

document.Pt(12)     // 12 points (1/72 inch each)
document.Mm(20)     // 20 millimeters
document.Cm(2.5)    // 2.5 centimeters
document.In(1)      // 1 inch (= 72 points)
document.Em(1.5)    // 1.5x current font size
document.Pct(50)    // 50% of parent dimension
document.Auto       // Auto-calculated by layout engine
```

### Conversion Reference

| Unit | To Points | Example |
|---|---|---|
| `Pt` | identity | `Pt(72)` = 72pt = 1 inch |
| `Mm` | x 2.83465 | `Mm(25.4)` = 72pt = 1 inch |
| `Cm` | x 28.3465 | `Cm(2.54)` = 72pt = 1 inch |
| `In` | x 72 | `In(1)` = 72pt |
| `Em` | x fontSize | `Em(1)` at 12pt = 12pt |
| `Pct` | / 100 x parent | `Pct(50)` = half of parent |

### Dimension Strings (JSON)

In JSON schemas, dimensions are specified as strings:

```json
"20mm"    // millimeters
"12pt"    // points
"2.5cm"   // centimeters
"1in"     // inches
"1.5em"   // relative to font size
"50%"     // percentage
"12"      // bare number = points
"auto"    // auto-calculated
```

### Margins and Edges

```go
// Uniform margins (same on all sides)
document.UniformEdges(document.Mm(20))

// Custom margins (Top, Right, Bottom, Left)
document.Edges{
    Top:    document.Mm(25),
    Right:  document.Mm(20),
    Bottom: document.Mm(25),
    Left:   document.Mm(20),
}
```

## Style Object (JSON Schema)

In JSON schemas, styles are defined as objects:

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

All fields are optional. Only specified fields override the defaults.

## Default Style

When no explicit style is set, gpdf uses these defaults:

| Property | Default |
|---|---|
| Font size | 12pt |
| Font weight | Normal (400) |
| Font style | Normal |
| Color | Black |
| Text alignment | Left |
| Line height | 1.2 |

The default font and size can be changed at the document level:

```go
doc := template.New(
    template.WithDefaultFont("NotoSansJP", 10),
)
```

## See Also

- [Elements](05-elements.md) -- How to apply styles to each element type
- [Fonts](09-fonts.md) -- Registering and using custom fonts
- [Layout](07-layout.md) -- Page sizes and margin configuration
