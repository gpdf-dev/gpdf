package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EncryptInfo holds parsed encryption dictionary values from a PDF.
type EncryptInfo struct {
	V     int    // encryption version (5 = AES-256)
	R     int    // revision (6 = ISO 32000-2)
	P     int32  // permission flags
	U     []byte // user password hash (48 bytes)
	O     []byte // owner password hash (48 bytes)
	UE    []byte // user encrypted key (32 bytes)
	OE    []byte // owner encrypted key (32 bytes)
	Perms []byte // encrypted permissions (16 bytes)
}

// ParseEncryptInfo extracts encryption dictionary values from raw PDF bytes.
func ParseEncryptInfo(pdfData []byte) (*EncryptInfo, error) {
	s := string(pdfData)

	// Find the Encrypt dictionary region
	idx := strings.Index(s, "/Filter /Standard")
	if idx < 0 {
		return nil, fmt.Errorf("encrypt: no /Filter /Standard found")
	}

	// Extract a window around the encrypt dict
	start := idx - 200
	if start < 0 {
		start = 0
	}
	end := idx + 2000
	if end > len(s) {
		end = len(s)
	}
	region := s[start:end]

	info := &EncryptInfo{}
	var err error

	info.V, err = parseIntField(region, "/V")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /V: %w", err)
	}
	info.R, err = parseIntField(region, "/R")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /R: %w", err)
	}

	pVal, err := parseIntField(region, "/P")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /P: %w", err)
	}
	info.P = int32(pVal)

	info.U, err = parseHexField(region, "/U")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /U: %w", err)
	}
	info.O, err = parseHexField(region, "/O")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /O: %w", err)
	}
	info.UE, err = parseHexField(region, "/UE")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /UE: %w", err)
	}
	info.OE, err = parseHexField(region, "/OE")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /OE: %w", err)
	}
	info.Perms, err = parseHexField(region, "/Perms")
	if err != nil {
		return nil, fmt.Errorf("encrypt: parse /Perms: %w", err)
	}

	return info, nil
}

// VerifyUserPassword checks if the given password matches the user password.
// Implements ISO 32000-2 Algorithm 2.A (step a) for Rev 6.
func (info *EncryptInfo) VerifyUserPassword(password string) bool {
	if len(info.U) < 48 {
		return false
	}
	pwd := truncatePassword(password)
	validationSalt := info.U[32:40]
	hash := computeHash(pwd, validationSalt, nil)
	return constantTimeEqual(hash[:32], info.U[:32])
}

// VerifyOwnerPassword checks if the given password matches the owner password.
// Implements ISO 32000-2 Algorithm 2.A (step b) for Rev 6.
func (info *EncryptInfo) VerifyOwnerPassword(password string) bool {
	if len(info.O) < 48 || len(info.U) < 48 {
		return false
	}
	pwd := truncatePassword(password)
	validationSalt := info.O[32:40]
	hash := computeHash(pwd, validationSalt, info.U)
	return constantTimeEqual(hash[:32], info.O[:32])
}

// DecryptFileKey recovers the file encryption key using the user password.
// Returns the 32-byte file encryption key.
func (info *EncryptInfo) DecryptFileKey(password string) ([]byte, error) {
	if len(info.U) < 48 || len(info.UE) < 32 {
		return nil, fmt.Errorf("encrypt: invalid U/UE length")
	}
	pwd := truncatePassword(password)
	keySalt := info.U[40:48]
	ueKey := computeHash(pwd, keySalt, nil)
	block, err := aes.NewCipher(ueKey)
	if err != nil {
		return nil, err
	}
	fileKey := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(fileKey, info.UE)
	return fileKey, nil
}

// VerifyPerms decrypts the /Perms entry using the file encryption key
// and checks that the decrypted permissions match /P.
func (info *EncryptInfo) VerifyPerms(fileKey []byte) error {
	if len(info.Perms) != 16 {
		return fmt.Errorf("encrypt: Perms must be 16 bytes, got %d", len(info.Perms))
	}
	block, err := aes.NewCipher(fileKey)
	if err != nil {
		return err
	}
	decrypted := make([]byte, 16)
	block.Decrypt(decrypted, info.Perms)

	// Check 'adb' marker at bytes 9-11
	if decrypted[9] != 'a' || decrypted[10] != 'd' || decrypted[11] != 'b' {
		return fmt.Errorf("encrypt: Perms decryption failed (bad marker: %q)", decrypted[9:12])
	}

	// Check permission bits match /P
	pFromPerms := int32(uint32(decrypted[0]) | uint32(decrypted[1])<<8 |
		uint32(decrypted[2])<<16 | uint32(decrypted[3])<<24)
	if pFromPerms != info.P {
		return fmt.Errorf("encrypt: Perms P=%d does not match /P=%d", pFromPerms, info.P)
	}

	return nil
}

// Permissions returns the permission flags from the encrypt dictionary.
func (info *EncryptInfo) Permissions() Permission {
	return Permission(uint32(info.P) & uint32(PermAll))
}

// HasPermission checks if a specific permission is granted.
func (info *EncryptInfo) HasPermission(perm Permission) bool {
	return info.Permissions()&perm == perm
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

func parseIntField(region, name string) (int, error) {
	// Look for /Name followed by integer
	pattern := regexp.MustCompile(regexp.QuoteMeta(name) + `\s+(-?\d+)`)
	m := pattern.FindStringSubmatch(region)
	if m == nil {
		return 0, fmt.Errorf("field %s not found", name)
	}
	return strconv.Atoi(m[1])
}

func parseHexField(region, name string) ([]byte, error) {
	// Look for /Name <HEXDATA>
	// The name might be followed by whitespace and then <hex>
	pattern := regexp.MustCompile(regexp.QuoteMeta(name) + `\s*<([0-9A-Fa-f]+)>`)
	m := pattern.FindStringSubmatch(region)
	if m == nil {
		return nil, fmt.Errorf("field %s not found", name)
	}
	return hex.DecodeString(m[1])
}
