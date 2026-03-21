package encrypt_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/encrypt"
	"github.com/gpdf-dev/gpdf/template"
)

// TestEncrypt_02_UserOwnerPassword creates a PDF with both user and owner passwords.
// The user password is required to open the PDF.
func TestEncrypt_02_UserOwnerPassword(t *testing.T) {
	doc := gpdf.NewDocument(
		gpdf.WithPageSize(document.A4),
		gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
		gpdf.WithEncryption(
			encrypt.WithOwnerPassword("owner-pass"),
			encrypt.WithUserPassword("user-pass"),
		),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("User & Owner Password Protected PDF", template.FontSize(24))
		})
	})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("This PDF requires a password to open. User password: 'user-pass', Owner password: 'owner-pass'.", template.FontSize(12))
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	testutil.AssertValidPDF(t, data)
	testutil.WritePDF(t, "encrypt_02_user_owner_password.pdf", data)

	// Verify both passwords
	info, err := encrypt.ParseEncryptInfo(data)
	if err != nil {
		t.Fatalf("ParseEncryptInfo failed: %v", err)
	}
	if !info.VerifyUserPassword("user-pass") {
		t.Error("user password verification failed")
	}
	if !info.VerifyOwnerPassword("owner-pass") {
		t.Error("owner password verification failed")
	}

	// Wrong passwords should fail
	if info.VerifyUserPassword("wrong") {
		t.Error("wrong user password should not verify")
	}
	if info.VerifyOwnerPassword("wrong") {
		t.Error("wrong owner password should not verify")
	}

	// Decrypt file key and verify Perms
	fileKey, err := info.DecryptFileKey("user-pass")
	if err != nil {
		t.Fatalf("DecryptFileKey failed: %v", err)
	}
	if err := info.VerifyPerms(fileKey); err != nil {
		t.Fatalf("VerifyPerms failed: %v", err)
	}
}
