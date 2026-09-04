package signature

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/gpdf-dev/gpdf/pdf"
)

// Helper: generate a minimal valid PDF for testing.
func generateTestPDF(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)
	err := pw.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
	if err := pw.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	return buf.Bytes()
}

func TestSign_RSA(t *testing.T) {
	pdfData := generateTestPDF(t)
	signer, err := GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate error: %v", err)
	}

	signed, err := Sign(pdfData, signer,
		WithReason("Test signing"),
		WithLocation("Tokyo"),
		WithSignTime(time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)),
	)
	if err != nil {
		t.Fatalf("Sign error: %v", err)
	}

	if len(signed) <= len(pdfData) {
		t.Error("signed PDF should be larger than original")
	}

	s := string(signed)
	if !strings.Contains(s, "/Sig") {
		t.Error("missing /Sig dictionary")
	}
	if !strings.Contains(s, "/ByteRange") {
		t.Error("missing /ByteRange")
	}
	if !strings.Contains(s, "/Contents") {
		t.Error("missing /Contents")
	}
	if !strings.Contains(s, "Adobe.PPKLite") {
		t.Error("missing /Filter Adobe.PPKLite")
	}
	if !strings.Contains(s, "adbe.pkcs7.detached") {
		t.Error("missing /SubFilter adbe.pkcs7.detached")
	}
	if !strings.Contains(s, "Test signing") {
		t.Error("missing reason")
	}
	if !strings.Contains(s, "Tokyo") {
		t.Error("missing location")
	}

	// Should still be a valid PDF (starts with header, ends with EOF)
	if !bytes.HasPrefix(signed, []byte("%PDF-")) {
		t.Error("missing PDF header")
	}
	eofMarker := []byte("%%EOF")
	if !bytes.HasSuffix(bytes.TrimRight(signed, "\n"), eofMarker) {
		t.Error("missing EOF marker")
	}
}

func TestSign_ECDSA(t *testing.T) {
	pdfData := generateTestPDF(t)
	signer, err := GenerateTestECCertificate()
	if err != nil {
		t.Fatalf("GenerateTestECCertificate error: %v", err)
	}

	signed, err := Sign(pdfData, signer,
		WithReason("EC Test"),
	)
	if err != nil {
		t.Fatalf("Sign error: %v", err)
	}

	if len(signed) <= len(pdfData) {
		t.Error("signed PDF should be larger than original")
	}
}

func TestSign_NoCertificate(t *testing.T) {
	pdfData := generateTestPDF(t)
	_, err := Sign(pdfData, Signer{})
	if err == nil {
		t.Error("expected error for missing certificate")
	}
}

func TestSign_NoPrivateKey(t *testing.T) {
	pdfData := generateTestPDF(t)
	signer, _ := GenerateTestCertificate()
	signer.PrivateKey = nil
	_, err := Sign(pdfData, signer)
	if err == nil {
		t.Error("expected error for missing private key")
	}
}

func TestSign_InvalidPDF(t *testing.T) {
	signer, _ := GenerateTestCertificate()
	_, err := Sign([]byte("not a pdf"), signer, WithReason("test"))
	if err == nil {
		t.Error("expected error for invalid PDF")
	}
}

func TestParseTrailerBasic(t *testing.T) {
	pdfData := generateTestPDF(t)
	rootRef, xrefOffset, size, err := parseTrailerBasic(pdfData)
	if err != nil {
		t.Fatalf("parseTrailerBasic error: %v", err)
	}
	if rootRef <= 0 {
		t.Errorf("rootRef = %d, want > 0", rootRef)
	}
	if xrefOffset <= 0 {
		t.Errorf("xrefOffset = %d, want > 0", xrefOffset)
	}
	if size <= 0 {
		t.Errorf("size = %d, want > 0", size)
	}
}

func TestComputeByteRangeHash(t *testing.T) {
	data := []byte("Hello World, this is a test document for hashing")
	br := [4]int64{0, 10, 20, int64(len(data)) - 20}

	hash, err := computeByteRangeHash(data, br)
	if err != nil {
		t.Fatalf("computeByteRangeHash error: %v", err)
	}
	if len(hash) != 32 {
		t.Errorf("hash length = %d, want 32", len(hash))
	}

	// Same input should produce same hash
	hash2, _ := computeByteRangeHash(data, br)
	if !bytes.Equal(hash, hash2) {
		t.Error("deterministic hash check failed")
	}
}

func TestInjectSignature(t *testing.T) {
	// Build a simple test case
	data := []byte("PREFIX<00000000>SUFFIX")
	sig := []byte{0xAB, 0xCD}

	result, err := injectSignature(data, 8, 8, sig)
	if err != nil {
		t.Fatalf("injectSignature error: %v", err)
	}

	// Check that signature was injected
	injected := string(result[8:16])
	if !strings.HasPrefix(injected, "ABCD") {
		t.Errorf("injected = %q, want prefix 'ABCD'", injected)
	}
}

func TestInjectSignature_TooLarge(t *testing.T) {
	data := make([]byte, 100)
	sig := make([]byte, 100) // way too large for 10 hex chars
	_, err := injectSignature(data, 10, 10, sig)
	if err == nil {
		t.Error("expected error for oversized signature")
	}
}

func TestEscapeParens(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a(b)c", `a\(b\)c`},
		{`back\slash`, `back\\slash`},
	}
	for _, tt := range tests {
		got := escapeParens(tt.input)
		if got != tt.want {
			t.Errorf("escapeParens(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateTestCertificate(t *testing.T) {
	signer, err := GenerateTestCertificate()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if signer.Certificate == nil {
		t.Error("certificate is nil")
	}
	if signer.PrivateKey == nil {
		t.Error("private key is nil")
	}
	if signer.Certificate.Subject.CommonName != "Test Signer" {
		t.Errorf("CN = %q, want 'Test Signer'", signer.Certificate.Subject.CommonName)
	}
}

func TestGenerateTestECCertificate(t *testing.T) {
	signer, err := GenerateTestECCertificate()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if signer.Certificate == nil {
		t.Error("certificate is nil")
	}
	if signer.PrivateKey == nil {
		t.Error("private key is nil")
	}
}

// The /Contents placeholder is fixed-width and zero-padded, but the padding
// cannot be stripped textually: a CMS blob whose own DER ends in 0x00 loses
// that byte and fails to parse as "data truncated". ECDSA hits this roughly
// once every 256 signatures because its DER length and final byte vary per
// signature, which made TestSign_WithTimestamp_ECDSA flaky in CI.
func TestExtractContentsHex_PayloadEndingInZeroByte(t *testing.T) {
	// A DER SEQUENCE holding one OCTET STRING that ends in 0x00.
	der := []byte{0x30, 0x06, 0x04, 0x04, 0xDE, 0xAD, 0xBE, 0x00}

	hexStr := strings.ToUpper(hex.EncodeToString(der))
	padded := hexStr + strings.Repeat("0", 40) // placeholder zero padding
	pdfLike := "/Contents <" + padded + ">"

	got, err := extractContentsHex(pdfLike)
	if err != nil {
		t.Fatalf("extractContentsHex: %v", err)
	}
	if !bytes.Equal(got, der) {
		t.Errorf("got % X, want % X", got, der)
	}
}

func TestExtractContentsHex_LongFormLength(t *testing.T) {
	// Long-form length: SEQUENCE with a 200-byte OCTET STRING payload of zeros.
	payload := make([]byte, 200)
	der := append([]byte{0x30, 0x81, 0xCA, 0x04, 0x81, 0xC7}, payload[:199]...)

	padded := strings.ToUpper(hex.EncodeToString(der)) + strings.Repeat("0", 64)
	got, err := extractContentsHex("/Contents <" + padded + ">")
	if err != nil {
		t.Fatalf("extractContentsHex: %v", err)
	}
	if len(got) != len(der) {
		t.Errorf("got %d bytes, want %d", len(got), len(der))
	}
}

func TestExtractContentsHex_Malformed(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{"all padding", strings.Repeat("0", 32)},
		{"too short", "30"},
		{"length exceeds contents", "3082FFFF" + strings.Repeat("0", 16)},
		{"indefinite length", "3080" + strings.Repeat("0", 16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := extractContentsHex("/Contents <" + tt.contents + ">"); err == nil {
				t.Error("expected an error")
			}
		})
	}
}
