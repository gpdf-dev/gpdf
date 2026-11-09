# gpdf JSON — Document Definition Spec (canonical)

> Everything you need to write the JSON that gpdf's `template.FromJSON` accepts, in one file: prose, the JSON Schema, and verified examples. To have an AI write gpdf JSON, hand it this file as is.

- Published at: `https://gpdf.dev/spec/gpdf-json-spec.en.md` (Japanese: `gpdf-json-spec.ja.md`)
- JSON Schema: `https://gpdf.dev/schema/gpdf.schema.json` (same as Appendix A)
- Generated file. The sources live under `schema/` in the gpdf repository

🔴 **Every JSON document in this file has been run through the implementation.** Tests check the Appendix A schema against the Go types gpdf decodes into, and every Appendix B example passes the schema and renders to a PDF.

---

## 1. Write, check, fix

A document is **one JSON file**. Four steps:

1. **Write.** Start from the closest example in Appendix B.
2. **Validate against the schema.** Run any JSON Schema validator against Appendix A (or the URL above). Unknown fields, invalid colors, invalid dimensions and out-of-range `span` values fail here.
3. **Render.** Turn it into a PDF with the program in "7. Rendering".
4. **Inspect.** Follow "8. Checking the output" to make sure no text is garbled and no element went missing.

- 🔴 **gpdf itself silently swallows many mistakes.** Elements with invalid values are dropped and a PDF is still produced without an error ("6. What gets silently swallowed"). So **never skip steps 2 and 4.** A successful render does not mean the document came out as intended.
- **Appendix A is the source of truth for field names, types, enums and formats.** This prose does not repeat them.

## 2. Document structure

The top level is `page` (required) plus the fields that hold content.

| Field | Role |
|---|---|
| `page` | Paper size, orientation, margins, and a background shared by every page |
| `metadata` | Title, author, etc. embedded in the PDF |
| `header` / `footer` | Rows repeated at the top / bottom of every page |
| `body` | Body rows. **Breaks to a new page automatically** when it does not fit |
| `pages` | An array of explicitly separated pages, each with its own `body` |
| `absolute` | Elements placed at given coordinates (stamps, watermarks, margin notes) |

- You can use both `body` and `pages`. `body` comes first as page 1, then `pages` follow.
- **Use `pages` only when you want to decide the page breaks yourself.** Long tables and text placed in `body` flow onto the next page on their own.

### Rows and columns (12-column grid)

The body is **a sequence of rows (`row`)**, and each row lays out **columns (`cols`)** left to right. A column's width is its `span`; **the spans in a row must add up to 12.**

- 🔴 **Nothing checks that the spans add up to 12** (the schema only checks that each `span` is 1–12). Above 12, the row runs off the right edge; below 12, the right side stays empty.
- To leave a gap, add a column with no content (`{"span": 4, "text": ""}`).
- Row height grows with its content by default. Writing a dimension in `height` fixes it.

### Column content: shorthand and `elements`

To put a single element in a column, write it directly on the column (**shorthand**).

```json
{ "span": 12, "text": "Heading", "style": { "size": 18, "bold": true } }
```

To stack several, list them in the `elements` array, each with a `type`.

```json
{ "span": 12, "elements": [
  { "type": "text", "content": "Heading", "style": { "size": 18, "bold": true } },
  { "type": "spacer", "height": "4mm" },
  { "type": "text", "content": "Body" }
]}
```

- 🔴 **When `elements` is present, every shorthand field on the same column (`text`, `table`, …) is ignored.** Do not mix them.
- The shorthand `text` string becomes `content` inside `elements`.

## 3. Elements

There are 13 `type` values. See Appendix A for each element's fields.

| `type` | What it places | Shorthand |
|---|---|---|
| `text` | Text. `style` sets size, bold, color, alignment, background, padding, etc. | `text` + `style` |
| `link` | Linked text. The URL goes in `style.link` (blue and underlined by default) | none |
| `richText` | One paragraph with per-fragment styling (a bold phrase, an inline link) | `richText` |
| `image` | PNG / JPEG / SVG. `src` is a data URI, file path, base64, or SVG markup | `image` |
| `table` | Header row plus data rows. Long tables continue across pages | `table` |
| `list` | Bulleted list (`"ordered"` for numbered) | `list` |
| `line` | Horizontal rule | `line` |
| `spacer` | Vertical space (`height` is a dimension) | `spacer` (dimension string) |
| `qrcode` | QR code | `qrcode` |
| `barcode` | Barcode (Code 128) | `barcode` |
| `box` | Container with background, border, rounded corners and padding. Stacks `elements` inside; can be nested | `box` |
| `pageNumber` | Current page number (use in `header` / `footer`) | none |
| `totalPages` | Total page count | none |

- **`header` / `footer` use the same rows and columns as the body.** For page numbers, put `pageNumber` in the `elements` of a `footer` column.
- `pageNumber` and `totalPages` output only the number. For "1 / 3", place them in adjacent columns.

## 4. Dimensions and colors

- **Dimensions are strings with a unit:** `mm` / `cm` / `in` / `pt` / `em` / `%`. A bare number means `pt`. The exact format is `dimension` in Appendix A.
- **Only font size (`size`) and letter spacing (`letterSpacing`) are numbers, in pt.**
- **Colors are strings:** `#RRGGBB` / `rgb(r, g, b)` (each 0.0–1.0) / `gray(v)` (0.0–1.0) / eight names. Note that `rgb()` is **not 0–255**.
- **An `opacity` of 0 means "unset = opaque", not "transparent".** For a faint watermark, use a value above 0 such as 0.1–0.3.
- Per-edge fields (`paddings` / `margins` / `widths`) use CSS order: **top, right, bottom, left**.

## 5. Fonts

- **Without a registered font, the PDF standard fonts (Helvetica family) are used.** In practice they cover **ASCII only**.
- 🔴 **Japanese, Chinese and Korean text becomes `?` unless a font is registered.** It is not an error. Non-ASCII symbols such as `—` (em dash) also disappear or turn into other characters with the standard fonts.
- **Registering fonts is the renderer's job (Go), not the JSON's.** `fontFamily` in JSON can only pick a registered name; an unregistered name **silently falls back to the default font**.
- 🔴 **Table cells have no `fontFamily` field and are drawn with the default font.** So for documents with CJK text, **register the font as the default font** ("7. Rendering"). Then the JSON needs no `fontFamily` at all.
- `bold` / `italic` render in the regular face unless a font file for that face is registered.

## 6. What gets silently swallowed

gpdf **does not raise an error** for the following; it drops them or falls back to defaults and still produces a PDF. Some fail at step 2 (schema); others only show up at step 4 (inspection).

| What you wrote | What gpdf does | Where you can catch it |
|---|---|---|
| An unknown field (typo, nonexistent feature) | Ignored | Schema |
| An invalid element color (`"#ggg"`, `"rgb(255,0,0)"`, …) | That color setting is ignored | Schema |
| An invalid dimension in `row.height` | **The whole row disappears** | Schema |
| An invalid dimension in `spacer` | The space disappears | Schema |
| An unknown `type` | The element disappears | Schema |
| Both `elements` and shorthand fields | The shorthand disappears | Visual check |
| An unreadable image (wrong path, broken base64) | **The image disappears** | Visual check |
| Spans not adding up to 12 | Overflows or leaves a gap | Visual check |
| An unregistered `fontFamily` | Falls back to the default font | Visual check |
| Non-ASCII text with the standard fonts | Becomes `?` or disappears | `pdftotext` |
| `bold` / `size` / `fontFamily` in a table's `headerStyle` | **Ignored** (only `color` and `background` apply) | Visual check |
| `type: "link"` without `style.link` | Rendered as plain text | Visual check |

**What does raise an error:** JSON syntax errors, an unknown paper size, an invalid dimension in `page.margins`, and an invalid color or image in the page background (`page.background`).

## 7. Rendering

Pass the JSON to `template.FromJSON` and call `Generate`. The program below works as is. Passing a font file as the third argument registers it as the default font.

```go
// render: go run . <in.json> <out.pdf> [font.ttf]
package main

import (
	"log"
	"os"

	"github.com/gpdf-dev/gpdf/template"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: render <in.json> <out.pdf> [font.ttf]")
	}
	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var opts []template.Option
	if len(os.Args) > 3 { // CJK etc.: register the font and make it the default
		font, err := os.ReadFile(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts,
			template.WithFont("Main", font),
			template.WithDefaultFont("Main", 10))
	}
	doc, err := template.FromJSON(src, nil, opts...)
	if err != nil {
		log.Fatal(err)
	}
	pdf, err := doc.Generate()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(os.Args[2], pdf, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

```sh
go mod init render && go get github.com/gpdf-dev/gpdf
go run . invoice.json invoice.pdf                       # ASCII-only document
go run . quote.json quote.pdf NotoSansJP-Regular.ttf    # document with Japanese text
```

- **Image file paths are relative to where the program runs.** When in doubt, embed the image in the JSON as a data URI.
- If you pass data as the second argument of `FromJSON`, the JSON is first expanded as a **Go template** (`{{.Name}}`, etc.). When rendering JSON an AI wrote, pass `nil`.

## 8. Checking the output

**A successful render does not mean the result is what you intended** ("6. What gets silently swallowed"). Once you have a PDF, check:

1. **Is the text there?** Extract it with `pdftotext out.pdf -` and make sure **there are no runs of `?`** and every sentence you wrote is present. A missing font registration shows up here.
2. **Is the page count as expected?** `Pages` in `pdfinfo out.pdf`. Two pages where you expected one means something overflowed onto a new page.
3. **How does it look?** Read the PDF directly if you can; otherwise turn it into images with `pdftoppm -png -r 60 out.pdf page`. Missing images, overflowing columns and header styles that did not apply only show up here.

When an AI writes the JSON, **the AI can run these three checks itself and fix what it finds.** With rendering and checking in the AI's hands, a person only needs to look once, at the end.
