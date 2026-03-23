package pdf

import (
	"testing"
)

func TestMergeTwoPDFs(t *testing.T) {
	pdf1 := buildTestPDF(t, 2)
	pdf2 := buildTestPDF(t, 3)

	merged, err := MergePDFs([]MergeSource{
		{Data: pdf1, FromPage: 0, ToPage: -1},
		{Data: pdf2, FromPage: 0, ToPage: -1},
	}, MergeConfig{})
	if err != nil {
		t.Fatalf("MergePDFs: %v", err)
	}

	r, err := NewReader(merged)
	if err != nil {
		t.Fatalf("NewReader on merged: %v", err)
	}
	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 5 {
		t.Errorf("PageCount = %d, want 5", count)
	}
}

func TestMergePageRange(t *testing.T) {
	pdf1 := buildTestPDF(t, 5)

	merged, err := MergePDFs([]MergeSource{
		{Data: pdf1, FromPage: 1, ToPage: 3}, // pages 2-4 (0-based)
	}, MergeConfig{})
	if err != nil {
		t.Fatalf("MergePDFs: %v", err)
	}

	r, err := NewReader(merged)
	if err != nil {
		t.Fatalf("NewReader on merged: %v", err)
	}
	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 3 {
		t.Errorf("PageCount = %d, want 3", count)
	}
}

func TestMergeEmptySources(t *testing.T) {
	_, err := MergePDFs(nil, MergeConfig{})
	if err == nil {
		t.Error("expected error for empty sources")
	}
}

func TestMergeSingleSource(t *testing.T) {
	pdf1 := buildTestPDF(t, 3)

	merged, err := MergePDFs([]MergeSource{
		{Data: pdf1, FromPage: 0, ToPage: -1},
	}, MergeConfig{})
	if err != nil {
		t.Fatalf("MergePDFs: %v", err)
	}

	r, err := NewReader(merged)
	if err != nil {
		t.Fatalf("NewReader on merged: %v", err)
	}
	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 3 {
		t.Errorf("PageCount = %d, want 3", count)
	}
}

func TestMergeWithMetadata(t *testing.T) {
	pdf1 := buildTestPDF(t, 1)

	merged, err := MergePDFs([]MergeSource{
		{Data: pdf1, FromPage: 0, ToPage: -1},
	}, MergeConfig{
		Info: DocumentInfo{Title: "Test Title", Author: "Test Author"},
	})
	if err != nil {
		t.Fatalf("MergePDFs: %v", err)
	}

	r, err := NewReader(merged)
	if err != nil {
		t.Fatalf("NewReader on merged: %v", err)
	}

	// Verify the Info dict exists in the trailer.
	infoObj, ok := r.Trailer()[Name("Info")]
	if !ok {
		t.Fatal("merged PDF has no /Info in trailer")
	}
	infoDict, err := r.ResolveDict(infoObj)
	if err != nil {
		t.Fatalf("resolve Info: %v", err)
	}
	if title, ok := infoDict[Name("Title")]; !ok {
		t.Error("Info dict missing /Title")
	} else if title.(LiteralString) != "Test Title" {
		t.Errorf("Title = %v, want 'Test Title'", title)
	}
}

func TestMergeInvalidPageRange(t *testing.T) {
	pdf1 := buildTestPDF(t, 2)

	_, err := MergePDFs([]MergeSource{
		{Data: pdf1, FromPage: 0, ToPage: 5},
	}, MergeConfig{})
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

func TestMergeRoundTrip(t *testing.T) {
	// Merge, read, merge again to verify output is valid PDF.
	pdf1 := buildTestPDF(t, 2)
	pdf2 := buildTestPDF(t, 1)

	merged1, err := MergePDFs([]MergeSource{
		{Data: pdf1, FromPage: 0, ToPage: -1},
		{Data: pdf2, FromPage: 0, ToPage: -1},
	}, MergeConfig{})
	if err != nil {
		t.Fatalf("first merge: %v", err)
	}

	// Merge the merged PDF with another.
	merged2, err := MergePDFs([]MergeSource{
		{Data: merged1, FromPage: 0, ToPage: -1},
		{Data: pdf2, FromPage: 0, ToPage: -1},
	}, MergeConfig{})
	if err != nil {
		t.Fatalf("second merge: %v", err)
	}

	r, err := NewReader(merged2)
	if err != nil {
		t.Fatalf("NewReader on double-merged: %v", err)
	}
	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 4 {
		t.Errorf("PageCount = %d, want 4", count)
	}
}
