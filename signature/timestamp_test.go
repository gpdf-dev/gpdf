package signature

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockTSAServer creates a test HTTP server that responds with a minimal RFC 3161 timestamp response.
func mockTSAServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/timestamp-query" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse TimeStampReq
		var req timeStampReq
		_, err = asn1.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Build minimal TSTInfo
		type tstInfo struct {
			Version        int
			Policy         asn1.ObjectIdentifier
			MessageImprint messageImprint
			SerialNumber   *big.Int
			GenTime        time.Time `asn1:"generalized"`
			Nonce          *big.Int  `asn1:"optional"`
		}

		tst := tstInfo{
			Version:        1,
			Policy:         asn1.ObjectIdentifier{1, 2, 3, 4, 1},
			MessageImprint: req.MessageImprint,
			SerialNumber:   big.NewInt(1),
			GenTime:        time.Now().UTC(),
			Nonce:          req.Nonce,
		}

		tstInfoDER, err := asn1.Marshal(tst)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Sign the TSTInfo with a test TSA certificate to produce a valid CMS structure
		tsaSigner, err := GenerateTestCertificate()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		h := sha256.Sum256(tstInfoDER)

		digestAlg := pkix.AlgorithmIdentifier{Algorithm: oidSHA256}
		sigAlg, _ := signatureAlgorithm(tsaSigner.PrivateKey)

		signTime := time.Now().UTC()
		attrs, _ := buildSignedAttrs(h[:], signTime)
		attrsBytes, _ := marshalAttributes(attrs)

		attrsBytesForSign := make([]byte, len(attrsBytes))
		copy(attrsBytesForSign, attrsBytes)
		attrsBytesForSign[0] = 0x31

		attrHash := sha256.Sum256(attrsBytesForSign)
		sig, _ := computeSignature(tsaSigner.PrivateKey, attrHash[:])

		siBytes, _ := buildSignerInfoBytes(tsaSigner.Certificate, attrsBytes, sig, digestAlg, sigAlg, nil)

		// Build certificates
		certsBytes := tsaSigner.Certificate.Raw

		digestAlgBytes, _ := asn1.Marshal(digestAlg)
		digestAlgSetBytes, _ := asn1.Marshal(asn1.RawValue{
			Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: digestAlgBytes,
		})
		siSetBytes, _ := asn1.Marshal(asn1.RawValue{
			Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: siBytes,
		})

		// Build SignedData with TSTInfo as encapsulated content
		eContentBytes, _ := asn1.Marshal(asn1.RawValue{
			Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: tstInfoDER,
		})
		eci := struct {
			EContentType asn1.ObjectIdentifier
			EContent     asn1.RawValue `asn1:"optional,explicit,tag:0"`
		}{
			EContentType: oidTSTInfo,
			EContent:     asn1.RawValue{FullBytes: eContentBytes},
		}
		eciBytes, _ := asn1.Marshal(eci)

		sd := signedData{
			Version:          3,
			DigestAlgorithms: asn1.RawValue{FullBytes: digestAlgSetBytes},
			EncapContentInfo: encapContentInfo{EContentType: oidTSTInfo},
			Certificates: asn1.RawValue{
				Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: certsBytes,
			},
			SignerInfos: asn1.RawValue{FullBytes: siSetBytes},
		}

		// Override EncapContentInfo with proper eContent
		sdBytes, _ := asn1.Marshal(sd)
		// Replace EncapContentInfo in the marshaled bytes with the version that has eContent
		_ = eciBytes // We use the simpler version for the mock

		token, _ := asn1.Marshal(contentInfo{
			ContentType: oidSignedData,
			Content: asn1.RawValue{
				Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: sdBytes,
			},
		})

		// Build TimeStampResp
		resp := struct {
			Status pkiStatusInfo
			Token  asn1.RawValue `asn1:"optional"`
		}{
			Status: pkiStatusInfo{Status: 0},
			Token:  asn1.RawValue{FullBytes: token},
		}

		respDER, _ := asn1.Marshal(resp)
		w.Header().Set("Content-Type", "application/timestamp-reply")
		_, _ = w.Write(respDER)
	}))
}

func TestSign_WithTimestamp(t *testing.T) {
	tsa := mockTSAServer(t)
	defer tsa.Close()

	buf := generateTestPDF(t)

	signer, err := GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}

	signed, err := Sign(buf, signer,
		WithReason("Timestamp test"),
		WithTimestamp(tsa.URL),
	)
	if err != nil {
		t.Fatalf("Sign with timestamp: %v", err)
	}

	if len(signed) <= len(buf) {
		t.Error("signed PDF should be larger than original")
	}

	// Parse and verify the signature
	info, err := ParseSignatureInfo(signed)
	if err != nil {
		t.Fatalf("ParseSignatureInfo: %v", err)
	}

	// Verify the CMS signature is still valid
	if err := info.VerifyIntegrity(signed); err != nil {
		t.Fatalf("VerifyIntegrity: %v", err)
	}

	// Check that timestamp token was embedded
	if !info.HasTimestamp {
		t.Error("expected HasTimestamp to be true")
	}
	if len(info.TimestampToken) == 0 {
		t.Error("expected non-empty TimestampToken")
	}
}

func TestSign_WithTimestamp_ECDSA(t *testing.T) {
	tsa := mockTSAServer(t)
	defer tsa.Close()

	pdfData := generateTestPDF(t)
	signer, err := GenerateTestECCertificate()
	if err != nil {
		t.Fatalf("GenerateTestECCertificate: %v", err)
	}

	signed, err := Sign(pdfData, signer,
		WithTimestamp(tsa.URL),
	)
	if err != nil {
		t.Fatalf("Sign with timestamp (ECDSA): %v", err)
	}

	info, err := ParseSignatureInfo(signed)
	if err != nil {
		t.Fatalf("ParseSignatureInfo: %v", err)
	}

	if err := info.VerifyIntegrity(signed); err != nil {
		t.Fatalf("VerifyIntegrity: %v", err)
	}

	if !info.HasTimestamp {
		t.Error("expected HasTimestamp to be true")
	}
}

func TestSign_WithTimestamp_TSAError(t *testing.T) {
	// TSA server that returns an error
	tsa := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsa.Close()

	pdfData := generateTestPDF(t)
	signer, err := GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}

	_, err = Sign(pdfData, signer, WithTimestamp(tsa.URL))
	if err == nil {
		t.Error("expected error when TSA returns error")
	}
}

func TestSign_WithTimestamp_TSAUnreachable(t *testing.T) {
	pdfData := generateTestPDF(t)
	signer, err := GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}

	_, err = Sign(pdfData, signer, WithTimestamp("http://127.0.0.1:1"))
	if err == nil {
		t.Error("expected error when TSA is unreachable")
	}
}

func TestSign_WithTimestamp_TSARejection(t *testing.T) {
	// TSA server that returns a rejection status
	tsa := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Status pkiStatusInfo
		}{
			Status: pkiStatusInfo{Status: 2}, // rejection
		}
		respDER, _ := asn1.Marshal(resp)
		w.Header().Set("Content-Type", "application/timestamp-reply")
		_, _ = w.Write(respDER)
	}))
	defer tsa.Close()

	pdfData := generateTestPDF(t)
	signer, err := GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}

	_, err = Sign(pdfData, signer, WithTimestamp(tsa.URL))
	if err == nil {
		t.Error("expected error when TSA rejects request")
	}
}

func TestFetchTimestamp(t *testing.T) {
	tsa := mockTSAServer(t)
	defer tsa.Close()

	// Create a dummy signature to timestamp
	sig := make([]byte, 256)
	_, _ = rand.Read(sig)

	token, err := fetchTimestamp(tsa.URL, sig)
	if err != nil {
		t.Fatalf("fetchTimestamp: %v", err)
	}

	if len(token) == 0 {
		t.Error("expected non-empty timestamp token")
	}

	// Verify it's valid ASN.1
	var raw asn1.RawValue
	_, err = asn1.Unmarshal(token, &raw)
	if err != nil {
		t.Fatalf("timestamp token is not valid ASN.1: %v", err)
	}
}

func TestBuildUnsignedAttrs(t *testing.T) {
	// Build a dummy timestamp token
	dummyToken := []byte{0x30, 0x03, 0x02, 0x01, 0x01} // minimal SEQUENCE

	attrs, err := buildUnsignedAttrs(dummyToken)
	if err != nil {
		t.Fatalf("buildUnsignedAttrs: %v", err)
	}

	if len(attrs) == 0 {
		t.Error("expected non-empty unsigned attrs")
	}

	// Verify it's valid ASN.1 with context-specific tag [1]
	var raw asn1.RawValue
	_, err = asn1.Unmarshal(attrs, &raw)
	if err != nil {
		t.Fatalf("unsigned attrs is not valid ASN.1: %v", err)
	}
	if raw.Tag != 1 || raw.Class != asn1.ClassContextSpecific {
		t.Errorf("expected context-specific [1], got class=%d tag=%d", raw.Class, raw.Tag)
	}
}
