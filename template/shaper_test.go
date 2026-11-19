package template

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/internal/testfont"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// monoShaper draws every rune as glyph 1 with a one-em advance, in face 0,
// and records the lines it shapes.
type monoShaper struct {
	mu    sync.Mutex
	lines []string
	dirs  []font.Direction
	faces [][]string
}

func (s *monoShaper) Shape(text string, faces []font.ShapeFace, dir font.Direction) []font.GlyphRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lines = append(s.lines, text)
	s.dirs = append(s.dirs, dir)
	var keys []string
	for _, f := range faces {
		keys = append(keys, f.Key)
	}
	s.faces = append(s.faces, keys)
	run := font.GlyphRun{}
	for _, r := range text {
		run.Glyphs = append(run.Glyphs, font.Glyph{ID: 1, XAdvance: 1000, Text: string(r)})
	}
	return []font.GlyphRun{run}
}

func TestWithShaper(t *testing.T) {
	data := testfont.Build(2, []uint16{0, 500}, map[rune]uint16{'a': 1})
	s := &monoShaper{}
	doc := New(
		WithFont("Main", data),
		WithFont("Alt", data),
		WithShaper(s),
		WithFallbackFonts("Alt"),
		WithDefaultFont("Main", 10),
		WithPageSize(document.Size{Width: 140, Height: 200}),
		WithMargins(document.UniformEdges(document.Pt(20))),
	)
	// 1 em per rune at 10pt: "aaaa bbbb cc" is 120pt, so a 100pt column
	// must wrap after "bbbb" (measured with the shaper, not with /W widths
	// of 5pt per rune, which would fit everything on one line).
	doc.AddPage().AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("aaaa bbbb cc", RTL())
		})
	})
	out, err := doc.Generate()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(out, []byte("/ActualText <FEFF006100610061006100200062006200620062>")) {
		t.Errorf("first line %q not drawn as a shaped span", "aaaa bbbb")
	}
	if !bytes.Contains(out, []byte("/ActualText <FEFF00630063>")) {
		t.Errorf("second line %q not drawn as its own span", "cc")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	var sawRTL bool
	for i, line := range s.lines {
		if line == "cc" && s.dirs[i] == font.DirectionRTL {
			sawRTL = true
		}
		if strings.Join(s.faces[i], ",") != "Main,Alt" {
			t.Errorf("faces for %q = %v, want [Main Alt]", line, s.faces[i])
		}
	}
	if !sawRTL {
		t.Errorf("RTL direction not passed to the shaper; lines shaped: %q", s.lines)
	}
}

func TestWithoutShaperNoActualText(t *testing.T) {
	data := testfont.Build(2, []uint16{0, 500}, map[rune]uint16{'a': 1})
	doc := New(WithFont("Main", data), WithDefaultFont("Main", 10))
	doc.AddPage().AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("aaaa", RTL()) })
	})
	out, err := doc.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("/ActualText")) {
		t.Error("text drawn through a shaper although none is configured")
	}
}
