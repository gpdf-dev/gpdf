package flatten_test

import (
	"bytes"
	"fmt"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/pdf"
)

// escapePDFString escapes special characters for PDF literal strings.
func escapePDFString(s string) string {
	var out []byte
	for _, ch := range []byte(s) {
		switch ch {
		case '(', ')', '\\':
			out = append(out, '\\', ch)
		default:
			out = append(out, ch)
		}
	}
	return string(out)
}

// formField describes a form field to add to the test PDF.
type formField struct {
	Name   string
	Type   string // "Tx", "Btn"
	Value  string // text value or "Yes"/"Off" for checkboxes
	X, Y   float64
	W, H   float64
	LabelX float64
}

// buildFilledFormPDF creates a PDF with multiple filled form fields.
func buildFilledFormPDF(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	fields := []formField{
		{Name: "Given Name", Type: "Tx", Value: "Taro", X: 200, Y: 700, W: 200, H: 20, LabelX: 50},
		{Name: "Family Name", Type: "Tx", Value: "Yamada", X: 200, Y: 665, W: 200, H: 20, LabelX: 50},
		{Name: "Address", Type: "Tx", Value: "1-2-3 Shibuya, Shibuya-ku", X: 200, Y: 630, W: 300, H: 20, LabelX: 50},
		{Name: "Postcode", Type: "Tx", Value: "150-0002", X: 200, Y: 595, W: 120, H: 20, LabelX: 50},
		{Name: "City", Type: "Tx", Value: "Tokyo", X: 380, Y: 595, W: 120, H: 20, LabelX: 330},
		{Name: "Country", Type: "Tx", Value: "Japan", X: 200, Y: 560, W: 200, H: 20, LabelX: 50},
		{Name: "Height", Type: "Tx", Value: "170", X: 200, Y: 525, W: 100, H: 20, LabelX: 50},
		{Name: "Shoe Size", Type: "Tx", Value: "26", X: 200, Y: 490, W: 100, H: 20, LabelX: 50},
		{Name: "Driving License", Type: "Btn", Value: "Yes", X: 200, Y: 450, W: 14, H: 14, LabelX: 50},
		{Name: "Language 1", Type: "Btn", Value: "Yes", X: 200, Y: 420, W: 14, H: 14, LabelX: 50},
		{Name: "Language 2", Type: "Btn", Value: "Off", X: 300, Y: 420, W: 14, H: 14, LabelX: 230},
		{Name: "Language 3", Type: "Btn", Value: "Yes", X: 400, Y: 420, W: 14, H: 14, LabelX: 330},
	}

	// Page labels.
	var content bytes.Buffer
	content.WriteString("BT /F1 16 Tf 50 770 Td (Form Flatten Test - Filled Fields) Tj ET\n")
	for _, f := range fields {
		label := f.Name + ":"
		if f.Type == "Btn" {
			label = f.Name
		}
		fmt.Fprintf(&content, "BT /F1 10 Tf %.1f %.1f Td (%s) Tj ET\n", f.LabelX, f.Y+4, label)
	}

	contentRef := w.AllocObject()
	if err := w.WriteObject(contentRef, pdf.Stream{
		Dict:    pdf.Dict{},
		Content: content.Bytes(),
	}); err != nil {
		t.Fatal(err)
	}

	// Form fields.
	var annotRefs pdf.Array
	var fieldRefs pdf.Array

	for _, f := range fields {
		bbox := pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(f.W), pdf.Real(f.H)}
		rect := pdf.Array{pdf.Real(f.X), pdf.Real(f.Y), pdf.Real(f.X + f.W), pdf.Real(f.Y + f.H)}

		annotRef := w.AllocObject()
		annotDict := pdf.Dict{
			pdf.Name("Type"):    pdf.Name("Annot"),
			pdf.Name("Subtype"): pdf.Name("Widget"),
			pdf.Name("FT"):      pdf.Name(f.Type),
			pdf.Name("T"):       pdf.LiteralString(f.Name),
			pdf.Name("Rect"):    rect,
		}

		switch f.Type {
		case "Tx":
			annotDict[pdf.Name("V")] = pdf.LiteralString(f.Value)

			apRef := w.AllocObject()
			apContent := fmt.Sprintf("/Tx BMC\nq\n0.95 0.95 0.95 rg\n0 0 %.1f %.1f re f\n0 0 0 rg\nBT /F1 10 Tf 2 5 Td (%s) Tj ET\nQ\nEMC",
				f.W, f.H, escapePDFString(f.Value))
			if err := w.WriteObject(apRef, pdf.Stream{
				Dict: pdf.Dict{
					pdf.Name("Type"):    pdf.Name("XObject"),
					pdf.Name("Subtype"): pdf.Name("Form"),
					pdf.Name("BBox"):    bbox,
				},
				Content: []byte(apContent),
			}); err != nil {
				t.Fatal(err)
			}
			annotDict[pdf.Name("AP")] = pdf.Dict{pdf.Name("N"): apRef}

		case "Btn":
			isChecked := f.Value == "Yes"
			if isChecked {
				annotDict[pdf.Name("V")] = pdf.Name("Yes")
				annotDict[pdf.Name("AS")] = pdf.Name("Yes")
			} else {
				annotDict[pdf.Name("V")] = pdf.Name("Off")
				annotDict[pdf.Name("AS")] = pdf.Name("Off")
			}

			yesRef := w.AllocObject()
			yesContent := fmt.Sprintf("q\n0 0 %.1f %.1f re S\n2 2 m %.1f %.1f l S\n%.1f 2 m 2 %.1f l S\nQ",
				f.W, f.H, f.W-2, f.H-2, f.W-2, f.H-2)
			if err := w.WriteObject(yesRef, pdf.Stream{
				Dict: pdf.Dict{
					pdf.Name("Type"):    pdf.Name("XObject"),
					pdf.Name("Subtype"): pdf.Name("Form"),
					pdf.Name("BBox"):    bbox,
				},
				Content: []byte(yesContent),
			}); err != nil {
				t.Fatal(err)
			}

			offRef := w.AllocObject()
			offContent := fmt.Sprintf("q\n0 0 %.1f %.1f re S\nQ", f.W, f.H)
			if err := w.WriteObject(offRef, pdf.Stream{
				Dict: pdf.Dict{
					pdf.Name("Type"):    pdf.Name("XObject"),
					pdf.Name("Subtype"): pdf.Name("Form"),
					pdf.Name("BBox"):    bbox,
				},
				Content: []byte(offContent),
			}); err != nil {
				t.Fatal(err)
			}

			annotDict[pdf.Name("AP")] = pdf.Dict{
				pdf.Name("N"): pdf.Dict{
					pdf.Name("Yes"): yesRef,
					pdf.Name("Off"): offRef,
				},
			}
		}

		if err := w.WriteObject(annotRef, annotDict); err != nil {
			t.Fatal(err)
		}
		annotRefs = append(annotRefs, annotRef)
		fieldRefs = append(fieldRefs, annotRef)
	}

	w.AddCatalogEntry(pdf.Name("AcroForm"), pdf.Dict{
		pdf.Name("Fields"): fieldRefs,
	})

	pageRef := w.AllocObject()
	if err := w.WriteObject(pageRef, pdf.Dict{
		pdf.Name("Type"):      pdf.Name("Page"),
		pdf.Name("Parent"):    w.PageTreeRef(),
		pdf.Name("MediaBox"):  pdf.Array{pdf.Real(0), pdf.Real(0), pdf.Real(595), pdf.Real(842)},
		pdf.Name("Contents"):  contentRef,
		pdf.Name("Resources"): pdf.Dict{},
		pdf.Name("Annots"):    annotRefs,
	}); err != nil {
		t.Fatal(err)
	}
	w.AddRawPage(pageRef)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExample_Flatten_04_FilledForm(t *testing.T) {
	source := buildFilledFormPDF(t)

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
	testutil.WritePDF(t, "04_filled_form_after.pdf", result)

	// Verify flattening.
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

	// Verify XObject resources exist for flattened fields.
	res, err := r.ResolveDict(pageDict[pdf.Name("Resources")])
	if err != nil {
		t.Fatalf("resolve resources: %v", err)
	}
	xobjDict, err := r.ResolveDict(res[pdf.Name("XObject")])
	if err != nil {
		t.Fatalf("resolve XObject: %v", err)
	}
	if len(xobjDict) < 12 {
		t.Errorf("expected at least 12 XObjects, got %d", len(xobjDict))
	}
}
