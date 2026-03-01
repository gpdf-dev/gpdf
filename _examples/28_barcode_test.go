package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_28_Barcode(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "Barcode Examples",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Barcode Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Basic barcode
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Code 128 barcode:")
			c.Spacer(document.Mm(2))
			c.Barcode("INV-2026-0001")
			c.Spacer(document.Mm(5))
		})
	})

	// Barcodes with explicit sizes
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("With display width (80mm):")
			c.Spacer(document.Mm(2))
			c.Barcode("PRODUCT-A-12345", template.BarcodeWidth(document.Mm(80)))
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("With display height (10mm):")
			c.Spacer(document.Mm(2))
			c.Barcode("SMALL-BAR", template.BarcodeHeight(document.Mm(10)))
			c.Spacer(document.Mm(5))
		})
	})

	// Numeric data (optimized with Code C)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Numeric data (Code C optimization):")
			c.Spacer(document.Mm(2))
			c.Barcode("1234567890")
			c.Spacer(document.Mm(5))
		})
	})

	// Barcodes in columns
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Barcodes in columns:")
			c.Spacer(document.Mm(2))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Item A", template.FontSize(9))
			c.Barcode("ITEM-A-001", template.BarcodeWidth(document.Mm(60)))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Item B", template.FontSize(9))
			c.Barcode("ITEM-B-002", template.BarcodeWidth(document.Mm(60)))
		})
	})

	generatePDF(t, "28_barcode.pdf", doc)
}
