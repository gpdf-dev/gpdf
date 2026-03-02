package htmlpdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// ─── Converter tests ─────────────────────────────────────────────

func TestFromHTML_BasicParagraph(t *testing.T) {
	result, err := FromHTML("<p>Hello, World!</p>")
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.doc == nil {
		t.Fatal("document is nil")
	}
	if len(result.doc.Pages) == 0 {
		t.Fatal("no pages in document")
	}
}

func TestFromHTML_MultipleElements(t *testing.T) {
	html := `
	<html>
	<body>
		<h1>Title</h1>
		<p>First paragraph.</p>
		<p>Second paragraph.</p>
	</body>
	</html>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) < 3 {
		t.Errorf("expected at least 3 content nodes (h1 + 2p), got %d", len(page.Content))
	}
}

func TestFromHTML_InlineElements(t *testing.T) {
	html := `<p>This is <strong>bold</strong> and <em>italic</em> text.</p>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}

	// The <p> should produce a Box containing a RichText with multiple fragments
	box, ok := page.Content[0].(*document.Box)
	if !ok {
		t.Fatalf("expected Box, got %T", page.Content[0])
	}
	if len(box.Content) == 0 {
		t.Fatal("box has no content")
	}

	rt, ok := box.Content[0].(*document.RichText)
	if !ok {
		t.Fatalf("expected RichText, got %T", box.Content[0])
	}
	if len(rt.Fragments) < 3 {
		t.Errorf("expected at least 3 fragments (text + bold + text + italic + text), got %d", len(rt.Fragments))
	}
}

func TestFromHTML_HeadingStyles(t *testing.T) {
	html := `<h1>Big</h1><h2>Medium</h2><h3>Small</h3>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) < 3 {
		t.Errorf("expected 3 heading nodes, got %d", len(page.Content))
	}
}

func TestFromHTML_UnorderedList(t *testing.T) {
	html := `<ul><li>One</li><li>Two</li><li>Three</li></ul>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}

	list, ok := page.Content[0].(*document.List)
	if !ok {
		t.Fatalf("expected List, got %T", page.Content[0])
	}
	if len(list.Items) != 3 {
		t.Errorf("expected 3 list items, got %d", len(list.Items))
	}
	if list.ListType != document.Unordered {
		t.Errorf("expected Unordered, got %v", list.ListType)
	}
}

func TestFromHTML_OrderedList(t *testing.T) {
	html := `<ol><li>First</li><li>Second</li></ol>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	list, ok := page.Content[0].(*document.List)
	if !ok {
		t.Fatalf("expected List, got %T", page.Content[0])
	}
	if list.ListType != document.Ordered {
		t.Errorf("expected Ordered, got %v", list.ListType)
	}
}

func TestFromHTML_Blockquote(t *testing.T) {
	html := `<blockquote>Quoted text</blockquote>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}
	// blockquote maps to a Box
	_, ok := page.Content[0].(*document.Box)
	if !ok {
		t.Fatalf("expected Box for blockquote, got %T", page.Content[0])
	}
}

func TestFromHTML_PreformattedCode(t *testing.T) {
	html := `<pre><code>func main() {}</code></pre>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}
}

func TestFromHTML_HorizontalRule(t *testing.T) {
	html := `<p>Before</p><hr><p>After</p>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) < 3 {
		t.Errorf("expected at least 3 nodes (p + hr + p), got %d", len(page.Content))
	}
}

func TestFromHTML_NestedStructure(t *testing.T) {
	// No whitespace between tags to avoid anonymous text nodes
	html := `<div><h2>Section</h2><div><p>Nested paragraph.</p></div></div>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}
	// Should be a Box (div) containing children
	box, ok := page.Content[0].(*document.Box)
	if !ok {
		t.Fatalf("expected Box for outer div, got %T", page.Content[0])
	}
	if len(box.Content) < 2 {
		t.Errorf("expected at least 2 children (h2 + div), got %d", len(box.Content))
	}
}

func TestFromHTML_DisplayNone(t *testing.T) {
	html := `<p>Visible</p><p style="display:none">Hidden</p>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	// Only 1 visible paragraph should be present
	if len(page.Content) != 1 {
		t.Errorf("expected 1 visible node, got %d", len(page.Content))
	}
}

func TestFromHTML_InlineStyle(t *testing.T) {
	html := `<p style="color: red; font-size: 20pt;">Styled</p>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}
}

func TestFromHTML_StyleElement(t *testing.T) {
	html := `
	<html>
	<head>
		<style>
		h1 { font-size: 36pt; color: blue; }
		.highlight { background-color: yellow; }
		</style>
	</head>
	<body>
		<h1>Title</h1>
		<p class="highlight">Highlighted text</p>
	</body>
	</html>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) < 2 {
		t.Errorf("expected at least 2 content nodes, got %d", len(page.Content))
	}
}

func TestFromHTML_MixedInlineAndBlock(t *testing.T) {
	// Mixed inline text and block elements inside a div
	html := `<div>Some text before<p>A paragraph</p>Some text after</div>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}

	// The div should contain: anonymous block (text), p block, anonymous block (text)
	box, ok := page.Content[0].(*document.Box)
	if !ok {
		t.Fatalf("expected Box, got %T", page.Content[0])
	}
	if len(box.Content) < 3 {
		t.Errorf("expected at least 3 children (anon-block + p + anon-block), got %d", len(box.Content))
	}
}

// ─── Config / Option tests ───────────────────────────────────────

func TestWithPageSize(t *testing.T) {
	result, err := FromHTML("<p>test</p>", WithPageSize(document.Letter))
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	if result.config.PageSize != document.Letter {
		t.Errorf("expected Letter page size, got %v", result.config.PageSize)
	}
}

func TestWithMargins(t *testing.T) {
	margins := document.UniformEdges(document.Mm(10))
	result, err := FromHTML("<p>test</p>", WithMargins(margins))
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	if result.config.Margins != margins {
		t.Errorf("margins mismatch")
	}
}

func TestWithStylesheet(t *testing.T) {
	result, err := FromHTML("<p>test</p>", WithStylesheet("p { color: red; }"))
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	if result.config.ExtraCSS != "p { color: red; }" {
		t.Errorf("ExtraCSS mismatch: %q", result.config.ExtraCSS)
	}
}

func TestWithDefaultFont(t *testing.T) {
	result, err := FromHTML("<p>test</p>", WithDefaultFont("Helvetica", 14))
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	if result.config.DefaultFont != "Helvetica" {
		t.Errorf("DefaultFont: got %q, want %q", result.config.DefaultFont, "Helvetica")
	}
	if result.config.DefaultSize != 14 {
		t.Errorf("DefaultSize: got %v, want 14", result.config.DefaultSize)
	}
}

// ─── Render pipeline tests ──────────────────────────────────────

func TestResult_Bytes_NoFonts(t *testing.T) {
	// Without fonts, the render should still produce valid PDF bytes
	// (using placeholder font metrics)
	result, err := FromHTML("<p>Hello, World!</p>")
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	data, err := result.Bytes()
	if err != nil {
		t.Fatalf("Bytes() failed: %v", err)
	}

	// Check PDF header
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("output does not start with %PDF- header")
	}

	// Check PDF trailer
	if !bytes.HasSuffix(bytes.TrimRight(data, "\n\r "), []byte("%%EOF")) {
		t.Error("output does not end with PDF EOF trailer")
	}
}

func TestResult_Write(t *testing.T) {
	result, err := FromHTML("<p>test</p>")
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	var buf bytes.Buffer
	if err := result.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Write produced empty output")
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("output does not start with %PDF- header")
	}
}

// ─── Font resolver tests ────────────────────────────────────────

func TestFontResolver_NoFonts(t *testing.T) {
	fr := &htmlFontResolver{
		fonts:   make(map[string]*font.TrueTypeFont),
		metrics: make(map[string]layout.FontMetrics),
	}

	resolved := fr.Resolve("Arial", document.WeightNormal, false)
	if resolved.ID != "default" {
		t.Errorf("expected 'default' ID for no-font resolver, got %q", resolved.ID)
	}

	width := fr.MeasureString(resolved, "Hello", 12)
	// 5 chars * 12 * 0.5 = 30
	expected := 30.0
	if width != expected {
		t.Errorf("MeasureString: got %v, want %v", width, expected)
	}
}

func TestFontResolver_Fallback(t *testing.T) {
	fr := &htmlFontResolver{
		fonts:   make(map[string]*font.TrueTypeFont),
		metrics: make(map[string]layout.FontMetrics),
	}

	lines := fr.LineBreak(layout.ResolvedFont{ID: "unknown"}, "Hello World", 12, 30)
	if len(lines) == 0 {
		t.Error("LineBreak returned no lines")
	}
}

// ─── Simple line break tests ────────────────────────────────────

func TestSimpleLineBreak(t *testing.T) {
	tests := []struct {
		text     string
		size     float64
		maxWidth float64
		minLines int
	}{
		{"Hello", 10, 100, 1},
		{"Hello World This Is A Test", 10, 25, 3},
		{"Short", 10, 0, 1},   // maxWidth <= 0 returns single line
		{"X", 10, 1, 1},       // very narrow
	}

	for _, tt := range tests {
		lines := simpleLineBreak(tt.text, tt.size, tt.maxWidth)
		if len(lines) < tt.minLines {
			t.Errorf("simpleLineBreak(%q, %v, %v): got %d lines, want at least %d",
				tt.text, tt.size, tt.maxWidth, len(lines), tt.minLines)
		}
		// All text should be preserved
		joined := strings.Join(lines, "")
		if joined != tt.text {
			t.Errorf("simpleLineBreak: text not preserved: %q vs %q", joined, tt.text)
		}
	}
}

// ─── Edge cases ─────────────────────────────────────────────────

func TestFromHTML_EmptyBody(t *testing.T) {
	result, err := FromHTML("<html><body></body></html>")
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	if result.doc == nil {
		t.Fatal("document is nil")
	}
}

func TestFromHTML_TextOnly(t *testing.T) {
	result, err := FromHTML("Just plain text")
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}
	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Error("expected content for plain text")
	}
}

func TestFromHTML_ComplexDocument(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<style>
		body { font-size: 12pt; }
		h1 { color: #333; margin-bottom: 20pt; }
		.info { background-color: #f0f0f0; padding: 10pt; }
		.highlight { font-weight: bold; color: red; }
		</style>
	</head>
	<body>
		<h1>Invoice #12345</h1>
		<div class="info">
			<p>Date: 2026-03-02</p>
			<p>Customer: <strong>John Doe</strong></p>
		</div>
		<hr>
		<h2>Items</h2>
		<ul>
			<li>Widget A — $10.00</li>
			<li>Widget B — $20.00</li>
			<li>Widget C — $30.00</li>
		</ul>
		<p class="highlight">Total: $60.00</p>
		<blockquote>Thank you for your business!</blockquote>
	</body>
	</html>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	// Generate PDF bytes
	data, err := result.Bytes()
	if err != nil {
		t.Fatalf("Bytes() failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("generated PDF is empty")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("not a valid PDF")
	}
}

func TestFromHTML_BRElement(t *testing.T) {
	html := `<p>Line one<br>Line two</p>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) == 0 {
		t.Fatal("no content nodes")
	}
}

func TestFromHTML_CJKText(t *testing.T) {
	html := `<p>日本語テキスト</p><p>中文文本</p><p>한국어 텍스트</p>`

	result, err := FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	page := result.doc.Pages[0]
	if len(page.Content) < 3 {
		t.Errorf("expected 3 paragraphs, got %d", len(page.Content))
	}
}
