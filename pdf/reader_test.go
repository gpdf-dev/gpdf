package pdf

import (
	"bytes"
	"strings"
	"testing"
)

// buildTestPDF creates a minimal valid PDF with the given number of pages for testing.
func buildTestPDF(t *testing.T, numPages int) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(false)

	for i := 0; i < numPages; i++ {
		contentRef := w.AllocObject()
		content := Stream{
			Dict:    Dict{},
			Content: []byte("BT /F1 12 Tf 100 700 Td (Hello) Tj ET"),
		}
		if err := w.WriteObject(contentRef, content); err != nil {
			t.Fatal(err)
		}
		if err := w.AddPage(PageObject{
			MediaBox: Rectangle{LLX: 0, LLY: 0, URX: 595, URY: 842},
			Contents: []ObjectRef{contentRef},
		}); err != nil {
			t.Fatal(err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestNewReader(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if r.RootRef().Number == 0 {
		t.Error("root ref should be non-zero")
	}
}

func TestReaderPageCount(t *testing.T) {
	tests := []struct {
		name  string
		pages int
	}{
		{"single page", 1},
		{"two pages", 2},
		{"five pages", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildTestPDF(t, tt.pages)
			r, err := NewReader(data)
			if err != nil {
				t.Fatalf("NewReader: %v", err)
			}
			count, err := r.PageCount()
			if err != nil {
				t.Fatalf("PageCount: %v", err)
			}
			if count != tt.pages {
				t.Errorf("PageCount = %d, want %d", count, tt.pages)
			}
		})
	}
}

func TestReaderPageMediaBox(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	info, err := r.Page(0)
	if err != nil {
		t.Fatalf("Page(0): %v", err)
	}
	if info.MediaBox.URX != 595 || info.MediaBox.URY != 842 {
		t.Errorf("MediaBox = %v, want A4-ish (595x842)", info.MediaBox)
	}
}

func TestReaderGetObject(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	// Object 1 should be the catalog.
	obj, err := r.GetObject(r.RootRef().Number)
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	d, ok := obj.(Dict)
	if !ok {
		t.Fatalf("catalog is %T, want Dict", obj)
	}
	if typeName, ok := d[Name("Type")].(Name); !ok || typeName != "Catalog" {
		t.Errorf("catalog /Type = %v, want Catalog", d[Name("Type")])
	}
}

func TestReaderPageOutOfRange(t *testing.T) {
	data := buildTestPDF(t, 2)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	_, err = r.Page(5)
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

func TestReaderResolve(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Resolve a non-ref should return as-is.
	n := Integer(42)
	resolved, err := r.Resolve(n)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved != n {
		t.Errorf("Resolve(Integer) = %v, want %v", resolved, n)
	}

	// Resolve a ref should return the object.
	resolved, err = r.Resolve(r.RootRef())
	if err != nil {
		t.Fatalf("Resolve(ref): %v", err)
	}
	if _, ok := resolved.(Dict); !ok {
		t.Errorf("Resolve(root) = %T, want Dict", resolved)
	}
}

func TestReaderString(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	s := r.String()
	if !strings.Contains(s, "Reader:") {
		t.Errorf("String() = %q, want to contain 'Reader:'", s)
	}
}

func TestParserObjects(t *testing.T) { //nolint:gocyclo
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, obj Object)
	}{
		{
			name:  "integer",
			input: "42",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Integer); !ok || v != 42 {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "negative integer",
			input: "-7",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Integer); !ok || v != -7 {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "real",
			input: "3.14",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Real); !ok || float64(v) < 3.13 || float64(v) > 3.15 {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "boolean true",
			input: "true ",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Boolean); !ok || !bool(v) {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "boolean false",
			input: "false ",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Boolean); !ok || bool(v) {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "null",
			input: "null ",
			check: func(t *testing.T, obj Object) {
				if _, ok := obj.(Null); !ok {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "name",
			input: "/Type",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Name); !ok || v != "Type" {
					t.Errorf("got %v (%T)", obj, obj)
				}
			},
		},
		{
			name:  "name with hex escape",
			input: "/A#20B",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(Name); !ok || v != "A B" {
					t.Errorf("got %q (%T)", obj, obj)
				}
			},
		},
		{
			name:  "literal string",
			input: "(Hello World)",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(LiteralString); !ok || v != "Hello World" {
					t.Errorf("got %q (%T)", obj, obj)
				}
			},
		},
		{
			name:  "literal string with escapes",
			input: `(Hello\nWorld)`,
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(LiteralString); !ok || v != "Hello\nWorld" {
					t.Errorf("got %q (%T)", obj, obj)
				}
			},
		},
		{
			name:  "literal string with nested parens",
			input: "(Hello (World))",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(LiteralString); !ok || v != "Hello (World)" {
					t.Errorf("got %q (%T)", obj, obj)
				}
			},
		},
		{
			name:  "hex string",
			input: "<48656C6C6F>",
			check: func(t *testing.T, obj Object) {
				if v, ok := obj.(HexString); !ok || v != "Hello" {
					t.Errorf("got %q (%T)", obj, obj)
				}
			},
		},
		{
			name:  "indirect reference",
			input: "10 0 R ",
			check: func(t *testing.T, obj Object) {
				ref, ok := obj.(ObjectRef)
				if !ok {
					t.Fatalf("got %T", obj)
				}
				if ref.Number != 10 || ref.Generation != 0 {
					t.Errorf("got %v", ref)
				}
			},
		},
		{
			name:  "array",
			input: "[1 2 /Name (str)]",
			check: func(t *testing.T, obj Object) {
				arr, ok := obj.(Array)
				if !ok {
					t.Fatalf("got %T", obj)
				}
				if len(arr) != 4 {
					t.Errorf("array len = %d, want 4", len(arr))
				}
			},
		},
		{
			name:  "dict",
			input: "<< /Type /Catalog /Pages 2 0 R >>",
			check: func(t *testing.T, obj Object) {
				d, ok := obj.(Dict)
				if !ok {
					t.Fatalf("got %T", obj)
				}
				if len(d) != 2 {
					t.Errorf("dict len = %d, want 2", len(d))
				}
				if d[Name("Type")] != Name("Catalog") {
					t.Errorf("/Type = %v", d[Name("Type")])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser([]byte(tt.input))
			obj, err := p.parseObject()
			if err != nil {
				t.Fatalf("parseObject: %v", err)
			}
			tt.check(t, obj)
		})
	}
}

func TestModifierOverlay(t *testing.T) {
	data := buildTestPDF(t, 2)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	m := NewModifier(r)

	// Overlay text on page 0.
	overlay := []byte("BT /F1 24 Tf 100 400 Td (OVERLAY) Tj ET")
	if err := m.OverlayPage(0, overlay, nil); err != nil {
		t.Fatalf("OverlayPage: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Verify the result is a valid PDF that can be re-read.
	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read modified PDF: %v", err)
	}
	count, err := r2.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 2 {
		t.Errorf("page count = %d, want 2", count)
	}

	// Verify the result contains the overlay content.
	if !bytes.Contains(result, []byte("OVERLAY")) {
		t.Error("modified PDF should contain overlay text")
	}

	// Result should be larger than original (incremental update appended).
	if len(result) <= len(data) {
		t.Errorf("modified PDF (%d bytes) should be larger than original (%d bytes)", len(result), len(data))
	}
}

func TestModifierOverlayAllPages(t *testing.T) {
	data := buildTestPDF(t, 3)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	m := NewModifier(r)
	count, _ := r.PageCount()
	for i := 0; i < count; i++ {
		overlay := []byte("BT /F1 10 Tf 50 50 Td (Footer) Tj ET")
		if err := m.OverlayPage(i, overlay, nil); err != nil {
			t.Fatalf("OverlayPage(%d): %v", i, err)
		}
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	r2, err := NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	count2, _ := r2.PageCount()
	if count2 != 3 {
		t.Errorf("page count = %d, want 3", count2)
	}
}

func TestModifierNoChanges(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	m := NewModifier(r)
	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.Equal(result, data) {
		t.Error("no-change modifier should produce identical output")
	}
}

func TestModifierWithResources(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	m := NewModifier(r)
	fontRef := m.AllocObject()
	m.SetObject(fontRef, Dict{
		Name("Type"):     Name("Font"),
		Name("Subtype"):  Name("Type1"),
		Name("BaseFont"): Name("Helvetica"),
	})

	resources := Dict{
		Name("Font"): Dict{
			Name("F99"): fontRef,
		},
	}
	overlay := []byte("BT /F99 18 Tf 100 100 Td (New Font) Tj ET")
	if err := m.OverlayPage(0, overlay, &resources); err != nil {
		t.Fatalf("OverlayPage: %v", err)
	}

	result, err := m.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Should be re-readable.
	_, err = NewReader(result)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
}
