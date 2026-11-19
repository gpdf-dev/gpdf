# Fonts

gpdf supports TrueType font embedding with automatic subsetting. Custom fonts enable CJK text rendering, brand-specific typography, and Unicode coverage beyond standard PDF fonts.

## Registering Fonts

### Builder API

```go
fontData, err := os.ReadFile("NotoSansJP-Regular.ttf")
if err != nil {
    panic(err)
}

doc := template.New(
    template.WithFont("NotoSansJP", fontData),
    template.WithDefaultFont("NotoSansJP", 12),
)
```

### Multiple Fonts

Register multiple font families and variants:

```go
regular, _ := os.ReadFile("NotoSansJP-Regular.ttf")
bold, _ := os.ReadFile("NotoSansJP-Bold.ttf")

doc := template.New(
    template.WithFont("NotoSansJP", regular),
    template.WithFont("NotoSansJP-Bold", bold),
    template.WithDefaultFont("NotoSansJP", 12),
)
```

Use the `FontFamily` text option to switch fonts within a document:

```go
c.Text("Regular text")  // uses default font
c.Text("Bold text", template.FontFamily("NotoSansJP-Bold"), template.Bold())
```

### With JSON/Template

Pass font options alongside `FromJSON` or `FromTemplate`:

```go
doc, err := template.FromJSON(schema, data,
    template.WithFont("NotoSansJP", fontData),
    template.WithDefaultFont("NotoSansJP", 12),
)
```

## Font Subsetting

gpdf automatically subsets embedded fonts. Only the glyphs actually used in the document are included in the PDF output. This dramatically reduces file size, especially for CJK fonts which can be 10-20 MB unsubsetted.

The subsetting process:
1. Track all Unicode codepoints used in the document
2. Map codepoints to glyph IDs via the font's `cmap` table
3. Resolve composite glyph dependencies
4. Extract only the required glyph outlines
5. Build a minimal TrueType font file
6. Generate a ToUnicode CMap for text extraction/search

## CJK Support

gpdf has first-class support for Japanese, Chinese, and Korean text:

### Line Breaking

- **Word wrapping** at spaces (for Latin text)
- **Character-level wrapping** for CJK characters
- **Kinsoku rules** for Japanese (prohibits certain characters at line start/end)

### Example with Japanese Text

```go
fontData, _ := os.ReadFile("NotoSansJP-Regular.ttf")

doc := template.New(
    template.WithPageSize(document.A4),
    template.WithFont("NotoSansJP", fontData),
    template.WithDefaultFont("NotoSansJP", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
    r.Col(12, func(c *template.ColBuilder) {
        c.Text("請求書", template.FontSize(24), template.Bold())
        c.Text("株式会社テスト")
        c.Text("東京都渋谷区1-2-3")
    })
})
```

### QR Codes with CJK

QR codes natively support Unicode content:

```go
c.QRCode("こんにちは世界", template.QRSize(document.Mm(30)))
```

## Complex Scripts (Arabic, Hebrew, Indic)

On its own, gpdf maps each character to one glyph and lays text out left to
right. That is enough for Latin, Cyrillic, Greek and CJK, but not for scripts
where letters change shape with their neighbours, combine into ligatures,
carry positioned marks, or run right to left. For those, plug in a text
shaper with `template.WithShaper`. gpdf stays dependency-free: the shaper is
an interface (`font.Shaper`) that you implement with a shaping engine such
as [go-text/typesetting](https://github.com/go-text/typesetting) (a pure-Go
HarfBuzz port with the Unicode Bidi Algorithm).

```go
doc := gpdf.NewDocument(
    gpdf.WithFont("NotoSans", sans),
    gpdf.WithFont("NotoNaskhArabic", naskh),
    template.WithShaper(myShaper),                 // implements font.Shaper
    template.WithFallbackFonts("NotoNaskhArabic"), // for characters NotoSans lacks
    gpdf.WithDefaultFont("NotoSans", 12),
)

page.AutoRow(func(r *template.RowBuilder) {
    r.Col(12, func(c *template.ColBuilder) {
        c.Text("رقم الطلب: 1024", template.RTL(), template.AlignRight())
    })
})
```

With a shaper configured, text in registered TrueType fonts is:

- **broken into lines by gpdf**, in logical order, measuring candidate lines
  with the shaper (`font.WrapWords`), as the Unicode Bidi Algorithm requires;
- **shaped per line** by the shaper, which returns glyph runs in visual order
  with advances and offsets (contextual forms, ligatures, mark positioning,
  kerning, bidi reordering);
- **drawn by glyph ID**, using `TJ` adjustments for advances and text rise
  (`Ts`) for vertical offsets;
- **searchable and copyable**: each line is wrapped in an `/ActualText`
  marked-content span holding its logical text, and the ToUnicode map gets
  the text of each glyph (all characters of a ligature).

`WithFallbackFonts` lists families tried, in order, for characters the
text's own font does not cover (the same bold/italic variant is used when
registered). `template.Direction` / `template.RTL` set the base direction;
by default the shaper detects it from the first strong character.

Without `WithShaper`, nothing changes. Overlays (`gpdf.Open`) and the JSON
schema do not use the shaper yet, and a `RichText` line is shaped per
fragment, so bidi reordering does not cross fragment boundaries: prefer a
single `Text` with fallback fonts for mixed-script lines.

## Font Metrics

gpdf uses TrueType font metrics for accurate text measurement:

- **Ascender/Descender**: Vertical extent of the font
- **Line gap**: Extra space between lines
- **Cap height / X-height**: Heights of uppercase/lowercase letters
- **Glyph widths**: Per-character advance widths for precise layout

These metrics are used internally by the layout engine for:
- Text width measurement
- Line breaking decisions
- Vertical spacing calculations

## How Font Resolution Works

When rendering text, gpdf resolves fonts in this order:

1. **Explicit font**: If `FontFamily("name")` is set, use that font
2. **Variant matching**: Try to match weight (bold) and style (italic) variants
   - E.g., if family is "Helvetica" and text is bold, try "Helvetica-Bold" first
3. **Base family fallback**: Fall back to the base family name
4. **Approximate metrics**: If no matching font is registered, use approximate metrics based on standard font characteristics

## Supported Font Formats

| Format | Supported | Notes |
|---|---|---|
| TrueType (.ttf) | Yes | Full support with subsetting |
| OpenType (.otf) | Partial | TrueType-outline OpenType fonts work |
| WOFF/WOFF2 | No | Decompress to TTF first |
| Type 1 | No | -- |

## Best Practices

1. **Always embed fonts for CJK**: Standard PDF fonts don't include CJK glyphs
2. **Use subsetting**: gpdf does this automatically -- no action needed
3. **Set a default font**: Use `WithDefaultFont` to avoid falling back to approximate metrics
4. **Register variants separately**: Bold and italic are separate font files in TrueType
5. **Prefer Noto fonts**: Google's Noto font family provides excellent Unicode coverage and is freely available

## Recommended Fonts

| Font | Coverage | License | Use Case |
|---|---|---|---|
| Noto Sans JP | Japanese + Latin | OFL | General Japanese documents |
| Noto Sans SC | Simplified Chinese + Latin | OFL | Chinese documents |
| Noto Sans KR | Korean + Latin | OFL | Korean documents |
| Noto Sans | Latin, Cyrillic, Greek | OFL | Western documents |
| Inter | Latin | OFL | Modern UI-style documents |

## See Also

- [Styling](06-styling.md) -- Font size, weight, and style options
- [Elements](05-elements.md) -- Using fonts with text elements
- [Architecture](10-architecture.md) -- Font subsystem internals
