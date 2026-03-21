package pdfa

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/pdf"
)

func TestGenerateXMP_A1b(t *testing.T) {
	xmp := generateXMP(LevelA1b, MetadataInfo{
		Title:    "Test Document",
		Author:   "Test Author",
		Producer: "gpdf-test",
		Creator:  "gpdf-test",
	})
	s := string(xmp)

	checks := []string{
		"<?xpacket begin=",
		"pdfaid:part>1</pdfaid:part",
		"pdfaid:conformance>B</pdfaid:conformance",
		"Test Document",
		"Test Author",
		"gpdf-test",
		"<?xpacket end=",
	}
	for _, c := range checks {
		if !strings.Contains(s, c) {
			t.Errorf("XMP missing %q", c)
		}
	}
}

func TestGenerateXMP_A2b(t *testing.T) {
	xmp := generateXMP(LevelA2b, MetadataInfo{
		Title:   "A2b Doc",
		Creator: "test",
	})
	s := string(xmp)
	if !strings.Contains(s, "pdfaid:part>2</pdfaid:part") {
		t.Error("expected pdfaid:part=2 for A2b")
	}
	if !strings.Contains(s, "pdfaid:conformance>B</pdfaid:conformance") {
		t.Error("expected conformance=B")
	}
}

func TestSRGBICCProfile(t *testing.T) {
	icc := sRGBICCProfile()
	if len(icc) < 128 {
		t.Fatalf("ICC profile too small: %d bytes", len(icc))
	}
	// Check header signature
	if string(icc[36:40]) != "acsp" {
		t.Errorf("ICC signature = %q, want 'acsp'", string(icc[36:40]))
	}
	// Check color space
	if string(icc[16:20]) != "RGB " {
		t.Errorf("color space = %q, want 'RGB '", string(icc[16:20]))
	}
	// Check device class
	if string(icc[12:16]) != "mntr" {
		t.Errorf("device class = %q, want 'mntr'", string(icc[12:16]))
	}
}

func TestApply_Integration(t *testing.T) {
	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)

	Apply(pw, WithLevel(LevelA1b), WithMetadata(MetadataInfo{
		Title:    "PDF/A Test",
		Author:   "Test",
		Producer: "gpdf",
		Creator:  "gpdf",
	}))

	// Add a minimal page
	err := pw.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}

	err = pw.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}

	got := buf.String()

	// Should contain OutputIntents in catalog
	if !strings.Contains(got, "/OutputIntents") {
		t.Error("missing /OutputIntents in PDF")
	}

	// Should contain Metadata in catalog
	if !strings.Contains(got, "/Metadata") {
		t.Error("missing /Metadata in PDF")
	}

	// Should contain XMP data
	if !strings.Contains(got, "pdfaid:part") {
		t.Error("missing pdfaid:part in XMP")
	}

	// Should contain GTS_PDFA1
	if !strings.Contains(got, "/GTS_PDFA1") {
		t.Error("missing /GTS_PDFA1 output intent")
	}

	// Should be valid PDF structure
	if !strings.Contains(got, "%PDF-1.7") {
		t.Error("missing PDF header")
	}
	if !strings.Contains(got, "%%EOF") {
		t.Error("missing EOF marker")
	}
}

func TestXMLEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a&b", "a&amp;b"},
		{"<tag>", "&lt;tag&gt;"},
		{"he said \"hi\"", "he said &quot;hi&quot;"},
	}
	for _, tt := range tests {
		got := xmlEscape(tt.input)
		if got != tt.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
