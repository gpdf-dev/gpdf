package encrypt_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/encrypt"
	"github.com/gpdf-dev/gpdf/template"
)

// TestEncrypt_03_Permissions creates PDFs with different permission settings.
func TestEncrypt_03_Permissions(t *testing.T) {
	t.Run("PrintAndCopyOnly", func(t *testing.T) {
		doc := gpdf.NewDocument(
			gpdf.WithPageSize(document.A4),
			gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
			gpdf.WithEncryption(
				encrypt.WithOwnerPassword("owner"),
				encrypt.WithUserPassword("user"),
				encrypt.WithPermissions(encrypt.PermPrint|encrypt.PermCopy|encrypt.PermPrintHighRes),
			),
		)

		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("Print & Copy Only", template.FontSize(24))
			})
		})
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("This PDF allows printing and copying, but not modification.", template.FontSize(12))
			})
		})

		data, err := doc.Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		testutil.AssertValidPDF(t, data)
		testutil.WritePDF(t, "encrypt_03_print_copy_only.pdf", data)

		info, err := encrypt.ParseEncryptInfo(data)
		if err != nil {
			t.Fatalf("ParseEncryptInfo failed: %v", err)
		}

		// Allowed permissions
		if !info.HasPermission(encrypt.PermPrint) {
			t.Error("PermPrint should be allowed")
		}
		if !info.HasPermission(encrypt.PermCopy) {
			t.Error("PermCopy should be allowed")
		}
		if !info.HasPermission(encrypt.PermPrintHighRes) {
			t.Error("PermPrintHighRes should be allowed")
		}

		// Denied permissions
		if info.HasPermission(encrypt.PermModify) {
			t.Error("PermModify should be denied")
		}
		if info.HasPermission(encrypt.PermAnnotate) {
			t.Error("PermAnnotate should be denied")
		}
	})

	t.Run("NoPermissions", func(t *testing.T) {
		doc := gpdf.NewDocument(
			gpdf.WithPageSize(document.A4),
			gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
			gpdf.WithEncryption(
				encrypt.WithOwnerPassword("owner"),
				encrypt.WithUserPassword("user"),
				encrypt.WithPermissions(0),
			),
		)

		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("No Permissions (View Only)", template.FontSize(24))
			})
		})

		data, err := doc.Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		testutil.AssertValidPDF(t, data)
		testutil.WritePDF(t, "encrypt_03_no_permissions.pdf", data)

		info, err := encrypt.ParseEncryptInfo(data)
		if err != nil {
			t.Fatalf("ParseEncryptInfo failed: %v", err)
		}

		if info.HasPermission(encrypt.PermPrint) {
			t.Error("PermPrint should be denied")
		}
		if info.HasPermission(encrypt.PermCopy) {
			t.Error("PermCopy should be denied")
		}
		if info.HasPermission(encrypt.PermModify) {
			t.Error("PermModify should be denied")
		}
	})
}
