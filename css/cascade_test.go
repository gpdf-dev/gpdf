package css

import (
	"math"
	"strings"
	"testing"
)

// ─── Mock node for testing ────────────────────────────────────────

type mockNode struct {
	tag     string
	id      string
	classes []string
	attrs   map[string]string
	parent  *mockNode
}

func (n *mockNode) NodeTag() string                     { return n.tag }
func (n *mockNode) NodeID() string                      { return n.id }
func (n *mockNode) NodeClasses() []string               { return n.classes }
func (n *mockNode) NodeAttr(name string) (string, bool) { v, ok := n.attrs[name]; return v, ok }
func (n *mockNode) NodeParent() Matchable {
	if n.parent == nil {
		return nil
	}
	return n.parent
}

// ─── Selector parsing tests ──────────────────────────────────────

func TestParseSelector_Tag(t *testing.T) {
	sel, ok := ParseSelector("div")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(sel.Parts))
	}
	if sel.Parts[0].Tag != "div" {
		t.Errorf("tag = %q", sel.Parts[0].Tag)
	}
}

func TestParseSelector_Class(t *testing.T) {
	sel, ok := ParseSelector(".main")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if len(sel.Parts[0].Classes) != 1 || sel.Parts[0].Classes[0] != "main" {
		t.Errorf("classes = %v", sel.Parts[0].Classes)
	}
}

func TestParseSelector_ID(t *testing.T) {
	sel, ok := ParseSelector("#title")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if sel.Parts[0].ID != "title" {
		t.Errorf("id = %q", sel.Parts[0].ID)
	}
}

func TestParseSelector_Universal(t *testing.T) {
	sel, ok := ParseSelector("*")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if !sel.Parts[0].Universal {
		t.Error("expected universal selector")
	}
}

func TestParseSelector_Compound(t *testing.T) {
	sel, ok := ParseSelector("div.main#app")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if len(sel.Parts) != 1 {
		t.Fatalf("expected 1 compound, got %d", len(sel.Parts))
	}
	p := sel.Parts[0]
	if p.Tag != "div" || p.ID != "app" || len(p.Classes) != 1 || p.Classes[0] != "main" {
		t.Errorf("compound = tag:%q id:%q classes:%v", p.Tag, p.ID, p.Classes)
	}
}

func TestParseSelector_Descendant(t *testing.T) {
	sel, ok := ParseSelector("div p")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if len(sel.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[0].Tag != "div" {
		t.Errorf("parts[0].tag = %q", sel.Parts[0].Tag)
	}
	if sel.Parts[1].Tag != "p" {
		t.Errorf("parts[1].tag = %q", sel.Parts[1].Tag)
	}
	if sel.Parts[1].Combinator != CombinatorDescendant {
		t.Errorf("expected descendant combinator, got %d", sel.Parts[1].Combinator)
	}
}

func TestParseSelector_Child(t *testing.T) {
	sel, ok := ParseSelector("div > p")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if len(sel.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != CombinatorChild {
		t.Errorf("expected child combinator, got %d", sel.Parts[1].Combinator)
	}
}

func TestParseSelectorGroup(t *testing.T) {
	sels := ParseSelectorGroup("h1, h2, h3")
	if len(sels) != 3 {
		t.Fatalf("expected 3 selectors, got %d", len(sels))
	}
	tags := make([]string, len(sels))
	for i, s := range sels {
		tags[i] = s.Parts[0].Tag
	}
	if tags[0] != "h1" || tags[1] != "h2" || tags[2] != "h3" {
		t.Errorf("tags = %v", tags)
	}
}

func TestParseSelector_MultiClass(t *testing.T) {
	sel, ok := ParseSelector(".foo.bar.baz")
	if !ok {
		t.Fatal("expected valid selector")
	}
	if len(sel.Parts[0].Classes) != 3 {
		t.Fatalf("expected 3 classes, got %d", len(sel.Parts[0].Classes))
	}
}

// ─── Selector matching tests ─────────────────────────────────────

func TestMatch_Tag(t *testing.T) {
	sel, _ := ParseSelector("div")
	node := &mockNode{tag: "div"}
	if !Match(sel, node) {
		t.Error("div should match div")
	}
	node2 := &mockNode{tag: "p"}
	if Match(sel, node2) {
		t.Error("div should not match p")
	}
}

func TestMatch_Class(t *testing.T) {
	sel, _ := ParseSelector(".main")
	node := &mockNode{tag: "div", classes: []string{"main", "container"}}
	if !Match(sel, node) {
		t.Error(".main should match div.main.container")
	}
	node2 := &mockNode{tag: "div", classes: []string{"other"}}
	if Match(sel, node2) {
		t.Error(".main should not match div.other")
	}
}

func TestMatch_ID(t *testing.T) {
	sel, _ := ParseSelector("#title")
	node := &mockNode{tag: "h1", id: "title"}
	if !Match(sel, node) {
		t.Error("#title should match h1#title")
	}
}

func TestMatch_Compound(t *testing.T) {
	sel, _ := ParseSelector("div.main")
	node := &mockNode{tag: "div", classes: []string{"main"}}
	if !Match(sel, node) {
		t.Error("div.main should match")
	}
	node2 := &mockNode{tag: "span", classes: []string{"main"}}
	if Match(sel, node2) {
		t.Error("div.main should not match span.main")
	}
}

func TestMatch_Universal(t *testing.T) {
	sel, _ := ParseSelector("*")
	node := &mockNode{tag: "anything"}
	if !Match(sel, node) {
		t.Error("* should match any element")
	}
}

func TestMatch_Descendant(t *testing.T) {
	sel, _ := ParseSelector("div p")
	grandparent := &mockNode{tag: "div"}
	parent := &mockNode{tag: "section", parent: grandparent}
	node := &mockNode{tag: "p", parent: parent}
	if !Match(sel, node) {
		t.Error("div p should match p inside div (with section in between)")
	}
}

func TestMatch_Child(t *testing.T) {
	sel, _ := ParseSelector("div > p")
	parent := &mockNode{tag: "div"}
	node := &mockNode{tag: "p", parent: parent}
	if !Match(sel, node) {
		t.Error("div > p should match direct child")
	}

	// Should NOT match if there's an intermediate element
	middle := &mockNode{tag: "section", parent: parent}
	node2 := &mockNode{tag: "p", parent: middle}
	if Match(sel, node2) {
		t.Error("div > p should not match p inside section inside div")
	}
}

func TestMatch_Complex(t *testing.T) {
	sel, _ := ParseSelector("div.main > p")
	parent := &mockNode{tag: "div", classes: []string{"main"}}
	node := &mockNode{tag: "p", parent: parent}
	if !Match(sel, node) {
		t.Error("div.main > p should match")
	}
}

func TestMatch_DeepNesting(t *testing.T) {
	sel, _ := ParseSelector("body div p")
	body := &mockNode{tag: "body"}
	div := &mockNode{tag: "div", parent: body}
	section := &mockNode{tag: "section", parent: div}
	p := &mockNode{tag: "p", parent: section}
	if !Match(sel, p) {
		t.Error("body div p should match p nested deep")
	}
}

// ─── Specificity tests ───────────────────────────────────────────

func TestSpecificity_Simple(t *testing.T) {
	tests := []struct {
		selector string
		spec     Specificity
	}{
		{"*", Specificity{0, 0, 0}},
		{"div", Specificity{0, 0, 1}},
		{".main", Specificity{0, 1, 0}},
		{"#title", Specificity{1, 0, 0}},
		{"div.main", Specificity{0, 1, 1}},
		{"div.main#app", Specificity{1, 1, 1}},
		{"div p", Specificity{0, 0, 2}},
		{".a.b.c", Specificity{0, 3, 0}},
	}
	for _, tt := range tests {
		sel, ok := ParseSelector(tt.selector)
		if !ok {
			t.Errorf("ParseSelector(%q) failed", tt.selector)
			continue
		}
		if sel.Spec != tt.spec {
			t.Errorf("selector %q: specificity = %v, want %v", tt.selector, sel.Spec, tt.spec)
		}
	}
}

func TestSpecificity_Less(t *testing.T) {
	a := Specificity{0, 0, 1} // type
	b := Specificity{0, 1, 0} // class
	c := Specificity{1, 0, 0} // id
	if !a.Less(b) {
		t.Error("type < class")
	}
	if !b.Less(c) {
		t.Error("class < id")
	}
	if c.Less(a) {
		t.Error("id should not be < type")
	}
}

// ─── Shorthand expansion tests ───────────────────────────────────

func TestExpandShorthands_Margin(t *testing.T) {
	tests := []struct {
		value string
		want  [4]string // top, right, bottom, left
	}{
		{"10px", [4]string{"10px", "10px", "10px", "10px"}},
		{"10px 20px", [4]string{"10px", "20px", "10px", "20px"}},
		{"10px 20px 30px", [4]string{"10px", "20px", "30px", "20px"}},
		{"10px 20px 30px 40px", [4]string{"10px", "20px", "30px", "40px"}},
	}
	for _, tt := range tests {
		decls := ExpandShorthands([]Declaration{{Property: "margin", Value: tt.value}})
		if len(decls) != 4 {
			t.Errorf("margin: %s → %d declarations", tt.value, len(decls))
			continue
		}
		props := [4]string{"margin-top", "margin-right", "margin-bottom", "margin-left"}
		for i, prop := range props {
			if decls[i].Property != prop || decls[i].Value != tt.want[i] {
				t.Errorf("margin: %s → %s: %q, want %q", tt.value, prop, decls[i].Value, tt.want[i])
			}
		}
	}
}

func TestExpandShorthands_Border(t *testing.T) {
	decls := ExpandShorthands([]Declaration{{Property: "border", Value: "2px solid red"}})
	// Should produce width/style/color for all 4 sides = 12 declarations
	if len(decls) != 12 {
		t.Fatalf("expected 12 declarations, got %d", len(decls))
	}
	// Check first side (top)
	found := make(map[string]string)
	for _, d := range decls {
		found[d.Property] = d.Value
	}
	if found["border-top-width"] != "2px" {
		t.Errorf("border-top-width = %q", found["border-top-width"])
	}
	if found["border-top-style"] != "solid" {
		t.Errorf("border-top-style = %q", found["border-top-style"])
	}
	if found["border-top-color"] != "red" {
		t.Errorf("border-top-color = %q", found["border-top-color"])
	}
}

func TestExpandShorthands_BorderSide(t *testing.T) {
	decls := ExpandShorthands([]Declaration{{Property: "border-bottom", Value: "1px dashed #ccc"}})
	found := make(map[string]string)
	for _, d := range decls {
		found[d.Property] = d.Value
	}
	if found["border-bottom-width"] != "1px" {
		t.Errorf("border-bottom-width = %q", found["border-bottom-width"])
	}
	if found["border-bottom-style"] != "dashed" {
		t.Errorf("border-bottom-style = %q", found["border-bottom-style"])
	}
}

func TestExpandShorthands_Background(t *testing.T) {
	decls := ExpandShorthands([]Declaration{{Property: "background", Value: "#f5f5f5"}})
	if len(decls) != 1 || decls[0].Property != "background-color" {
		t.Errorf("background expansion: %v", decls)
	}
}

func TestExpandShorthands_Important(t *testing.T) {
	decls := ExpandShorthands([]Declaration{{Property: "margin", Value: "10px", Important: true}})
	for _, d := range decls {
		if !d.Important {
			t.Errorf("%s should be !important", d.Property)
		}
	}
}

// ─── Cascade tests ───────────────────────────────────────────────

func TestCascade_BasicRule(t *testing.T) {
	ss, _ := ParseStylesheet(`p { color: red; font-size: 14pt; }`)
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, nil, nil)

	if styles["color"] != "red" {
		t.Errorf("color = %q, want 'red'", styles["color"])
	}
	if styles["font-size"] != "14pt" {
		t.Errorf("font-size = %q, want '14pt'", styles["font-size"])
	}
}

func TestCascade_SpecificityWins(t *testing.T) {
	ss, _ := ParseStylesheet(`
		p { color: red; }
		p.special { color: blue; }
	`)
	node := &mockNode{tag: "p", classes: []string{"special"}}
	styles := Cascade(node, []*Stylesheet{ss}, nil, nil, nil)
	if styles["color"] != "blue" {
		t.Errorf("color = %q, want 'blue' (higher specificity)", styles["color"])
	}
}

func TestCascade_SourceOrderWins(t *testing.T) {
	ss, _ := ParseStylesheet(`
		p { color: red; }
		p { color: green; }
	`)
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, nil, nil)
	if styles["color"] != "green" {
		t.Errorf("color = %q, want 'green' (later source order)", styles["color"])
	}
}

func TestCascade_ImportantWins(t *testing.T) {
	ss, _ := ParseStylesheet(`
		p.special { color: blue; }
		p { color: red !important; }
	`)
	node := &mockNode{tag: "p", classes: []string{"special"}}
	styles := Cascade(node, []*Stylesheet{ss}, nil, nil, nil)
	if styles["color"] != "red" {
		t.Errorf("color = %q, want 'red' (!important)", styles["color"])
	}
}

func TestCascade_InlineOverrides(t *testing.T) {
	ss, _ := ParseStylesheet(`p { color: red; }`)
	inline := ParseDeclarations("color: blue")
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, inline, nil, nil)
	if styles["color"] != "blue" {
		t.Errorf("color = %q, want 'blue' (inline)", styles["color"])
	}
}

func TestCascade_UAStylesheet(t *testing.T) {
	ua, _ := ParseStylesheet(`h1 { font-size: 2em; font-weight: bold; }`)
	node := &mockNode{tag: "h1"}
	styles := Cascade(node, nil, nil, nil, ua)
	if styles["font-weight"] != "bold" {
		t.Errorf("font-weight = %q, want 'bold'", styles["font-weight"])
	}
}

func TestCascade_AuthorOverridesUA(t *testing.T) {
	ua, _ := ParseStylesheet(`h1 { color: black; }`)
	author, _ := ParseStylesheet(`h1 { color: red; }`)
	node := &mockNode{tag: "h1"}
	styles := Cascade(node, []*Stylesheet{author}, nil, nil, ua)
	if styles["color"] != "red" {
		t.Errorf("color = %q, want 'red' (author > UA)", styles["color"])
	}
}

func TestCascade_Inheritance(t *testing.T) {
	parentStyles := ComputedStyles{
		"color":       "blue",
		"font-size":   "16pt",
		"font-family": "Arial",
	}
	node := &mockNode{tag: "span"}
	styles := Cascade(node, nil, nil, parentStyles, nil)
	// color is inherited
	if styles["color"] != "blue" {
		t.Errorf("color = %q, want 'blue' (inherited)", styles["color"])
	}
	// font-family is inherited
	if styles["font-family"] != "Arial" {
		t.Errorf("font-family = %q, want 'Arial' (inherited)", styles["font-family"])
	}
}

func TestCascade_InheritKeyword(t *testing.T) {
	ss, _ := ParseStylesheet(`p { margin-top: inherit; }`)
	parentStyles := ComputedStyles{"margin-top": "20pt"}
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, parentStyles, nil)
	if styles["margin-top"] != "20pt" {
		t.Errorf("margin-top = %q, want '20pt' (inherit)", styles["margin-top"])
	}
}

func TestCascade_InitialKeyword(t *testing.T) {
	ss, _ := ParseStylesheet(`p { color: initial; }`)
	parentStyles := ComputedStyles{"color": "blue"}
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, parentStyles, nil)
	if styles["color"] != "#000000" {
		t.Errorf("color = %q, want '#000000' (initial)", styles["color"])
	}
}

func TestCascade_UnsetInherited(t *testing.T) {
	ss, _ := ParseStylesheet(`p { color: unset; }`)
	parentStyles := ComputedStyles{"color": "green"}
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, parentStyles, nil)
	// color is inherited, so unset = inherit
	if styles["color"] != "green" {
		t.Errorf("color = %q, want 'green' (unset on inherited)", styles["color"])
	}
}

func TestCascade_UnsetNonInherited(t *testing.T) {
	ss, _ := ParseStylesheet(`p { margin-top: unset; }`)
	parentStyles := ComputedStyles{"margin-top": "50pt"}
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, parentStyles, nil)
	// margin-top is NOT inherited, so unset = initial
	if styles["margin-top"] != "0" {
		t.Errorf("margin-top = %q, want '0' (unset on non-inherited)", styles["margin-top"])
	}
}

func TestCascade_ShorthandInCascade(t *testing.T) {
	ss, _ := ParseStylesheet(`p { margin: 10px 20px; }`)
	node := &mockNode{tag: "p"}
	styles := Cascade(node, []*Stylesheet{ss}, nil, nil, nil)
	if styles["margin-top"] != "10px" {
		t.Errorf("margin-top = %q", styles["margin-top"])
	}
	if styles["margin-right"] != "20px" {
		t.Errorf("margin-right = %q", styles["margin-right"])
	}
	if styles["margin-bottom"] != "10px" {
		t.Errorf("margin-bottom = %q", styles["margin-bottom"])
	}
	if styles["margin-left"] != "20px" {
		t.Errorf("margin-left = %q", styles["margin-left"])
	}
}

// ─── Computed value tests ─────────────────────────────────────────

func TestResolveComputed_EmToPoint(t *testing.T) {
	styles := ComputedStyles{
		"font-size":  "12pt",
		"margin-top": "2em",
	}
	resolved := ResolveComputed(styles, 12, 595)
	mt := ParsePt(resolved["margin-top"], 0)
	if math.Abs(mt-24) > 0.01 {
		t.Errorf("margin-top = %gpt, want 24pt (2em * 12pt)", mt)
	}
}

func TestResolveComputed_PercentToPoint(t *testing.T) {
	styles := ComputedStyles{
		"width": "50%",
	}
	resolved := ResolveComputed(styles, 12, 500)
	w := ParsePt(resolved["width"], 0)
	if math.Abs(w-250) > 0.01 {
		t.Errorf("width = %gpt, want 250pt", w)
	}
}

func TestResolveComputed_PxToPoint(t *testing.T) {
	styles := ComputedStyles{
		"padding-top": "16px",
	}
	resolved := ResolveComputed(styles, 12, 500)
	pt := ParsePt(resolved["padding-top"], 0)
	if math.Abs(pt-12) > 0.01 {
		t.Errorf("padding-top = %gpt, want 12pt (16px * 0.75)", pt)
	}
}

func TestResolveComputed_MmToPoint(t *testing.T) {
	styles := ComputedStyles{
		"margin-top": "25.4mm",
	}
	resolved := ResolveComputed(styles, 12, 500)
	pt := ParsePt(resolved["margin-top"], 0)
	if math.Abs(pt-72) > 0.1 {
		t.Errorf("margin-top = %gpt, want 72pt (25.4mm = 1in = 72pt)", pt)
	}
}

func TestResolveComputed_LineHeight(t *testing.T) {
	styles := ComputedStyles{
		"font-size":   "12pt",
		"line-height": "1.5",
	}
	resolved := ResolveComputed(styles, 12, 500)
	lh := ParsePt(resolved["line-height"], 0)
	if math.Abs(lh-18) > 0.01 {
		t.Errorf("line-height = %gpt, want 18pt (1.5 * 12pt)", lh)
	}
}

func TestResolveComputed_FontWeight(t *testing.T) {
	styles := ComputedStyles{"font-weight": "bold"}
	resolved := ResolveComputed(styles, 12, 500)
	if resolved["font-weight"] != "700" {
		t.Errorf("font-weight = %q, want '700'", resolved["font-weight"])
	}
}

func TestResolveComputed_BorderWidth(t *testing.T) {
	styles := ComputedStyles{
		"border-top-width": "thin",
		"border-top-style": "solid",
	}
	resolved := ResolveComputed(styles, 12, 500)
	if !strings.HasPrefix(resolved["border-top-width"], "0.75") {
		t.Errorf("border-top-width = %q, want '0.75pt'", resolved["border-top-width"])
	}
}

func TestResolveComputed_Auto(t *testing.T) {
	styles := ComputedStyles{"width": "auto"}
	resolved := ResolveComputed(styles, 12, 500)
	if resolved["width"] != "auto" {
		t.Errorf("width = %q, want 'auto'", resolved["width"])
	}
}

func TestDimensionToPt(t *testing.T) {
	tests := []struct {
		num  float64
		unit string
		want float64
	}{
		{12, "pt", 12},
		{16, "px", 12},
		{2, "em", 24},
		{25.4, "mm", 72},
		{1, "in", 72},
		{2.54, "cm", 72},
		{1, "pc", 12},
	}
	for _, tt := range tests {
		got := DimensionToPt(tt.num, tt.unit, 12)
		if math.Abs(got-tt.want) > 0.1 {
			t.Errorf("DimensionToPt(%g, %q, 12) = %g, want %g", tt.num, tt.unit, got, tt.want)
		}
	}
}

// ─── Properties tests ────────────────────────────────────────────

func TestIsInherited(t *testing.T) {
	inherited := []string{"color", "font-family", "font-size", "line-height", "text-align", "list-style-type"}
	for _, p := range inherited {
		if !IsInherited(p) {
			t.Errorf("%q should be inherited", p)
		}
	}
	notInherited := []string{"margin-top", "padding-top", "width", "display", "background-color"}
	for _, p := range notInherited {
		if IsInherited(p) {
			t.Errorf("%q should not be inherited", p)
		}
	}
}

func TestInitialValue(t *testing.T) {
	if InitialValue("display") != "inline" {
		t.Errorf("display initial = %q", InitialValue("display"))
	}
	if InitialValue("color") != "#000000" {
		t.Errorf("color initial = %q", InitialValue("color"))
	}
	if InitialValue("margin-top") != "0" {
		t.Errorf("margin-top initial = %q", InitialValue("margin-top"))
	}
}
