package pdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestAddPageWritesAnnots(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	annot := w.AllocObject()
	if err := w.WriteObject(annot, Dict{Name("Type"): Name("Annot")}); err != nil {
		t.Fatal(err)
	}
	if err := w.AddPage(PageObject{MediaBox: Rectangle{URX: 100, URY: 100}, Annots: []ObjectRef{annot}}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "/Annots [") {
		t.Errorf("page dict missing /Annots:\n%s", buf.String())
	}
}

func annotsOf(t *testing.T, data []byte) Array {
	t.Helper()
	r, err := NewReader(data)
	if err != nil {
		t.Fatal(err)
	}
	info, err := r.Page(0)
	if err != nil {
		t.Fatal(err)
	}
	pd, err := r.ResolveDict(info.Ref)
	if err != nil {
		t.Fatal(err)
	}
	arr, _ := pd[Name("Annots")].(Array)
	return arr
}

func TestModifierAddAnnotations(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModifier(r)
	if err := m.AddAnnotations(0); err != nil {
		t.Fatalf("no-op AddAnnotations: %v", err)
	}
	if err := m.OverlayPage(0, []byte("q Q"), nil); err != nil {
		t.Fatal(err)
	}
	link := Dict{Name("Type"): Name("Annot"), Name("Subtype"): Name("Link")}
	if err := m.AddAnnotations(0, link); err != nil {
		t.Fatal(err)
	}
	out, err := m.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(annotsOf(t, out)); got != 1 {
		t.Fatalf("annots = %d, want 1", got)
	}

	// A second incremental update appends rather than replaces.
	r2, err := NewReader(out)
	if err != nil {
		t.Fatal(err)
	}
	m2 := NewModifier(r2)
	if err := m2.AddAnnotations(0, link, link); err != nil {
		t.Fatal(err)
	}
	out2, err := m2.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(annotsOf(t, out2)); got != 3 {
		t.Errorf("annots after second update = %d, want 3", got)
	}

	if err := m2.AddAnnotations(99, link); err == nil {
		t.Error("expected error for out-of-range page")
	}
}
