package css

import (
	"math"
	"strings"
	"testing"
)

// ─── Tokenizer tests ──────────────────────────────────────────────

func TestTokenizer_Ident(t *testing.T) {
	tokens := tokenize("color font-size -webkit-transform _custom")
	idents := filterType(tokens, TokenIdent)
	want := []string{"color", "font-size", "-webkit-transform", "_custom"}
	if len(idents) != len(want) {
		t.Fatalf("got %d idents, want %d", len(idents), len(want))
	}
	for i, tok := range idents {
		if tok.Value != want[i] {
			t.Errorf("ident[%d] = %q, want %q", i, tok.Value, want[i])
		}
	}
}

func TestTokenizer_Hash(t *testing.T) {
	tokens := tokenize("#fff #main #2980b9")
	hashes := filterType(tokens, TokenHash)
	want := []string{"fff", "main", "2980b9"}
	if len(hashes) != len(want) {
		t.Fatalf("got %d hashes, want %d", len(hashes), len(want))
	}
	for i, tok := range hashes {
		if tok.Value != want[i] {
			t.Errorf("hash[%d] = %q, want %q", i, tok.Value, want[i])
		}
	}
}

func TestTokenizer_String(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello world"`, "hello world"},
		{`'single quotes'`, "single quotes"},
		{`"escaped \"quote\""`, `escaped "quote"`},
		{`"line\ncontinuation"`, "linencontinuation"}, // CSS: \n is escaped literal 'n'
		{`""`, ""},
	}
	for _, tt := range tests {
		tokens := tokenize(tt.input)
		strs := filterType(tokens, TokenString)
		if len(strs) != 1 {
			t.Errorf("input %s: got %d strings, want 1", tt.input, len(strs))
			continue
		}
		if strs[0].Value != tt.want {
			t.Errorf("input %s: got %q, want %q", tt.input, strs[0].Value, tt.want)
		}
	}
}

func TestTokenizer_Number(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"-1", -1},
		{"+0.5", 0.5},
		{"0", 0},
		{".5", 0.5},
	}
	for _, tt := range tests {
		tokens := tokenize(tt.input)
		nums := filterType(tokens, TokenNumber)
		if len(nums) != 1 {
			t.Errorf("input %q: got %d numbers, want 1", tt.input, len(nums))
			continue
		}
		if math.Abs(nums[0].Num-tt.want) > 1e-10 {
			t.Errorf("input %q: got %g, want %g", tt.input, nums[0].Num, tt.want)
		}
	}
}

func TestTokenizer_Dimension(t *testing.T) {
	tests := []struct {
		input string
		num   float64
		unit  string
	}{
		{"12px", 12, "px"},
		{"1.5em", 1.5, "em"},
		{"20mm", 20, "mm"},
		{"100vw", 100, "vw"},
		{"-2pt", -2, "pt"},
		{"0.67em", 0.67, "em"},
	}
	for _, tt := range tests {
		tokens := tokenize(tt.input)
		dims := filterType(tokens, TokenDimension)
		if len(dims) != 1 {
			t.Errorf("input %q: got %d dimensions, want 1 (tokens: %v)", tt.input, len(dims), tokens)
			continue
		}
		if math.Abs(dims[0].Num-tt.num) > 1e-10 {
			t.Errorf("input %q: num = %g, want %g", tt.input, dims[0].Num, tt.num)
		}
		if dims[0].Unit != tt.unit {
			t.Errorf("input %q: unit = %q, want %q", tt.input, dims[0].Unit, tt.unit)
		}
	}
}

func TestTokenizer_Percentage(t *testing.T) {
	tests := []struct {
		input string
		num   float64
	}{
		{"50%", 50},
		{"100%", 100},
		{"0%", 0},
		{"33.3%", 33.3},
	}
	for _, tt := range tests {
		tokens := tokenize(tt.input)
		pcts := filterType(tokens, TokenPercentage)
		if len(pcts) != 1 {
			t.Errorf("input %q: got %d percentages, want 1", tt.input, len(pcts))
			continue
		}
		if math.Abs(pcts[0].Num-tt.num) > 1e-10 {
			t.Errorf("input %q: got %g, want %g", tt.input, pcts[0].Num, tt.num)
		}
	}
}

func TestTokenizer_Function(t *testing.T) {
	tokens := tokenize("rgb(255, 0, 0)")
	fns := filterType(tokens, TokenFunction)
	if len(fns) != 1 {
		t.Fatalf("got %d functions, want 1", len(fns))
	}
	if fns[0].Value != "rgb" {
		t.Errorf("function = %q, want 'rgb'", fns[0].Value)
	}
}

func TestTokenizer_AtKeyword(t *testing.T) {
	tokens := tokenize("@media @page @import")
	ats := filterType(tokens, TokenAtKeyword)
	want := []string{"media", "page", "import"}
	if len(ats) != len(want) {
		t.Fatalf("got %d at-keywords, want %d", len(ats), len(want))
	}
	for i, tok := range ats {
		if tok.Value != want[i] {
			t.Errorf("at[%d] = %q, want %q", i, tok.Value, want[i])
		}
	}
}

func TestTokenizer_Punctuation(t *testing.T) {
	tokens := tokenize(": ; , { } ( ) [ ]")
	types := make([]TokenType, 0)
	for _, tok := range tokens {
		if tok.Type != TokenWhitespace {
			types = append(types, tok.Type)
		}
	}
	want := []TokenType{
		TokenColon, TokenSemicolon, TokenComma,
		TokenOpenBrace, TokenCloseBrace,
		TokenOpenParen, TokenCloseParen,
		TokenOpenBracket, TokenCloseBracket,
	}
	if len(types) != len(want) {
		t.Fatalf("got %d punctuation tokens, want %d: %v", len(types), len(want), types)
	}
	for i, typ := range types {
		if typ != want[i] {
			t.Errorf("token[%d] type = %d, want %d", i, typ, want[i])
		}
	}
}

func TestTokenizer_Comment(t *testing.T) {
	tokens := tokenize("color /* comment */ : red")
	// Comments should be skipped
	for _, tok := range tokens {
		if tok.Type == TokenEOF {
			break
		}
		if strings.Contains(tok.Value, "comment") {
			t.Fatal("comment should be skipped")
		}
	}
	idents := filterType(tokens, TokenIdent)
	if len(idents) != 2 {
		t.Fatalf("expected 2 idents (color, red), got %d", len(idents))
	}
}

func TestTokenizer_Delimiters(t *testing.T) {
	tokens := tokenize("> + ~ * . !")
	delims := filterType(tokens, TokenDelim)
	want := []string{">", "+", "~", "*", ".", "!"}
	if len(delims) != len(want) {
		t.Fatalf("got %d delims, want %d", len(delims), len(want))
	}
	for i, tok := range delims {
		if tok.Value != want[i] {
			t.Errorf("delim[%d] = %q, want %q", i, tok.Value, want[i])
		}
	}
}

func TestTokenizer_ComplexSelector(t *testing.T) {
	tokens := tokenize("div.main > p:first-child")
	// Should produce: ident(div) delim(.) ident(main) ws delim(>) ws ident(p) colon ident(first-child)
	nonWS := filterNonWhitespace(tokens)
	if len(nonWS) < 7 {
		t.Fatalf("expected at least 7 non-ws tokens, got %d: %v", len(nonWS), nonWS)
	}
}

func TestTokenizer_FullRule(t *testing.T) {
	input := `h1 { font-size: 2em; color: #2980b9; }`
	tokens := tokenize(input)
	// Should not be empty
	if len(tokens) == 0 {
		t.Fatal("expected tokens")
	}
	// First non-ws token should be ident "h1"
	nonWS := filterNonWhitespace(tokens)
	if nonWS[0].Type != TokenIdent || nonWS[0].Value != "h1" {
		t.Errorf("first token: %v", nonWS[0])
	}
}

func TestTokenizer_EscapeSequence(t *testing.T) {
	// \41 is hex for 'A'
	tokens := tokenize(`\41 bc`)
	idents := filterType(tokens, TokenIdent)
	if len(idents) < 1 {
		t.Fatal("expected at least 1 ident")
	}
	if idents[0].Value != "Abc" {
		t.Errorf("escaped ident = %q, want 'Abc'", idents[0].Value)
	}
}

func TestTokenizer_HexEscapeInString(t *testing.T) {
	// \2603 is Unicode snowman ☃
	tokens := tokenize(`"\2603 "`)
	strs := filterType(tokens, TokenString)
	if len(strs) != 1 {
		t.Fatalf("expected 1 string, got %d", len(strs))
	}
	if strs[0].Value != "☃" {
		t.Errorf("got %q, want '☃'", strs[0].Value)
	}
}

// ─── Value parsing tests ──────────────────────────────────────────

func TestParseColor_Hex(t *testing.T) {
	tests := []struct {
		input string
		r, g, b float64
	}{
		{"#000", 0, 0, 0},
		{"#fff", 1, 1, 1},
		{"#f00", 1, 0, 0},
		{"#000000", 0, 0, 0},
		{"#ffffff", 1, 1, 1},
		{"#ff0000", 1, 0, 0},
		{"#2980b9", 0x29 / 255.0, 0x80 / 255.0, 0xB9 / 255.0},
	}
	for _, tt := range tests {
		c, ok := ParseColor(tt.input)
		if !ok {
			t.Errorf("ParseColor(%q) failed", tt.input)
			continue
		}
		if !approxEq(c.R, tt.r) || !approxEq(c.G, tt.g) || !approxEq(c.B, tt.b) {
			t.Errorf("ParseColor(%q) = rgb(%g,%g,%g), want rgb(%g,%g,%g)", tt.input, c.R, c.G, c.B, tt.r, tt.g, tt.b)
		}
		if c.A != 1.0 {
			t.Errorf("ParseColor(%q) alpha = %g, want 1.0", tt.input, c.A)
		}
	}
}

func TestParseColor_HexAlpha(t *testing.T) {
	c, ok := ParseColor("#ff000080")
	if !ok {
		t.Fatal("ParseColor failed")
	}
	if !approxEq(c.R, 1) || !approxEq(c.G, 0) || !approxEq(c.B, 0) {
		t.Errorf("rgb = (%g,%g,%g)", c.R, c.G, c.B)
	}
	if !approxEq(c.A, 0x80/255.0) {
		t.Errorf("alpha = %g, want %g", c.A, 0x80/255.0)
	}
}

func TestParseColor_Named(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b float64
	}{
		{"red", 1, 0, 0},
		{"green", 0, 0x80 / 255.0, 0},
		{"blue", 0, 0, 1},
		{"white", 1, 1, 1},
		{"black", 0, 0, 0},
		{"orange", 0xFF / 255.0, 0xA5 / 255.0, 0},
		{"rebeccapurple", 0x66 / 255.0, 0x33 / 255.0, 0x99 / 255.0},
	}
	for _, tt := range tests {
		c, ok := ParseColor(tt.name)
		if !ok {
			t.Errorf("ParseColor(%q) failed", tt.name)
			continue
		}
		if !approxEq(c.R, tt.r) || !approxEq(c.G, tt.g) || !approxEq(c.B, tt.b) {
			t.Errorf("ParseColor(%q) = rgb(%g,%g,%g), want rgb(%g,%g,%g)", tt.name, c.R, c.G, c.B, tt.r, tt.g, tt.b)
		}
	}
}

func TestParseColor_Transparent(t *testing.T) {
	c, ok := ParseColor("transparent")
	if !ok {
		t.Fatal("ParseColor(transparent) failed")
	}
	if c.A != 0 {
		t.Errorf("transparent alpha = %g, want 0", c.A)
	}
}

func TestParseColorFromTokens_RGB(t *testing.T) {
	tokens := tokenize("rgb(255, 128, 0)")
	// Remove whitespace for simpler parsing
	c, ok := ParseColorFromTokens(tokens)
	if !ok {
		t.Fatal("ParseColorFromTokens failed")
	}
	if !approxEq(c.R, 1) || !approxEq(c.G, 128.0/255) || !approxEq(c.B, 0) {
		t.Errorf("rgb = (%g,%g,%g)", c.R, c.G, c.B)
	}
}

func TestParseColorFromTokens_RGBA(t *testing.T) {
	tokens := tokenize("rgba(255, 0, 0, 0.5)")
	c, ok := ParseColorFromTokens(tokens)
	if !ok {
		t.Fatal("ParseColorFromTokens failed")
	}
	if !approxEq(c.R, 1) || !approxEq(c.A, 0.5) {
		t.Errorf("rgba = (%g,%g,%g,%g)", c.R, c.G, c.B, c.A)
	}
}

func TestParseColorFromTokens_Hash(t *testing.T) {
	tokens := []Token{{Type: TokenHash, Value: "ff0000"}}
	c, ok := ParseColorFromTokens(tokens)
	if !ok {
		t.Fatal("ParseColorFromTokens failed")
	}
	if !approxEq(c.R, 1) || !approxEq(c.G, 0) || !approxEq(c.B, 0) {
		t.Errorf("color = (%g,%g,%g)", c.R, c.G, c.B)
	}
}

func TestParseColor_Invalid(t *testing.T) {
	invalids := []string{"", "notacolor", "#gg", "#12345"}
	for _, s := range invalids {
		_, ok := ParseColor(s)
		if ok {
			t.Errorf("ParseColor(%q) should fail", s)
		}
	}
}

func TestCSSValue_String(t *testing.T) {
	tests := []struct {
		val  CSSValue
		want string
	}{
		{CSSValue{Type: ValueIdent, Str: "auto"}, "auto"},
		{CSSValue{Type: ValueNumber, Num: 1.5}, "1.5"},
		{CSSValue{Type: ValueDimension, Num: 12, Unit: "px"}, "12px"},
		{CSSValue{Type: ValuePercentage, Num: 50}, "50%"},
	}
	for _, tt := range tests {
		got := tt.val.String()
		if got != tt.want {
			t.Errorf("CSSValue.String() = %q, want %q", got, tt.want)
		}
	}
}

// ─── Parser tests ─────────────────────────────────────────────────

func TestParseStylesheet_SingleRule(t *testing.T) {
	ss, err := ParseStylesheet(`h1 { font-size: 2em; color: red; }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(ss.Rules))
	}
	rule := ss.Rules[0]
	if rule.Selectors != "h1" {
		t.Errorf("selector = %q, want 'h1'", rule.Selectors)
	}
	if len(rule.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(rule.Declarations))
	}
	assertDecl(t, rule.Declarations[0], "font-size", "2em")
	assertDecl(t, rule.Declarations[1], "color", "red")
}

func TestParseStylesheet_MultipleRules(t *testing.T) {
	ss, err := ParseStylesheet(`
		h1 { color: red; }
		p { margin: 1em 0; }
		.bold { font-weight: bold; }
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(ss.Rules))
	}
	if ss.Rules[0].Selectors != "h1" {
		t.Errorf("rule[0] selector = %q", ss.Rules[0].Selectors)
	}
	if ss.Rules[1].Selectors != "p" {
		t.Errorf("rule[1] selector = %q", ss.Rules[1].Selectors)
	}
	if ss.Rules[2].Selectors != ".bold" {
		t.Errorf("rule[2] selector = %q", ss.Rules[2].Selectors)
	}
}

func TestParseStylesheet_ComplexSelectors(t *testing.T) {
	ss, err := ParseStylesheet(`div.main > p { color: blue; }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(ss.Rules))
	}
	sel := strings.TrimSpace(ss.Rules[0].Selectors)
	if !strings.Contains(sel, "div") || !strings.Contains(sel, ".main") || !strings.Contains(sel, ">") {
		t.Errorf("selector = %q, expected to contain 'div', '.main', '>'", sel)
	}
}

func TestParseStylesheet_GroupedSelectors(t *testing.T) {
	ss, err := ParseStylesheet(`h1, h2, h3 { font-weight: bold; }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(ss.Rules))
	}
	sel := ss.Rules[0].Selectors
	if !strings.Contains(sel, "h1") || !strings.Contains(sel, "h2") || !strings.Contains(sel, "h3") {
		t.Errorf("grouped selector = %q", sel)
	}
}

func TestParseStylesheet_Important(t *testing.T) {
	ss, err := ParseStylesheet(`p { color: red !important; font-size: 12px; }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 || len(ss.Rules[0].Declarations) != 2 {
		t.Fatalf("expected 1 rule with 2 declarations")
	}
	if !ss.Rules[0].Declarations[0].Important {
		t.Error("color should be !important")
	}
	if ss.Rules[0].Declarations[1].Important {
		t.Error("font-size should not be !important")
	}
}

func TestParseStylesheet_Comments(t *testing.T) {
	ss, err := ParseStylesheet(`
		/* Header styles */
		h1 {
			/* Title color */
			color: blue;
			font-size: 2em; /* big */
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(ss.Rules))
	}
	if len(ss.Rules[0].Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(ss.Rules[0].Declarations))
	}
}

func TestParseStylesheet_AtRuleSkip(t *testing.T) {
	ss, err := ParseStylesheet(`
		@import url("style.css");
		@media print { body { font-size: 12pt; } }
		h1 { color: red; }
	`)
	if err != nil {
		t.Fatal(err)
	}
	// @import and @media should be skipped; only h1 rule remains
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule (h1), got %d", len(ss.Rules))
	}
	if ss.Rules[0].Selectors != "h1" {
		t.Errorf("selector = %q", ss.Rules[0].Selectors)
	}
}

func TestParseStylesheet_Empty(t *testing.T) {
	ss, err := ParseStylesheet("")
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(ss.Rules))
	}
}

func TestParseStylesheet_NoDeclarations(t *testing.T) {
	ss, err := ParseStylesheet(`p { }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(ss.Rules))
	}
	if len(ss.Rules[0].Declarations) != 0 {
		t.Fatalf("expected 0 declarations, got %d", len(ss.Rules[0].Declarations))
	}
}

func TestParseStylesheet_ValueWithFunction(t *testing.T) {
	ss, err := ParseStylesheet(`p { color: rgb(255, 0, 0); background: rgba(0, 0, 0, 0.5); }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 || len(ss.Rules[0].Declarations) != 2 {
		t.Fatal("expected 1 rule with 2 declarations")
	}
	colorVal := ss.Rules[0].Declarations[0].Value
	if !strings.Contains(colorVal, "rgb") || !strings.Contains(colorVal, "255") {
		t.Errorf("color value = %q", colorVal)
	}
}

func TestParseStylesheet_MultiValueProperty(t *testing.T) {
	ss, err := ParseStylesheet(`p { margin: 10px 20px 30px 40px; }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 1 || len(ss.Rules[0].Declarations) != 1 {
		t.Fatal("expected 1 rule with 1 declaration")
	}
	val := ss.Rules[0].Declarations[0].Value
	if !strings.Contains(val, "10px") || !strings.Contains(val, "40px") {
		t.Errorf("margin value = %q", val)
	}
}

func TestParseStylesheet_BorderShorthand(t *testing.T) {
	ss, err := ParseStylesheet(`div { border: 2px solid #2980b9; }`)
	if err != nil {
		t.Fatal(err)
	}
	decl := ss.Rules[0].Declarations[0]
	if decl.Property != "border" {
		t.Errorf("property = %q", decl.Property)
	}
	val := decl.Value
	if !strings.Contains(val, "2px") || !strings.Contains(val, "solid") || !strings.Contains(val, "#2980b9") {
		t.Errorf("border value = %q", val)
	}
}

func TestParseStylesheet_MissingSemicolon(t *testing.T) {
	// Last declaration without semicolon before }
	ss, err := ParseStylesheet(`p { color: red; font-size: 12px }`)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules[0].Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(ss.Rules[0].Declarations))
	}
}

func TestParseStylesheet_CaseSensitivity(t *testing.T) {
	ss, err := ParseStylesheet(`P { Font-Size: 12PX; COLOR: RED; }`)
	if err != nil {
		t.Fatal(err)
	}
	// Property names should be lowercased
	if ss.Rules[0].Declarations[0].Property != "font-size" {
		t.Errorf("property = %q, want 'font-size'", ss.Rules[0].Declarations[0].Property)
	}
	if ss.Rules[0].Declarations[1].Property != "color" {
		t.Errorf("property = %q, want 'color'", ss.Rules[0].Declarations[1].Property)
	}
}

func TestParseDeclarations_InlineStyle(t *testing.T) {
	decls := ParseDeclarations("color: red; font-size: 14px; margin: 10px 20px")
	if len(decls) != 3 {
		t.Fatalf("expected 3 declarations, got %d", len(decls))
	}
	assertDecl(t, decls[0], "color", "red")
	assertDecl(t, decls[1], "font-size", "14px")
	if decls[2].Property != "margin" {
		t.Errorf("property = %q", decls[2].Property)
	}
}

func TestParseDeclarations_Empty(t *testing.T) {
	decls := ParseDeclarations("")
	if len(decls) != 0 {
		t.Fatalf("expected 0 declarations, got %d", len(decls))
	}
}

func TestParseDeclarations_InvalidSkip(t *testing.T) {
	// Invalid declarations should be skipped
	decls := ParseDeclarations("color red; font-size: 12px; : value; valid: yes")
	// "color red" has no colon — skip
	// ": value" has no property — skip
	found := make(map[string]bool)
	for _, d := range decls {
		found[d.Property] = true
	}
	if !found["font-size"] {
		t.Error("expected font-size declaration")
	}
	if !found["valid"] {
		t.Error("expected valid declaration")
	}
}

func TestParseValue_Basic(t *testing.T) {
	tokens := ParseValue("12px solid #333")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenDimension || tokens[0].Unit != "px" {
		t.Errorf("token[0] = %v", tokens[0])
	}
	if tokens[1].Type != TokenIdent || tokens[1].Value != "solid" {
		t.Errorf("token[1] = %v", tokens[1])
	}
	if tokens[2].Type != TokenHash || tokens[2].Value != "333" {
		t.Errorf("token[2] = %v", tokens[2])
	}
}

func TestParseStylesheet_InvoiceCSS(t *testing.T) {
	css := `
		body { font-family: NotoSansJP; font-size: 12pt; }
		h1 { color: #2980b9; border-bottom: 2px solid #2980b9; padding-bottom: 5mm; }
		.container { max-width: 170mm; margin: 0 auto; }
		table { width: 100%; border-collapse: collapse; }
		th { background-color: #2980b9; color: white; padding: 8pt; }
		td { border: 1px solid #ddd; padding: 8pt; }
	`
	ss, err := ParseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Rules) != 6 {
		t.Fatalf("expected 6 rules, got %d", len(ss.Rules))
	}

	// Verify body rule
	body := ss.Rules[0]
	if body.Selectors != "body" {
		t.Errorf("rule[0] selector = %q", body.Selectors)
	}
	if len(body.Declarations) != 2 {
		t.Errorf("body declarations = %d", len(body.Declarations))
	}

	// Verify table rule
	table := ss.Rules[3]
	if table.Selectors != "table" {
		t.Errorf("rule[3] selector = %q", table.Selectors)
	}
	assertDecl(t, table.Declarations[0], "width", "100%")
	assertDecl(t, table.Declarations[1], "border-collapse", "collapse")
}

func TestToken_IsIdent(t *testing.T) {
	tok := Token{Type: TokenIdent, Value: "bold"}
	if !tok.IsIdent("bold") {
		t.Error("IsIdent(bold) should be true")
	}
	if !tok.IsIdent("BOLD") {
		t.Error("IsIdent(BOLD) should be true (case-insensitive)")
	}
	if tok.IsIdent("italic") {
		t.Error("IsIdent(italic) should be false")
	}
}

// ─── Test helpers ─────────────────────────────────────────────────

func tokenize(input string) []Token {
	tok := NewTokenizer([]byte(input))
	var tokens []Token
	for {
		t := tok.Next()
		if t.Type == TokenEOF {
			break
		}
		tokens = append(tokens, t)
	}
	return tokens
}

func filterType(tokens []Token, typ TokenType) []Token {
	var result []Token
	for _, t := range tokens {
		if t.Type == typ {
			result = append(result, t)
		}
	}
	return result
}

func filterNonWhitespace(tokens []Token) []Token {
	var result []Token
	for _, t := range tokens {
		if t.Type != TokenWhitespace {
			result = append(result, t)
		}
	}
	return result
}

func assertDecl(t *testing.T, d Declaration, prop, val string) {
	t.Helper()
	if d.Property != prop {
		t.Errorf("declaration property = %q, want %q", d.Property, prop)
	}
	trimmed := strings.TrimSpace(d.Value)
	if trimmed != val {
		t.Errorf("declaration %s value = %q, want %q", prop, trimmed, val)
	}
}

func approxEq(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}
