package htmlpdf

import (
	"strings"

	"github.com/gpdf-dev/gpdf/css"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/html"
	"github.com/gpdf-dev/gpdf/pdf"
)

// isInlineTag reports whether the given HTML tag is inline by default.
var inlineTags = map[string]bool{
	"span": true, "strong": true, "b": true, "em": true, "i": true,
	"u": true, "s": true, "del": true, "strike": true,
	"code": true, "a": true, "sup": true, "sub": true,
	"small": true, "mark": true, "abbr": true, "cite": true,
	"q": true, "dfn": true, "time": true, "var": true, "samp": true,
	"kbd": true, "data": true, "br": true, "wbr": true,
}

// isBlockDisplay returns true if the computed display value is block-like.
func isBlockDisplay(computed css.ComputedStyles) bool {
	display := strings.ToLower(strings.TrimSpace(computed["display"]))
	switch display {
	case "block", "list-item", "table", "table-row", "table-cell",
		"table-header-group", "table-row-group", "table-footer-group":
		return true
	case "inline", "inline-block":
		return false
	case "none":
		return false // will be filtered out
	}
	// Default to block for unknown values
	return true
}

// isDisplayNone returns true if display: none.
func isDisplayNone(computed css.ComputedStyles) bool {
	return strings.ToLower(strings.TrimSpace(computed["display"])) == "none"
}

// isInlineElement checks if an HTML node should be treated as inline.
func isInlineElement(node *html.Node, computed css.ComputedStyles) bool {
	// Text nodes are always inline
	if node.Type == html.TextNode {
		return true
	}
	if node.Type != html.ElementNode {
		return false
	}

	// Check explicit display property
	if display, ok := computed["display"]; ok {
		d := strings.ToLower(strings.TrimSpace(display))
		if d == "inline" || d == "inline-block" {
			return true
		}
		if d == "block" || d == "list-item" {
			return false
		}
	}

	// Fall back to HTML default
	return inlineTags[node.Tag]
}

// mapBlockElement maps a block-level HTML element to a DocumentNode.
func mapBlockElement(node *html.Node, computed css.ComputedStyles, children []document.DocumentNode) document.DocumentNode {
	style := applyStyle(computed)
	boxStyle := applyBoxStyle(computed)

	switch node.Tag {
	case "ul":
		return mapList(node, children, style, document.Unordered)
	case "ol":
		return mapList(node, children, style, document.Ordered)
	case "hr":
		return mapHR(computed)
	case "img":
		return mapImage(node, style)
	default:
		return &document.Box{
			Content:  children,
			BoxStyle: boxStyle,
		}
	}
}

// mapList creates a document.List from list item children.
func mapList(node *html.Node, children []document.DocumentNode, style document.Style, listType document.ListType) document.DocumentNode {
	var items []document.ListItem
	for _, child := range children {
		switch c := child.(type) {
		case *document.Box:
			items = append(items, document.ListItem{
				Content:   c.Content,
				ItemStyle: c.Style(),
			})
		default:
			items = append(items, document.ListItem{
				Content:   []document.DocumentNode{child},
				ItemStyle: style,
			})
		}
	}
	if len(items) == 0 {
		items = append(items, document.ListItem{ItemStyle: style})
	}
	return &document.List{
		Items:     items,
		ListType:  listType,
		ListStyle: style,
	}
}

// mapHR creates a horizontal rule as a Box with a top border.
func mapHR(computed css.ComputedStyles) document.DocumentNode {
	style := applyBoxStyle(computed)
	// If no explicit border, add a default top border
	if style.Border.Top.Style == document.BorderNone {
		style.Border.Top = document.BorderSide{
			Width: document.Pt(1),
			Style: document.BorderSolid,
			Color: pdf.Black,
		}
	}
	return &document.Box{
		BoxStyle: style,
	}
}

// mapImage creates a document.Image from an <img> element.
func mapImage(node *html.Node, style document.Style) document.DocumentNode {
	src, _ := node.Attr("src")
	_ = src // Image data loading will be handled by the converter with WithBaseURL

	img := &document.Image{
		ImgStyle: style,
		FitMode:  document.FitContain,
	}

	// Parse width/height attributes
	if w, ok := node.Attr("width"); ok {
		img.DisplayWidth = parseCSSValue(w + "px")
	}
	if h, ok := node.Attr("height"); ok {
		img.DisplayHeight = parseCSSValue(h + "px")
	}

	return img
}

// collectInlineFragments collects inline content as RichText fragments.
func collectInlineFragments(node *html.Node, computed css.ComputedStyles, cascade cascadeFn) []document.RichTextFragment {
	if node.Type == html.TextNode {
		content := node.Content
		if content == "" {
			return nil
		}
		style := applyStyle(computed)
		return []document.RichTextFragment{
			{Content: content, FragmentStyle: style},
		}
	}

	if node.Type != html.ElementNode {
		return nil
	}

	// Handle <br> as newline
	if node.Tag == "br" {
		return []document.RichTextFragment{
			{Content: "\n", FragmentStyle: applyStyle(computed)},
		}
	}

	// Collect fragments from children
	var fragments []document.RichTextFragment
	for _, child := range node.Children {
		childComputed := cascade(child)
		frags := collectInlineFragments(child, childComputed, cascade)
		fragments = append(fragments, frags...)
	}
	return fragments
}

// cascadeFn is a function type for computing CSS styles for a node.
type cascadeFn func(*html.Node) css.ComputedStyles

// wrapInlineAsRichText wraps inline fragments into a RichText node.
func wrapInlineAsRichText(fragments []document.RichTextFragment, blockStyle document.Style) document.DocumentNode {
	if len(fragments) == 0 {
		return nil
	}
	return &document.RichText{
		Fragments:  fragments,
		BlockStyle: blockStyle,
	}
}
