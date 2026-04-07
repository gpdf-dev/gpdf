package flatten_test

import (
	"bytes"
	"os"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/pdf"
)

// buildFormPDF creates a PDF with a text field and a checkbox for flatten testing.
func buildFormPDF(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	contentRef := w.AllocObject()
	if err := w.WriteObject(contentRef, pdf.Stream{
		Dict:    pdf.Dict{},
		Content: []byte("BT /F1 12 Tf 50 780 Td (Form Flatten Example) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}

	apRef := w.AllocObject()
	if err := w.WriteObject(apRef, pdf.Stream{
		Dict: pdf.Dict{
			pdf.Name("Type"):      pdf.Name("XObject"),
			pdf.Name("Subtype"):   pdf.Name("Form"),
			pdf.Name("BBox"):      pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(200), pdf.Real(20)},
			pdf.Name("Resources"): pdf.Dict{},
		},
		Content: []byte("BT /Helv 12 Tf 2 5 Td (John Doe) Tj ET"),
	}); err != nil {
		t.Fatal(err)
	}

	textFieldRef := w.AllocObject()
	if err := w.WriteObject(textFieldRef, pdf.Dict{
		pdf.Name("Type"):    pdf.Name("Annot"),
		pdf.Name("Subtype"): pdf.Name("Widget"),
		pdf.Name("FT"):      pdf.Name("Tx"),
		pdf.Name("T"):       pdf.LiteralString("Name"),
		pdf.Name("V"):       pdf.LiteralString("John Doe"),
		pdf.Name("Rect"):    pdf.Array{pdf.Real(100), pdf.Real(700), pdf.Real(300), pdf.Real(720)},
		pdf.Name("AP"):      pdf.Dict{pdf.Name("N"): apRef},
	}); err != nil {
		t.Fatal(err)
	}

	yesRef := w.AllocObject()
	if err := w.WriteObject(yesRef, pdf.Stream{
		Dict: pdf.Dict{
			pdf.Name("Type"):      pdf.Name("XObject"),
			pdf.Name("Subtype"):   pdf.Name("Form"),
			pdf.Name("BBox"):      pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(14), pdf.Real(14)},
			pdf.Name("Resources"): pdf.Dict{},
		},
		Content: []byte("0 0 14 14 re S 2 2 m 12 12 l S 12 2 m 2 12 l S"),
	}); err != nil {
		t.Fatal(err)
	}

	offRef := w.AllocObject()
	if err := w.WriteObject(offRef, pdf.Stream{
		Dict: pdf.Dict{
			pdf.Name("Type"):      pdf.Name("XObject"),
			pdf.Name("Subtype"):   pdf.Name("Form"),
			pdf.Name("BBox"):      pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(14), pdf.Real(14)},
			pdf.Name("Resources"): pdf.Dict{},
		},
		Content: []byte("0 0 14 14 re S"),
	}); err != nil {
		t.Fatal(err)
	}

	checkboxRef := w.AllocObject()
	if err := w.WriteObject(checkboxRef, pdf.Dict{
		pdf.Name("Type"):    pdf.Name("Annot"),
		pdf.Name("Subtype"): pdf.Name("Widget"),
		pdf.Name("FT"):      pdf.Name("Btn"),
		pdf.Name("T"):       pdf.LiteralString("Agree"),
		pdf.Name("V"):       pdf.Name("Yes"),
		pdf.Name("AS"):      pdf.Name("Yes"),
		pdf.Name("Rect"):    pdf.Array{pdf.Real(100), pdf.Real(650), pdf.Real(114), pdf.Real(664)},
		pdf.Name("AP"): pdf.Dict{
			pdf.Name("N"): pdf.Dict{
				pdf.Name("Yes"): yesRef,
				pdf.Name("Off"): offRef,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	w.AddCatalogEntry(pdf.Name("AcroForm"), pdf.Dict{
		pdf.Name("Fields"): pdf.Array{textFieldRef, checkboxRef},
	})

	pageRef := w.AllocObject()
	if err := w.WriteObject(pageRef, pdf.Dict{
		pdf.Name("Type"):      pdf.Name("Page"),
		pdf.Name("Parent"):    w.PageTreeRef(),
		pdf.Name("MediaBox"):  pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(595), pdf.Real(842)},
		pdf.Name("Contents"):  contentRef,
		pdf.Name("Resources"): pdf.Dict{},
		pdf.Name("Annots"):    pdf.Array{textFieldRef, checkboxRef},
	}); err != nil {
		t.Fatal(err)
	}
	w.AddRawPage(pageRef)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExample_Flatten_01_BasicFlatten(t *testing.T) {
	source, err := os.ReadFile(testdataDir + "/01_basic_flatten_before.pdf")
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
	testutil.WritePDF(t, "01_basic_flatten_after.pdf", result)

	// Verify: AcroForm removed, annotations removed.
	r, err := pdf.NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	catalog, err := r.ResolveDict(r.RootRef())
	if err != nil {
		t.Fatalf("resolve catalog: %v", err)
	}
	if _, ok := catalog[pdf.Name("AcroForm")]; ok {
		t.Error("AcroForm should be removed after flattening")
	}
	pageDict, err := r.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}
	if _, ok := pageDict[pdf.Name("Annots")]; ok {
		t.Error("Annots should be removed after flattening")
	}
}
