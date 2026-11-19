package font

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFallbackKeys(t *testing.T) {
	registered := map[string]bool{
		"Sans": true, "Sans-Bold": true, "Naskh": true, "Naskh-Bold": true, "Serif-Bold": true,
	}
	has := func(k string) bool { return registered[k] }

	tests := []struct {
		name      string
		primary   string
		fallbacks []string
		want      []string
	}{
		{name: "regular", primary: "Sans", fallbacks: []string{"Naskh"}, want: []string{"Naskh"}},
		{name: "bold variant used", primary: "Sans-Bold", fallbacks: []string{"Naskh"}, want: []string{"Naskh-Bold"}},
		{name: "primary skipped", primary: "Sans", fallbacks: []string{"Sans", "Naskh"}, want: []string{"Naskh"}},
		{name: "unregistered skipped", primary: "Sans", fallbacks: []string{"Missing", "Naskh"}, want: []string{"Naskh"}},
		{name: "base family when variant missing", primary: "Serif-Bold", fallbacks: []string{"Mono"}, want: nil},
		{name: "italic falls back to base", primary: "Sans-Italic", fallbacks: []string{"Naskh"}, want: []string{"Naskh"}},
		{name: "duplicates removed", primary: "Serif-Bold", fallbacks: []string{"Sans", "Sans"}, want: []string{"Sans-Bold"}},
		{name: "none", primary: "Sans", fallbacks: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FallbackKeys(tt.primary, tt.fallbacks, has); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FallbackKeys(%q, %v) = %v, want %v", tt.primary, tt.fallbacks, got, tt.want)
			}
		})
	}
}

func TestShapedWidth(t *testing.T) {
	runs := []GlyphRun{
		{Face: 0, Glyphs: []Glyph{{XAdvance: 500}, {XAdvance: 500}}}, // 1000 units of 1000/em
		{Face: 1, Glyphs: []Glyph{{XAdvance: 1024}}},                 // 1024 units of 2048/em
		{Face: 2, Glyphs: []Glyph{{XAdvance: 999}}},                  // unknown face: ignored
	}
	upem := func(face int) int { return []int{1000, 2048, 0}[face] }
	if got, want := ShapedWidth(runs, upem, 10), 15.0; got != want {
		t.Errorf("ShapedWidth = %v, want %v", got, want)
	}
}

func TestWrapWords(t *testing.T) {
	// one unit per rune keeps expectations readable
	measure := func(s string) float64 { return float64(utf8.RuneCountInString(s)) }

	tests := []struct {
		name     string
		in       string
		maxWidth float64
		want     []string
	}{
		{name: "empty", in: "", maxWidth: 10, want: []string{""}},
		{name: "fits", in: "aa bb", maxWidth: 10, want: []string{"aa bb"}},
		{name: "breaks between words", in: "aa bb cc", maxWidth: 5, want: []string{"aa bb", "cc"}},
		{name: "explicit newline", in: "aa\nbb", maxWidth: 10, want: []string{"aa", "bb"}},
		{name: "blank line kept", in: "aa\n\nbb", maxWidth: 10, want: []string{"aa", "", "bb"}},
		{name: "extra spaces collapse", in: "  aa   bb  ", maxWidth: 10, want: []string{"aa bb"}},
		{name: "long word cut", in: "a bbbbbbbb c", maxWidth: 3, want: []string{"a", "bbb", "bbb", "bb", "c"}},
		// U+0301 combining acute stays with its base letter
		{name: "mark not split from base", in: "abe\u0301c", maxWidth: 3, want: []string{"ab", "e\u0301c"}},
		{name: "logical order kept", in: "رقم الطلب 1024", maxWidth: 9, want: []string{"رقم الطلب", "1024"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapWords(tt.in, tt.maxWidth, measure)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WrapWords(%q, %v) = %q, want %q", tt.in, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestGenerateToUnicodeCMapText(t *testing.T) {
	tests := []struct {
		name     string
		mapping  map[uint16]string
		contains []string
		absent   []string
	}{
		{name: "empty", mapping: map[uint16]string{}},
		{name: "single rune", mapping: map[uint16]string{1: "A"}, contains: []string{"<0001> <0041>"}},
		{name: "ligature maps to all its characters", mapping: map[uint16]string{7: "لا"}, contains: []string{"<0007> <06440627>"}},
		{name: "supplementary plane", mapping: map[uint16]string{2: "\U0001F600"}, contains: []string{"<0002> <D83DDE00>"}},
		{name: "glyph without text left out", mapping: map[uint16]string{3: "", 4: "B"}, contains: []string{"1 beginbfchar", "<0004> <0042>"}, absent: []string{"<0003>"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(GenerateToUnicodeCMapText(tt.mapping))
			if len(tt.contains) == 0 && got != "" {
				t.Fatalf("want no CMap, got %q", got)
			}
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("CMap missing %q:\n%s", s, got)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(got, s) {
					t.Errorf("CMap has %q:\n%s", s, got)
				}
			}
		})
	}
}

func TestEncodeGlyph(t *testing.T) {
	ttf, _ := buildTestFont(t) // 'A'=1 (600), 'B'=2 (700), ' '=3 (250)

	if got := ttf.EncodeGlyph(2, "x"); !reflect.DeepEqual(got, []byte{0, 2}) {
		t.Errorf("EncodeGlyph = %v, want [0 2]", got)
	}
	ttf.EncodeGlyph(2, "y")  // first text wins
	ttf.EncodeGlyph(1, "")   // glyph without text (a mark of a cluster)
	ttf.EncodeGlyph(1, "AB") // text recorded later fills the gap

	want := map[uint16]string{1: "AB", 2: "x"}
	if got := ttf.UsedGlyphs(); !reflect.DeepEqual(got, want) {
		t.Errorf("UsedGlyphs = %v, want %v", got, want)
	}
	if len(ttf.UsedRunes()) != 0 {
		t.Error("EncodeGlyph must not record runes")
	}

	if w, ok := ttf.GlyphAdvance(2); !ok || w != 700 {
		t.Errorf("GlyphAdvance(2) = %d, %v; want 700, true", w, ok)
	}
	if _, ok := ttf.GlyphAdvance(99); ok {
		t.Error("GlyphAdvance(99) should not exist")
	}

	if _, err := ttf.Subset(nil); err != nil {
		t.Errorf("Subset with glyphs only: %v", err)
	}
}
