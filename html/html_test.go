package html

import (
	"strings"
	"testing"
)

// ─── Node helpers ─────────────────────────────────────────────────

func TestNode_AppendChild(t *testing.T) {
	parent := &Node{Type: ElementNode, Tag: "div"}
	child := &Node{Type: TextNode, Content: "hello"}
	parent.AppendChild(child)
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if child.Parent != parent {
		t.Fatal("parent not set")
	}
}

func TestNode_RemoveChild(t *testing.T) {
	parent := &Node{Type: ElementNode, Tag: "div"}
	child := &Node{Type: TextNode, Content: "hello"}
	parent.AppendChild(child)
	parent.RemoveChild(child)
	if len(parent.Children) != 0 {
		t.Fatalf("expected 0 children, got %d", len(parent.Children))
	}
	if child.Parent != nil {
		t.Fatal("parent should be nil")
	}
}

func TestNode_Attr(t *testing.T) {
	n := &Node{Type: ElementNode, Tag: "div", Attrs: map[string]string{"class": "main", "id": "app"}}
	if v, ok := n.Attr("class"); !ok || v != "main" {
		t.Fatalf("expected class=main, got %q %v", v, ok)
	}
	if _, ok := n.Attr("missing"); ok {
		t.Fatal("expected missing attr to return false")
	}
}

func TestNode_HasClass(t *testing.T) {
	n := &Node{Type: ElementNode, Tag: "div", Attrs: map[string]string{"class": "foo bar baz"}}
	if !n.HasClass("bar") {
		t.Fatal("expected HasClass(bar) to be true")
	}
	if n.HasClass("qux") {
		t.Fatal("expected HasClass(qux) to be false")
	}
}

func TestNode_TextContent(t *testing.T) {
	// <div>Hello <b>world</b></div>
	div := &Node{Type: ElementNode, Tag: "div"}
	div.AppendChild(&Node{Type: TextNode, Content: "Hello "})
	b := &Node{Type: ElementNode, Tag: "b"}
	b.AppendChild(&Node{Type: TextNode, Content: "world"})
	div.AppendChild(b)

	got := div.TextContent()
	if got != "Hello world" {
		t.Fatalf("expected 'Hello world', got %q", got)
	}
}

func TestNode_StyleElements(t *testing.T) {
	root, err := ParseString(`<html><head><style>h1{}</style></head><body><style>p{}</style></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	styles := root.StyleElements()
	if len(styles) != 2 {
		t.Fatalf("expected 2 style elements, got %d", len(styles))
	}
}

// ─── Entity decoding ──────────────────────────────────────────────

func TestDecodeEntities_Basic(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&apos;", "'"},
		{"&nbsp;", "\u00A0"},
		{"&copy;", "\u00A9"},
		{"&yen;", "\u00A5"},
		{"&euro;", "\u20AC"},
		{"&mdash;", "\u2014"},
	}
	for _, tt := range tests {
		got := decodeEntities(tt.input)
		if got != tt.want {
			t.Errorf("decodeEntities(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDecodeEntities_Numeric(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"&#123;", "{"},
		{"&#x7B;", "{"},
		{"&#x7b;", "{"},
		{"&#65;", "A"},
		{"&#x41;", "A"},
		{"&#12354;", "あ"},   // hiragana 'a'
		{"&#x3042;", "あ"},   // hiragana 'a' in hex
		{"&#x1F600;", "😀"}, // emoji
	}
	for _, tt := range tests {
		got := decodeEntities(tt.input)
		if got != tt.want {
			t.Errorf("decodeEntities(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDecodeEntities_Invalid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"&;", "&;"},                 // empty entity
		{"&unknown;", "&unknown;"},   // unknown named entity
		{"&#;", "&#;"},               // empty numeric
		{"&#x;", "&#x;"},            // empty hex
		{"&#999999999;", "&#999999999;"}, // overflow
		{"&", "&"},                   // bare ampersand
		{"&no-semicolon", "&no-semicolon"}, // no semicolon
	}
	for _, tt := range tests {
		got := decodeEntities(tt.input)
		if got != tt.want {
			t.Errorf("decodeEntities(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDecodeEntities_Mixed(t *testing.T) {
	input := "Hello &amp; &lt;world&gt; &#x263A;"
	want := "Hello & <world> \u263A"
	got := decodeEntities(input)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ─── Tokenizer ────────────────────────────────────────────────────

func TestTokenizer_BasicTags(t *testing.T) {
	tok := newTokenizer([]byte(`<div class="main"><p>Hello</p></div>`))
	tokens := collectTokens(tok)

	assertToken(t, tokens, 0, tokenStartTag, "div")
	assertToken(t, tokens, 1, tokenStartTag, "p")
	assertToken(t, tokens, 2, tokenText, "")
	assertToken(t, tokens, 3, tokenEndTag, "p")
	assertToken(t, tokens, 4, tokenEndTag, "div")
}

func TestTokenizer_SelfClosing(t *testing.T) {
	tok := newTokenizer([]byte(`<br/><img src="x.png" />`))
	tokens := collectTokens(tok)

	assertToken(t, tokens, 0, tokenSelfClose, "br")
	assertToken(t, tokens, 1, tokenSelfClose, "img")
}

func TestTokenizer_Comment(t *testing.T) {
	tok := newTokenizer([]byte(`<!-- comment -->`))
	tokens := collectTokens(tok)

	if len(tokens) < 1 || tokens[0].typ != tokenComment {
		t.Fatal("expected comment token")
	}
	if strings.TrimSpace(tokens[0].data) != "comment" {
		t.Fatalf("expected 'comment', got %q", tokens[0].data)
	}
}

func TestTokenizer_Doctype(t *testing.T) {
	tok := newTokenizer([]byte(`<!DOCTYPE html><html></html>`))
	tokens := collectTokens(tok)

	assertToken(t, tokens, 0, tokenDoctype, "")
	assertToken(t, tokens, 1, tokenStartTag, "html")
	assertToken(t, tokens, 2, tokenEndTag, "html")
}

func TestTokenizer_Attributes(t *testing.T) {
	tok := newTokenizer([]byte(`<input type="text" disabled value='hello' data-x=42>`))
	tokens := collectTokens(tok)

	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	attrs := tokens[0].attrs
	if attrs["type"] != "text" {
		t.Errorf("type attr: got %q", attrs["type"])
	}
	if attrs["disabled"] != "" {
		t.Errorf("disabled attr: got %q, want empty string", attrs["disabled"])
	}
	if attrs["value"] != "hello" {
		t.Errorf("value attr: got %q", attrs["value"])
	}
	if attrs["data-x"] != "42" {
		t.Errorf("data-x attr: got %q", attrs["data-x"])
	}
}

func TestTokenizer_EntityInText(t *testing.T) {
	tok := newTokenizer([]byte(`<p>&lt;hello&gt;</p>`))
	tokens := collectTokens(tok)
	// tokens: <p>, text, </p>
	if len(tokens) < 2 {
		t.Fatal("expected at least 2 tokens")
	}
	if tokens[1].data != "<hello>" {
		t.Errorf("expected '<hello>', got %q", tokens[1].data)
	}
}

func TestTokenizer_EntityInAttr(t *testing.T) {
	tok := newTokenizer([]byte(`<a href="x&amp;y">`))
	tokens := collectTokens(tok)
	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	if tokens[0].attrs["href"] != "x&y" {
		t.Errorf("expected 'x&y', got %q", tokens[0].attrs["href"])
	}
}

func TestTokenizer_CaseFolding(t *testing.T) {
	tok := newTokenizer([]byte(`<DIV Class="X"><P></P></DIV>`))
	tokens := collectTokens(tok)
	assertToken(t, tokens, 0, tokenStartTag, "div")
	assertToken(t, tokens, 1, tokenStartTag, "p")
	if tokens[0].attrs["class"] != "X" {
		t.Errorf("attr name should be lowercase but value preserved: got %q", tokens[0].attrs["class"])
	}
}

// ─── Parser: Basic structure ──────────────────────────────────────

func TestParse_SimpleDocument(t *testing.T) {
	root, err := ParseString(`<html><head><title>Test</title></head><body><p>Hello</p></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if root.Type != DocumentNode {
		t.Fatal("expected DocumentNode as root")
	}
	html := findElement(root, "html")
	if html == nil {
		t.Fatal("expected <html> element")
	}
	body := findElement(html, "body")
	if body == nil {
		t.Fatal("expected <body> element")
	}
	p := findElement(body, "p")
	if p == nil {
		t.Fatal("expected <p> element")
	}
	if p.TextContent() != "Hello" {
		t.Fatalf("expected 'Hello', got %q", p.TextContent())
	}
}

func TestParse_NestedElements(t *testing.T) {
	root, err := ParseString(`<div><div><div><span>deep</span></div></div></div>`)
	if err != nil {
		t.Fatal(err)
	}
	div1 := findElement(root, "div")
	if div1 == nil {
		t.Fatal("expected outer div")
	}
	div2 := findElement(div1, "div")
	if div2 == nil {
		t.Fatal("expected middle div")
	}
	div3 := findElement(div2, "div")
	if div3 == nil {
		t.Fatal("expected inner div")
	}
	span := findElement(div3, "span")
	if span == nil {
		t.Fatal("expected span")
	}
	if span.TextContent() != "deep" {
		t.Fatalf("expected 'deep', got %q", span.TextContent())
	}
}

func TestParse_MixedContent(t *testing.T) {
	root, err := ParseString(`<p>Hello <strong>world</strong> and <em>more</em></p>`)
	if err != nil {
		t.Fatal(err)
	}
	p := findElement(root, "p")
	if p == nil {
		t.Fatal("expected <p>")
	}
	if len(p.Children) != 4 {
		t.Fatalf("expected 4 children, got %d", len(p.Children))
	}
	if p.Children[0].Content != "Hello " {
		t.Errorf("first text: %q", p.Children[0].Content)
	}
	if p.Children[1].Tag != "strong" {
		t.Errorf("expected strong, got %q", p.Children[1].Tag)
	}
	if p.Children[2].Content != " and " {
		t.Errorf("between text: %q", p.Children[2].Content)
	}
	if p.Children[3].Tag != "em" {
		t.Errorf("expected em, got %q", p.Children[3].Tag)
	}
}

// ─── Parser: Void elements ────────────────────────────────────────

func TestParse_VoidElements(t *testing.T) {
	root, err := ParseString(`<p>line1<br>line2<hr><img src="x.png"></p>`)
	if err != nil {
		t.Fatal(err)
	}
	p := findElement(root, "p")
	if p == nil {
		t.Fatal("expected <p>")
	}
	// Should have: text, br, text, hr, img
	br := findElement(p, "br")
	if br == nil {
		t.Fatal("expected <br>")
	}
	if len(br.Children) != 0 {
		t.Fatal("void element should have no children")
	}
}

func TestParse_VoidWithoutSlash(t *testing.T) {
	root, err := ParseString(`<br><input type="text"><hr>`)
	if err != nil {
		t.Fatal(err)
	}
	// All should be children of root, not nested
	elements := collectElements(root)
	tags := make([]string, len(elements))
	for i, e := range elements {
		tags[i] = e.Tag
	}
	for _, expected := range []string{"br", "input", "hr"} {
		found := false
		for _, tag := range tags {
			if tag == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected <%s> in tree", expected)
		}
	}
}

// ─── Parser: Implicit close ───────────────────────────────────────

func TestParse_ImplicitClose_PinP(t *testing.T) {
	root, err := ParseString(`<p>first<p>second<p>third`)
	if err != nil {
		t.Fatal(err)
	}
	ps := findAllElements(root, "p")
	if len(ps) != 3 {
		t.Fatalf("expected 3 <p> elements, got %d", len(ps))
	}
	// Each <p> should be a sibling, not nested
	for i, p := range ps {
		if p.TextContent() != []string{"first", "second", "third"}[i] {
			t.Errorf("p[%d] text = %q", i, p.TextContent())
		}
	}
}

func TestParse_ImplicitClose_DivInP(t *testing.T) {
	root, err := ParseString(`<p>text<div>block</div>`)
	if err != nil {
		t.Fatal(err)
	}
	p := findElement(root, "p")
	div := findElement(root, "div")
	if p == nil || div == nil {
		t.Fatal("expected both <p> and <div>")
	}
	// <div> should NOT be inside <p>
	if findElement(p, "div") != nil {
		t.Fatal("<div> should not be nested inside <p>")
	}
}

func TestParse_ImplicitClose_Li(t *testing.T) {
	root, err := ParseString(`<ul><li>a<li>b<li>c</ul>`)
	if err != nil {
		t.Fatal(err)
	}
	lis := findAllElements(root, "li")
	if len(lis) != 3 {
		t.Fatalf("expected 3 <li>, got %d", len(lis))
	}
}

func TestParse_ImplicitClose_Td(t *testing.T) {
	root, err := ParseString(`<table><tr><td>a<td>b<td>c</tr></table>`)
	if err != nil {
		t.Fatal(err)
	}
	tds := findAllElements(root, "td")
	if len(tds) != 3 {
		t.Fatalf("expected 3 <td>, got %d", len(tds))
	}
}

// ─── Parser: Unclosed tags ────────────────────────────────────────

func TestParse_UnclosedTags(t *testing.T) {
	// Parser should not error on unclosed tags
	root, err := ParseString(`<div><p>text<span>inner`)
	if err != nil {
		t.Fatal(err)
	}
	div := findElement(root, "div")
	if div == nil {
		t.Fatal("expected <div>")
	}
	if div.TextContent() != "textinner" {
		t.Fatalf("expected 'textinner', got %q", div.TextContent())
	}
}

// ─── Parser: Cross-nesting ────────────────────────────────────────

func TestParse_CrossNesting(t *testing.T) {
	// <b><i></b></i> — should recover gracefully
	root, err := ParseString(`<p><b><i>text</b></i></p>`)
	if err != nil {
		t.Fatal(err)
	}
	// Should not panic; exact tree structure depends on recovery strategy
	if root.TextContent() != "text" {
		t.Fatalf("expected 'text', got %q", root.TextContent())
	}
}

// ─── Parser: <style> tag ──────────────────────────────────────────

func TestParse_StyleTag(t *testing.T) {
	root, err := ParseString(`<html><head><style>h1 { color: red; } p > .class { margin: 0; }</style></head><body></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	styles := root.StyleElements()
	if len(styles) != 1 {
		t.Fatalf("expected 1 style element, got %d", len(styles))
	}
	css := styles[0].TextContent()
	if !strings.Contains(css, "color: red") {
		t.Errorf("style content should contain 'color: red', got %q", css)
	}
	if !strings.Contains(css, "p > .class") {
		t.Errorf("style content should preserve selectors, got %q", css)
	}
}

func TestParse_StyleTagInBody(t *testing.T) {
	root, err := ParseString(`<body><style>.x{}</style><p>text</p></body>`)
	if err != nil {
		t.Fatal(err)
	}
	styles := root.StyleElements()
	if len(styles) != 1 {
		t.Fatalf("expected 1 style, got %d", len(styles))
	}
}

func TestParse_MultipleStyleTags(t *testing.T) {
	root, err := ParseString(`<style>a{}</style><style>b{}</style><style>c{}</style>`)
	if err != nil {
		t.Fatal(err)
	}
	styles := root.StyleElements()
	if len(styles) != 3 {
		t.Fatalf("expected 3 style elements, got %d", len(styles))
	}
}

// ─── Parser: Attributes ──────────────────────────────────────────

func TestParse_Attributes(t *testing.T) {
	root, err := ParseString(`<div id="main" class="container large" data-value="123"></div>`)
	if err != nil {
		t.Fatal(err)
	}
	div := findElement(root, "div")
	if div == nil {
		t.Fatal("expected <div>")
	}
	if div.ID() != "main" {
		t.Errorf("id: got %q", div.ID())
	}
	if !div.HasClass("container") {
		t.Error("expected HasClass(container)")
	}
	if !div.HasClass("large") {
		t.Error("expected HasClass(large)")
	}
	v, _ := div.Attr("data-value")
	if v != "123" {
		t.Errorf("data-value: got %q", v)
	}
}

// ─── Parser: Comments ─────────────────────────────────────────────

func TestParse_Comments(t *testing.T) {
	root, err := ParseString(`<div><!-- comment --><p>text</p></div>`)
	if err != nil {
		t.Fatal(err)
	}
	div := findElement(root, "div")
	if div == nil {
		t.Fatal("expected <div>")
	}
	hasComment := false
	for _, child := range div.Children {
		if child.Type == CommentNode {
			hasComment = true
			if strings.TrimSpace(child.Content) != "comment" {
				t.Errorf("comment content: %q", child.Content)
			}
		}
	}
	if !hasComment {
		t.Fatal("expected comment node")
	}
}

// ─── Parser: Unknown tags ─────────────────────────────────────────

func TestParse_UnknownTags(t *testing.T) {
	root, err := ParseString(`<custom-element>content</custom-element>`)
	if err != nil {
		t.Fatal(err)
	}
	el := findElement(root, "custom-element")
	if el == nil {
		t.Fatal("expected custom-element")
	}
	if el.TextContent() != "content" {
		t.Fatalf("expected 'content', got %q", el.TextContent())
	}
}

// ─── Parser: Safety limits ────────────────────────────────────────

func TestParse_MaxDepth(t *testing.T) {
	// Build deeply nested HTML
	var sb strings.Builder
	depth := 300
	for i := 0; i < depth; i++ {
		sb.WriteString("<div>")
	}
	sb.WriteString("deep")
	for i := 0; i < depth; i++ {
		sb.WriteString("</div>")
	}

	_, err := ParseString(sb.String(), WithMaxDepth(256))
	if err == nil {
		t.Fatal("expected depth limit error")
	}
	if !strings.Contains(err.Error(), "nesting depth") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParse_MaxNodeCount(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("<p>text</p>")
	}
	_, err := ParseString(sb.String(), WithMaxNodeCount(50))
	if err == nil {
		t.Fatal("expected node count limit error")
	}
	if !strings.Contains(err.Error(), "node count") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParse_MaxInputSize(t *testing.T) {
	input := strings.Repeat("x", 1024)
	_, err := ParseString(input, WithMaxInputSize(512))
	if err == nil {
		t.Fatal("expected input size limit error")
	}
	if !strings.Contains(err.Error(), "input size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Parser: Edge cases ───────────────────────────────────────────

func TestParse_EmptyInput(t *testing.T) {
	root, err := ParseString("")
	if err != nil {
		t.Fatal(err)
	}
	if root.Type != DocumentNode {
		t.Fatal("expected DocumentNode")
	}
	if len(root.Children) != 0 {
		t.Fatalf("expected 0 children, got %d", len(root.Children))
	}
}

func TestParse_WhitespaceOnly(t *testing.T) {
	root, err := ParseString("   \n\t  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 text child, got %d", len(root.Children))
	}
	if root.Children[0].Type != TextNode {
		t.Fatal("expected TextNode")
	}
}

func TestParse_TextOnly(t *testing.T) {
	root, err := ParseString("just text")
	if err != nil {
		t.Fatal(err)
	}
	if root.TextContent() != "just text" {
		t.Fatalf("expected 'just text', got %q", root.TextContent())
	}
}

func TestParse_DOCTYPE(t *testing.T) {
	root, err := ParseString(`<!DOCTYPE html><html><body></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	html := findElement(root, "html")
	if html == nil {
		t.Fatal("expected <html>")
	}
}

func TestParse_CJKContent(t *testing.T) {
	root, err := ParseString(`<p>日本語テスト</p><p>中文测试</p><p>한국어 테스트</p>`)
	if err != nil {
		t.Fatal(err)
	}
	ps := findAllElements(root, "p")
	if len(ps) != 3 {
		t.Fatalf("expected 3 <p> elements, got %d", len(ps))
	}
	if ps[0].TextContent() != "日本語テスト" {
		t.Errorf("Japanese: %q", ps[0].TextContent())
	}
	if ps[1].TextContent() != "中文测试" {
		t.Errorf("Chinese: %q", ps[1].TextContent())
	}
	if ps[2].TextContent() != "한국어 테스트" {
		t.Errorf("Korean: %q", ps[2].TextContent())
	}
}

func TestParse_HeadingElements(t *testing.T) {
	root, err := ParseString(`<h1>Title</h1><h2>Sub</h2><h3>Sub2</h3><h4>Sub3</h4><h5>Sub4</h5><h6>Sub5</h6>`)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		el := findElement(root, tag)
		if el == nil {
			t.Errorf("expected <%s>", tag)
		}
	}
}

func TestParse_ListElements(t *testing.T) {
	root, err := ParseString(`<ul><li>a</li><li>b</li></ul><ol><li>1</li><li>2</li></ol>`)
	if err != nil {
		t.Fatal(err)
	}
	ul := findElement(root, "ul")
	ol := findElement(root, "ol")
	if ul == nil || ol == nil {
		t.Fatal("expected <ul> and <ol>")
	}
}

func TestParse_TableElements(t *testing.T) {
	root, err := ParseString(`<table><thead><tr><th>H1</th><th>H2</th></tr></thead><tbody><tr><td>A</td><td>B</td></tr></tbody></table>`)
	if err != nil {
		t.Fatal(err)
	}
	table := findElement(root, "table")
	if table == nil {
		t.Fatal("expected <table>")
	}
	thead := findElement(table, "thead")
	if thead == nil {
		t.Fatal("expected <thead>")
	}
	tbody := findElement(table, "tbody")
	if tbody == nil {
		t.Fatal("expected <tbody>")
	}
}

func TestParse_BlockquotePreCode(t *testing.T) {
	root, err := ParseString(`<blockquote>quote</blockquote><pre>preformatted</pre><code>inline code</code>`)
	if err != nil {
		t.Fatal(err)
	}
	bq := findElement(root, "blockquote")
	if bq == nil {
		t.Fatal("expected <blockquote>")
	}
	pre := findElement(root, "pre")
	if pre == nil {
		t.Fatal("expected <pre>")
	}
	code := findElement(root, "code")
	if code == nil {
		t.Fatal("expected <code>")
	}
}

func TestParse_ImgAttributes(t *testing.T) {
	root, err := ParseString(`<img src="photo.png" alt="Photo" width="100" height="50">`)
	if err != nil {
		t.Fatal(err)
	}
	img := findElement(root, "img")
	if img == nil {
		t.Fatal("expected <img>")
	}
	src, _ := img.Attr("src")
	if src != "photo.png" {
		t.Errorf("src: %q", src)
	}
	alt, _ := img.Attr("alt")
	if alt != "Photo" {
		t.Errorf("alt: %q", alt)
	}
}

func TestParse_Walk(t *testing.T) {
	root, err := ParseString(`<div><p>a</p><p>b</p></div>`)
	if err != nil {
		t.Fatal(err)
	}
	var tags []string
	root.Walk(func(n *Node) bool {
		if n.Type == ElementNode {
			tags = append(tags, n.Tag)
		}
		return true
	})
	if len(tags) != 3 { // div, p, p
		t.Fatalf("expected 3 elements, got %v", tags)
	}
}

func TestParse_WalkSkipSubtree(t *testing.T) {
	root, err := ParseString(`<div><p><span>inner</span></p><p>other</p></div>`)
	if err != nil {
		t.Fatal(err)
	}
	var tags []string
	root.Walk(func(n *Node) bool {
		if n.Type == ElementNode {
			tags = append(tags, n.Tag)
			if n.Tag == "p" {
				return false // skip children of first <p> encountered
			}
		}
		return true
	})
	// Should visit div, first p (skip its children), second p (skip its children)
	if len(tags) != 3 {
		t.Fatalf("expected 3 elements (div, p, p), got %v", tags)
	}
}

func TestParse_FullInvoiceHTML(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: NotoSansJP; font-size: 12pt; }
        h1 { color: #2980b9; border-bottom: 2px solid #2980b9; }
        table { width: 100%; border-collapse: collapse; }
        th { background-color: #2980b9; color: white; padding: 8pt; }
        td { border: 1px solid #ddd; padding: 8pt; }
    </style>
</head>
<body>
    <h1>請求書</h1>
    <p>株式会社サンプル 御中</p>
    <table>
        <thead>
            <tr><th>品目</th><th>数量</th><th>単価</th><th>金額</th></tr>
        </thead>
        <tbody>
            <tr><td>コンサルティング</td><td>10</td><td>¥50,000</td><td>¥500,000</td></tr>
            <tr><td>開発費</td><td>1</td><td>¥1,000,000</td><td>¥1,000,000</td></tr>
        </tbody>
    </table>
</body>
</html>`

	root, err := ParseString(html)
	if err != nil {
		t.Fatal(err)
	}

	// Verify structure
	htmlEl := findElement(root, "html")
	if htmlEl == nil {
		t.Fatal("expected <html>")
	}
	body := findElement(htmlEl, "body")
	if body == nil {
		t.Fatal("expected <body>")
	}
	h1 := findElement(body, "h1")
	if h1 == nil || h1.TextContent() != "請求書" {
		t.Fatal("expected <h1>請求書</h1>")
	}

	// Verify style extraction
	styles := root.StyleElements()
	if len(styles) != 1 {
		t.Fatalf("expected 1 style, got %d", len(styles))
	}
	css := styles[0].TextContent()
	if !strings.Contains(css, "border-collapse") {
		t.Error("expected CSS to contain 'border-collapse'")
	}

	// Verify table structure
	table := findElement(body, "table")
	if table == nil {
		t.Fatal("expected <table>")
	}
	ths := findAllElements(table, "th")
	if len(ths) != 4 {
		t.Fatalf("expected 4 <th>, got %d", len(ths))
	}
	tds := findAllElements(table, "td")
	if len(tds) != 8 {
		t.Fatalf("expected 8 <td>, got %d", len(tds))
	}
}

// ─── Parser: io.Reader interface ──────────────────────────────────

func TestParse_Reader(t *testing.T) {
	r := strings.NewReader("<p>hello</p>")
	root, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}
	p := findElement(root, "p")
	if p == nil || p.TextContent() != "hello" {
		t.Fatal("expected <p>hello</p>")
	}
}

// ─── Test helpers ─────────────────────────────────────────────────

func collectTokens(t *tokenizer) []token {
	var tokens []token
	for {
		tok := t.next()
		if tok.typ == tokenEOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func assertToken(t *testing.T, tokens []token, idx int, typ tokenType, tag string) {
	t.Helper()
	if idx >= len(tokens) {
		t.Fatalf("token[%d]: out of range (have %d tokens)", idx, len(tokens))
	}
	tok := tokens[idx]
	if tok.typ != typ {
		t.Errorf("token[%d]: type = %v, want %v (%s)", idx, tok.typ, typ, tok)
	}
	if tag != "" && tok.tag != tag {
		t.Errorf("token[%d]: tag = %q, want %q (%s)", idx, tok.tag, tag, tok)
	}
}

func findElement(root *Node, tag string) *Node {
	if root.Type == ElementNode && root.Tag == tag {
		return root
	}
	for _, child := range root.Children {
		if found := findElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func findAllElements(root *Node, tag string) []*Node {
	var result []*Node
	root.Walk(func(n *Node) bool {
		if n.Type == ElementNode && n.Tag == tag {
			result = append(result, n)
		}
		return true
	})
	return result
}

func collectElements(root *Node) []*Node {
	var result []*Node
	root.Walk(func(n *Node) bool {
		if n.Type == ElementNode {
			result = append(result, n)
		}
		return true
	})
	return result
}
