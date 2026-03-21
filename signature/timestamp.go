package signature

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"io"
	"math/big"
	"net/http"
)

// OIDs for RFC 3161 timestamping.
var (
	oidTSTInfo                 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 1, 4}
	oidAttributeTimeStampToken = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14}
)

// ASN.1 structures for RFC 3161.

type timeStampReq struct {
	Version        int
	MessageImprint messageImprint
	Nonce          *big.Int `asn1:"optional"`
	CertReq        bool     `asn1:"optional,default:false"`
}

type messageImprint struct {
	HashAlgorithm pkix.AlgorithmIdentifier
	HashedMessage []byte
}

type timeStampResp struct {
	Status         pkiStatusInfo
	TimeStampToken asn1.RawValue `asn1:"optional"`
}

type pkiStatusInfo struct {
	Status int
}

// fetchTimestamp sends an RFC 3161 timestamp request to the given TSA URL.
// It hashes the provided signature value, sends the request, and returns
// the raw DER-encoded timestamp token (ContentInfo).
func fetchTimestamp(tsaURL string, sig []byte) ([]byte, error) {
	h := sha256.Sum256(sig)

	nonce, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))
	if err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	req := timeStampReq{
		Version: 1,
		MessageImprint: messageImprint{
			HashAlgorithm: pkix.AlgorithmIdentifier{Algorithm: oidSHA256},
			HashedMessage: h[:],
		},
		Nonce:   nonce,
		CertReq: true,
	}

	reqDER, err := asn1.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal timestamp request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", tsaURL, bytes.NewReader(reqDER))
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("TSA request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TSA returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read TSA response: %w", err)
	}

	var tsResp timeStampResp
	_, err = asn1.Unmarshal(body, &tsResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal timestamp response: %w", err)
	}

	if tsResp.Status.Status > 1 {
		return nil, fmt.Errorf("TSA error status: %d", tsResp.Status.Status)
	}

	if len(tsResp.TimeStampToken.FullBytes) == 0 {
		return nil, fmt.Errorf("no timestamp token in response")
	}

	return tsResp.TimeStampToken.FullBytes, nil
}

// buildUnsignedAttrs creates the unsigned attributes containing the timestamp token.
// Returns DER-encoded IMPLICIT [1] SET OF Attribute.
func buildUnsignedAttrs(timestampToken []byte) ([]byte, error) {
	attr := attribute{
		Type: oidAttributeTimeStampToken,
		Values: asn1.RawValue{
			Class:      asn1.ClassUniversal,
			Tag:        asn1.TagSet,
			IsCompound: true,
			Bytes:      timestampToken,
		},
	}

	attrBytes, err := asn1.Marshal(attr)
	if err != nil {
		return nil, fmt.Errorf("marshal timestamp attribute: %w", err)
	}

	return asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassContextSpecific,
		Tag:        1,
		IsCompound: true,
		Bytes:      attrBytes,
	})
}
