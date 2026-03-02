package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_34_ComponentLetter(t *testing.T) {
	doc := template.Letter(template.LetterData{
		From: template.LetterParty{
			Name:    "ACME Corporation",
			Address: []string{"123 Business Street", "Suite 100", "San Francisco, CA 94105", "contact@acme.com"},
		},
		To: template.LetterParty{
			Name:    "Mr. John Smith",
			Address: []string{"Tech Solutions Inc.", "456 Client Avenue", "New York, NY 10001"},
		},
		Date:     "March 1, 2026",
		Subject:  "Partnership Proposal",
		Greeting: "Dear Mr. Smith,",
		Body: []string{
			"I am writing to express our interest in establishing a strategic partnership " +
				"between ACME Corporation and Tech Solutions Inc. Over the past year, we have " +
				"observed the remarkable growth of your organization and believe that a collaboration " +
				"would be mutually beneficial.",
			"Our proposal includes joint development of cloud-based solutions targeting " +
				"the enterprise market. ACME Corporation brings extensive experience in PDF " +
				"generation and document processing, while Tech Solutions Inc. has demonstrated " +
				"excellence in frontend technologies and user experience design.",
			"We would like to schedule a meeting at your earliest convenience to discuss " +
				"the details of this proposal. Please feel free to contact me directly at " +
				"ceo@acme.com or call our office at (415) 555-0100.",
		},
		Closing:     "Sincerely,",
		Signature:   "Jane Doe",
		SignerTitle: "Chief Executive Officer",
	})

	generatePDF(t, "34_component_letter.pdf", doc)
}
