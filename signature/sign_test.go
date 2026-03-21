package signature

import (
	"bytes"
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
