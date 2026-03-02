package template

import (
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// ReportData holds all the information needed to generate a report PDF.
type ReportData struct {
	// Title is the main report title.
	Title string
	// Subtitle is displayed below the title.
	Subtitle string
	// Author is the report author.
	Author string
	// Date is the report date.
	Date string
	// Sections is the list of content sections.
	Sections []ReportSection
}

// ReportSection represents a section within a report.
type ReportSection struct {
	// Title is the section heading.
	Title string
	// Content is the paragraph text for this section.
	Content string
	// Table is an optional data table for this section.
	Table *ReportTable
	// Metrics is an optional list of key metrics displayed in a grid.
	Metrics []ReportMetric
}

// ReportTable defines a simple data table within a report section.
type ReportTable struct {
	// Header is the list of column headers.
	Header []string
	// Rows is the table data.
	Rows [][]string
	// ColumnWidths is optional column width percentages.
	ColumnWidths []float64
}

// ReportMetric represents a single key metric displayed as a card.
type ReportMetric struct {
	// Label is the metric description (e.g., "Revenue").
	Label string
	// Value is the metric value (e.g., "$12.5M").
	Value string
	// ColorHex is the hex color for the value (e.g., 0x2E7D32). Zero uses default.
	ColorHex uint32
}

// Report creates a ready-to-generate report Document from the given data.
// Additional options (WithFont, WithPageSize, etc.) can customize the output.
func Report(data ReportData, opts ...Option) *Document {
	// Theme colors.
	primary := pdf.RGBHex(0x1A237E)
	accent := pdf.RGBHex(0x1565C0)
	stripe := pdf.RGBHex(0xF5F5F5)

	doc := New(append([]Option{
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(20))),
		WithMetadata(document.DocumentMetadata{
			Title:   data.Title,
			Author:  data.Author,
			Subject: data.Subtitle,
		}),
	}, opts...)...)

	// ── Header / Footer ──
	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(6, func(c *ColBuilder) {
				authorText := data.Author
				if authorText == "" {
					authorText = data.Title
				}
				c.Text(authorText, Bold(), FontSize(9), TextColor(accent))
			})
			r.Col(6, func(c *ColBuilder) {
				c.Text(data.Title, AlignRight(), FontSize(9), TextColor(pdf.Gray(0.5)))
			})
		})
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Line(LineColor(accent))
				c.Spacer(document.Mm(3))
			})
		})
	})

	doc.Footer(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Spacer(document.Mm(3))
				c.Line(LineColor(pdf.Gray(0.8)))
				c.Spacer(document.Mm(2))
			})
		})
		p.AutoRow(func(r *RowBuilder) {
			r.Col(6, func(c *ColBuilder) {
				c.Text("Confidential", FontSize(7), TextColor(pdf.Gray(0.5)))
			})
			r.Col(6, func(c *ColBuilder) {
				c.PageNumber(AlignRight(), FontSize(7), TextColor(pdf.Gray(0.5)))
			})
		})
	})

	// ── Title page (first page) ──
	page := doc.AddPage()

	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Spacer(document.Mm(20))
			c.Text(data.Title, FontSize(28), Bold(), AlignCenter(), TextColor(primary))
			if data.Subtitle != "" {
				c.Text(data.Subtitle, FontSize(16), AlignCenter(), TextColor(pdf.Gray(0.4)))
			}
			if data.Date != "" {
				c.Spacer(document.Mm(3))
				c.Text(data.Date, FontSize(11), AlignCenter(), TextColor(pdf.Gray(0.5)))
			}
			c.Spacer(document.Mm(10))
			c.Line(LineColor(primary), LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(10))
		})
	})

	// ── Sections ──
	for _, sec := range data.Sections {
		// Section title
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text(sec.Title, FontSize(16), Bold())
				c.Spacer(document.Mm(3))
			})
		})

		// Section text content
		if sec.Content != "" {
			page.AutoRow(func(r *RowBuilder) {
				r.Col(12, func(c *ColBuilder) {
					c.Text(sec.Content)
					c.Spacer(document.Mm(5))
				})
			})
		}

		// Metrics grid
		if len(sec.Metrics) > 0 {
			page.AutoRow(func(r *RowBuilder) {
				span := max(1, min(12, 12/len(sec.Metrics)))
				for _, m := range sec.Metrics {
					r.Col(span, func(c *ColBuilder) {
						c.Text(m.Label, TextColor(pdf.Gray(0.5)), FontSize(9))
						valueColor := accent
						if m.ColorHex != 0 {
							valueColor = pdf.RGBHex(m.ColorHex)
						}
						c.Text(m.Value, FontSize(18), Bold(), TextColor(valueColor))
					})
				}
			})
			page.AutoRow(func(r *RowBuilder) {
				r.Col(12, func(c *ColBuilder) {
					c.Spacer(document.Mm(5))
				})
			})
		}

		// Section table
		if sec.Table != nil {
			page.AutoRow(func(r *RowBuilder) {
				r.Col(12, func(c *ColBuilder) {
					var tblOpts []TableOption
					if len(sec.Table.ColumnWidths) > 0 {
						tblOpts = append(tblOpts, ColumnWidths(sec.Table.ColumnWidths...))
					}
					tblOpts = append(tblOpts,
						TableHeaderStyle(TextColor(pdf.White), BgColor(primary)),
						TableStripe(stripe),
					)
					c.Table(sec.Table.Header, sec.Table.Rows, tblOpts...)
					c.Spacer(document.Mm(8))
				})
			})
		}
	}

	return doc
}
