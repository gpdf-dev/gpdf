package htmlpdf

import (
	"strings"

	"github.com/gpdf-dev/gpdf/css"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/html"
)

// nodeAdapter adapts html.Node to the css.Matchable interface.
type nodeAdapter struct {
	node *html.Node
}

func (a *nodeAdapter) NodeTag() string { return a.node.Tag }
func (a *nodeAdapter) NodeID() string  { return a.node.ID() }
func (a *nodeAdapter) NodeClasses() []string {
	classAttr, ok := a.node.Attr("class")
	if !ok {
		return nil
	}
	return strings.Fields(classAttr)
}
func (a *nodeAdapter) NodeAttr(name string) (string, bool) { return a.node.Attr(name) }
func (a *nodeAdapter) NodeParent() css.Matchable {
	if a.node.Parent == nil {
		return nil
	}
	return &nodeAdapter{node: a.node.Parent}
}

// UA stylesheet (built-in browser defaults for PDF rendering).
const uaCSS = `
html, body          { display: block; }
h1                  { font-size: 24pt; font-weight: bold; margin-top: 8pt; margin-bottom: 8pt; }
h2                  { font-size: 18pt; font-weight: bold; margin-top: 6pt; margin-bottom: 6pt; }
h3                  { font-size: 14pt; font-weight: bold; margin-top: 5pt; margin-bottom: 5pt; }
h4                  { font-weight: bold; margin-top: 5pt; margin-bottom: 5pt; }
h5                  { font-size: 10pt; font-weight: bold; margin-top: 4pt; margin-bottom: 4pt; }
h6                  { font-size: 8pt; font-weight: bold; margin-top: 4pt; margin-bottom: 4pt; }
p                   { display: block; margin-top: 12pt; margin-bottom: 12pt; }
div                 { display: block; }
strong, b           { font-weight: bold; }
em, i               { font-style: italic; }
u                   { text-decoration: underline; }
s, del, strike      { text-decoration: line-through; }
hr                  { margin-top: 6pt; margin-bottom: 6pt; }
ul, ol              { display: block; margin-top: 12pt; margin-bottom: 12pt; padding-left: 40pt; }
li                  { display: list-item; }
blockquote          { display: block; margin-top: 12pt; margin-bottom: 12pt; margin-left: 40pt; }
pre                 { display: block; margin-top: 12pt; margin-bottom: 12pt; }
code                { font-family: monospace; }
table               { display: block; }
`

// converter handles the HTML→DocumentNode conversion.
type converter struct {
	config       *Config
	uaStylesheet *css.Stylesheet
	stylesheets  []*css.Stylesheet
	styleCache   map[*html.Node]css.ComputedStyles
}

// newConverter creates a new converter with parsed stylesheets.
func newConverter(config *Config, htmlRoot *html.Node) *converter {
	ua, _ := css.ParseStylesheet(uaCSS)

	// Collect <style> elements from the document
	var stylesheets []*css.Stylesheet
	for _, styleNode := range htmlRoot.StyleElements() {
		cssText := styleNode.TextContent()
		if ss, err := css.ParseStylesheet(cssText); err == nil {
			stylesheets = append(stylesheets, ss)
		}
	}

	// Add user-provided additional CSS
	if config.ExtraCSS != "" {
		if ss, err := css.ParseStylesheet(config.ExtraCSS); err == nil {
			stylesheets = append(stylesheets, ss)
		}
	}

	return &converter{
		config:       config,
		uaStylesheet: ua,
		stylesheets:  stylesheets,
		styleCache:   make(map[*html.Node]css.ComputedStyles),
	}
}

// convert transforms the HTML DOM tree into a document.Document.
func (c *converter) convert(root *html.Node) *document.Document {
	// Find <body> or use root
	body := findBody(root)
	if body == nil {
		body = root
	}

	// Build the default style
	defaultStyle := c.config.defaultStyle()

	// Compute styles for body
	bodyComputed := c.computeStyles(body, nil)

	// Convert body's children to DocumentNode tree
	children := c.convertChildren(body, bodyComputed)

	// Wrap in a page
	page := &document.Page{
		Size:    c.config.PageSize,
		Margins: c.config.Margins,
		Content: children,
		PageStyle: document.Style{
			FontFamily: defaultStyle.FontFamily,
			FontSize:   defaultStyle.FontSize,
			LineHeight: defaultStyle.FontSize * 1.2,
			Color:      defaultStyle.Color,
		},
	}

	return &document.Document{
		Pages:        []*document.Page{page},
		DefaultStyle: defaultStyle,
	}
}

// convertChildren converts child nodes into DocumentNode slices,
// handling inline/block mixing and anonymous block boxes.
func (c *converter) convertChildren(parent *html.Node, parentComputed css.ComputedStyles) []document.DocumentNode {
	var result []document.DocumentNode
	var inlineBuffer []*html.Node

	flushInline := func() {
		if len(inlineBuffer) == 0 {
			return
		}
		// Collect inline content as RichText fragments
		var fragments []document.RichTextFragment
		for _, inlineNode := range inlineBuffer {
			computed := c.computeStyles(inlineNode, parentComputed)
			frags := collectInlineFragments(inlineNode, computed, func(n *html.Node) css.ComputedStyles {
				return c.computeStyles(n, computed)
			})
			fragments = append(fragments, frags...)
		}
		if len(fragments) > 0 {
			blockStyle := applyStyle(parentComputed)
			rt := wrapInlineAsRichText(fragments, blockStyle)
			if rt != nil {
				result = append(result, rt)
			}
		}
		inlineBuffer = nil
	}

	for _, child := range parent.Children {
		// Skip comments
		if child.Type == html.CommentNode {
			continue
		}

		computed := c.computeStyles(child, parentComputed)

		// Skip display: none
		if child.Type == html.ElementNode && isDisplayNone(computed) {
			continue
		}

		if isInlineElement(child, computed) {
			inlineBuffer = append(inlineBuffer, child)
		} else {
			// Block element encountered — flush any pending inline content
			flushInline()
			// Recursively convert the block element
			blockNode := c.convertElement(child, computed)
			if blockNode != nil {
				result = append(result, blockNode)
			}
		}
	}

	// Flush remaining inline content
	flushInline()

	return result
}

// convertElement converts a single block-level HTML element to a DocumentNode.
func (c *converter) convertElement(node *html.Node, computed css.ComputedStyles) document.DocumentNode {
	if node.Type == html.TextNode {
		// Standalone text (shouldn't happen for block, but handle gracefully)
		style := applyStyle(computed)
		return &document.Text{
			Content:   node.Content,
			TextStyle: style,
		}
	}

	if node.Type != html.ElementNode {
		return nil
	}

	// Special handling for certain elements
	switch node.Tag {
	case "br", "wbr":
		return nil // handled inline
	case "img":
		return mapImage(node, applyStyle(computed))
	case "hr":
		return mapHR(computed)
	}

	// Recurse into children
	children := c.convertChildren(node, computed)

	return mapBlockElement(node, computed, children)
}

// computeStyles computes CSS cascade for a node.
func (c *converter) computeStyles(node *html.Node, parentComputed css.ComputedStyles) css.ComputedStyles {
	if cached, ok := c.styleCache[node]; ok {
		return cached
	}

	var inlineDecls []css.Declaration
	if node.Type == html.ElementNode {
		if styleAttr, ok := node.Attr("style"); ok {
			inlineDecls = css.ParseDeclarations(styleAttr)
		}
	}

	matchable := &nodeAdapter{node: node}
	computed := css.Cascade(matchable, c.stylesheets, inlineDecls, parentComputed, c.uaStylesheet)

	// Resolve computed values
	parentFontSize := 12.0
	containerWidth := c.config.PageSize.Width - c.config.Margins.Left.Resolve(0, 0) - c.config.Margins.Right.Resolve(0, 0)
	if parentComputed != nil {
		parentFontSize = css.ParsePt(parentComputed["font-size"], 12)
	}
	computed = css.ResolveComputed(computed, parentFontSize, containerWidth)

	c.styleCache[node] = computed
	return computed
}

// findBody finds the <body> element in the DOM tree.
func findBody(root *html.Node) *html.Node {
	var body *html.Node
	root.Walk(func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Tag == "body" {
			body = n
			return false
		}
		return true
	})
	return body
}
