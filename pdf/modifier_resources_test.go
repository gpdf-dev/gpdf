package pdf

import (
	"bytes"
	"testing"
)

// LibreOffice and Ghostscript write the page's /Font as an indirect object.
func TestModifierOverlayKeepsIndirectFontDict(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(false)

	fontRef := w.AllocObject()
	if err := w.WriteObject(fontRef, Dict{
		Name("Type"):     Name("Font"),
		Name("Subtype"):  Name("Type1"),
		Name("BaseFont"): Name("Helvetica"),
	}); err != nil {
		t.Fatal(err)
	}
	fontsRef := w.AllocObject()
	if err := w.WriteObject(fontsRef, Dict{Name("F1"): fontRef}); err != nil {
		t.Fatal(err)
	}
	contentRef := w.AllocObject()
	if err := w.WriteObject(contentRef, Stream{
		Dict:    Dict{},
		Content: []byte("BT /F1 12 Tf 100 700 Td (Page Content) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}
	pageRef := w.AllocObject()
	if err := w.WriteObject(pageRef, Dict{
		Name("Type"):      Name("Page"),
		Name("Parent"):    w.PageTreeRef(),
		Name("MediaBox"):  Array{Real(0), Real(0), Real(595), Real(842)},
		Name("Contents"):  contentRef,
		Name("Resources"): Dict{Name("Font"): fontsRef},
	}); err != nil {
		t.Fatal(err)
	}
	w.AddRawPage(pageRef)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)
	resources := Dict{Name("Font"): Dict{Name("OvF1"): fontRef}}
	if err := m.OverlayPage(0, []byte("BT /OvF1 24 Tf 100 400 Td (OVERLAY) Tj ET"), &resources); err != nil {
		t.Fatalf("OverlayPage: %v", err)
	}
	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	page, err := r2.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}
	res, err := r2.ResolveDict(page[Name("Resources")])
	if err != nil {
		t.Fatalf("Resources: %v", err)
	}
	fonts, err := r2.ResolveDict(res[Name("Font")])
	if err != nil {
		t.Fatalf("Font: %v", err)
	}
	for _, name := range []Name{"F1", "OvF1"} {
		if _, ok := fonts[name]; !ok {
			t.Errorf("page fonts lack /%s: %v", name, fonts)
		}
	}
}
