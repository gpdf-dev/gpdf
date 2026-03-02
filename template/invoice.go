package template

import (
	"fmt"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// InvoiceData holds all the information needed to generate an invoice PDF.
type InvoiceData struct {
	// Number is the invoice identifier (e.g., "INV-2026-001").
	Number string
	// Date is the invoice issue date (e.g., "March 1, 2026").
	Date string
	// DueDate is the payment due date (e.g., "March 31, 2026").
	DueDate string
	// From is the sender/seller information.
	From InvoiceParty
	// To is the recipient/buyer information.
	To InvoiceParty
	// Items is the list of line items.
	Items []InvoiceItem
	// TaxRate is the tax percentage (e.g., 10 for 10%). Zero means no tax.
	TaxRate float64
	// Currency is the currency symbol (e.g., "$", "EUR"). Defaults to "$".
	Currency string
	// Notes is optional text displayed at the bottom of the invoice.
	Notes string
	// Payment holds optional payment information.
	Payment *InvoicePayment
}

// InvoiceParty represents a party (sender or recipient) on an invoice.
type InvoiceParty struct {
	// Name is the company or person name.
	Name string
	// Address is a list of address lines.
	Address []string
}

// InvoiceItem represents a single line item on an invoice.
type InvoiceItem struct {
	// Description is the item description.
	Description string
	// Quantity is the quantity string (e.g., "40 hrs", "5").
	Quantity string
	// UnitPrice is the price per unit.
	UnitPrice float64
	// Amount is the total for this line item. If zero, it is calculated
	// as UnitPrice * quantity (parsed as float, or left as zero).
	Amount float64
}

// InvoicePayment holds payment information displayed on the invoice.
type InvoicePayment struct {
	// BankName is the bank name.
	BankName string
	// Account is the account number.
	Account string
	// Routing is the routing number.
	Routing string
}

// Invoice creates a ready-to-generate invoice Document from the given data.
// Additional options (WithFont, WithPageSize, etc.) can customize the output.
func Invoice(data InvoiceData, opts ...Option) *Document {
	currency := data.Currency
	if currency == "" {
		currency = "$"
	}

	// Theme colors.
	primary := pdf.RGBHex(0x1A237E)
	stripe := pdf.RGBHex(0xF5F5F5)
	muted := pdf.Gray(0.4)

	doc := New(append([]Option{
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(20))),
		WithMetadata(document.DocumentMetadata{
			Title:  "Invoice " + data.Number,
			Author: data.From.Name,
		}),
	}, opts...)...)

	page := doc.AddPage()

	// ── Company header + Invoice title ──
	page.AutoRow(func(r *RowBuilder) {
		r.Col(6, func(c *ColBuilder) {
			c.Text(data.From.Name, FontSize(20), Bold(), TextColor(primary))
			for _, line := range data.From.Address {
				c.Text(line)
			}
		})
		r.Col(6, func(c *ColBuilder) {
			c.Text("INVOICE", FontSize(24), Bold(), AlignRight(), TextColor(primary))
			c.Spacer(document.Mm(3))
			c.Text(data.Number, AlignRight(), FontSize(12))
			if data.Date != "" {
				c.Text("Date: "+data.Date, AlignRight())
			}
			if data.DueDate != "" {
				c.Text("Due: "+data.DueDate, AlignRight())
			}
		})
	})

	// ── Separator ──
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Spacer(document.Mm(5))
			c.Line(LineColor(primary), LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(5))
		})
	})

	// ── Bill To + Payment Info ──
	page.AutoRow(func(r *RowBuilder) {
		r.Col(6, func(c *ColBuilder) {
			c.Text("Bill To:", Bold(), TextColor(muted))
			c.Spacer(document.Mm(2))
			c.Text(data.To.Name, Bold())
			for _, line := range data.To.Address {
				c.Text(line)
			}
		})
		r.Col(6, func(c *ColBuilder) {
			if data.Payment != nil {
				c.Text("Payment Info:", Bold(), TextColor(muted))
				c.Spacer(document.Mm(2))
				if data.Payment.BankName != "" {
					c.Text("Bank: " + data.Payment.BankName)
				}
				if data.Payment.Account != "" {
					c.Text("Account: " + data.Payment.Account)
				}
				if data.Payment.Routing != "" {
					c.Text("Routing: " + data.Payment.Routing)
				}
			}
		})
	})

	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Spacer(document.Mm(10))
		})
	})

	// ── Items table ──
	header := []string{"Description", "Qty", "Unit Price", "Amount"}
	rows := make([][]string, len(data.Items))
	var subtotal float64
	for i, item := range data.Items {
		amount := item.Amount
		if amount == 0 {
			amount = item.UnitPrice // fallback
		}
		subtotal += amount
		rows[i] = []string{
			item.Description,
			item.Quantity,
			formatMoney(currency, item.UnitPrice),
			formatMoney(currency, amount),
		}
	}

	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Table(header, rows,
				ColumnWidths(40, 15, 20, 25),
				TableHeaderStyle(
					TextColor(pdf.White),
					BgColor(primary),
				),
				TableStripe(stripe),
			)
		})
	})

	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// ── Totals ──
	tax := subtotal * data.TaxRate / 100
	total := subtotal + tax

	page.AutoRow(func(r *RowBuilder) {
		r.Col(8, func(c *ColBuilder) {})
		r.Col(4, func(c *ColBuilder) {
			c.Text("Subtotal:    "+formatMoney(currency, subtotal), AlignRight())
			if data.TaxRate > 0 {
				c.Text(fmt.Sprintf("Tax (%.0f%%):    %s", data.TaxRate, formatMoney(currency, tax)), AlignRight())
			}
			c.Spacer(document.Mm(2))
			c.Line(LineThickness(document.Pt(1)))
			c.Spacer(document.Mm(2))
			c.Text("Total:       "+formatMoney(currency, total), AlignRight(), Bold(), FontSize(14))
		})
	})

	// ── Notes / Footer ──
	if data.Notes != "" {
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Spacer(document.Mm(15))
				c.Line(LineColor(pdf.Gray(0.8)))
				c.Spacer(document.Mm(3))
				c.Text(data.Notes, AlignCenter(), Italic(), TextColor(pdf.Gray(0.5)))
			})
		})
	}

	return doc
}

// formatMoney formats a float as a monetary value with the given currency symbol.
func formatMoney(currency string, amount float64) string {
	return fmt.Sprintf("%s%.2f", currency, amount)
}
