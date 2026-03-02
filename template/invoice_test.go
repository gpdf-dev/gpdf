package template

import (
	"testing"
)

func TestInvoice_Generate(t *testing.T) {
	doc := Invoice(InvoiceData{
		Number:  "#INV-001",
		Date:    "2026-03-01",
		DueDate: "2026-03-31",
		From: InvoiceParty{
			Name:    "Seller Inc.",
			Address: []string{"123 Main St", "City, ST 00000"},
		},
		To: InvoiceParty{
			Name:    "Buyer Corp.",
			Address: []string{"456 Side St"},
		},
		Items: []InvoiceItem{
			{Description: "Service A", Quantity: "10", UnitPrice: 100, Amount: 1000},
			{Description: "Service B", Quantity: "5", UnitPrice: 200, Amount: 1000},
		},
		TaxRate:  10,
		Currency: "$",
		Notes:    "Thank you",
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Invoice.Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Invoice generated empty PDF")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

func TestInvoice_NoTax(t *testing.T) {
	doc := Invoice(InvoiceData{
		Number: "#INV-002",
		From:   InvoiceParty{Name: "Seller"},
		To:     InvoiceParty{Name: "Buyer"},
		Items: []InvoiceItem{
			{Description: "Item", Quantity: "1", UnitPrice: 500, Amount: 500},
		},
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Invoice.Generate failed: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

func TestInvoice_WithPayment(t *testing.T) {
	doc := Invoice(InvoiceData{
		Number: "#INV-003",
		From:   InvoiceParty{Name: "Seller"},
		To:     InvoiceParty{Name: "Buyer"},
		Items: []InvoiceItem{
			{Description: "Item", Quantity: "1", UnitPrice: 100, Amount: 100},
		},
		Payment: &InvoicePayment{
			BankName: "Test Bank",
			Account:  "12345",
			Routing:  "67890",
		},
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Invoice.Generate failed: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		currency string
		amount   float64
		want     string
	}{
		{"$", 1234.50, "$1234.50"},
		{"EUR", 99.99, "EUR99.99"},
		{"$", 0, "$0.00"},
	}
	for _, tt := range tests {
		got := formatMoney(tt.currency, tt.amount)
		if got != tt.want {
			t.Errorf("formatMoney(%q, %v) = %q, want %q", tt.currency, tt.amount, got, tt.want)
		}
	}
}
