package template

import (
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// LetterData holds all the information needed to generate a business letter PDF.
type LetterData struct {
	// From is the sender information.
	From LetterParty
	// To is the recipient information.
	To LetterParty
	// Date is the letter date (e.g., "March 1, 2026").
	Date string
	// Subject is the letter subject line. Empty means no subject.
	Subject string
	// Greeting is the opening salutation (e.g., "Dear Mr. Smith,").
	Greeting string
	// Body is the list of paragraphs composing the letter body.
	Body []string
	// Closing is the valediction (e.g., "Sincerely,").
	Closing string
	// Signature is the signer's name.
	Signature string
	// SignerTitle is the signer's title or role (e.g., "CEO").
	SignerTitle string
}

// LetterParty represents a party (sender or recipient) on a business letter.
type LetterParty struct {
	// Name is the person or company name.
	Name string
	// Address is a list of address lines.
	Address []string
}

// Letter creates a ready-to-generate business letter Document from the given data.
// Additional options (WithFont, WithPageSize, etc.) can customize the output.
func Letter(data LetterData, opts ...Option) *Document {
	primary := pdf.RGBHex(0x1A237E)

	doc := New(append([]Option{
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(25))),
		WithMetadata(document.DocumentMetadata{
			Title:  data.Subject,
			Author: data.From.Name,
		}),
	}, opts...)...)

	page := doc.AddPage()

	// ── Sender header ──
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text(data.From.Name, FontSize(16), Bold(), TextColor(primary))
			for _, line := range data.From.Address {
				c.Text(line, FontSize(10), TextColor(pdf.Gray(0.4)))
			}
			c.Spacer(document.Mm(5))
			c.Line(LineColor(primary), LineThickness(document.Pt(1)))
			c.Spacer(document.Mm(10))
		})
	})

	// ── Date ──
	if data.Date != "" {
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text(data.Date, AlignRight())
				c.Spacer(document.Mm(10))
			})
		})
	}

	// ── Recipient ──
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text(data.To.Name, Bold())
			for _, line := range data.To.Address {
				c.Text(line)
			}
			c.Spacer(document.Mm(10))
		})
	})

	// ── Subject ──
	if data.Subject != "" {
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("Re: "+data.Subject, Bold())
				c.Spacer(document.Mm(8))
			})
		})
	}

	// ── Greeting ──
	if data.Greeting != "" {
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text(data.Greeting)
				c.Spacer(document.Mm(5))
			})
		})
	}

	// ── Body paragraphs ──
	for _, para := range data.Body {
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text(para)
				c.Spacer(document.Mm(4))
			})
		})
	}

	// ── Closing ──
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Spacer(document.Mm(8))
			if data.Closing != "" {
				c.Text(data.Closing)
			}
			c.Spacer(document.Mm(15))
			if data.Signature != "" {
				c.Text(data.Signature, Bold())
			}
			if data.SignerTitle != "" {
				c.Text(data.SignerTitle, TextColor(pdf.Gray(0.4)))
			}
		})
	})

	return doc
}
