package pdf

import (
	"bytes"
	"testing"
)

func TestModifierReader(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)
	if m.Reader() != r {
		t.Error("Reader() should return the same reader")
	}
}

func TestMergeResources(t *testing.T) {
	t.Run("non-overlapping keys", func(t *testing.T) {
		existing := Dict{
			Name("Font"): Dict{
				Name("F1"): ObjectRef{Number: 1},
			},
		}
		overlay := Dict{
			Name("XObject"): Dict{
				Name("Im1"): ObjectRef{Number: 2},
			},
		}
		result := mergeResources(existing, overlay)
		if _, ok := result[Name("Font")]; !ok {
			t.Error("missing /Font from existing")
		}
		if _, ok := result[Name("XObject")]; !ok {
			t.Error("missing /XObject from overlay")
		}
	})

	t.Run("overlapping sub-dicts are merged", func(t *testing.T) {
		existing := Dict{
			Name("Font"): Dict{
				Name("F1"): ObjectRef{Number: 1},
			},
		}
		overlay := Dict{
			Name("Font"): Dict{
				Name("F2"): ObjectRef{Number: 2},
			},
		}
		result := mergeResources(existing, overlay)
		fontDict, ok := result[Name("Font")].(Dict)
		if !ok {
			t.Fatal("Font is not a Dict")
		}
		if _, ok := fontDict[Name("F1")]; !ok {
			t.Error("missing F1 from existing")
		}
		if _, ok := fontDict[Name("F2")]; !ok {
			t.Error("missing F2 from overlay")
		}
	})

	t.Run("overlay overrides non-dict values", func(t *testing.T) {
		existing := Dict{
			Name("ProcSet"): Array{Name("PDF"), Name("Text")},
		}
		overlay := Dict{
			Name("ProcSet"): Array{Name("PDF"), Name("ImageB")},
		}
		result := mergeResources(existing, overlay)
		arr, ok := result[Name("ProcSet")].(Array)
		if !ok {
			t.Fatal("ProcSet is not an Array")
		}
		// Overlay replaces non-dict values entirely.
		if len(arr) != 2 {
			t.Errorf("ProcSet len = %d, want 2", len(arr))
		}
	})

	t.Run("overlay dict over non-dict existing", func(t *testing.T) {
		existing := Dict{
			Name("Font"): Integer(42), // unusual, but test the path
		}
		overlay := Dict{
			Name("Font"): Dict{
				Name("F1"): ObjectRef{Number: 1},
			},
		}
		result := mergeResources(existing, overlay)
		// Overlay should win because existing value is not a Dict.
		fontDict, ok := result[Name("Font")].(Dict)
		if !ok {
			t.Fatal("Font should be Dict from overlay")
		}
		if _, ok := fontDict[Name("F1")]; !ok {
			t.Error("missing F1 from overlay")
		}
	})

	t.Run("existing dict but overlay non-dict", func(t *testing.T) {
		existing := Dict{
			Name("Font"): Dict{
				Name("F1"): ObjectRef{Number: 1},
			},
		}
		overlay := Dict{
			Name("Font"): Integer(99),
		}
		result := mergeResources(existing, overlay)
		// Overlay non-dict overrides existing dict.
		if _, ok := result[Name("Font")].(Integer); !ok {
			t.Error("Font should be Integer from overlay")
		}
	})
}

func TestModifierOverlayWithContentsArray(t *testing.T) {
	// Build a PDF where the page /Contents is already an array.
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	m := NewModifier(r)
	overlay := []byte("BT /F1 12 Tf 50 50 Td (Test) Tj ET")
	resources := Dict{
		Name("Font"): Dict{
			Name("F2"): ObjectRef{Number: 100},
		},
	}
	if err := m.OverlayPage(0, overlay, &resources); err != nil {
		t.Fatalf("OverlayPage: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Verify the result is valid.
	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	count, _ := r2.PageCount()
	if count != 1 {
		t.Errorf("page count = %d, want 1", count)
	}
}

func TestModifierOverlayOutOfRange(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)
	err = m.OverlayPage(99, []byte("test"), nil)
	if err == nil {
		t.Error("expected error for out-of-range page overlay")
	}
}

// TestModifierIncrementalXRefEntryWidth guards the fixed-width xref entry format
// of the incremental update. Every entry line must be exactly 20 bytes: readers
// index into the table by offset, so a single extra byte shifts every following
// entry and makes the update look damaged (ISO 32000-2 §7.5.4).
func TestModifierIncrementalXRefEntryWidth(t *testing.T) {
	data := buildTestPDF(t, 2)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	m := NewModifier(r)
	if err := m.OverlayPage(0, []byte("BT /F1 24 Tf 100 400 Td (OVERLAY) Tj ET"), nil); err != nil {
		t.Fatalf("OverlayPage: %v", err)
	}
	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Take the appended xref section: "\nxref\n<subsections>trailer".
	// Match on the leading newline so this does not hit "startxref".
	xrefStart := bytes.LastIndex(result, []byte("\nxref\n"))
	trailerStart := bytes.LastIndex(result, []byte("trailer"))
	if xrefStart < 0 || trailerStart < xrefStart {
		t.Fatalf("no incremental xref section found")
	}
	section := result[xrefStart+len("\nxref\n") : trailerStart]

	entries := 0
	for _, line := range bytes.SplitAfter(section, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		// Subsection headers are "start count\n"; entry lines begin with a
		// zero-padded 10-digit offset. Match on the offset rather than the
		// terminator so a wrong terminator is reported as a width error.
		if !isXRefEntryLine(line) {
			continue
		}
		entries++
		if len(line) != 20 {
			t.Errorf("xref entry %q is %d bytes, want exactly 20", line, len(line))
		}
	}
	if entries == 0 {
		t.Fatal("no xref entry lines found in the incremental section")
	}
}

// isXRefEntryLine reports whether a line from an xref section is an entry
// (a 10-digit zero-padded offset followed by a space) rather than a
// "start count" subsection header.
func isXRefEntryLine(line []byte) bool {
	if len(line) < 11 || line[10] != ' ' {
		return false
	}
	for _, c := range line[:10] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
