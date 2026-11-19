package render

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/internal/testfont"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// fakeShaper returns fixed runs and records what it was asked.
type fakeShaper struct {
	mu    sync.Mutex
	runs  []font.GlyphRun
	text  string
	faces []string
	dir   font.Direction
}

func (s *fakeShaper) Shape(text string, faces []font.ShapeFace, dir font.Direction) []font.GlyphRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.text, s.dir, s.faces = text, dir, nil
	for _, f := range faces {
		s.faces = append(s.faces, f.Key)
	}
	return s.runs
}

// shapedTestRenderer registers two in-memory fonts (1000 units/em):
// "Main" with A=1 (600), B=2 (700), mark=3 (0); "Alt" with C=1 (500).
func shapedTestRenderer(t *testing.T, s font.Shaper) (*PDFRenderer, *bytes.Buffer) {
	t.Helper()
	r, buf := newTestRenderer(t)
	for _, f := range []struct {
		key    string
		widths []uint16
		runes  map[rune]uint16
	}{
		{key: "Main", widths: []uint16{0, 600, 700, 0}, runes: map[rune]uint16{'A': 1, 'B': 2, '\u0301': 3}},
		{key: "Alt", widths: []uint16{0, 500}, runes: map[rune]uint16{'C': 1}},
	} {
		data := testfont.Build(len(f.widths), f.widths, f.runes)
		ttf, err := font.ParseTrueType(data)
		if err != nil {
			t.Fatalf("parse %s: %v", f.key, err)
		}
		r.RegisterTTFont(f.key, ttf, data)
	}
	r.SetShaper(s, []string{"Alt", "Missing"})
	if err := r.BeginPage(document.Size{Width: 200, Height: 200}); err != nil {
		t.Fatal(err)
	}
	return r, buf
}

func TestRenderShapedText(t *testing.T) {
	s := &fakeShaper{runs: []font.GlyphRun{
		{Face: 1, Glyphs: []font.Glyph{{ID: 1, XAdvance: 500, Text: "C"}}},
		{Face: 0, Glyphs: []font.Glyph{
			{ID: 1, XAdvance: 650, Text: "A"},                    // advance 50 wider than /W
			{ID: 3, XOffset: -300, YOffset: 200, Text: "\u0301"}, // positioned mark
			{ID: 2, XAdvance: 700, Text: "B"},                    // advance equals /W
		}},
	}}
	r, _ := shapedTestRenderer(t, s)

	style := document.Style{FontFamily: "Main", FontSize: 10, Color: pdf.Black, Direction: document.DirectionRTL}
	if err := r.RenderText("A\u0301BC", document.Point{X: 20, Y: 30}, style); err != nil {
		t.Fatal(err)
	}
	got := string(r.pageContent)

	if s.text != "A\u0301BC" || s.dir != font.DirectionRTL {
		t.Errorf("shaper got %q dir %v, want logical text and RTL", s.text, s.dir)
	}
	if strings.Join(s.faces, ",") != "Main,Alt" {
		t.Errorf("faces = %v, want [Main Alt] (unregistered fallback skipped)", s.faces)
	}

	for _, want := range []string{
		"/Span <</ActualText <FEFF0041030100420043>>> BDC", // logical text, UTF-16BE
		"20 160 Td",            // 200 - 30 - fontSize
		"[<0001>] TJ",          // run in Alt: advance matches /W
		"[<0001> -50 ",         // pen moves 650 instead of 600
		"2 Ts\n",               // mark raised by 200/1000 em at 10pt
		"[300 <0003> -300] TJ", // mark moved left 300, pen restored
		"0 Ts\n",               // rise reset after the mark
		"<0002>] TJ",           // no adjustment when advance equals /W
		"ET\nEMC\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("content stream missing %q:\n%s", want, got)
		}
	}
	if n := strings.Count(got, " Tf\n"); n != 2 {
		t.Errorf("font selected %d times, want 2 (one per face change):\n%s", n, got)
	}
}

func TestRenderShapedTextEmbedsShapedGlyphs(t *testing.T) {
	s := &fakeShaper{runs: []font.GlyphRun{{Face: 0, Glyphs: []font.Glyph{
		{ID: 2, XAdvance: 700, Text: "AB"}, // a "ligature" glyph for two characters
		{ID: 3, XAdvance: 0, Text: ""},     // a glyph with no rune of its own
	}}}}
	r, buf := shapedTestRenderer(t, s)
	style := document.Style{FontFamily: "Main", FontSize: 12, Color: pdf.Black}
	if err := r.RenderText("AB", document.Point{}, style); err != nil {
		t.Fatal(err)
	}
	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}

	used := r.ttFonts["Main"].UsedGlyphs()
	if used[2] != "AB" {
		t.Errorf("glyph 2 text = %q, want %q", used[2], "AB")
	}
	if _, ok := used[3]; !ok {
		t.Error("glyph 3 not recorded for subsetting")
	}
	if !bytes.Contains(buf.Bytes(), []byte("/W [2 [700 0]]")) {
		t.Errorf("widths of shaped glyphs missing from /W")
	}
}

func TestRenderTextWithoutShaperUnchanged(t *testing.T) {
	r, _ := shapedTestRenderer(t, nil)
	r.shaper = nil
	style := document.Style{FontFamily: "Main", FontSize: 10, Color: pdf.Black}
	if err := r.RenderText("AB", document.Point{}, style); err != nil {
		t.Fatal(err)
	}
	got := string(r.pageContent)
	if strings.Contains(got, "ActualText") || strings.Contains(got, "TJ") {
		t.Errorf("rune path should not use shaping operators:\n%s", got)
	}
	if !strings.Contains(got, "<00010002> Tj") {
		t.Errorf("rune path output changed:\n%s", got)
	}
}

func TestUTF16BEHex(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{in: "", want: "FEFF"},
		{in: "A", want: "FEFF0041"},
		{in: "لا", want: "FEFF06440627"},
		{in: "\U0001F600", want: "FEFFD83DDE00"},
	}
	for _, tt := range tests {
		if got := utf16BEHex(tt.in); got != tt.want {
			t.Errorf("utf16BEHex(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestShapeDirection(t *testing.T) {
	tests := []struct {
		in   document.TextDirection
		want font.Direction
	}{
		{in: document.DirectionAuto, want: font.DirectionAuto},
		{in: document.DirectionLTR, want: font.DirectionLTR},
		{in: document.DirectionRTL, want: font.DirectionRTL},
	}
	for _, tt := range tests {
		if got := shapeDirection(tt.in); got != tt.want {
			t.Errorf("shapeDirection(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestToUnicodeText(t *testing.T) {
	tests := []struct {
		name   string
		runes  map[rune]uint16
		glyphs map[uint16]string
		want   map[uint16]string
	}{
		{name: "runes only", runes: map[rune]uint16{'A': 1}, want: map[uint16]string{1: "A"}},
		{name: "shaped only", glyphs: map[uint16]string{5: "لا", 6: ""}, want: map[uint16]string{5: "لا", 6: ""}},
		{name: "rune mapping wins", runes: map[rune]uint16{'A': 1}, glyphs: map[uint16]string{1: "X"}, want: map[uint16]string{1: "A"}},
		{name: "lowest rune for a shared glyph", runes: map[rune]uint16{'b': 2, 'a': 2}, want: map[uint16]string{2: "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toUnicodeText(tt.runes, tt.glyphs)
			if len(got) != len(tt.want) {
				t.Fatalf("toUnicodeText = %v, want %v", got, tt.want)
			}
			for gid, s := range tt.want {
				if got[gid] != s {
					t.Errorf("glyph %d = %q, want %q", gid, got[gid], s)
				}
			}
		})
	}
}
