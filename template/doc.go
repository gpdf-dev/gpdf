// Package template provides a declarative, builder-pattern API for
// constructing PDF documents. It sits on top of the document model
// (Layer 2) and provides high-level constructs such as grids, headers,
// footers, and reusable components. This is Layer 3 of the gpdf
// architecture and the recommended entry point for most users.
//
// # Quick Start
//
// Create a document, add a page, and generate PDF bytes:
//
//	doc := template.New(
//	    template.WithPageSize(document.A4),
//	    template.WithMargins(document.UniformEdges(document.Mm(15))),
//	)
//	page := doc.AddPage()
//	page.AutoRow(func(r *template.RowBuilder) {
//	    r.Col(12, func(c *template.ColBuilder) {
//	        c.Text("Hello, World!", template.FontSize(24))
//	    })
//	})
//	data, err := doc.Generate()
//
// # 12-Column Grid System
//
// Content is organized using a Bootstrap-style 12-column grid:
//
//   - [PageBuilder] manages rows within a page
//   - [RowBuilder] divides a row into columns (span values sum to 12)
//   - [ColBuilder] populates a column with text, images, tables, and more
//
// Example two-column layout:
//
//	page.AutoRow(func(r *template.RowBuilder) {
//	    r.Col(6, func(c *template.ColBuilder) { c.Text("Left") })
//	    r.Col(6, func(c *template.ColBuilder) { c.Text("Right") })
//	})
//
// # Content Elements
//
// [ColBuilder] provides methods for all supported content types:
//
//   - [ColBuilder.Text] — styled text with font, color, and alignment options
//   - [ColBuilder.Image] — JPEG/PNG images with fit modes
//   - [ColBuilder.Table] — tabular data with headers, striping, and column widths
//   - [ColBuilder.List] / [ColBuilder.OrderedList] — bulleted or numbered lists
//   - [ColBuilder.Line] — horizontal rules
//   - [ColBuilder.Spacer] — vertical whitespace
//   - [ColBuilder.QRCode] — QR code images
//   - [ColBuilder.Barcode] — Code 128 barcode images
//   - [ColBuilder.RichText] — mixed inline styles in a single paragraph
//   - [ColBuilder.PageNumber] / [ColBuilder.TotalPages] — page numbering
//
// # Functional Options
//
// Text and element styling uses the functional options pattern:
//
//	c.Text("Title", template.FontSize(24), template.Bold(), template.TextColor(pdf.Blue))
//
// Available text options: [FontSize], [Bold], [Italic], [TextColor],
// [BgColor], [AlignLeft], [AlignCenter], [AlignRight], [FontFamily],
// [Underline], [Strikethrough], [LetterSpacing], [TextIndent].
//
// # JSON Schema
//
// Documents can be defined declaratively as JSON using [FromJSON]:
//
//	doc, err := template.FromJSON(jsonBytes, nil)
//
// The JSON schema supports Go template expressions for dynamic data binding:
//
//	doc, err := template.FromJSON(schemaBytes, map[string]string{"name": "gpdf"})
//
// See [Schema] for the complete JSON structure definition.
//
// # Go Template Integration
//
// [FromTemplate] executes a pre-parsed Go text/template that produces JSON
// schema output. Use [TemplateFuncMap] to register helper functions:
//
//	tmpl := gotemplate.New("").Funcs(template.TemplateFuncMap())
//	tmpl, _ = tmpl.Parse(templateString)
//	doc, err := template.FromTemplate(tmpl, data)
//
// # Reusable Components
//
// Pre-built document components generate complete, styled PDFs:
//
//   - [Invoice] — professional invoice with line items, totals, and payment info
//   - [Report] — structured report with sections, tables, and metrics
//   - [Letter] — formal business letter
//
// Each component accepts a typed data struct and returns a ready-to-generate
// [Document].
package template
