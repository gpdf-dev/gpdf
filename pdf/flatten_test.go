package pdf

import (
	"bytes"
	"strings"
	"testing"
)

// buildTestPDFWithForm creates a simple PDF with an AcroForm text field.
func buildTestPDFWithForm(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(false)

	// Create page content stream.
	contentRef := w.AllocObject()
	content := Stream{
		Dict:    Dict{},
		Content: []byte("BT /F1 12 Tf 100 700 Td (Page Content) Tj ET"),
	}
	if err := w.WriteObject(contentRef, content); err != nil {
		t.Fatal(err)
	}

	// Create appearance stream for the form field.
	apRef := w.AllocObject()
	apStream := Stream{
		Dict: Dict{
			Name("Type"):      Name("XObject"),
			Name("Subtype"):   Name("Form"),
			Name("BBox"):      Array{Real(0), Real(0), Real(200), Real(20)},
			Name("Resources"): Dict{},
		},
		Content: []byte("BT /Helv 12 Tf 2 5 Td (Hello World) Tj ET"),
	}
	if err := w.WriteObject(apRef, apStream); err != nil {
		t.Fatal(err)
	}

	// Create widget annotation / form field.
	annotRef := w.AllocObject()
	annotDict := Dict{
		Name("Type"):    Name("Annot"),
		Name("Subtype"): Name("Widget"),
		Name("FT"):      Name("Tx"),
		Name("T"):       LiteralString("TextField1"),
		Name("V"):       LiteralString("Hello World"),
		Name("Rect"):    Array{Real(100), Real(600), Real(300), Real(620)},
		Name("AP"): Dict{
			Name("N"): apRef,
		},
	}
	if err := w.WriteObject(annotRef, annotDict); err != nil {
		t.Fatal(err)
	}

	// Add AcroForm to catalog.
	w.AddCatalogEntry(Name("AcroForm"), Dict{
		Name("Fields"): Array{annotRef},
	})

	// Add page with the annotation.
	pageRef := w.AllocObject()
	pageDict := Dict{
		Name("Type"):      Name("Page"),
		Name("MediaBox"):  Array{Real(0), Real(0), Real(595), Real(842)},
		Name("Contents"):  contentRef,
		Name("Resources"): Dict{},
		Name("Annots"):    Array{annotRef},
	}
	pageDict[Name("Parent")] = w.PageTreeRef()
	if err := w.WriteObject(pageRef, pageDict); err != nil {
		t.Fatal(err)
	}
	w.AddRawPage(pageRef)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// buildTestPDFWithCheckbox creates a PDF with a checkbox field.
func buildTestPDFWithCheckbox(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(false)

	contentRef := w.AllocObject()
	if err := w.WriteObject(contentRef, Stream{
		Dict:    Dict{},
		Content: []byte("BT /F1 12 Tf 100 700 Td (Checkbox Test) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}

	// Appearance streams for checked and unchecked states.
	yesRef := w.AllocObject()
	if err := w.WriteObject(yesRef, Stream{
		Dict: Dict{
			Name("Type"):      Name("XObject"),
			Name("Subtype"):   Name("Form"),
			Name("BBox"):      Array{Real(0), Real(0), Real(14), Real(14)},
			Name("Resources"): Dict{},
		},
		Content: []byte("0 0 14 14 re S 2 2 m 12 12 l S 12 2 m 2 12 l S"),
	}); err != nil {
		t.Fatal(err)
	}

	offRef := w.AllocObject()
	if err := w.WriteObject(offRef, Stream{
		Dict: Dict{
			Name("Type"):      Name("XObject"),
			Name("Subtype"):   Name("Form"),
			Name("BBox"):      Array{Real(0), Real(0), Real(14), Real(14)},
			Name("Resources"): Dict{},
		},
		Content: []byte("0 0 14 14 re S"),
	}); err != nil {
		t.Fatal(err)
	}

	annotRef := w.AllocObject()
	annotDict := Dict{
		Name("Type"):    Name("Annot"),
		Name("Subtype"): Name("Widget"),
		Name("FT"):      Name("Btn"),
		Name("T"):       LiteralString("Checkbox1"),
		Name("V"):       Name("Yes"),
		Name("AS"):      Name("Yes"),
		Name("Rect"):    Array{Real(100), Real(500), Real(114), Real(514)},
		Name("AP"): Dict{
			Name("N"): Dict{
				Name("Yes"): yesRef,
				Name("Off"): offRef,
			},
		},
	}
	if err := w.WriteObject(annotRef, annotDict); err != nil {
		t.Fatal(err)
	}

	w.AddCatalogEntry(Name("AcroForm"), Dict{
		Name("Fields"): Array{annotRef},
	})

	pageRef := w.AllocObject()
	pageDict := Dict{
		Name("Type"):      Name("Page"),
		Name("Parent"):    w.PageTreeRef(),
		Name("MediaBox"):  Array{Real(0), Real(0), Real(595), Real(842)},
		Name("Contents"):  contentRef,
		Name("Resources"): Dict{},
		Name("Annots"):    Array{annotRef},
	}
	if err := w.WriteObject(pageRef, pageDict); err != nil {
		t.Fatal(err)
	}
	w.AddRawPage(pageRef)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFlattenFormsNoForms(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)

	// Should be a no-op.
	if err := m.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	// Verify output is still valid.
	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	count, _ := r2.PageCount()
	if count != 1 {
		t.Errorf("page count = %d, want 1", count)
	}
}

func TestFlattenFormsTextField(t *testing.T) {
	data := buildTestPDFWithForm(t)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)

	if err := m.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Re-read and verify.
	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	// Verify page count.
	count, _ := r2.PageCount()
	if count != 1 {
		t.Errorf("page count = %d, want 1", count)
	}

	// Verify /AcroForm is removed from catalog.
	catalog, err := r2.ResolveDict(r2.RootRef())
	if err != nil {
		t.Fatalf("resolve catalog: %v", err)
	}
	if _, ok := catalog[Name("AcroForm")]; ok {
		t.Error("AcroForm should be removed from catalog after flattening")
	}

	// Verify /Annots is removed from the page.
	pageDict, err := r2.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}
	if _, ok := pageDict[Name("Annots")]; ok {
		t.Error("Annots should be removed from page after flattening")
	}

	// Verify content has been added (Contents should be an array now).
	contentsObj, ok := pageDict[Name("Contents")]
	if !ok {
		t.Fatal("page missing /Contents")
	}
	if _, ok := contentsObj.(Array); !ok {
		t.Error("Contents should be an array after flattening")
	}

	// Verify XObject resources were added.
	res, err := r2.ResolveDict(pageDict[Name("Resources")])
	if err != nil {
		t.Fatalf("resolve resources: %v", err)
	}
	xobjDict, err := r2.ResolveDict(res[Name("XObject")])
	if err != nil {
		t.Fatalf("resolve XObject: %v", err)
	}
	if _, ok := xobjDict[Name("_Flat0")]; !ok {
		t.Error("XObject /_Flat0 should exist in resources")
	}
}

func TestFlattenFormsCheckbox(t *testing.T) {
	data := buildTestPDFWithCheckbox(t)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)

	if err := m.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	catalog, err := r2.ResolveDict(r2.RootRef())
	if err != nil {
		t.Fatalf("resolve catalog: %v", err)
	}
	if _, ok := catalog[Name("AcroForm")]; ok {
		t.Error("AcroForm should be removed from catalog after flattening")
	}

	pageDict, err := r2.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}
	if _, ok := pageDict[Name("Annots")]; ok {
		t.Error("Annots should be removed from page after flattening")
	}
}

func TestFlattenFormsMixedAnnotations(t *testing.T) {
	// Build a PDF with both a widget and a non-widget annotation.
	data := buildTestPDFWithMixedAnnotations(t)

	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)

	if err := m.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	pageDict, err := r2.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}

	// /Annots should still exist with only the link annotation.
	annotsObj, ok := pageDict[Name("Annots")]
	if !ok {
		t.Fatal("Annots should still exist (link annotation)")
	}
	annotsResolved, err := r2.Resolve(annotsObj)
	if err != nil {
		t.Fatalf("resolve annots: %v", err)
	}
	annotsArr, ok := annotsResolved.(Array)
	if !ok {
		t.Fatal("Annots should be an array")
	}
	if len(annotsArr) != 1 {
		t.Errorf("Annots length = %d, want 1 (only link)", len(annotsArr))
	}

	// Verify the remaining annotation is the link.
	linkDict, err := r2.ResolveDict(annotsArr[0])
	if err != nil {
		t.Fatalf("resolve remaining annot: %v", err)
	}
	if subtype, ok := linkDict[Name("Subtype")].(Name); !ok || string(subtype) != "Link" {
		t.Error("remaining annotation should be the Link annotation")
	}
}

// buildTestPDFWithMixedAnnotations creates a PDF with a widget and a link annotation.
func buildTestPDFWithMixedAnnotations(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(false)

	contentRef := w.AllocObject()
	writeObj(t, w, contentRef, Stream{
		Dict:    Dict{},
		Content: []byte("BT /F1 12 Tf 100 700 Td (Mixed) Tj ET"),
	})

	apRef := w.AllocObject()
	writeObj(t, w, apRef, Stream{
		Dict: Dict{
			Name("Type"):      Name("XObject"),
			Name("Subtype"):   Name("Form"),
			Name("BBox"):      Array{Real(0), Real(0), Real(100), Real(20)},
			Name("Resources"): Dict{},
		},
		Content: []byte("BT /Helv 12 Tf 2 5 Td (Test) Tj ET"),
	})

	widgetRef := w.AllocObject()
	writeObj(t, w, widgetRef, Dict{
		Name("Type"):    Name("Annot"),
		Name("Subtype"): Name("Widget"),
		Name("FT"):      Name("Tx"),
		Name("T"):       LiteralString("Field1"),
		Name("Rect"):    Array{Real(100), Real(600), Real(200), Real(620)},
		Name("AP"):      Dict{Name("N"): apRef},
	})

	linkRef := w.AllocObject()
	writeObj(t, w, linkRef, Dict{
		Name("Type"):    Name("Annot"),
		Name("Subtype"): Name("Link"),
		Name("Rect"):    Array{Real(50), Real(400), Real(200), Real(420)},
	})

	w.AddCatalogEntry(Name("AcroForm"), Dict{
		Name("Fields"): Array{widgetRef},
	})

	pageRef := w.AllocObject()
	writeObj(t, w, pageRef, Dict{
		Name("Type"):      Name("Page"),
		Name("Parent"):    w.PageTreeRef(),
		Name("MediaBox"):  Array{Real(0), Real(0), Real(595), Real(842)},
		Name("Contents"):  contentRef,
		Name("Resources"): Dict{},
		Name("Annots"):    Array{widgetRef, linkRef},
	})
	w.AddRawPage(pageRef)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// writeObj is a test helper that writes an object and fails on error.
func writeObj(t *testing.T, w *Writer, ref ObjectRef, obj Object) {
	t.Helper()
	if err := w.WriteObject(ref, obj); err != nil {
		t.Fatal(err)
	}
}

func TestFlattenFormsContentStream(t *testing.T) {
	// Verify the flattened content stream references the XObject.
	data := buildTestPDFWithForm(t)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)

	if err := m.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// The flattened content should contain "/_Flat0 Do".
	if !strings.Contains(string(result), "/_Flat0 Do") {
		t.Error("flattened PDF should contain /_Flat0 Do operator")
	}
}

func TestFlattenFormsWidgetWithoutAppearance(t *testing.T) {
	// Widget without /AP should be removed silently.
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(false)

	contentRef := w.AllocObject()
	if err := w.WriteObject(contentRef, Stream{
		Dict:    Dict{},
		Content: []byte("BT /F1 12 Tf 100 700 Td (NoAP) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}

	widgetRef := w.AllocObject()
	if err := w.WriteObject(widgetRef, Dict{
		Name("Type"):    Name("Annot"),
		Name("Subtype"): Name("Widget"),
		Name("FT"):      Name("Tx"),
		Name("T"):       LiteralString("NoAppearance"),
		Name("Rect"):    Array{Real(100), Real(600), Real(200), Real(620)},
		// No /AP entry.
	}); err != nil {
		t.Fatal(err)
	}

	w.AddCatalogEntry(Name("AcroForm"), Dict{
		Name("Fields"): Array{widgetRef},
	})

	pageRef := w.AllocObject()
	if err := w.WriteObject(pageRef, Dict{
		Name("Type"):      Name("Page"),
		Name("Parent"):    w.PageTreeRef(),
		Name("MediaBox"):  Array{Real(0), Real(0), Real(595), Real(842)},
		Name("Contents"):  contentRef,
		Name("Resources"): Dict{},
		Name("Annots"):    Array{widgetRef},
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

	if err := m.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	catalog, err := r2.ResolveDict(r2.RootRef())
	if err != nil {
		t.Fatalf("resolve catalog: %v", err)
	}
	if _, ok := catalog[Name("AcroForm")]; ok {
		t.Error("AcroForm should be removed")
	}

	pageDict, err := r2.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}
	if _, ok := pageDict[Name("Annots")]; ok {
		t.Error("Annots should be removed")
	}
}
