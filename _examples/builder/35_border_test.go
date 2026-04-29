package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// TestExample_35_Border exercises the template.Border / WithTableBorder /
// WithTableCellBorder / WithImageBorder API surface added for issue #22.
//
// The page contains four tables that exercise each border pattern in turn,
// followed by a bordered image. The same golden file is shared with the
// JSON and Go-template entry points to ensure the three input methods
// produce byte-identical output.
func TestExample_35_Border(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	header := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"Alice", "30", "Tokyo"},
		{"Bob", "25", "New York"},
		{"Charlie", "35", "London"},
	}

	addLabel := func(label string) {
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text(label, template.FontSize(11), template.Bold())
				c.Spacer(document.Mm(2))
			})
		})
	}

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Borders & backgrounds", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	outerBorder := template.Border(
		template.BorderWidth(document.Pt(1)),
		template.BorderColor(pdf.RGBHex(0x1A237E)),
	)
	cellBorder := template.Border(
		template.BorderWidth(document.Pt(0.5)),
		template.BorderColor(pdf.Gray(0.5)),
	)
	dashedCellBorder := template.Border(
		template.BorderWidth(document.Pt(0.75)),
		template.BorderColor(pdf.RGBHex(0x0D47A1)),
		template.BorderLine(document.BorderDashed),
	)

	// Pattern A — outer frame only.
	addLabel("A. Outer border only")
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(header, rows,
				template.WithTableBorder(outerBorder),
				template.WithTableBackground(pdf.RGBHex(0xF5F5F5)),
				template.ColumnWidths(40, 20, 40),
			)
			c.Spacer(document.Mm(5))
		})
	})

	// Pattern B — cell grid only.
	addLabel("B. Cell grid only")
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(header, rows,
				template.WithTableCellBorder(cellBorder),
				template.ColumnWidths(40, 20, 40),
			)
			c.Spacer(document.Mm(5))
		})
	})

	// Pattern C — outer frame + cell grid (Excel-style).
	addLabel("C. Outer frame + cell grid")
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(header, rows,
				template.WithTableBorder(outerBorder),
				template.WithTableCellBorder(cellBorder),
				template.WithTableBackground(pdf.RGBHex(0xFAFAFA)),
				template.ColumnWidths(40, 20, 40),
			)
			c.Spacer(document.Mm(5))
		})
	})

	// Pattern D — dashed cell border with row stripes.
	addLabel("D. Dashed cell grid with stripes")
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(header, rows,
				template.WithTableCellBorder(dashedCellBorder),
				template.TableStripe(pdf.RGBHex(0xE3F2FD)),
				template.ColumnWidths(40, 20, 40),
			)
			c.Spacer(document.Mm(5))
		})
	})

	// Bordered image.
	imageBorder := template.Border(
		template.BorderWidths(
			document.Pt(2),
			document.Pt(2),
			document.Pt(2),
			document.Pt(2),
		),
		template.BorderColor(pdf.RGBHex(0xE53935)),
		template.BorderLine(document.BorderSolid),
	)

	pngData := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})

	addLabel("E. Image with border + background")
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Image(
				pngData,
				template.FitWidth(document.Mm(60)),
				template.WithImageBorder(imageBorder),
				template.WithImageBackground(pdf.RGBHex(0xFFF8E1)),
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "35_border.pdf", doc)
}

// TestExample_35b_BoxBorder demonstrates the new ColBuilder.Box container with
// border, background, and padding. Builder-only because the JSON / Go-template
// schema does not yet expose a "box" element.
func TestExample_35b_BoxBorder(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Box container", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("This paragraph lives inside a Box with a 1pt border, a", template.FontSize(11))
				c.Text("light fill, and 4mm of padding on every side.", template.FontSize(11))
			},
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(1)),
					template.BorderColor(pdf.RGBHex(0x1A237E)),
				)),
				template.WithBoxBackground(pdf.RGBHex(0xE8EAF6)),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
			)
		})
	})

	testutil.GeneratePDF(t, "35_box_border.pdf", doc)
}
