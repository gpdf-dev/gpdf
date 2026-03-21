package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"time"
)

// OIDs for CMS/PKCS#7
var (
	oidSignedData             = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidData                   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidAttributeContentType   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidAttributeMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	oidAttributeSigningTime   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
	oidSHA256                 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidRSAWithSHA256          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}
	oidECDSAWithSHA256        = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}
)

// CMS ASN.1 structures
type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,tag:0"`
}

type signedData struct {
	Version          int
	DigestAlgorithms asn1.RawValue `asn1:"set"`
	EncapContentInfo encapContentInfo
	Certificates     asn1.RawValue `asn1:"optional,tag:0"`
	SignerInfos      asn1.RawValue `asn1:"set"`
}

type encapContentInfo struct {
	EContentType asn1.ObjectIdentifier
}

type signerInfo struct {
	Version            int
	SID                issuerAndSerialNumber
	DigestAlgorithm    pkix.AlgorithmIdentifier
	SignedAttrs        asn1.RawValue `asn1:"optional,tag:0"`
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Signature          []byte
}

type issuerAndSerialNumber struct {
	Issuer       asn1.RawValue
	SerialNumber *big.Int
}

type attribute struct {
	Type   asn1.ObjectIdentifier
	Values asn1.RawValue `asn1:"set"`
}

// signatureAlgorithm returns the AlgorithmIdentifier for the given private key.
func signatureAlgorithm(key crypto.PrivateKey) (pkix.AlgorithmIdentifier, error) {
	switch key.(type) {
	case *rsa.PrivateKey:
		return pkix.AlgorithmIdentifier{Algorithm: oidRSAWithSHA256}, nil
	case *ecdsa.PrivateKey:
		return pkix.AlgorithmIdentifier{Algorithm: oidECDSAWithSHA256}, nil
	default:
		return pkix.AlgorithmIdentifier{}, fmt.Errorf("unsupported key type: %T", key)
	}
}

// computeSignature signs the digest with the given private key.
func computeSignature(key crypto.PrivateKey, digest []byte) ([]byte, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, digest)
	case *ecdsa.PrivateKey:
		return ecdsa.SignASN1(rand.Reader, k, digest)
	default:
		return nil, fmt.Errorf("unsupported key type: %T", key)
	}
}

// buildSignerInfoBytes builds and marshals the SignerInfo structure.
func buildSignerInfoBytes(cert *x509.Certificate, attrsBytes []byte, sig []byte, digestAlg, sigAlg pkix.AlgorithmIdentifier) ([]byte, error) {
	issuerRaw := asn1.RawValue{FullBytes: cert.RawIssuer}

	innerAttrsBytes, err := extractInnerBytes(attrsBytes)
	if err != nil {
		return nil, fmt.Errorf("extract inner attrs: %w", err)
	}
	signedAttrsRaw := asn1.RawValue{
		Class:      asn1.ClassContextSpecific,
		Tag:        0,
		IsCompound: true,
		Bytes:      innerAttrsBytes,
	}

	signedAttrsEncoded, err := asn1.Marshal(signedAttrsRaw)
	if err != nil {
		return nil, fmt.Errorf("marshal signed attrs raw: %w", err)
	}

	si := signerInfo{
		Version: 1,
		SID: issuerAndSerialNumber{
			Issuer:       issuerRaw,
			SerialNumber: cert.SerialNumber,
		},
		DigestAlgorithm:    digestAlg,
		SignedAttrs:        asn1.RawValue{FullBytes: signedAttrsEncoded},
		SignatureAlgorithm: sigAlg,
		Signature:          sig,
	}

	return asn1.Marshal(si)
}

// createCMSSignature creates a CMS/PKCS#7 SignedData structure.
func createCMSSignature(hash []byte, signer Signer, cfg *signConfig) ([]byte, error) {
	cert := signer.Certificate

	sigAlg, err := signatureAlgorithm(signer.PrivateKey)
	if err != nil {
		return nil, err
	}

	digestAlg := pkix.AlgorithmIdentifier{Algorithm: oidSHA256}

	signTime := cfg.signTime
	if signTime.IsZero() {
		signTime = time.Now()
	}

	attrs, err := buildSignedAttrs(hash, signTime)
	if err != nil {
		return nil, fmt.Errorf("build signed attrs: %w", err)
	}

	attrsBytes, err := marshalAttributes(attrs)
	if err != nil {
		return nil, fmt.Errorf("marshal attrs: %w", err)
	}

	// For signing, authenticated attributes must be encoded as SET (0x31)
	attrsBytesForSign := make([]byte, len(attrsBytes))
	copy(attrsBytesForSign, attrsBytes)
	attrsBytesForSign[0] = 0x31

	attrHash := crypto.SHA256.New()
	attrHash.Write(attrsBytesForSign)
	attrDigest := attrHash.Sum(nil)

	sig, err := computeSignature(signer.PrivateKey, attrDigest)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	siBytes, err := buildSignerInfoBytes(cert, attrsBytes, sig, digestAlg, sigAlg)
	if err != nil {
		return nil, fmt.Errorf("marshal signer info: %w", err)
	}

	// Build certificates
	var certsBytes []byte
	allCerts := append([]*x509.Certificate{cert}, signer.Chain...)
	for _, c := range allCerts {
		certsBytes = append(certsBytes, c.Raw...)
	}

	// Marshal digest algorithms set
	digestAlgBytes, err := asn1.Marshal(digestAlg)
	if err != nil {
		return nil, err
	}
	digestAlgSetBytes, err := asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSet,
		IsCompound: true,
		Bytes:      digestAlgBytes,
	})
	if err != nil {
		return nil, err
	}

	// Build signer infos set
	siSetBytes, err := asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSet,
		IsCompound: true,
		Bytes:      siBytes,
	})
	if err != nil {
		return nil, err
	}

	// Build SignedData
	sd := signedData{
		Version:          1,
		DigestAlgorithms: asn1.RawValue{FullBytes: digestAlgSetBytes},
		EncapContentInfo: encapContentInfo{
			EContentType: oidData,
		},
		Certificates: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      certsBytes,
		},
		SignerInfos: asn1.RawValue{FullBytes: siSetBytes},
	}

	sdBytes, err := asn1.Marshal(sd)
	if err != nil {
		return nil, fmt.Errorf("marshal signed data: %w", err)
	}

	// Wrap in ContentInfo
	ci := contentInfo{
		ContentType: oidSignedData,
		Content: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      sdBytes,
		},
	}

	result, err := asn1.Marshal(ci)
	if err != nil {
		return nil, fmt.Errorf("marshal content info: %w", err)
	}

	return result, nil
}

// marshalAttributes marshals a list of attributes as a SEQUENCE.
func marshalAttributes(attrs []attribute) ([]byte, error) {
	var attrBytes []byte
	for _, attr := range attrs {
		b, err := asn1.Marshal(attr)
		if err != nil {
			return nil, err
		}
		attrBytes = append(attrBytes, b...)
	}
	// Wrap in a SET tag (0x31)
	return asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSet,
		IsCompound: true,
		Bytes:      attrBytes,
	})
}

// extractInnerBytes extracts the inner content bytes from a TLV-encoded ASN.1 value.
func extractInnerBytes(data []byte) ([]byte, error) {
	var raw asn1.RawValue
	_, err := asn1.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	return raw.Bytes, nil
}

func buildSignedAttrs(messageDigest []byte, signTime time.Time) ([]attribute, error) {
	// Content Type attribute
	ctBytes, err := asn1.Marshal(oidData)
	if err != nil {
		return nil, err
	}

	// Message Digest attribute
	mdBytes, err := asn1.Marshal(messageDigest)
	if err != nil {
		return nil, err
	}

	// Signing Time attribute
	stBytes, err := asn1.Marshal(signTime.UTC())
	if err != nil {
		return nil, err
	}

	return []attribute{
		{
			Type:   oidAttributeContentType,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: ctBytes},
		},
		{
			Type:   oidAttributeSigningTime,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: stBytes},
		},
		{
			Type:   oidAttributeMessageDigest,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: mdBytes},
		},
	}, nil
}

// GenerateTestCertificate creates a self-signed RSA certificate for testing.
func GenerateTestCertificate() (Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return Signer{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Signer",
			Organization: []string{"gpdf Test"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return Signer{}, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return Signer{}, err
	}

	return Signer{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

// GenerateTestECCertificate creates a self-signed EC certificate for testing.
func GenerateTestECCertificate() (Signer, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Signer{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "Test EC Signer",
			Organization: []string{"gpdf Test"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	pub := key.Public()
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, pub, key)
	if err != nil {
		return Signer{}, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return Signer{}, err
	}

	return Signer{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}
