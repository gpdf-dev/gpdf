// Package encrypt provides AES-256 PDF encryption (ISO 32000-2, Rev 6).
package encrypt

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/gpdf-dev/gpdf/pdf"
)

// Option configures encryption.
type Option func(*encryptConfig)

type encryptConfig struct {
	ownerPassword string
	userPassword  string
	permissions   Permission
}

// WithOwnerPassword sets the owner password.
func WithOwnerPassword(pw string) Option {
	return func(c *encryptConfig) { c.ownerPassword = pw }
}

// WithUserPassword sets the user password.
func WithUserPassword(pw string) Option {
	return func(c *encryptConfig) { c.userPassword = pw }
}

// WithPermissions sets the document permissions.
func WithPermissions(perm Permission) Option {
	return func(c *encryptConfig) { c.permissions = perm }
}

// Apply configures a pdf.Writer for AES-256 encryption (ISO 32000-2, Rev 6).
func Apply(pw *pdf.Writer, opts ...Option) error {
	cfg := &encryptConfig{
		permissions: PermAll,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.ownerPassword == "" && cfg.userPassword == "" {
		return fmt.Errorf("encrypt: at least one password is required")
	}
	if cfg.ownerPassword == "" {
		cfg.ownerPassword = cfg.userPassword
	}

	// Generate encryption key
	fileKey, err := generateEncryptionKey()
	if err != nil {
		return err
	}

	// Compute U, UE
	u, ue, err := computeU(fileKey, cfg.userPassword)
	if err != nil {
		return fmt.Errorf("encrypt: compute U: %w", err)
	}

	// Compute O, OE
	o, oe, err := computeO(fileKey, cfg.ownerPassword, u)
	if err != nil {
		return fmt.Errorf("encrypt: compute O: %w", err)
	}

	// Compute Perms
	perms, err := computePerms(fileKey, cfg.permissions, true)
	if err != nil {
		return fmt.Errorf("encrypt: compute Perms: %w", err)
	}

	// Generate document ID
	docID := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, docID); err != nil {
		return fmt.Errorf("encrypt: document ID: %w", err)
	}

	// Pre-allocate encrypt dict ref
	h := &handler{
		encryptionKey: fileKey,
	}

	// Set up object hook for encryption
	pw.SetObjectHook(h.encryptObject)

	pw.OnBeforeClose(func(pw *pdf.Writer) error {
		// Write Encrypt dictionary
		h.encryptRef = pw.AllocObject()
		encryptDict := pdf.Dict{
			pdf.Name("Filter"): pdf.Name("Standard"),
			pdf.Name("V"):      pdf.Integer(5),
			pdf.Name("R"):      pdf.Integer(6),
			pdf.Name("Length"):  pdf.Integer(256),
			pdf.Name("CF"): pdf.Dict{
				pdf.Name("StdCF"): pdf.Dict{
					pdf.Name("Type"):   pdf.Name("CryptFilter"),
					pdf.Name("CFM"):    pdf.Name("AESV3"),
					pdf.Name("Length"): pdf.Integer(32),
				},
			},
			pdf.Name("StmF"):            pdf.Name("StdCF"),
			pdf.Name("StrF"):            pdf.Name("StdCF"),
			pdf.Name("O"):               pdf.HexString(o),
			pdf.Name("U"):               pdf.HexString(u),
			pdf.Name("OE"):              pdf.HexString(oe),
			pdf.Name("UE"):              pdf.HexString(ue),
			pdf.Name("P"):               pdf.Integer(int32(uint32(cfg.permissions) | 0xFFFFF000)),
			pdf.Name("Perms"):           pdf.HexString(perms),
			pdf.Name("EncryptMetadata"): pdf.Boolean(true),
		}

		// Temporarily disable hook to write Encrypt dict unencrypted,
		// then restore for remaining Close() objects (page tree, catalog, info).
		pw.SetObjectHook(nil)
		if err := pw.WriteObject(h.encryptRef, encryptDict); err != nil {
			return err
		}
		pw.SetObjectHook(h.encryptObject)

		// Add to trailer
		pw.AddTrailerEntry(pdf.Name("Encrypt"), h.encryptRef)
		pw.AddTrailerEntry(pdf.Name("ID"), pdf.Array{
			pdf.HexString(docID),
			pdf.HexString(docID),
		})

		return nil
	})

	return nil
}
