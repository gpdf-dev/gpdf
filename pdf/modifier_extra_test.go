package pdf

import (
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
