package template

import (
	"os"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		want  document.Value
	}{
		{"15mm", document.Mm(15)},
		{"20pt", document.Pt(20)},
		{"2.5cm", document.Cm(2.5)},
		{"1in", document.In(1)},
		{"1.5em", document.Em(1.5)},
		{"50%", document.Pct(50)},
		{"72", document.Pt(72)},
		{"auto", document.Auto},
		{"", document.Auto},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseValue(tt.input)
			if err != nil {
				t.Fatalf("parseValue(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseValue(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseValueError(t *testing.T) {
	bad := []string{"abc", "12xyz", "mm15"}
	for _, s := range bad {
		if _, err := parseValue(s); err == nil {
			t.Errorf("parseValue(%q) expected error, got nil", s)
		}
	}
}

func TestParsePageSize(t *testing.T) {
	tests := []struct {
		input string
		want  document.Size
	}{
		{"A4", document.A4},
		{"a4", document.A4},
		{"A3", document.A3},
		{"Letter", document.Letter},
		{"legal", document.Legal},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parsePageSize(tt.input)
			if err != nil {
				t.Fatalf("parsePageSize(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parsePageSize(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}

	if _, err := parsePageSize("Unknown"); err == nil {
		t.Error("parsePageSize(Unknown) expected error")
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input string
		want  pdf.Color
	}{
		{"black", pdf.Black},
		{"white", pdf.White},
		{"red", pdf.Red},
		{"#FF0000", pdf.RGBHex(0xFF0000)},
		{"#1A237E", pdf.RGBHex(0x1A237E)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseColor(tt.input)
			if err != nil {
				t.Fatalf("parseColor(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseColor(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}

	badColors := []string{"#GG0000", "unknown", "#FFF"}
	for _, s := range badColors {
		if _, err := parseColor(s); err == nil {
			t.Errorf("parseColor(%q) expected error, got nil", s)
		}
	}
}

func TestApplySchemaStyle(t *testing.T) {
	// nil style should return nil.
	if opts := applySchemaStyle(nil); opts != nil {
		t.Error("applySchemaStyle(nil) should return nil")
	}

	ss := &SchemaStyle{
		Size:          18,
		Bold:          true,
		Italic:        true,
		Align:         "center",
		Color:         "red",
		Background:    "#F5F5F5",
		Underline:     true,
		Strikethrough: true,
		LetterSpacing: 1.5,
	}
	opts := applySchemaStyle(ss)
	if len(opts) == 0 {
		t.Fatal("expected non-empty options from full SchemaStyle")
	}

	// Apply to a default style and verify some fields.
	style := document.DefaultStyle()
	for _, opt := range opts {
		opt(&style)
	}
	if style.FontSize != 18 {
		t.Errorf("FontSize = %v, want 18", style.FontSize)
	}
	if style.FontWeight != document.WeightBold {
		t.Error("expected bold")
	}
	if style.FontStyle != document.StyleItalic {
		t.Error("expected italic")
	}
	if style.TextAlign != document.AlignCenter {
		t.Error("expected center align")
	}
	if style.TextDecoration&document.DecorationUnderline == 0 {
		t.Error("expected underline")
	}
	if style.TextDecoration&document.DecorationStrikethrough == 0 {
		t.Error("expected strikethrough")
	}
	if style.LetterSpacing != 1.5 {
		t.Errorf("LetterSpacing = %v, want 1.5", style.LetterSpacing)
	}
}

func TestBuildFromSchema_Basic(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{
			Size:    "A4",
			Margins: "20mm",
		},
		Metadata: &SchemaMeta{
			Title:  "Test Document",
			Author: "gpdf",
		},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Text: "Hello, World!", Style: &SchemaStyle{Size: 24, Bold: true}},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("generated PDF is empty")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("invalid PDF header: %q", string(data[:5]))
	}
}

func TestBuildFromSchema_FixedHeightRow(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "Letter"},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Height: "25mm",
				Cols: []SchemaCol{
					{Span: 6, Text: "Left"},
					{Span: 6, Text: "Right"},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_HeaderFooter(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4", Margins: "15mm"},
		Header: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Text: "Page Header", Style: &SchemaStyle{Bold: true}},
				},
			}},
		},
		Footer: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Elements: []SchemaElement{
						{Type: "pageNumber", Style: &SchemaStyle{Align: "center"}},
					}},
				},
			}},
		},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Text: "Body content"},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_Elements(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4"},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Elements: []SchemaElement{
						{Type: "text", Content: "Title", Style: &SchemaStyle{Size: 24, Bold: true}},
						{Type: "spacer", Height: "10mm"},
						{Type: "line", Line: &SchemaLine{Color: "#CCCCCC", Thickness: "2pt"}},
						{Type: "spacer", Height: "5mm"},
						{Type: "text", Content: "Body text"},
						{Type: "list", List: &SchemaList{Items: []string{"Item 1", "Item 2", "Item 3"}}},
						{Type: "list", List: &SchemaList{Type: "ordered", Items: []string{"First", "Second"}}},
					}},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_Table(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4"},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Table: &SchemaTable{
						Header:       []string{"Name", "Qty", "Price"},
						Rows:         [][]string{{"Widget", "10", "$5.00"}, {"Gadget", "3", "$15.00"}},
						ColumnWidths: []float64{50, 25, 25},
						HeaderStyle:  &SchemaStyle{Bold: true, Color: "white", Background: "#1A237E"},
						StripeColor:  "#F5F5F5",
					}},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_QRCodeBarcode(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4"},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 6, QRCode: &SchemaQRCode{Data: "https://gpdf.dev", Size: "30mm"}},
					{Span: 6, Barcode: &SchemaBarcode{Data: "ABC-123", Width: "50mm", Height: "20mm", Format: "code128"}},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_Line(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4"},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Line: &SchemaLine{Color: "red", Thickness: "2pt"}},
				},
			}},
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Elements: []SchemaElement{
						{Type: "line"}, // default line
					}},
				},
			}},
		},
	}

	doc, err := buildFromSchema(schema, nil)
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_WithOptions(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4"},
		Body: []SchemaRow{
			{Row: SchemaRowDef{
				Cols: []SchemaCol{
					{Span: 12, Text: "With custom metadata"},
				},
			}},
		},
	}

	// Options passed to buildFromSchema override schema settings.
	doc, err := buildFromSchema(schema, []Option{
		WithMetadata(document.DocumentMetadata{Title: "Override Title"}),
	})
	if err != nil {
		t.Fatalf("buildFromSchema error: %v", err)
	}
	if _, err := doc.Generate(); err != nil {
		t.Fatalf("Generate error: %v", err)
	}
}

func TestBuildFromSchema_InvalidPageSize(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "B5"},
		Body: []SchemaRow{},
	}
	if _, err := buildFromSchema(schema, nil); err == nil {
		t.Error("expected error for unknown page size")
	}
}

func TestBuildFromSchema_InvalidMargins(t *testing.T) {
	schema := &Schema{
		Page: SchemaPage{Size: "A4", Margins: "abc"},
		Body: []SchemaRow{},
	}
	if _, err := buildFromSchema(schema, nil); err == nil {
		t.Error("expected error for invalid margins")
	}
}

func TestParseFitMode(t *testing.T) {
	tests := []struct {
		input string
		want  document.ImageFitMode
		ok    bool
	}{
		{"contain", document.FitContain, true},
		{"cover", document.FitCover, true},
		{"stretch", document.FitStretch, true},
		{"original", document.FitOriginal, true},
		{"Contain", document.FitContain, true},
		{"COVER", document.FitCover, true},
		{"invalid", document.FitContain, false},
		{"", document.FitContain, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parseFitMode(tt.input)
			if ok != tt.ok {
				t.Errorf("parseFitMode(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("parseFitMode(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseImageAlign(t *testing.T) {
	tests := []struct {
		input string
		want  document.TextAlign
		ok    bool
	}{
		{"left", document.AlignLeft, true},
		{"center", document.AlignCenter, true},
		{"right", document.AlignRight, true},
		{"Center", document.AlignCenter, true},
		{"invalid", document.AlignLeft, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parseImageAlign(tt.input)
			if ok != tt.ok {
				t.Errorf("parseImageAlign(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("parseImageAlign(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"/absolute/path.png", true},
		{"./relative/path.png", true},
		{"../parent/path.png", true},
		{"C:/windows/path.png", true},
		{"D:\\data\\image.png", true},
		{"iVBORw0KGgo=", false},  // base64 string
		{"data:image/png;", false}, // data URI (handled before isFilePath)
		{"some-string", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isFilePath(tt.input); got != tt.want {
				t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadImageData_FilePath(t *testing.T) {
	// Write a temporary file and load it.
	tmpDir := t.TempDir()
	testData := []byte{0x89, 'P', 'N', 'G', 1, 2, 3}
	path := tmpDir + "/test.png"
	if err := os.WriteFile(path, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// Absolute path
	got, err := loadImageData(path)
	if err != nil {
		t.Fatalf("loadImageData(%q) error: %v", path, err)
	}
	if len(got) != len(testData) {
		t.Errorf("loadImageData returned %d bytes, want %d", len(got), len(testData))
	}

	// file:// URI
	got, err = loadImageData("file://" + path)
	if err != nil {
		t.Fatalf("loadImageData(file://) error: %v", err)
	}
	if len(got) != len(testData) {
		t.Errorf("file:// loadImageData returned %d bytes, want %d", len(got), len(testData))
	}
}

func TestLoadImageData_DataURI(t *testing.T) {
	// data URI with valid base64
	data, err := loadImageData("data:image/png;base64,AQID")
	if err != nil {
		t.Fatalf("loadImageData(data URI) error: %v", err)
	}
	if len(data) != 3 || data[0] != 1 || data[1] != 2 || data[2] != 3 {
		t.Errorf("unexpected data: %v", data)
	}
}

func TestLoadImageData_RawBase64(t *testing.T) {
	data, err := loadImageData("AQID")
	if err != nil {
		t.Fatalf("loadImageData(raw base64) error: %v", err)
	}
	if len(data) != 3 {
		t.Errorf("expected 3 bytes, got %d", len(data))
	}
}

func TestSchemaImage_FitAndAlign(t *testing.T) {
	img := SchemaImage{
		Src:   "AQID",
		Fit:   "cover",
		Align: "center",
	}
	if img.Fit != "cover" {
		t.Errorf("Fit = %q, want %q", img.Fit, "cover")
	}
	if img.Align != "center" {
		t.Errorf("Align = %q, want %q", img.Align, "center")
	}
}
