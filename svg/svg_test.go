package svg

import (
	"math"
	"strings"
	"testing"
)

func mustParse(t *testing.T, src string) *Image {
	t.Helper()
	img, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return img
}

func render(t *testing.T, src string) string {
	t.Helper()
	return string(mustParse(t, src).Render().Content)
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestParseErrors(t *testing.T) {
	for _, src := range []string{"", "not xml", "<html></html>", "<svg"} {
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", src)
		}
	}
	deep := strings.Repeat("<g>", maxDepth+10)
	if _, err := Parse([]byte("<svg>" + deep)); err == nil {
		t.Error("deeply nested input should be rejected")
	}
}

func TestIntrinsicSize(t *testing.T) {
	tests := []struct {
		src  string
		w, h float64
	}{
		{`<svg width="24" height="12"/>`, 24, 12},
		{`<svg width="10mm" height="1in"/>`, 72 / 2.54, 72},
		{`<svg viewBox="0 0 200 100"/>`, 200, 100},
		{`<svg width="50" viewBox="0 0 200 100"/>`, 50, 25},
		{`<svg height="50" viewBox="0 0 200 100"/>`, 100, 50},
		{`<svg width="100%" height="100%" viewBox="0 0 30 40"/>`, 30, 40},
		{`<svg/>`, 300, 150},
		{`<svg viewBox="0 0 -1 10"/>`, 300, 150},
		{`<svg width="1e9" height="5"/>`, 300, 5},
	}
	for _, tc := range tests {
		img := mustParse(t, tc.src)
		if !near(img.Width, tc.w) || !near(img.Height, tc.h) {
			t.Errorf("%s: size = %gx%g, want %gx%g", tc.src, img.Width, img.Height, tc.w, tc.h)
		}
	}
}

func TestUnitMatrix(t *testing.T) {
	// viewBox wider than viewport aspect: meet scales by height... width.
	img := mustParse(t, `<svg width="100" height="100" viewBox="0 0 200 100"/>`)
	m := img.unitMatrix()
	// Point (0,0) maps to (0, 0.75): content is centered vertically.
	x, y := m[4], m[5]
	if !near(x, 0) || !near(y, 0.75) {
		t.Errorf("origin maps to (%g,%g), want (0,0.75)", x, y)
	}
	// Point (200,100) maps to (1, 0.25).
	x = m[0]*200 + m[2]*100 + m[4]
	y = m[1]*200 + m[3]*100 + m[5]
	if !near(x, 1) || !near(y, 0.25) {
		t.Errorf("corner maps to (%g,%g), want (1,0.25)", x, y)
	}

	img = mustParse(t, `<svg width="100" height="100" viewBox="0 0 200 100" preserveAspectRatio="none"/>`)
	m = img.unitMatrix()
	if !near(m[0], 1.0/200) || !near(m[3], -1.0/100) || !near(m[5], 1) {
		t.Errorf("none matrix = %v", m)
	}
}

func segOps(segs []segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteByte(s.op)
	}
	return b.String()
}

func TestPathData(t *testing.T) {
	tests := []struct {
		d    string
		ops  string
		last [2]float64 // end point of last drawing segment
	}{
		{"M10 10 L20 20", "ML", [2]float64{20, 20}},
		{"M10,10 20,20 30,10", "MLL", [2]float64{30, 10}},
		{"m10 10 10 10", "ML", [2]float64{20, 20}},
		{"M0 0H10V10h-5v-5z", "MLLLLZ", [2]float64{5, 5}},
		{"M0 0C1 1 2 2 3 3S5 5 6 6", "MCC", [2]float64{6, 6}},
		{"M0 0Q5 5 10 0T20 0", "MCC", [2]float64{20, 0}},
		{"M-1-2.5L1e1 1E-1", "ML", [2]float64{10, 0.1}},
		{"M.5.5l.5.5", "ML", [2]float64{1, 1}},
		{"M0 0 L 10", "M", [2]float64{0, 0}}, // malformed: stop at error
		{"L10 10", "ML", [2]float64{10, 10}}, // implicit start at origin
		{"M0 0z l5 5", "MZML", [2]float64{5, 5}},
	}
	for _, tc := range tests {
		segs := parsePathData(tc.d)
		if got := segOps(segs); got != tc.ops {
			t.Errorf("%q: ops = %q, want %q", tc.d, got, tc.ops)
			continue
		}
		for i := len(segs) - 1; i >= 0; i-- {
			s := segs[i]
			var end [2]float64
			switch s.op {
			case 'Z':
				continue
			case 'C':
				end = [2]float64{s.p[4], s.p[5]}
			default:
				end = [2]float64{s.p[0], s.p[1]}
			}
			if !near(end[0], tc.last[0]) || !near(end[1], tc.last[1]) {
				t.Errorf("%q: end = %v, want %v", tc.d, end, tc.last)
			}
			break
		}
	}
}

func TestSmoothCurveReflection(t *testing.T) {
	segs := parsePathData("M0 0C0 10 10 10 10 0S20 -10 20 0")
	s := segs[2]
	// Reflection of (10,10) about (10,0) is (10,-10).
	if !near(s.p[0], 10) || !near(s.p[1], -10) {
		t.Errorf("reflected control = (%g,%g), want (10,-10)", s.p[0], s.p[1])
	}
}

func TestArc(t *testing.T) {
	// Half circle of radius 10 from (0,0) to (20,0); flags packed together.
	segs := parsePathData("M0 0A10 10 0 0120 0")
	if len(segs) < 3 {
		t.Fatalf("arc produced %d segments: %v", len(segs), segs)
	}
	last := segs[len(segs)-1]
	if last.op != 'C' || !near(last.p[4], 20) || !near(last.p[5], 0) {
		t.Errorf("arc end = %v, want (20,0)", last.p)
	}
	// Midpoint of the sweep lies on the circle (sweep=1: y negative).
	mid := segs[1]
	if math.Abs(math.Hypot(mid.p[4]-10, mid.p[5])-10) > 1e-6 || mid.p[5] > 0 {
		t.Errorf("arc midpoint %v not on the expected half circle", mid.p)
	}

	// Degenerate radius becomes a line; identical endpoints draw nothing.
	if got := segOps(parsePathData("M0 0A0 5 0 0 1 10 0")); got != "ML" {
		t.Errorf("zero-radius arc ops = %q, want ML", got)
	}
	if got := segOps(parsePathData("M0 0A5 5 0 0 1 0 0")); got != "M" {
		t.Errorf("zero-length arc ops = %q, want M", got)
	}
	// Radii too small are scaled up so the arc still reaches the end.
	segs = parsePathData("M0 0A1 1 0 0 1 100 0")
	if last := segs[len(segs)-1]; !near(last.p[4], 100) {
		t.Errorf("scaled arc end = %v", last.p)
	}
}

func TestTransform(t *testing.T) {
	m := parseTransform("translate(10,20) scale(2)")
	if m != (matrix{2, 0, 0, 2, 10, 20}) {
		t.Errorf("translate+scale = %v", m)
	}
	m = parseTransform("rotate(90 10 10)")
	// (20,10) rotated 90° about (10,10) is (10,20).
	x := m[0]*20 + m[2]*10 + m[4]
	y := m[1]*20 + m[3]*10 + m[5]
	if !near(x, 10) || !near(y, 20) {
		t.Errorf("rotate about center maps (20,10) to (%g,%g)", x, y)
	}
	for _, s := range []string{"matrix(1 0 0 1 5 6)", "skewX(45)", "skewY(45)", "translate(5)", "scale(1 2)", "rotate(30)"} {
		if parseTransform(s) == identity {
			t.Errorf("%s parsed as identity", s)
		}
	}
	if parseTransform("bogus(1) translate(5 5)") != identity {
		t.Error("unknown transform should stop parsing")
	}
}

func TestColors(t *testing.T) {
	cur := rgb{0.5, 0.5, 0.5, 1}
	tests := []struct {
		in   string
		want rgb
	}{
		{"#f00", rgb{1, 0, 0, 1}},
		{"#00FF00", rgb{0, 1, 0, 1}},
		{"#0000ff80", rgb{0, 0, 1, 128.0 / 255}},
		{"rgb(255, 0, 0)", rgb{1, 0, 0, 1}},
		{"rgb(100%,50%,0%)", rgb{1, 0.5, 0, 1}},
		{"rgba(0,0,255,0.5)", rgb{0, 0, 1, 0.5}},
		{"RebeccaPurple", rgb{0x66 / 255.0, 0x33 / 255.0, 0x99 / 255.0, 1}},
		{"currentColor", cur},
		{"transparent", rgb{}},
	}
	for _, tc := range tests {
		got, ok := parseColor(tc.in, cur)
		if !ok || !near(got.r, tc.want.r) || !near(got.g, tc.want.g) || !near(got.b, tc.want.b) || !near(got.a, tc.want.a) {
			t.Errorf("parseColor(%q) = %v, %v; want %v", tc.in, got, ok, tc.want)
		}
	}
	for _, bad := range []string{"#12", "#ggg", "rgb(1,2)", "notacolor", "rgb(a,b,c)"} {
		if _, ok := parseColor(bad, cur); ok {
			t.Errorf("parseColor(%q) succeeded", bad)
		}
	}
}

func TestLengths(t *testing.T) {
	tests := []struct {
		in   string
		ref  float64
		want float64
		ok   bool
	}{
		{"10", 0, 10, true},
		{"10px", 0, 10, true},
		{"1in", 0, 72, true},
		{"2.54cm", 0, 72, true},
		{"50%", 200, 100, true},
		{"1pc", 0, 12, true},
		{"5furlongs", 0, 0, false},
		{"abc", 0, 0, false},
		{"", 0, 0, false},
	}
	for _, tc := range tests {
		got, ok := parseLength(tc.in, tc.ref)
		if ok != tc.ok || (ok && !near(got, tc.want)) {
			t.Errorf("parseLength(%q) = %g, %v; want %g, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
	if fmtNum(math.Inf(1), 4) != "0" || fmtNum(math.NaN(), 4) != "0" {
		t.Error("non-finite numbers must format as 0")
	}
	if fmtNum(-0.00001, 4) != "0" || fmtNum(1.5, 4) != "1.5" || fmtNum(2, 4) != "2" || fmtNum(1e-7, 6) != "0" {
		t.Error("fmtNum formatting")
	}
}

func TestRenderBasicShapes(t *testing.T) {
	out := render(t, `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
		<rect x="10" y="10" width="30" height="20" fill="red"/>
		<rect x="50" y="10" width="30" height="20" rx="5" fill="blue"/>
		<circle cx="50" cy="60" r="10" fill="none" stroke="black" stroke-width="2"/>
		<ellipse cx="20" cy="80" rx="10" ry="5"/>
		<line x1="0" y1="0" x2="100" y2="100" stroke="green" stroke-linecap="round" stroke-dasharray="4 2 1"/>
		<polygon points="0,0 10,0 10,10" fill="yellow" fill-rule="evenodd"/>
		<polyline points="0,0 10 10 20" fill="none" stroke="#000" stroke-linejoin="bevel"/>
	</svg>`)
	for _, want := range []string{
		"0 0 1 1 re W n", "0.01 0 0 -0.01 0 1 cm",
		"1 0 0 rg", "0 0 1 rg", "2 w", "0 0.502 0 RG", "1 J", "[4 2 1 4 2 1] 0 d",
		"f*\n", "2 j\n", "10 10 m", " c\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "q\n") != strings.Count(out, "Q\n") {
		t.Errorf("unbalanced q/Q:\n%s", out)
	}
}

func TestRenderStylingCascade(t *testing.T) {
	out := render(t, `<svg viewBox="0 0 10 10">
		<style>/* c */ circle { fill: blue } .warn { fill: orange } #one { fill: lime } g > rect { fill: pink } @media print { rect { fill: red } }</style>
		<rect id="one" class="warn" width="1" height="1" fill="black"/>
		<rect class="warn" width="2" height="2"/>
		<rect width="3" height="3" style="fill: #800000 !important"/>
		<g fill="purple" color="#008080"><rect width="4" height="4"/><rect width="5" height="5" fill="currentColor"/></g>
		<circle r="1" fill="red"/>
	</svg>`)
	for _, want := range []string{"0 1 0 rg", "1 0.6471 0 rg", "0.502 0 0 rg", "0.502 0 0.502 rg", "0 0.502 0.502 rg", "0 0 1 rg"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "1 0 0 rg") || strings.Contains(out, "1 0.7529 0.7961 rg") {
		t.Errorf("unsupported selectors should be ignored:\n%s", out)
	}
}

func TestRenderOpacityAndVisibility(t *testing.T) {
	r := mustParse(t, `<svg viewBox="0 0 10 10">
		<g opacity="0.5"><rect width="5" height="5" fill="red" fill-opacity="50%"/></g>
		<rect width="5" height="5" fill="rgba(0,0,255,0.5)" stroke="black" stroke-opacity="0.25"/>
		<rect width="5" height="5" fill="red" opacity="0"/>
		<rect width="5" height="5" fill="red" display="none"/>
		<g style="display:none"><rect width="5" height="5" fill="red"/></g>
		<g visibility="hidden"><rect width="5" height="5" fill="green"/><rect width="5" height="5" fill="blue" visibility="visible"/></g>
		<rect width="0" height="5" fill="red"/>
		<rect width="5" height="5" fill="none" stroke="black" stroke-width="0"/>
	</svg>`).Render()
	out := string(r.Content)
	if len(r.ExtGStates) != 2 {
		t.Errorf("ExtGStates = %v, want 2 entries", r.ExtGStates)
	}
	var found bool
	for _, a := range r.ExtGStates {
		if near(a.Fill, 0.25) && near(a.Stroke, 1) {
			found = true
		}
	}
	if !found {
		t.Errorf("missing fill 0.25 state: %v", r.ExtGStates)
	}
	if strings.Count(out, " rg\n") != 3 {
		t.Errorf("expected 3 filled shapes (group-opacity rect, rgba rect, visible-in-hidden rect):\n%s", out)
	}
	if strings.Contains(out, "0 0.502 0 rg") {
		t.Error("hidden rect was painted")
	}
}

func TestRenderUseSymbolAndGradient(t *testing.T) {
	out := render(t, `<svg xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 10 10">
		<defs>
			<linearGradient id="base"><stop offset="0" stop-color="#ff0000"/><stop offset="1" stop-color="blue"/></linearGradient>
			<linearGradient id="ref" xlink:href="#base"/>
			<radialGradient id="empty"/>
			<symbol id="sym"><rect width="1" height="1" fill="url(#ref)"/></symbol>
			<rect id="r" width="2" height="2" fill="url('#empty') green"/>
		</defs>
		<use href="#sym" x="3" y="4"/>
		<use xlink:href="#r"/>
		<use href="#missing"/>
		<rect width="1" height="1" fill="url(#nope)"/>
		<svg x="5" y="5"><circle r="1" fill="navy"/></svg>
		<switch><rect width="1" height="1" fill="gold"/><rect width="1" height="1" fill="silver"/></switch>
		<text>ignored</text>
	</svg>`)
	for _, want := range []string{"1 0 0 1 3 4 cm", "1 0 0 rg", "0 0.502 0 rg", "1 0 0 1 5 5 cm", "0 0 0.502 rg", "1 0.8431 0 rg"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "1 0 0 1 0 0 cm") {
		t.Error("identity translation should not be emitted")
	}
	if strings.Contains(out, "0.7529 0.7529 0.7529 rg") {
		t.Error("switch rendered more than its first child")
	}
	if strings.Count(out, " rg\n") != 4 {
		t.Errorf("unexpected number of fills:\n%s", out)
	}
}

func TestRenderTransformAndStroke(t *testing.T) {
	out := render(t, `<svg viewBox="0 0 10 10">
		<g transform="translate(1 2)">
			<path d="M0 0L5 5" stroke="black" stroke-width="2" stroke-miterlimit="10" stroke-linecap="square" stroke-linejoin="round" stroke-dasharray="none" stroke-dashoffset="1"/>
			<path d="M0 0L5 5" stroke="black" stroke-dasharray="1,2" stroke-dashoffset="1" fill="red"/>
		</g>
	</svg>`)
	for _, want := range []string{"1 0 0 1 1 2 cm", "2 J", "1 j", "[1 2] 1 d", "B\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "10 M") {
		t.Error("miter limit equal to the PDF default should not be emitted")
	}
}

func FuzzParseRender(f *testing.F) {
	f.Add(`<svg viewBox="0 0 10 10"><path d="M0 0A5 5 0 1 1 10 10z"/></svg>`)
	f.Add(`<svg><use href="#a"/><g id="a"><use href="#a"/></g></svg>`)
	f.Add(`<svg width="1e309"><rect width="-1" height="NaN"/></svg>`)
	f.Fuzz(func(t *testing.T, src string) {
		img, err := Parse([]byte(src))
		if err != nil {
			return
		}
		out := img.Render()
		if strings.Contains(string(out.Content), "e+") || strings.Contains(string(out.Content), "e-") {
			t.Fatalf("exponent in content stream: %s", out.Content)
		}
	})
}
