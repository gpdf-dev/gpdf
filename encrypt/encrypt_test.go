package encrypt

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/pdf"
)

func TestAESEncryptCBC(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("Hello, World!")

	encrypted, err := aesEncryptCBC(key, plaintext)
	if err != nil {
		t.Fatalf("aesEncryptCBC error: %v", err)
	}

	// Should be IV (16) + padded ciphertext (16)
	if len(encrypted) != 32 {
		t.Errorf("encrypted length = %d, want 32", len(encrypted))
	}
	// Should not be plaintext
	if bytes.Contains(encrypted, plaintext) {
		t.Error("encrypted data contains plaintext")
	}
}

func TestAESEncryptECB(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := make([]byte, 16)
	for i := range plaintext {
		plaintext[i] = byte(i + 100)
	}

	encrypted, err := aesEncryptECB(key, plaintext)
	if err != nil {
		t.Fatalf("aesEncryptECB error: %v", err)
	}
	if len(encrypted) != 16 {
		t.Errorf("encrypted length = %d, want 16", len(encrypted))
	}
	if bytes.Equal(encrypted, plaintext) {
		t.Error("encrypted equals plaintext")
	}
}

func TestAESEncryptECB_WrongSize(t *testing.T) {
	key := make([]byte, 32)
	_, err := aesEncryptECB(key, []byte("short"))
	if err == nil {
		t.Error("expected error for wrong-size input")
	}
}

func TestTruncatePassword(t *testing.T) {
	short := truncatePassword("hello")
	if string(short) != "hello" {
		t.Errorf("truncatePassword(short) = %q, want 'hello'", short)
	}

	long := strings.Repeat("a", 200)
	got := truncatePassword(long)
	if len(got) != 127 {
		t.Errorf("truncatePassword(long) length = %d, want 127", len(got))
	}
}

func TestPermissionFlags(t *testing.T) {
	// All permissions should have unique bits
	perms := []Permission{PermPrint, PermModify, PermCopy, PermAnnotate, PermFillForms, PermExtract, PermAssemble, PermPrintHighRes}
	for i := 0; i < len(perms); i++ {
		for j := i + 1; j < len(perms); j++ {
			if perms[i]&perms[j] != 0 {
				t.Errorf("permission %d and %d overlap", i, j)
			}
		}
	}

	// PermAll should include all
	for i, p := range perms {
		if PermAll&p == 0 {
			t.Errorf("PermAll missing permission %d", i)
		}
	}
}

func TestComputeU(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	u, ue, err := computeU(key, "test")
	if err != nil {
		t.Fatalf("computeU error: %v", err)
	}
	if len(u) != 48 {
		t.Errorf("U length = %d, want 48", len(u))
	}
	if len(ue) != 32 {
		t.Errorf("UE length = %d, want 32", len(ue))
	}
}

func TestComputeO(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	u := make([]byte, 48)
	o, oe, err := computeO(key, "owner", u)
	if err != nil {
		t.Fatalf("computeO error: %v", err)
	}
	if len(o) != 48 {
		t.Errorf("O length = %d, want 48", len(o))
	}
	if len(oe) != 32 {
		t.Errorf("OE length = %d, want 32", len(oe))
	}
}

func TestComputePerms(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	perms, err := computePerms(key, PermAll, true)
	if err != nil {
		t.Fatalf("computePerms error: %v", err)
	}
	if len(perms) != 16 {
		t.Errorf("Perms length = %d, want 16", len(perms))
	}
}

func TestApply_Integration(t *testing.T) {
	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)

	err := Apply(pw,
		WithOwnerPassword("owner123"),
		WithUserPassword("user456"),
		WithPermissions(PermPrint|PermCopy),
	)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	// Add a page
	err = pw.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}

	err = pw.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}

	got := buf.String()

	// Should contain Encrypt in trailer
	if !strings.Contains(got, "/Encrypt") {
		t.Error("missing /Encrypt in PDF")
	}

	// Should contain /ID in trailer
	if !strings.Contains(got, "/ID") {
		t.Error("missing /ID in PDF")
	}

	// Should contain V=5, R=6 (AES-256)
	if !strings.Contains(got, "/V 5") {
		t.Error("missing /V 5")
	}
	if !strings.Contains(got, "/R 6") {
		t.Error("missing /R 6")
	}

	// Should contain AESV3 crypt filter
	if !strings.Contains(got, "/AESV3") {
		t.Error("missing /AESV3")
	}

	// Should be valid PDF structure
	if !strings.Contains(got, "%PDF-1.7") {
		t.Error("missing PDF header")
	}
	if !strings.Contains(got, "%%EOF") {
		t.Error("missing EOF marker")
	}
}

func TestApply_NoPassword(t *testing.T) {
	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)
	err := Apply(pw)
	if err == nil {
		t.Error("expected error when no password provided")
	}
}

func TestApply_OwnerOnly(t *testing.T) {
	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)
	err := Apply(pw, WithOwnerPassword("secret"))
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	_ = pw.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	err = pw.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestHandler_TransformObject(t *testing.T) {
	h := &handler{
		encryptionKey: make([]byte, 32),
		encryptRef:    pdf.ObjectRef{Number: 99}, // won't match
	}

	ref := pdf.ObjectRef{Number: 5}

	// String should be transformed
	result := h.transformObject(ref, pdf.LiteralString("hello"))
	if _, ok := result.(pdf.HexString); !ok {
		t.Errorf("expected HexString after encryption, got %T", result)
	}

	// Integer should not be transformed
	result = h.transformObject(ref, pdf.Integer(42))
	if v, ok := result.(pdf.Integer); !ok || v != 42 {
		t.Errorf("Integer should not be transformed: got %v", result)
	}

	// Dict entries should be transformed recursively
	dict := pdf.Dict{
		pdf.Name("Key"): pdf.LiteralString("value"),
	}
	result = h.transformObject(ref, dict)
	if d, ok := result.(pdf.Dict); ok {
		if _, isHex := d[pdf.Name("Key")].(pdf.HexString); !isHex {
			t.Error("dict string value should be encrypted")
		}
	} else {
		t.Errorf("expected Dict, got %T", result)
	}
}

func TestHandler_SkipEncryptDict(t *testing.T) {
	h := &handler{
		encryptionKey: make([]byte, 32),
		encryptRef:    pdf.ObjectRef{Number: 10},
	}

	// Object matching encryptRef should be returned as-is
	obj := pdf.Dict{pdf.Name("Test"): pdf.LiteralString("secret")}
	result := h.encryptObject(pdf.ObjectRef{Number: 10}, obj)
	if d, ok := result.(pdf.Dict); ok {
		if _, isLit := d[pdf.Name("Test")].(pdf.LiteralString); !isLit {
			t.Error("encrypt dict should not be encrypted")
		}
	}
}
