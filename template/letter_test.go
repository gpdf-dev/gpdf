package template

import (
	"testing"
)

func TestLetter_Generate(t *testing.T) {
	doc := Letter(LetterData{
		From: LetterParty{
			Name:    "ACME Corp",
			Address: []string{"123 Main St", "City, ST 00000"},
		},
		To: LetterParty{
			Name:    "Mr. John Smith",
			Address: []string{"456 Side St", "Other City, ST 11111"},
		},
		Date:     "March 1, 2026",
		Subject:  "Test Subject",
		Greeting: "Dear Mr. Smith,",
		Body: []string{
			"First paragraph of the letter.",
			"Second paragraph with more details.",
		},
		Closing:     "Sincerely,",
		Signature:   "Jane Doe",
		SignerTitle: "CEO",
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Letter.Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Letter generated empty PDF")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}

func TestLetter_MinimalData(t *testing.T) {
	doc := Letter(LetterData{
		From: LetterParty{Name: "Sender"},
		To:   LetterParty{Name: "Recipient"},
		Body: []string{"Hello."},
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Letter.Generate failed: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("Invalid PDF header: %q", string(data[:5]))
	}
}
