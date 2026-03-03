package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_19_Report(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:   "Quarterly Report Q1 2026",
			Author:  "ACME Corporation",
			Subject: "Q1 2026 Financial Summary",
		}),
	)

	// Header for all pages
	doc.Header(func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(6, func(c *template.ColBuilder) {
				c.Text("ACME Corp", template.Bold(), template.FontSize(9),
					template.TextColor(pdf.RGBHex(0x1565C0)))
			})
			r.Col(6, func(c *template.ColBuilder) {
				c.Text("Q1 2026 Report", template.AlignRight(), template.FontSize(9),
					template.TextColor(pdf.Gray(0.5)))
			})
		})
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Line(template.LineColor(pdf.RGBHex(0x1565C0)))
				c.Spacer(document.Mm(3))
			})
		})
	})

	// Footer for all pages
	doc.Footer(func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Spacer(document.Mm(3))
				c.Line(template.LineColor(pdf.Gray(0.8)))
				c.Spacer(document.Mm(2))
			})
		})
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("Confidential - For Internal Use Only",
					template.AlignCenter(), template.FontSize(7), template.TextColor(pdf.Gray(0.5)))
			})
		})
	})

	// --- Page 1: Title & Executive Summary ---
	page1 := doc.AddPage()

	page1.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(20))
			c.Text("Quarterly Report", template.FontSize(28), template.Bold(),
				template.AlignCenter(), template.TextColor(pdf.RGBHex(0x1A237E)))
			c.Text("Q1 2026 - Financial Summary", template.FontSize(16),
				template.AlignCenter(), template.TextColor(pdf.Gray(0.4)))
			c.Spacer(document.Mm(15))
			c.Line(template.LineColor(pdf.RGBHex(0x1A237E)), template.LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(10))
		})
	})

	page1.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Executive Summary", template.FontSize(16), template.Bold())
			c.Spacer(document.Mm(3))
			c.Text("This report presents the financial performance of ACME Corporation " +
				"for the first quarter of 2026. Revenue increased by 15% compared to Q4 2025, " +
				"driven primarily by strong growth in the cloud services division. " +
				"Operating margins improved to 22%, up from 19% in the previous quarter.")
			c.Spacer(document.Mm(5))
			c.Text("Key highlights include the successful launch of three new product lines, " +
				"expansion into the European market, and a 20% reduction in customer churn rate. " +
				"The company remains well-positioned for continued growth throughout 2026.")
		})
	})

	// Key metrics in grid
	page1.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(10))
			c.Text("Key Metrics", template.FontSize(14), template.Bold())
			c.Spacer(document.Mm(3))
		})
	})

	page1.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Revenue", template.TextColor(pdf.Gray(0.5)), template.FontSize(9))
			c.Text("$12.5M", template.FontSize(18), template.Bold(),
				template.TextColor(pdf.RGBHex(0x2E7D32)))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Growth", template.TextColor(pdf.Gray(0.5)), template.FontSize(9))
			c.Text("+15%", template.FontSize(18), template.Bold(),
				template.TextColor(pdf.RGBHex(0x2E7D32)))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Customers", template.TextColor(pdf.Gray(0.5)), template.FontSize(9))
			c.Text("2,450", template.FontSize(18), template.Bold(),
				template.TextColor(pdf.RGBHex(0x1565C0)))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Margin", template.TextColor(pdf.Gray(0.5)), template.FontSize(9))
			c.Text("22%", template.FontSize(18), template.Bold(),
				template.TextColor(pdf.RGBHex(0x1565C0)))
		})
	})

	// --- Page 2: Financial Details ---
	page2 := doc.AddPage()

	page2.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Revenue Breakdown", template.FontSize(16), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	darkHeader := pdf.RGBHex(0x1A237E)
	stripe := pdf.RGBHex(0xF5F5F5)

	page2.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Division", "Q1 2026", "Q4 2025", "Change"},
				[][]string{
					{"Cloud Services", "$5,200,000", "$4,100,000", "+26.8%"},
					{"Enterprise Software", "$3,800,000", "$3,500,000", "+8.6%"},
					{"Consulting", "$2,100,000", "$1,900,000", "+10.5%"},
					{"Support & Maintenance", "$1,400,000", "$1,350,000", "+3.7%"},
				},
				template.ColumnWidths(35, 22, 22, 21),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(darkHeader),
				),
				template.TableStripe(stripe),
			)
			c.Spacer(document.Mm(10))
		})
	})

	page2.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Expense Summary", template.FontSize(16), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	page2.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Category", "Amount", "% of Revenue"},
				[][]string{
					{"Personnel", "$5,500,000", "44.0%"},
					{"Infrastructure", "$1,800,000", "14.4%"},
					{"Marketing", "$1,200,000", "9.6%"},
					{"R&D", "$950,000", "7.6%"},
					{"General & Admin", "$300,000", "2.4%"},
				},
				template.ColumnWidths(40, 30, 30),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(darkHeader),
				),
				template.TableStripe(stripe),
			)
			c.Spacer(document.Mm(10))
		})
	})

	// Two-column commentary
	page2.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Highlights", template.Bold(), template.TextColor(pdf.RGBHex(0x2E7D32)))
			c.Spacer(document.Mm(2))
			c.Text("Cloud services revenue grew 26.8%, exceeding projections by 5%. " +
				"New enterprise clients added: 47.")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Challenges", template.Bold(), template.TextColor(pdf.RGBHex(0xC62828)))
			c.Spacer(document.Mm(2))
			c.Text("Infrastructure costs rose 12% due to scaling needs. " +
				"Two major client renewals deferred to Q2.")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "19_report.pdf", doc)
}
