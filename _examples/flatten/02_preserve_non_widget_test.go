package flatten_test

import (
	"bytes"
	"os"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/pdf"
)

// buildMixedAnnotPDF creates a PDF with both a widget and a link annotation.
func buildMixedAnnotPDF(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	contentRef := w.AllocObject()
	if err := w.WriteObject(contentRef, pdf.Stream{
		Dict:    pdf.Dict{},
		Content: []byte("BT /F1 12 Tf 50 780 Td (Mixed Annotations) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}

	apRef := w.AllocObject()
	if err := w.WriteObject(apRef, pdf.Stream{
		Dict: pdf.Dict{
			pdf.Name("Type"):      pdf.Name("XObject"),
			pdf.Name("Subtype"):   pdf.Name("Form"),
			pdf.Name("BBox"):      pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(150), pdf.Real(20)},
			pdf.Name("Resources"): pdf.Dict{},
		},
		Content: []byte("BT /Helv 10 Tf 2 5 Td (form field) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}

	widgetRef := w.AllocObject()
	if err := w.WriteObject(widgetRef, pdf.Dict{
		pdf.Name("Type"):    pdf.Name("Annot"),
		pdf.Name("Subtype"): pdf.Name("Widget"),
		pdf.Name("FT"):      pdf.Name("Tx"),
		pdf.Name("T"):       pdf.LiteralString("Field1"),
		pdf.Name("Rect"):    pdf.Array{pdf.Real(100), pdf.Real(700), pdf.Real(250), pdf.Real(720)},
		pdf.Name("AP"):      pdf.Dict{pdf.Name("N"): apRef},
	}); err != nil {
		t.Fatal(err)
	}

	linkRef := w.AllocObject()
	if err := w.WriteObject(linkRef, pdf.Dict{
		pdf.Name("Type"):    pdf.Name("Annot"),
		pdf.Name("Subtype"): pdf.Name("Link"),
		pdf.Name("Rect"):    pdf.Array{pdf.Real(50), pdf.Real(400), pdf.Real(200), pdf.Real(420)},
		pdf.Name("A"): pdf.Dict{
			pdf.Name("Type"): pdf.Name("Action"),
			pdf.Name("S"):    pdf.Name("URI"),
			pdf.Name("URI"):  pdf.LiteralString("https://github.com/gpdf-dev/gpdf"),
		},
	}); err != nil {
		t.Fatal(err)
	}

	w.AddCatalogEntry(pdf.Name("AcroForm"), pdf.Dict{
		pdf.Name("Fields"): pdf.Array{widgetRef},
	})

	pageRef := w.AllocObject()
	if err := w.WriteObject(pageRef, pdf.Dict{
		pdf.Name("Type"):      pdf.Name("Page"),
		pdf.Name("Parent"):    w.PageTreeRef(),
		pdf.Name("MediaBox"):  pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(595), pdf.Real(842)},
		pdf.Name("Contents"):  contentRef,
		pdf.Name("Resources"): pdf.Dict{},
		pdf.Name("Annots"):    pdf.Array{widgetRef, linkRef},
	}); err != nil {
		t.Fatal(err)
	}
	w.AddRawPage(pageRef)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExample_Flatten_02_PreserveNonWidget(t *testing.T) {
	source, err := os.ReadFile("testdata/02_preserve_non_widget_before.pdf")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	doc, err := gpdf.Open(source)
	if err != nil {
		t.Fatalf("gpdf.Open: %v", err)
	}
	if err := doc.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "02_preserve_non_widget_after.pdf", result)

	// Verify link annotation is preserved.
	r, err := pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	pageDict, err := r.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}

	annotsObj, ok := pageDict[pdf.Name("Annots")]
	if !ok {
		t.Fatal("Annots should still exist (link annotation preserved)")
	}
	resolved, err := r.Resolve(annotsObj)
	if err != nil {
		t.Fatalf("resolve annots: %v", err)
	}
	arr, ok := resolved.(pdf.Array)
	if !ok {
		t.Fatal("Annots should be an array")
	}
	if len(arr) != 1 {
		t.Errorf("Annots length = %d, want 1", len(arr))
	}
	annotDict, err := r.ResolveDict(arr[0])
	if err != nil {
		t.Fatalf("resolve annot: %v", err)
	}
	if subtype, _ := annotDict[pdf.Name("Subtype")].(pdf.Name); string(subtype) != "Link" {
		t.Errorf("remaining annotation subtype = %q, want Link", subtype)
	}
}
