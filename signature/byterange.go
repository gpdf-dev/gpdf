package signature

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

const (
	// signatureMaxLength is the maximum size of the hex-encoded CMS signature.
	// 8192 hex chars = 4096 bytes of raw signature data, sufficient for RSA-4096 + cert chain.
	signatureMaxLength = 8192
	// signatureMaxLengthTSA is the maximum size when RFC 3161 timestamping is enabled.
	// 20480 hex chars = 10240 bytes, accommodates signature + TSA certificate + timestamp token.
	signatureMaxLengthTSA = 20480
)

// buildSignedPDF prepares a PDF with signature placeholder.
// It appends an incremental update with the signature dictionary.
func buildSignedPDF(pdfData []byte, signer Signer, cfg *signConfig) (*signResult, error) {
	// We use a simplified approach: append signature objects after %%EOF
	// using an incremental update strategy.

	// Find the last %%EOF
	eofIdx := bytes.LastIndex(pdfData, []byte("%%EOF"))
	if eofIdx < 0 {
		return nil, fmt.Errorf("no %%EOF found in PDF")
	}

	// Use original data up to end
	data := make([]byte, len(pdfData))
	copy(data, pdfData)

	// Find the existing catalog and page refs by parsing the trailer
	rootRef, lastXrefOffset, trailerSize, err := parseTrailerBasic(data)
	if err != nil {
		return nil, fmt.Errorf("parse trailer: %w", err)
	}

	// Determine next object number
	nextObj := trailerSize

	// Build signature dictionary content
	sigDictRef := nextObj
	nextObj++

	// Build the appended content
	var appendBuf bytes.Buffer
	appendBuf.WriteByte('\n')

	// Signature dictionary object
	sigObjOffset := len(data) + appendBuf.Len()
	fmt.Fprintf(&appendBuf, "%d 0 obj\n", sigDictRef)
	appendBuf.WriteString("<< /Type /Sig /Filter /Adobe.PPKLite /SubFilter /adbe.pkcs7.detached\n")

	// Signing time
	signTimeStr := cfg.signTime.Format("20060102150405-07'00'")
	fmt.Fprintf(&appendBuf, "/M (D:%s)\n", signTimeStr)

	if cfg.reason != "" {
		fmt.Fprintf(&appendBuf, "/Reason (%s)\n", escapeParens(cfg.reason))
	}
	if cfg.location != "" {
		fmt.Fprintf(&appendBuf, "/Location (%s)\n", escapeParens(cfg.location))
	}

	// ByteRange placeholder - will be filled in later
	byteRangePlaceholder := "/ByteRange [0000000000 0000000000 0000000000 0000000000]"
	byteRangeOffset := len(data) + appendBuf.Len()
	appendBuf.WriteString(byteRangePlaceholder)
	appendBuf.WriteByte('\n')

	// Contents placeholder (hex string with zeros)
	maxLen := signatureMaxLength
	if cfg.tsaURL != "" {
		maxLen = signatureMaxLengthTSA
	}
	contentsPrefix := "/Contents <"
	contentsOffset := len(data) + appendBuf.Len() + len(contentsPrefix)
	appendBuf.WriteString(contentsPrefix)
	placeholder := strings.Repeat("0", maxLen)
	appendBuf.WriteString(placeholder)
	appendBuf.WriteString(">\n")

	appendBuf.WriteString(">>\nendobj\n")

	// AcroForm with signature field
	sigFieldRef := nextObj
	nextObj++
	sigFieldObjOffset := len(data) + appendBuf.Len()
	fmt.Fprintf(&appendBuf, "%d 0 obj\n", sigFieldRef)
	fmt.Fprintf(&appendBuf, "<< /Type /Annot /Subtype /Widget /FT /Sig /T (Signature1) /V %d 0 R /F 132 /Rect [0 0 0 0] >>\n", sigDictRef)
	appendBuf.WriteString("endobj\n")

	// New xref
	xrefOffset := len(data) + appendBuf.Len()
	appendBuf.WriteString("xref\n")
	fmt.Fprintf(&appendBuf, "%d 2\n", sigDictRef)
	fmt.Fprintf(&appendBuf, "%010d 00000 n \r\n", sigObjOffset)
	fmt.Fprintf(&appendBuf, "%010d 00000 n \r\n", sigFieldObjOffset)

	// Trailer
	appendBuf.WriteString("trailer\n")
	fmt.Fprintf(&appendBuf, "<< /Size %d /Root %d 0 R /Prev %d >>\n", nextObj, rootRef, lastXrefOffset)
	fmt.Fprintf(&appendBuf, "startxref\n%d\n%%%%EOF\n", xrefOffset)

	// Combine original + appended
	result := make([]byte, len(data)+appendBuf.Len())
	copy(result, data)
	copy(result[len(data):], appendBuf.Bytes())

	// Now fix the ByteRange values
	// The Contents hex string is: <XXXX...XXXX> where XXXX is signatureMaxLength hex chars
	// contentsOffset points to the first hex char
	sigStart := contentsOffset                    // first hex char
	sigEnd := contentsOffset + signatureMaxLength // after last hex char (the '>' is at sigEnd)

	br := [4]int64{
		0,
		int64(sigStart) - 1, // up to and excluding '<'
		int64(sigEnd) + 1,   // after '>'
		int64(len(result)) - int64(sigEnd) - 1,
	}

	// Write ByteRange values
	brStr := fmt.Sprintf("/ByteRange [%010d %010d %010d %010d]", br[0], br[1], br[2], br[3])
	copy(result[byteRangeOffset:byteRangeOffset+len(brStr)], brStr)

	return &signResult{
		pdf:            result,
		byteRange:      br,
		contentsOffset: sigStart,
		contentsLength: signatureMaxLength,
	}, nil
}

// computeByteRangeHash computes SHA-256 hash over the byte ranges.
func computeByteRangeHash(data []byte, br [4]int64) ([]byte, error) {
	h := sha256.New()

	// First range
	start1 := br[0]
	end1 := br[0] + br[1]
	if end1 > int64(len(data)) {
		return nil, fmt.Errorf("byte range 1 exceeds data length")
	}
	h.Write(data[start1:end1])

	// Second range
	start2 := br[2]
	end2 := br[2] + br[3]
	if end2 > int64(len(data)) {
		return nil, fmt.Errorf("byte range 2 exceeds data length")
	}
	h.Write(data[start2:end2])

	return h.Sum(nil), nil
}

// injectSignature injects the hex-encoded CMS signature into the placeholder.
func injectSignature(data []byte, contentsOffset, contentsLength int, sig []byte) ([]byte, error) {
	hexSig := fmt.Sprintf("%X", sig)
	if len(hexSig) > contentsLength {
		return nil, fmt.Errorf("signature too large: %d > %d", len(hexSig), contentsLength)
	}

	// Pad with zeros
	for len(hexSig) < contentsLength {
		hexSig += "0"
	}

	result := make([]byte, len(data))
	copy(result, data)
	copy(result[contentsOffset:contentsOffset+contentsLength], hexSig)

	return result, nil
}

// parseTrailerBasic extracts the Root ref, last xref offset, and Size from the trailer.
func parseTrailerBasic(data []byte) (rootRef, lastXrefOffset int, trailerSize int, err error) {
	// Find "startxref"
	idx := bytes.LastIndex(data, []byte("startxref"))
	if idx < 0 {
		return 0, 0, 0, fmt.Errorf("no startxref found")
	}

	// Parse the xref offset number after "startxref\n"
	rest := string(data[idx+len("startxref"):])
	rest = strings.TrimSpace(rest)
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return 0, 0, 0, fmt.Errorf("no xref offset after startxref")
	}
	lastXrefOffset, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid xref offset: %w", err)
	}

	// Find "trailer" and parse Root and Size
	trailerIdx := bytes.LastIndex(data, []byte("trailer"))
	if trailerIdx < 0 {
		return 0, 0, 0, fmt.Errorf("no trailer found")
	}

	trailerStr := string(data[trailerIdx:])

	// Parse /Root N G R
	rootRef, err = parseRef(trailerStr, "/Root")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse Root: %w", err)
	}

	// Parse /Size N
	trailerSize, err = parseInt(trailerStr, "/Size")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse Size: %w", err)
	}

	return rootRef, lastXrefOffset, trailerSize, nil
}

func parseRef(s, key string) (int, error) {
	idx := strings.Index(s, key)
	if idx < 0 {
		return 0, fmt.Errorf("%s not found", key)
	}
	rest := strings.TrimSpace(s[idx+len(key):])
	parts := strings.Fields(rest)
	if len(parts) < 3 || parts[2] != "R" {
		return 0, fmt.Errorf("invalid ref after %s", key)
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	return n, nil
}

func parseInt(s, key string) (int, error) {
	idx := strings.Index(s, key)
	if idx < 0 {
		return 0, fmt.Errorf("%s not found", key)
	}
	rest := strings.TrimSpace(s[idx+len(key):])
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return 0, fmt.Errorf("no value after %s", key)
	}
	return strconv.Atoi(parts[0])
}

func escapeParens(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	return s
}
