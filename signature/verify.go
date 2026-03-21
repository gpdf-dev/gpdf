package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SignatureInfo holds parsed signature information from a signed PDF.
type SignatureInfo struct {
	// PDF-level fields
	Filter    string // e.g. "Adobe.PPKLite"
	SubFilter string // e.g. "adbe.pkcs7.detached"
	Reason    string
	Location  string
	SignTime  string // raw /M value
	ByteRange [4]int64

	// CMS-level fields
	CMSData         []byte            // raw CMS/PKCS#7 DER data
	Certificate     *x509.Certificate // signer certificate
	DigestAlgorithm asn1.ObjectIdentifier
	SignatureAlg    asn1.ObjectIdentifier
	MessageDigest   []byte // from signed attributes
	RawSignature    []byte // the cryptographic signature value
	SignedAttrsRaw  []byte // DER-encoded signed attributes (for verification)
}

// ParseSignatureInfo extracts signature information from a signed PDF.
func ParseSignatureInfo(pdfData []byte) (*SignatureInfo, error) {
	s := string(pdfData)
	info := &SignatureInfo{}

	// Parse /Filter
	if m := reField(`/Filter\s*/(\S+)`).FindStringSubmatch(s); m != nil {
		info.Filter = m[1]
	}
	// Parse /SubFilter
	if m := reField(`/SubFilter\s*/(\S+)`).FindStringSubmatch(s); m != nil {
		info.SubFilter = m[1]
	}
	// Parse /Reason
	if m := reField(`/Reason\s*\(([^)]*)\)`).FindStringSubmatch(s); m != nil {
		info.Reason = m[1]
	}
	// Parse /Location
	if m := reField(`/Location\s*\(([^)]*)\)`).FindStringSubmatch(s); m != nil {
		info.Location = m[1]
	}
	// Parse /M (signing time)
	if m := reField(`/M\s*\(([^)]*)\)`).FindStringSubmatch(s); m != nil {
		info.SignTime = m[1]
	}

	// Parse /ByteRange
	br, err := parseByteRange(s)
	if err != nil {
		return nil, fmt.Errorf("signature: parse ByteRange: %w", err)
	}
	info.ByteRange = br

	// Extract CMS data from /Contents hex string
	cmsData, err := extractContentsHex(s)
	if err != nil {
		return nil, fmt.Errorf("signature: extract Contents: %w", err)
	}
	info.CMSData = cmsData

	// Parse CMS structure
	if err := info.parseCMS(); err != nil {
		return nil, fmt.Errorf("signature: parse CMS: %w", err)
	}

	return info, nil
}

// VerifyByteRangeHash recomputes the SHA-256 hash over the ByteRange regions
// and compares it with the messageDigest in the CMS signed attributes.
func (info *SignatureInfo) VerifyByteRangeHash(pdfData []byte) error {
	hash, err := computeByteRangeHash(pdfData, info.ByteRange)
	if err != nil {
		return fmt.Errorf("compute hash: %w", err)
	}

	if len(info.MessageDigest) == 0 {
		return fmt.Errorf("no messageDigest in signed attributes")
	}

	if !constantTimeEqual(hash, info.MessageDigest) {
		return fmt.Errorf("hash mismatch: ByteRange hash does not match messageDigest")
	}

	return nil
}

// VerifySignature cryptographically verifies the CMS signature using the
// certificate's public key.
func (info *SignatureInfo) VerifySignature() error {
	if info.Certificate == nil {
		return fmt.Errorf("no certificate found")
	}
	if len(info.SignedAttrsRaw) == 0 {
		return fmt.Errorf("no signed attributes for verification")
	}
	if len(info.RawSignature) == 0 {
		return fmt.Errorf("no signature value")
	}

	// Hash the signed attributes (encoded as SET, tag 0x31)
	attrHash := sha256.Sum256(info.SignedAttrsRaw)

	pub := info.Certificate.PublicKey
	switch key := pub.(type) {
	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(key, crypto.SHA256, attrHash[:], info.RawSignature)
	case *ecdsa.PublicKey:
		if !ecdsa.VerifyASN1(key, attrHash[:], info.RawSignature) {
			return fmt.Errorf("ECDSA signature verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unsupported public key type: %T", pub)
	}
}

// VerifyIntegrity performs full integrity verification:
// 1. Recomputes ByteRange hash and checks against messageDigest
// 2. Cryptographically verifies the CMS signature
func (info *SignatureInfo) VerifyIntegrity(pdfData []byte) error {
	if err := info.VerifyByteRangeHash(pdfData); err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}
	if err := info.VerifySignature(); err != nil {
		return fmt.Errorf("signature check failed: %w", err)
	}
	return nil
}

// --- CMS parsing ---

// Minimal ASN.1 structures for parsing CMS SignedData.
type cmsContentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,tag:0"`
}

type cmsSignedData struct {
	Version          int
	DigestAlgorithms asn1.RawValue `asn1:"set"`
	EncapContentInfo asn1.RawValue
	Certificates     asn1.RawValue `asn1:"optional,tag:0"`
	SignerInfos      asn1.RawValue `asn1:"set"`
}

type cmsSignerInfo struct {
	Version         int
	SID             asn1.RawValue
	DigestAlgorithm asn1.RawValue
	SignedAttrs     asn1.RawValue `asn1:"optional,tag:0"`
	SignatureAlg    asn1.RawValue
	Signature       []byte
}

func (info *SignatureInfo) parseCMS() error {
	// Parse ContentInfo
	var ci cmsContentInfo
	_, err := asn1.Unmarshal(info.CMSData, &ci)
	if err != nil {
		return fmt.Errorf("unmarshal ContentInfo: %w", err)
	}

	// Parse SignedData
	var sd cmsSignedData
	_, err = asn1.Unmarshal(ci.Content.Bytes, &sd)
	if err != nil {
		return fmt.Errorf("unmarshal SignedData: %w", err)
	}

	// Parse certificate from Certificates [0]
	if len(sd.Certificates.Bytes) > 0 {
		cert, err := x509.ParseCertificate(sd.Certificates.Bytes)
		if err != nil {
			return fmt.Errorf("parse certificate: %w", err)
		}
		info.Certificate = cert
	}

	// Parse SignerInfo from the SET
	siBytes, err := extractSetContent(sd.SignerInfos.FullBytes)
	if err != nil {
		return fmt.Errorf("extract SignerInfos: %w", err)
	}

	var si cmsSignerInfo
	_, err = asn1.Unmarshal(siBytes, &si)
	if err != nil {
		return fmt.Errorf("unmarshal SignerInfo: %w", err)
	}

	info.RawSignature = si.Signature

	// Parse DigestAlgorithm from SignerInfo
	var digestAlgID struct {
		Algorithm asn1.ObjectIdentifier
	}
	if _, err := asn1.Unmarshal(si.DigestAlgorithm.FullBytes, &digestAlgID); err == nil {
		info.DigestAlgorithm = digestAlgID.Algorithm
	}

	// Parse SignatureAlgorithm from SignerInfo
	var sigAlgID struct {
		Algorithm asn1.ObjectIdentifier
	}
	if _, err := asn1.Unmarshal(si.SignatureAlg.FullBytes, &sigAlgID); err == nil {
		info.SignatureAlg = sigAlgID.Algorithm
	}

	// Parse signed attributes to extract messageDigest
	if len(si.SignedAttrs.Bytes) > 0 {
		// For verification, signed attrs must be re-encoded as SET (0x31)
		info.SignedAttrsRaw = marshalAsSet(si.SignedAttrs)

		// Extract messageDigest from attributes
		info.MessageDigest = extractMessageDigest(si.SignedAttrs.Bytes)
	}

	return nil
}

// marshalAsSet re-encodes the IMPLICIT [0] tagged signed attributes as a SET.
func marshalAsSet(raw asn1.RawValue) []byte {
	setVal := asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSet,
		IsCompound: true,
		Bytes:      raw.Bytes,
	}
	encoded, err := asn1.Marshal(setVal)
	if err != nil {
		return nil
	}
	return encoded
}

// extractMessageDigest walks the signed attributes to find messageDigest (OID 1.2.840.113549.1.9.4).
func extractMessageDigest(attrsBytes []byte) []byte {
	rest := attrsBytes
	for len(rest) > 0 {
		var attr struct {
			Type   asn1.ObjectIdentifier
			Values asn1.RawValue `asn1:"set"`
		}
		var err error
		rest, err = asn1.Unmarshal(rest, &attr)
		if err != nil {
			break
		}
		if attr.Type.Equal(oidAttributeMessageDigest) {
			// Values is a SET containing the OCTET STRING
			var digest []byte
			if _, err := asn1.Unmarshal(attr.Values.Bytes, &digest); err == nil {
				return digest
			}
		}
	}
	return nil
}

// extractSetContent extracts the first element from a SET.
func extractSetContent(setBytes []byte) ([]byte, error) {
	var raw asn1.RawValue
	_, err := asn1.Unmarshal(setBytes, &raw)
	if err != nil {
		return nil, err
	}
	return raw.Bytes, nil
}

// --- PDF-level parsing helpers ---

func reField(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

var reByteRange = regexp.MustCompile(`/ByteRange\s*\[\s*(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s*\]`)

func parseByteRange(s string) ([4]int64, error) {
	m := reByteRange.FindStringSubmatch(s)
	if m == nil {
		return [4]int64{}, fmt.Errorf("ByteRange not found")
	}
	var br [4]int64
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseInt(m[i+1], 10, 64)
		if err != nil {
			return [4]int64{}, err
		}
		br[i] = v
	}
	return br, nil
}

func extractContentsHex(s string) ([]byte, error) {
	// Find /Contents <HEXDATA>
	re := regexp.MustCompile(`/Contents\s*<([0-9A-Fa-f]+)>`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("contents hex string not found")
	}
	hexStr := strings.TrimRight(m[1], "0") // remove trailing zero padding
	if len(hexStr)%2 != 0 {
		hexStr += "0"
	}
	return hex.DecodeString(hexStr)
}

func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// IsRSA returns true if the signature uses RSA.
func (info *SignatureInfo) IsRSA() bool {
	if info.Certificate == nil {
		return false
	}
	_, ok := info.Certificate.PublicKey.(*rsa.PublicKey)
	return ok
}

// IsECDSA returns true if the signature uses ECDSA.
func (info *SignatureInfo) IsECDSA() bool {
	if info.Certificate == nil {
		return false
	}
	_, ok := info.Certificate.PublicKey.(*ecdsa.PublicKey)
	return ok
}

// RSAKeySize returns the RSA key size in bits, or 0 if not RSA.
func (info *SignatureInfo) RSAKeySize() int {
	if info.Certificate == nil {
		return 0
	}
	if key, ok := info.Certificate.PublicKey.(*rsa.PublicKey); ok {
		return key.N.BitLen()
	}
	return 0
}

// ECDSACurve returns the ECDSA curve name, or "" if not ECDSA.
func (info *SignatureInfo) ECDSACurve() string {
	if info.Certificate == nil {
		return ""
	}
	if key, ok := info.Certificate.PublicKey.(*ecdsa.PublicKey); ok {
		return key.Curve.Params().Name
	}
	return ""
}

// CertSubject returns the certificate subject CommonName.
func (info *SignatureInfo) CertSubject() string {
	if info.Certificate == nil {
		return ""
	}
	return info.Certificate.Subject.CommonName
}

// CertOrganization returns the certificate subject Organization.
func (info *SignatureInfo) CertOrganization() []string {
	if info.Certificate == nil {
		return nil
	}
	return info.Certificate.Subject.Organization
}

// --- Tamper detection helper ---

// TamperByte modifies a single byte in the PDF data outside the signature
// Contents region, simulating document tampering.
// Returns a copy of the data with the byte at the given offset XORed with 0xFF.
func TamperByte(pdfData []byte, offset int) []byte {
	if offset < 0 || offset >= len(pdfData) {
		return pdfData
	}
	tampered := make([]byte, len(pdfData))
	copy(tampered, pdfData)
	tampered[offset] ^= 0xFF
	return tampered
}

// FindSafeByteRange computes offsets from ByteRange that are safe to tamper
// (i.e., within the signed region but outside the Contents placeholder).
// Returns an offset in the first byte range (the PDF header region).
func (info *SignatureInfo) FindSafeByteRange() int {
	// Use an offset in the first byte range, near the middle
	mid := info.ByteRange[1] / 2
	if mid < 10 {
		mid = 10
	}
	return int(mid)
}

// --- ECDSA signature helpers for testing ---

// CorruptECDSASignature creates a copy of CMS data with a corrupted ECDSA signature value.
// This is used for negative testing of signature verification.
func CorruptECDSASignature(cmsData []byte) []byte {
	corrupted := make([]byte, len(cmsData))
	copy(corrupted, cmsData)
	// Find and flip a byte near the end (where the signature value lives)
	if len(corrupted) > 50 {
		corrupted[len(corrupted)-30] ^= 0xFF
	}
	return corrupted
}

// CorruptRSASignature creates a copy of CMS data with a corrupted RSA signature value.
func CorruptRSASignature(cmsData []byte) []byte {
	return CorruptECDSASignature(cmsData) // same approach works
}

// --- ECDSA P-256 curve verification helper ---

// VerifyECDSACurveP256 checks that the ECDSA key uses the P-256 curve.
func (info *SignatureInfo) VerifyECDSACurveP256() bool {
	if info.Certificate == nil {
		return false
	}
	key, ok := info.Certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false
	}
	return key.Curve == elliptic.P256()
}
