package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// ============================================================
// parseSVGDimensions
// ============================================================

func TestParseSVGDimensions_ViewBox(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100"><rect/></svg>`)
	w, h := parseSVGDimensions(svg)
	if w != 200 || h != 100 {
		t.Errorf("got (%g, %g), want (200, 100)", w, h)
	}
}

func TestParseSVGDimensions_WidthHeight(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="300" height="150"></svg>`)
	w, h := parseSVGDimensions(svg)
	if w != 300 || h != 150 {
		t.Errorf("got (%g, %g), want (300, 150)", w, h)
	}
}

func TestParseSVGDimensions_ViewBoxPreferred(t *testing.T) {
	// viewBox takes priority over width/height.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="500" height="500" viewBox="0 0 100 50"></svg>`)
	w, h := parseSVGDimensions(svg)
	if w != 100 || h != 50 {
		t.Errorf("got (%g, %g), want (100, 50)", w, h)
	}
}

func TestParseSVGDimensions_Fallback(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	w, h := parseSVGDimensions(svg)
	if w != 100 || h != 100 {
		t.Errorf("got (%g, %g), want (100, 100)", w, h)
	}
}

func TestParseSVGDimensions_XMLDeclaration(t *testing.T) {
	svg := []byte(`<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"></svg>`)
	w, h := parseSVGDimensions(svg)
	if w != 64 || h != 64 {
		t.Errorf("got (%g, %g), want (64, 64)", w, h)
	}
}

// ============================================================
// parseColor
// ============================================================

func TestParseColor_Named(t *testing.T) {
	tests := []struct {
		input string
		r, g, b float64
	}{
		{"black", 0, 0, 0},
		{"white", 1, 1, 1},
		{"red", 1, 0, 0},
		{"blue", 0, 0, 1},
	}
	for _, tt := range tests {
		c, ok := parseColor(tt.input)
		if !ok {
			t.Errorf("parseColor(%q): got not-ok", tt.input)
			continue
		}
		if c.None {
			t.Errorf("parseColor(%q): unexpected None", tt.input)
			continue
		}
		if abs(c.R-tt.r) > 0.01 || abs(c.G-tt.g) > 0.01 || abs(c.B-tt.b) > 0.01 {
			t.Errorf("parseColor(%q): got (%g,%g,%g), want (%g,%g,%g)",
				tt.input, c.R, c.G, c.B, tt.r, tt.g, tt.b)
		}
	}
}

func TestParseColor_Hex6(t *testing.T) {
	c, ok := parseColor("#ff8000")
	if !ok || c.None {
		t.Fatal("expected valid color")
	}
	if abs(c.R-1) > 0.01 || abs(c.G-0.502) > 0.01 || abs(c.B-0) > 0.01 {
		t.Errorf("got (%g,%g,%g)", c.R, c.G, c.B)
	}
}

func TestParseColor_Hex3(t *testing.T) {
	c, ok := parseColor("#f80")
	if !ok || c.None {
		t.Fatal("expected valid color")
	}
	// #f80 = #ff8800
	if abs(c.R-1) > 0.01 || abs(c.G-0.533) > 0.01 || abs(c.B-0) > 0.01 {
		t.Errorf("got (%g,%g,%g)", c.R, c.G, c.B)
	}
}

func TestParseColor_RGB(t *testing.T) {
	c, ok := parseColor("rgb(255, 128, 0)")
	if !ok || c.None {
		t.Fatal("expected valid color")
	}
	if abs(c.R-1) > 0.01 || abs(c.G-0.502) > 0.01 || abs(c.B-0) > 0.01 {
		t.Errorf("got (%g,%g,%g)", c.R, c.G, c.B)
	}
}

func TestParseColor_None(t *testing.T) {
	c, ok := parseColor("none")
	if !ok || !c.None {
		t.Error("expected None color")
	}
}

func TestParseColor_Unknown(t *testing.T) {
	_, ok := parseColor("notacolor")
	if ok {
		t.Error("expected not-ok for unknown color")
	}
}

// ============================================================
// parseTransform
// ============================================================

func TestParseTransform_Translate(t *testing.T) {
	m := parseTransform("translate(10, 20)")
	// Expected: [1 0 0 1 10 20]
	if abs(m[0]-1) > 1e-9 || abs(m[3]-1) > 1e-9 || abs(m[4]-10) > 1e-9 || abs(m[5]-20) > 1e-9 {
		t.Errorf("translate: got %v", m)
	}
}

func TestParseTransform_Scale(t *testing.T) {
	m := parseTransform("scale(2)")
	// Expected: [2 0 0 2 0 0]
	if abs(m[0]-2) > 1e-9 || abs(m[3]-2) > 1e-9 {
		t.Errorf("scale: got %v", m)
	}
}

func TestParseTransform_ScaleXY(t *testing.T) {
	m := parseTransform("scale(3, 4)")
	if abs(m[0]-3) > 1e-9 || abs(m[3]-4) > 1e-9 {
		t.Errorf("scale(3,4): got %v", m)
	}
}

func TestParseTransform_Matrix(t *testing.T) {
	m := parseTransform("matrix(1 2 3 4 5 6)")
	want := matrix6{1, 2, 3, 4, 5, 6}
	for i := range m {
		if abs(m[i]-want[i]) > 1e-9 {
			t.Errorf("matrix: got %v, want %v", m, want)
			break
		}
	}
}

func TestParseTransform_Combined(t *testing.T) {
	// translate(10,0) scale(2) → matrix = T * S = [2 0 0 2 10 0]
	m := parseTransform("translate(10,0) scale(2)")
	if abs(m[0]-2) > 1e-9 || abs(m[3]-2) > 1e-9 || abs(m[4]-10) > 1e-9 || abs(m[5]-0) > 1e-9 {
		t.Errorf("combined: got %v", m)
	}
}

// ============================================================
// convertPathData
// ============================================================

func TestConvertPathData_MoveTo(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 10 20", &buf)
	if !strings.Contains(buf.String(), "10 20 m") {
		t.Errorf("got: %q", buf.String())
	}
}

func TestConvertPathData_LineTo(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 L 10 20", &buf)
	out := buf.String()
	if !strings.Contains(out, "10 20 l") {
		t.Errorf("got: %q", out)
	}
}

func TestConvertPathData_ClosePath(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 L 10 0 L 5 10 Z", &buf)
	if !strings.Contains(buf.String(), "h\n") {
		t.Errorf("expected closepath (h), got: %q", buf.String())
	}
}

func TestConvertPathData_HorizontalVertical(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 H 50 V 30", &buf)
	out := buf.String()
	if !strings.Contains(out, "50 0 l") {
		t.Errorf("H expected '50 0 l', got: %q", out)
	}
	if !strings.Contains(out, "50 30 l") {
		t.Errorf("V expected '50 30 l', got: %q", out)
	}
}

func TestConvertPathData_CubicBezier(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 C 10 20 30 40 50 60", &buf)
	if !strings.Contains(buf.String(), "10 20 30 40 50 60 c") {
		t.Errorf("got: %q", buf.String())
	}
}

func TestConvertPathData_RelativeCommands(t *testing.T) {
	var buf strings.Builder
	convertPathData("m 5 10 l 20 0 l 0 15 z", &buf)
	out := buf.String()
	if !strings.Contains(out, "5 10 m") {
		t.Errorf("relative m: got %q", out)
	}
}

func TestConvertPathData_ImplicitRepeat(t *testing.T) {
	var buf strings.Builder
	// M with multiple coordinate pairs: first is moveto, rest are lineto.
	convertPathData("M 0 0 10 20 30 40", &buf)
	out := buf.String()
	if !strings.Contains(out, "0 0 m") {
		t.Errorf("expected initial moveto, got: %q", out)
	}
	if !strings.Contains(out, "10 20 l") {
		t.Errorf("expected implicit lineto, got: %q", out)
	}
}

func TestConvertPathData_Arc(t *testing.T) {
	var buf strings.Builder
	// Simple semicircle arc.
	convertPathData("M 0 0 A 50 50 0 0 1 100 0", &buf)
	out := buf.String()
	// Arc should produce at least one Bézier curve.
	if !strings.Contains(out, " c\n") {
		t.Errorf("arc should produce Bézier curves, got: %q", out)
	}
}

// ============================================================
// svgToFormContent
// ============================================================

func TestSVGToFormContent_Rect(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">
		<rect x="10" y="10" width="80" height="30" fill="red"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	if fc.ViewW != 100 || fc.ViewH != 50 {
		t.Errorf("dimensions: got (%g, %g), want (100, 50)", fc.ViewW, fc.ViewH)
	}
	content := string(fc.Content)
	if !strings.Contains(content, "re\n") {
		t.Errorf("expected rectangle operator, got:\n%s", content)
	}
	if !strings.Contains(content, "1 0 0 rg") {
		t.Errorf("expected red fill (1 0 0 rg), got:\n%s", content)
	}
}

func TestSVGToFormContent_Circle(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<circle cx="50" cy="50" r="40" fill="blue" stroke="black" stroke-width="2"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	// Circle should produce Bézier curves.
	if !strings.Contains(content, " c\n") {
		t.Errorf("circle should use Bézier curves, got:\n%s", content)
	}
	// Should have both fill and stroke → B operator.
	if !strings.Contains(content, "\nB\n") {
		t.Errorf("expected fill+stroke (B), got:\n%s", content)
	}
}

func TestSVGToFormContent_Path(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<path d="M 10 10 L 90 10 L 50 90 Z" fill="green"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	if !strings.Contains(content, "m\n") {
		t.Errorf("expected moveto, got:\n%s", content)
	}
}

func TestSVGToFormContent_StrokeOnly(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<line x1="0" y1="0" x2="100" y2="100" stroke="black" stroke-width="2"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	// Line has no fill → S operator.
	if !strings.Contains(content, "\nS\n") {
		t.Errorf("expected stroke-only (S), got:\n%s", content)
	}
}

func TestSVGToFormContent_GroupTransform(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<g transform="translate(10 10)">
			<rect x="0" y="0" width="50" height="50" fill="blue"/>
		</g>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	// Group transform should emit cm operator.
	if !strings.Contains(content, " cm\n") {
		t.Errorf("expected transform (cm), got:\n%s", content)
	}
}

func TestSVGToFormContent_Opacity(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="0" y="0" width="100" height="100" fill="red" opacity="0.5"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	// Should have ExtGState resources for opacity.
	if fc.Resources == nil {
		t.Error("expected Resources for opacity")
	}
	content := string(fc.Content)
	if !strings.Contains(content, " gs\n") {
		t.Errorf("expected ExtGState reference (gs), got:\n%s", content)
	}
}

func TestSVGToFormContent_FillNone(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="0" y="0" width="100" height="100" fill="none" stroke="black"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	// fill=none with stroke → S operator.
	if !strings.Contains(content, "\nS\n") {
		t.Errorf("expected stroke-only (S) for fill=none, got:\n%s", content)
	}
}

func TestSVGToFormContent_InlineStyle(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="0" y="0" width="100" height="100" style="fill:red;stroke:blue;stroke-width:3"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	if !strings.Contains(content, "1 0 0 rg") {
		t.Errorf("expected red fill from style attr, got:\n%s", content)
	}
	if !strings.Contains(content, "0 0 1 RG") {
		t.Errorf("expected blue stroke from style attr, got:\n%s", content)
	}
}

func TestSVGToFormContent_Polygon(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<polygon points="50,10 90,90 10,90" fill="yellow"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	if !strings.Contains(content, "h\n") {
		t.Errorf("polygon should have closepath (h), got:\n%s", content)
	}
}

func TestSVGToFormContent_RoundedRect(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="10" y="10" width="80" height="80" rx="10" ry="10" fill="purple"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	// Rounded rect uses Bézier curves.
	if !strings.Contains(content, " c\n") {
		t.Errorf("rounded rect should use Bézier curves, got:\n%s", content)
	}
}

func TestSVGToFormContent_Ellipse(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100">
		<ellipse cx="100" cy="50" rx="90" ry="40" fill="teal"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	if !strings.Contains(string(fc.Content), " c\n") {
		t.Errorf("ellipse should use Bézier curves")
	}
}

func TestSVGToFormContent_Polyline(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<polyline points="10,10 50,50 90,10" stroke="black" fill="none"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	content := string(fc.Content)
	// Polyline should not close the path and should stroke only.
	if !strings.Contains(content, "\nS\n") {
		t.Errorf("polyline: expected stroke-only (S), got:\n%s", content)
	}
}

func TestSVGToFormContent_NoFillNoStroke(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="0" y="0" width="100" height="100" fill="none" stroke="none"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	if !strings.Contains(string(fc.Content), "\nn\n") {
		t.Errorf("no fill/stroke: expected no-op (n)")
	}
}

func TestSVGToFormContent_IgnoredElements(t *testing.T) {
	// defs, title, desc should be silently ignored.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<defs><marker id="m"/></defs>
		<title>Test</title>
		<desc>A description</desc>
		<rect x="0" y="0" width="100" height="100" fill="red"/>
	</svg>`)
	fc, err := svgToFormContent(svg)
	if err != nil {
		t.Fatalf("svgToFormContent: %v", err)
	}
	if len(fc.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

// ============================================================
// Path data: additional command coverage
// ============================================================

func TestConvertPathData_SmoothCubic(t *testing.T) {
	var buf strings.Builder
	// S uses reflection of previous C control point.
	convertPathData("M 0 0 C 10 -10 20 10 30 0 S 50 -10 60 0", &buf)
	out := buf.String()
	if strings.Count(out, " c\n") < 2 {
		t.Errorf("expected 2 cubic segments, got:\n%s", out)
	}
}

func TestConvertPathData_QuadraticBezier(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 Q 50 -50 100 0", &buf)
	// Q converts to cubic.
	if !strings.Contains(buf.String(), " c\n") {
		t.Errorf("quadratic should produce cubic, got:\n%s", buf.String())
	}
}

func TestConvertPathData_SmoothQuadratic(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 Q 50 -50 100 0 T 200 0", &buf)
	if strings.Count(buf.String(), " c\n") < 2 {
		t.Errorf("expected 2 cubic segments for Q+T, got:\n%s", buf.String())
	}
}

func TestConvertPathData_RelativeSmoothCubic(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 c 10 -10 20 10 30 0 s 20 10 30 0", &buf)
	if strings.Count(buf.String(), " c\n") < 2 {
		t.Errorf("expected 2 cubic segments (relative), got:\n%s", buf.String())
	}
}

func TestConvertPathData_RelativeQuadratic(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 0 0 q 25 -25 50 0 t 50 0", &buf)
	if strings.Count(buf.String(), " c\n") < 2 {
		t.Errorf("expected 2 cubic segments (relative q+t), got:\n%s", buf.String())
	}
}

func TestConvertPathData_ScientificNotation(t *testing.T) {
	var buf strings.Builder
	convertPathData("M 1e2 2e1", &buf)
	if !strings.Contains(buf.String(), "100 20 m") {
		t.Errorf("scientific notation: got %q", buf.String())
	}
}

// ============================================================
// parseTransform: additional coverage
// ============================================================

func TestParseTransform_Rotate(t *testing.T) {
	m := parseTransform("rotate(90)")
	// 90° rotation: [cos90 sin90 -sin90 cos90 0 0] ≈ [0 1 -1 0 0 0]
	if abs(m[0]-0) > 0.001 || abs(m[1]-1) > 0.001 || abs(m[2]+1) > 0.001 || abs(m[3]-0) > 0.001 {
		t.Errorf("rotate(90): got %v", m)
	}
}

func TestParseTransform_RotateAroundPoint(t *testing.T) {
	m := parseTransform("rotate(90, 50, 50)")
	// Rotating 90° around (50,50): translation should be non-zero.
	if abs(m[4]-100) > 0.001 || abs(m[5]-0) > 0.001 {
		t.Errorf("rotate(90,50,50): unexpected translation (%g,%g)", m[4], m[5])
	}
}

func TestParseTransform_SkewX(t *testing.T) {
	m := parseTransform("skewX(45)")
	// skewX(45): [1 0 tan(45) 1 0 0] ≈ [1 0 1 1 0 0]
	if abs(m[0]-1) > 0.001 || abs(m[2]-1) > 0.01 {
		t.Errorf("skewX(45): got %v", m)
	}
}

func TestParseTransform_SkewY(t *testing.T) {
	m := parseTransform("skewY(45)")
	if abs(m[3]-1) > 0.001 || abs(m[1]-1) > 0.01 {
		t.Errorf("skewY(45): got %v", m)
	}
}

// ============================================================
// parseColorComponent: percentage
// ============================================================

func TestParseColorComponent_Percentage(t *testing.T) {
	c, ok := parseColor("rgb(100%, 50%, 0%)")
	if !ok {
		t.Fatal("expected valid color")
	}
	if abs(c.R-1) > 0.01 || abs(c.G-0.5) > 0.01 || abs(c.B-0) > 0.01 {
		t.Errorf("rgb percentages: got (%g,%g,%g)", c.R, c.G, c.B)
	}
}

// ============================================================
// parseSVGLength: unit conversions
// ============================================================

func TestParseSVGLength_Units(t *testing.T) {
	tests := []struct {
		input string
		min   float64 // expected minimum value
	}{
		{"96px", 95},
		{"72pt", 95},  // 72pt = 96px
		{"1in", 95},   // 1in = 96px
		{"25.4mm", 95}, // 25.4mm = 1in = 96px
	}
	for _, tt := range tests {
		got := parseSVGLength(tt.input)
		if got < tt.min {
			t.Errorf("parseSVGLength(%q): got %g, want >= %g", tt.input, got, tt.min)
		}
	}
}

// ============================================================
// pdf.Writer RegisterFormXObject
// ============================================================

func TestRegisterFormXObject(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	content := []byte("q 1 0 0 rg 0 0 100 100 re f Q\n")
	bbox := pdf.Rectangle{LLX: 0, LLY: 0, URX: 100, URY: 100}
	matrix := [6]float64{0.01, 0, 0, -0.01, 0, 1}

	resName, ref, err := w.RegisterFormXObject("testform", content, bbox, matrix, nil)
	if err != nil {
		t.Fatalf("RegisterFormXObject: %v", err)
	}
	if resName != "Fm1" {
		t.Errorf("resName = %q, want Fm1", resName)
	}
	if ref.Number == 0 {
		t.Error("expected non-zero object ref")
	}
}

func TestParseColor_RGBA(t *testing.T) {
	c, ok := parseColor("rgba(255, 0, 0, 0.5)")
	if !ok || c.None {
		t.Fatal("expected valid color from rgba()")
	}
	if abs(c.R-1) > 0.01 {
		t.Errorf("rgba red channel: got %g, want 1", c.R)
	}
}

func TestArcToBezier_ZeroRadius(t *testing.T) {
	var buf strings.Builder
	// Arc with zero radius should fall back to a line.
	convertPathData("M 0 0 A 0 0 0 0 1 50 50", &buf)
	if !strings.Contains(buf.String(), "50 50 l") {
		t.Errorf("zero-radius arc should produce lineto, got:\n%s", buf.String())
	}
}

func TestArcToBezier_SamePoint(t *testing.T) {
	var buf strings.Builder
	// Arc where start == end should be a no-op.
	convertPathData("M 50 50 A 25 25 0 0 1 50 50", &buf)
	// Should only have the moveto, no extra lines.
	out := buf.String()
	if strings.Contains(out, " l\n") || strings.Contains(out, " c\n") {
		t.Errorf("same-point arc should produce no extra drawing ops, got:\n%s", out)
	}
}

func TestRegisterFormXObject_Deduplication(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	content := []byte("q Q\n")
	bbox := pdf.Rectangle{LLX: 0, LLY: 0, URX: 10, URY: 10}
	matrix := [6]float64{1, 0, 0, 1, 0, 0}

	resName1, ref1, _ := w.RegisterFormXObject("key1", content, bbox, matrix, nil)
	resName2, ref2, _ := w.RegisterFormXObject("key1", content, bbox, matrix, nil)

	if resName1 != resName2 {
		t.Errorf("duplicate: got different names %q and %q", resName1, resName2)
	}
	if ref1.Number != ref2.Number {
		t.Errorf("duplicate: got different object numbers %d and %d", ref1.Number, ref2.Number)
	}
}

// ============================================================
// Integration: RenderImage with SVG
// ============================================================

func TestRenderImage_SVG(t *testing.T) {
	r, buf := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="10" y="10" width="80" height="80" fill="red"/>
	</svg>`)

	src := document.ImageSource{
		Data:   svgData,
		Format: document.ImageSVG,
		Width:  100,
		Height: 100,
	}

	err := r.RenderImage(src, document.Point{X: 50, Y: 100}, document.Size{Width: 200, Height: 200})
	if err != nil {
		t.Fatalf("RenderImage SVG: %v", err)
	}

	content := string(r.pageContent)
	if !strings.Contains(content, "Do\n") {
		t.Errorf("expected Do operator, got:\n%s", content)
	}
	if !strings.Contains(content, " cm\n") {
		t.Errorf("expected cm operator, got:\n%s", content)
	}

	// Complete the page and produce a PDF.
	if err := r.EndPage(); err != nil {
		t.Fatalf("EndPage: %v", err)
	}
	pw := getPDFWriter(r)
	if err := pw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data := buf.Bytes()
	if len(data) == 0 {
		t.Fatal("empty PDF")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("invalid PDF header: %q", string(data[:5]))
	}
}

func TestRenderImage_SVGDeduplication(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 50 50">
		<circle cx="25" cy="25" r="20" fill="blue"/>
	</svg>`)

	src := document.ImageSource{
		Data:   svgData,
		Format: document.ImageSVG,
		Width:  50,
		Height: 50,
	}

	if err := r.RenderImage(src, document.Point{X: 0, Y: 0}, document.Size{Width: 100, Height: 100}); err != nil {
		t.Fatalf("first render: %v", err)
	}
	if err := r.RenderImage(src, document.Point{X: 200, Y: 0}, document.Size{Width: 100, Height: 100}); err != nil {
		t.Fatalf("second render: %v", err)
	}

	// Only one form XObject should be registered.
	if len(r.imageMap) != 1 {
		t.Errorf("expected 1 form XObject, got %d", len(r.imageMap))
	}
}

// getPDFWriter extracts the writer from a PDFRenderer (for test teardown).
func getPDFWriter(r *PDFRenderer) *pdf.Writer {
	return r.writer
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
