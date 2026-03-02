// Package html provides a zero-dependency HTML tokenizer and parser
// for converting HTML documents into a DOM tree structure.
//
// The parser operates in a tolerant mode that handles common HTML
// patterns without requiring full HTML5 specification compliance.
// It is designed for use with the gpdf PDF generation library.
package html

import "strings"

// NodeType represents the type of an HTML DOM node.
type NodeType int

const (
	// ElementNode represents an HTML element (e.g., <div>, <p>).
	ElementNode NodeType = iota
	// TextNode represents a text content node.
	TextNode
	// CommentNode represents an HTML comment (<!-- ... -->).
	CommentNode
	// DocumentNode represents the root document node.
	DocumentNode
)

// Node represents a node in the HTML DOM tree.
type Node struct {
	Type     NodeType
	Tag      string // Lowercase tag name for ElementNode.
	Attrs    map[string]string
	Children []*Node
	Content  string // Text content for TextNode and CommentNode.
	Parent   *Node  // Parent node (nil for root).
}

// AppendChild adds a child node to this node.
func (n *Node) AppendChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// RemoveChild removes a child node from this node.
func (n *Node) RemoveChild(child *Node) {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return
		}
	}
}

// Attr returns the value of the named attribute and whether it exists.
func (n *Node) Attr(name string) (string, bool) {
	if n.Attrs == nil {
		return "", false
	}
	v, ok := n.Attrs[strings.ToLower(name)]
	return v, ok
}

// SetAttr sets the value of the named attribute.
func (n *Node) SetAttr(name, value string) {
	if n.Attrs == nil {
		n.Attrs = make(map[string]string)
	}
	n.Attrs[strings.ToLower(name)] = value
}

// HasClass reports whether the node has the given CSS class.
func (n *Node) HasClass(class string) bool {
	classAttr, ok := n.Attr("class")
	if !ok {
		return false
	}
	for _, c := range strings.Fields(classAttr) {
		if c == class {
			return true
		}
	}
	return false
}

// ID returns the id attribute of the node, or empty string if not set.
func (n *Node) ID() string {
	id, _ := n.Attr("id")
	return id
}

// FirstChild returns the first child node, or nil if there are no children.
func (n *Node) FirstChild() *Node {
	if len(n.Children) == 0 {
		return nil
	}
	return n.Children[0]
}

// LastChild returns the last child node, or nil if there are no children.
func (n *Node) LastChild() *Node {
	if len(n.Children) == 0 {
		return nil
	}
	return n.Children[len(n.Children)-1]
}

// TextContent returns the concatenated text content of this node
// and all its descendants.
func (n *Node) TextContent() string {
	if n.Type == TextNode {
		return n.Content
	}
	var sb strings.Builder
	for _, child := range n.Children {
		sb.WriteString(child.TextContent())
	}
	return sb.String()
}

// StyleElements collects all <style> elements from the tree in document order.
func (n *Node) StyleElements() []*Node {
	var styles []*Node
	n.walkStyles(&styles)
	return styles
}

func (n *Node) walkStyles(styles *[]*Node) {
	if n.Type == ElementNode && n.Tag == "style" {
		*styles = append(*styles, n)
	}
	for _, child := range n.Children {
		child.walkStyles(styles)
	}
}

// Walk traverses the tree in depth-first order, calling fn for each node.
// If fn returns false, the subtree rooted at that node is skipped.
func (n *Node) Walk(fn func(*Node) bool) {
	if !fn(n) {
		return
	}
	for _, child := range n.Children {
		child.Walk(fn)
	}
}
