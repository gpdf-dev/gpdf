package encrypt_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/encrypt"
	"github.com/gpdf-dev/gpdf/template"
)

// TestEncrypt_01_OwnerPassword creates a PDF with owner password only.
// The PDF can be opened without a password but editing is restricted.
func TestEncrypt_01_OwnerPassword(t *testing.T) {
	doc := gpdf.NewDocument(
		gpdf.WithPageSize(document.A4),
		gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
		gpdf.WithEncryption(
			encrypt.WithOwnerPassword("owner-secret-123"),
		),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Owner Password Protected PDF", template.FontSize(24))
		})
	})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("This PDF has an owner password. It can be opened without a password, but editing requires the owner password.", template.FontSize(12))
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	testutil.AssertValidPDF(t, data)
	testutil.WritePDF(t, "encrypt_01_owner_password.pdf", data)

	// Verify encryption was applied
	info, err := encrypt.ParseEncryptInfo(data)
	if err != nil {
		t.Fatalf("ParseEncryptInfo failed: %v", err)
	}
	if info.V != 5 {
		t.Errorf("V = %d, want 5", info.V)
	}
	if info.R != 6 {
		t.Errorf("R = %d, want 6", info.R)
	}
	// Owner password should verify
	if !info.VerifyOwnerPassword("owner-secret-123") {
		t.Error("owner password verification failed")
	}
}
