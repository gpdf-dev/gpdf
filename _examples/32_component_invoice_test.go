package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_32_ComponentInvoice(t *testing.T) {
	doc := template.Invoice(template.InvoiceData{
		Number:  "#INV-2026-001",
		Date:    "March 1, 2026",
		DueDate: "March 31, 2026",
		From: template.InvoiceParty{
			Name:    "ACME Corporation",
			Address: []string{"123 Business Street", "Suite 100", "San Francisco, CA 94105"},
		},
		To: template.InvoiceParty{
			Name:    "John Smith",
			Address: []string{"Tech Solutions Inc.", "456 Client Avenue", "New York, NY 10001"},
		},
		Items: []template.InvoiceItem{
			{Description: "Web Development - Frontend", Quantity: "40 hrs", UnitPrice: 150.00, Amount: 6000.00},
			{Description: "Web Development - Backend", Quantity: "60 hrs", UnitPrice: 150.00, Amount: 9000.00},
			{Description: "UI/UX Design", Quantity: "20 hrs", UnitPrice: 120.00, Amount: 2400.00},
			{Description: "Database Design", Quantity: "15 hrs", UnitPrice: 130.00, Amount: 1950.00},
			{Description: "QA Testing", Quantity: "25 hrs", UnitPrice: 100.00, Amount: 2500.00},
			{Description: "Project Management", Quantity: "10 hrs", UnitPrice: 140.00, Amount: 1400.00},
		},
		TaxRate:  10,
		Currency: "$",
		Payment: &template.InvoicePayment{
			BankName: "First National Bank",
			Account:  "1234-5678-9012",
			Routing:  "021000021",
		},
		Notes: "Thank you for your business!",
	})

	generatePDF(t, "32_component_invoice.pdf", doc)
}
