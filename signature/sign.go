// Package signature provides CMS/PKCS#7 digital signatures for PDF documents.
package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"
)

// Signer represents a signing identity (certificate + private key).
type Signer struct {
	Certificate *x509.Certificate
	PrivateKey  crypto.PrivateKey
	Chain       []*x509.Certificate // intermediate certificates (optional)
}

// Option configures the signing process.
type Option func(*signConfig)

type signConfig struct {
	reason   string
	location string
	tsaURL   string
	signTime time.Time
}

// WithReason sets the reason for signing.
func WithReason(reason string) Option {
	return func(c *signConfig) { c.reason = reason }
}

// WithLocation sets the location of signing.
func WithLocation(location string) Option {
	return func(c *signConfig) { c.location = location }
}

// WithTimestamp enables RFC 3161 timestamping from the given TSA URL.
func WithTimestamp(tsaURL string) Option {
	return func(c *signConfig) { c.tsaURL = tsaURL }
}

// WithSignTime sets the signing time (default: current time).
func WithSignTime(t time.Time) Option {
	return func(c *signConfig) { c.signTime = t }
}

// Sign adds a digital signature to a PDF document.
// It returns the signed PDF as a new byte slice.
func Sign(pdfData []byte, signer Signer, opts ...Option) ([]byte, error) {
	cfg := &signConfig{
		signTime: time.Now(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if signer.Certificate == nil {
		return nil, fmt.Errorf("signature: certificate is required")
	}
	if signer.PrivateKey == nil {
		return nil, fmt.Errorf("signature: private key is required")
	}

	// Validate key type
	switch signer.PrivateKey.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		// OK
	default:
		return nil, fmt.Errorf("signature: unsupported key type %T", signer.PrivateKey)
	}

	// 1. Build the signature dictionary and placeholder
	sigResult, err := buildSignedPDF(pdfData, signer, cfg)
	if err != nil {
		return nil, fmt.Errorf("signature: build: %w", err)
	}

	// 2. Compute hash over ByteRange
	hash, err := computeByteRangeHash(sigResult.pdf, sigResult.byteRange)
	if err != nil {
		return nil, fmt.Errorf("signature: hash: %w", err)
	}

	// 3. Create CMS signature
	cmsData, err := createCMSSignature(hash, signer, cfg)
	if err != nil {
		return nil, fmt.Errorf("signature: cms: %w", err)
	}

	// 4. Inject signature into placeholder
	result, err := injectSignature(sigResult.pdf, sigResult.contentsOffset, sigResult.contentsLength, cmsData)
	if err != nil {
		return nil, fmt.Errorf("signature: inject: %w", err)
	}

	return result, nil
}

// signResult holds the intermediate result of PDF preparation.
type signResult struct {
	pdf            []byte
	byteRange      [4]int64 // [offset1, length1, offset2, length2]
	contentsOffset int      // byte offset of the hex string content
	contentsLength int      // length of the hex string content area
}
