package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"

	"github.com/gpdf-dev/gpdf/pdf"
)

// handler holds the encryption state.
type handler struct {
	encryptionKey []byte        // 32-byte file encryption key
	encryptRef    pdf.ObjectRef // ref of the Encrypt dict (skip during encryption)
}

// computeHash implements Algorithm 2.B from ISO 32000-2.
// This is a complex iterative hash using SHA-256/384/512.
func computeHash(password, salt, userKey []byte) []byte {
	// Initial hash
	h := sha256.New()
	h.Write(password)
	h.Write(salt)
	if len(userKey) > 0 {
		h.Write(userKey)
	}
	k := h.Sum(nil)

	// Round 0-63 (at minimum), continue while last byte of E > round number
	for round := 0; ; round++ {
		// Step a: build K1
		var k1 bytes.Buffer
		k1.Write(password)
		k1.Write(k)
		if len(userKey) > 0 {
			k1.Write(userKey)
		}
		k1seq := k1.Bytes()

		// Repeat K1 64 times
		repeated := make([]byte, 0, len(k1seq)*64)
		for i := 0; i < 64; i++ {
			repeated = append(repeated, k1seq...)
		}

		// Step b: AES-128-CBC encrypt with key=first 16 bytes of K, iv=next 16 bytes
		aesKey := k[:16]
		aesIV := k[16:32]
		block, err := aes.NewCipher(aesKey)
		if err != nil {
			// Fallback: should not happen with valid key
			return k
		}
		// Pad repeated to multiple of 16
		if len(repeated)%aes.BlockSize != 0 {
			padLen := aes.BlockSize - len(repeated)%aes.BlockSize
			repeated = append(repeated, make([]byte, padLen)...)
		}
		e := make([]byte, len(repeated))
		mode := newCBCEncrypterNoPad(block, aesIV)
		mode.CryptBlocks(e, repeated)

		// Step c: select hash based on sum of first 16 bytes mod 3
		var sum int
		for _, b := range e[:16] {
			sum += int(b)
		}
		switch sum % 3 {
		case 0:
			h2 := sha256.Sum256(e)
			k = h2[:]
		case 1:
			h2 := sha512.Sum384(e)
			k = h2[:]
		case 2:
			h2 := sha512.Sum512(e)
			k = h2[:]
		}

		// Step d: check termination (ISO 32000-2 Algorithm 2.B)
		// Exit when round >= 63 AND last byte of E <= round - 32.
		if round >= 63 && int(e[len(e)-1]) <= (round-32) {
			break
		}
	}

	return k[:32]
}

// newCBCEncrypterNoPad creates a CBC encrypter (no PKCS7 padding).
func newCBCEncrypterNoPad(block cipher.Block, iv []byte) cipher.BlockMode {
	return cipher.NewCBCEncrypter(block, iv)
}

// generateEncryptionKey generates a random 32-byte file encryption key.
func generateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("encrypt: random key: %w", err)
	}
	return key, nil
}

// computeU computes the U (user password hash) and UE (user encrypted key).
// Returns U (48 bytes: hash32 + validationSalt8 + keySalt8), UE (32 bytes).
func computeU(fileKey []byte, userPwd string) ([]byte, []byte, error) {
	password := truncatePassword(userPwd)

	// Generate random salts
	validationSalt := make([]byte, 8)
	keySalt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, validationSalt); err != nil {
		return nil, nil, err
	}
	if _, err := io.ReadFull(rand.Reader, keySalt); err != nil {
		return nil, nil, err
	}

	// U = hash(password, validationSalt) + validationSalt + keySalt
	hash := computeHash(password, validationSalt, nil)
	u := make([]byte, 48)
	copy(u[:32], hash)
	copy(u[32:40], validationSalt)
	copy(u[40:48], keySalt)

	// UE = AES-256-CBC encrypt fileKey with hash(password, keySalt) as key, zero IV
	ueKey := computeHash(password, keySalt, nil)
	block, err := aes.NewCipher(ueKey)
	if err != nil {
		return nil, nil, err
	}
	ue := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ue, fileKey)

	return u, ue, nil
}

// computeO computes the O (owner password hash) and OE (owner encrypted key).
// Returns O (48 bytes), OE (32 bytes).
func computeO(fileKey []byte, ownerPwd string, u []byte) ([]byte, []byte, error) {
	password := truncatePassword(ownerPwd)

	validationSalt := make([]byte, 8)
	keySalt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, validationSalt); err != nil {
		return nil, nil, err
	}
	if _, err := io.ReadFull(rand.Reader, keySalt); err != nil {
		return nil, nil, err
	}

	// O = hash(password, validationSalt, U) + validationSalt + keySalt
	hash := computeHash(password, validationSalt, u)
	o := make([]byte, 48)
	copy(o[:32], hash)
	copy(o[32:40], validationSalt)
	copy(o[40:48], keySalt)

	// OE = AES-256-CBC encrypt fileKey with hash(password, keySalt, U) as key, zero IV
	oeKey := computeHash(password, keySalt, u)
	block, err := aes.NewCipher(oeKey)
	if err != nil {
		return nil, nil, err
	}
	oe := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(oe, fileKey)

	return o, oe, nil
}

// computePerms computes the Perms entry (16 bytes encrypted).
func computePerms(fileKey []byte, perms Permission, encryptMetadata bool) ([]byte, error) {
	// Build 16-byte plaintext
	p := make([]byte, 16)
	pVal := uint32(perms) | 0xFFFFF000 // set upper bits
	p[0] = byte(pVal)
	p[1] = byte(pVal >> 8)
	p[2] = byte(pVal >> 16)
	p[3] = byte(pVal >> 24)
	// bytes 4-7: 0xFFFFFFFF
	p[4], p[5], p[6], p[7] = 0xFF, 0xFF, 0xFF, 0xFF
	// byte 8: 'T' or 'F' for EncryptMetadata
	if encryptMetadata {
		p[8] = 'T'
	} else {
		p[8] = 'F'
	}
	// bytes 9-11: 'a', 'd', 'b'
	p[9], p[10], p[11] = 'a', 'd', 'b'
	// bytes 12-15: random
	if _, err := io.ReadFull(rand.Reader, p[12:]); err != nil {
		return nil, err
	}

	return aesEncryptECB(fileKey, p)
}

func truncatePassword(pwd string) []byte {
	b := []byte(pwd)
	if len(b) > 127 {
		b = b[:127]
	}
	return b
}

// encryptObject encrypts strings and streams in a PDF object.
func (h *handler) encryptObject(ref pdf.ObjectRef, obj pdf.Object) pdf.Object {
	// Don't encrypt the Encrypt dictionary itself
	if ref.Number == h.encryptRef.Number {
		return obj
	}
	return h.transformObject(ref, obj)
}

func (h *handler) transformObject(ref pdf.ObjectRef, obj pdf.Object) pdf.Object {
	switch v := obj.(type) {
	case pdf.LiteralString:
		encrypted, err := aesEncryptCBC(h.encryptionKey, []byte(v))
		if err != nil {
			return obj // fallback: leave unencrypted
		}
		return pdf.HexString(encrypted)
	case pdf.HexString:
		encrypted, err := aesEncryptCBC(h.encryptionKey, []byte(v))
		if err != nil {
			return obj
		}
		return pdf.HexString(encrypted)
	case pdf.Dict:
		result := make(pdf.Dict, len(v))
		for k, val := range v {
			result[k] = h.transformObject(ref, val)
		}
		return result
	case pdf.Array:
		result := make(pdf.Array, len(v))
		for i, val := range v {
			result[i] = h.transformObject(ref, val)
		}
		return result
	case pdf.Stream:
		encrypted, err := aesEncryptCBC(h.encryptionKey, v.Content)
		if err != nil {
			return obj
		}
		newDict := make(pdf.Dict, len(v.Dict))
		for k, val := range v.Dict {
			newDict[k] = h.transformObject(ref, val)
		}
		return pdf.Stream{Dict: newDict, Content: encrypted}
	default:
		return obj
	}
}
