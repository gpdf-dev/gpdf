package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_12_SVG(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG Image Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Basic filled rectangle SVG.
	blueSVG := testutil.TestImageSVG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG rectangle (blue):")
			c.Spacer(document.Mm(2))
			c.Image(blueSVG, template.FitWidth(document.Mm(80)))
			c.Spacer(document.Mm(5))
		})
	})

	// SVG with circle and stroke.
	circleSVG := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<circle cx="50" cy="50" r="45" fill="#ea4335" stroke="#333333" stroke-width="3"/>
	</svg>`)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG circle with stroke (red):")
			c.Spacer(document.Mm(2))
			c.Image(circleSVG, template.FitWidth(document.Mm(50)))
			c.Spacer(document.Mm(5))
		})
	})

	// SVG with path data (triangle).
	triangleSVG := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<path d="M 50 5 L 95 95 L 5 95 Z" fill="#34a853" stroke="#1a5c2a" stroke-width="2"/>
	</svg>`)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG path triangle (green):")
			c.Spacer(document.Mm(2))
			c.Image(triangleSVG, template.FitWidth(document.Mm(50)))
			c.Spacer(document.Mm(5))
		})
	})

	// SVG with multiple shapes and groups.
	complexSVG := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 80">
		<rect x="0" y="0" width="200" height="80" fill="#f8f9fa"/>
		<g transform="translate(10,10)">
			<rect x="0" y="0" width="60" height="60" rx="8" ry="8" fill="#4285f4"/>
			<rect x="70" y="0" width="60" height="60" rx="8" ry="8" fill="#ea4335"/>
			<rect x="140" y="0" width="60" height="60" rx="8" ry="8" fill="#34a853"/>
		</g>
	</svg>`)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG group with rounded rectangles:")
			c.Spacer(document.Mm(2))
			c.Image(complexSVG, template.FitWidth(document.Mm(100)))
			c.Spacer(document.Mm(5))
		})
	})

	// SVG with XML declaration.
	xmlSVG := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">
	<ellipse cx="50" cy="25" rx="45" ry="20" fill="#fbbc05"/>
</svg>`)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG with XML declaration (ellipse, yellow):")
			c.Spacer(document.Mm(2))
			c.Image(xmlSVG, template.FitWidth(document.Mm(80)))
			c.Spacer(document.Mm(5))
		})
	})

	// Two SVGs side by side (deduplication test: same SVG used twice).
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Same SVG used twice (deduplication):")
			c.Spacer(document.Mm(2))
			c.Image(circleSVG, template.FitWidth(document.Mm(40)))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("(second reference)")
			c.Spacer(document.Mm(2))
			c.Image(circleSVG, template.FitWidth(document.Mm(40)))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "12_svg.pdf", doc)
}
