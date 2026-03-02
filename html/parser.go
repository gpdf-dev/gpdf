package html

import (
	"fmt"
	"io"
	"strings"
)

// Default limits for parser safety.
const (
	DefaultMaxDepth     = 256
	DefaultMaxInputSize = 50 * 1024 * 1024 // 50MB
	DefaultMaxNodeCount = 1_000_000
)

// voidElements are HTML elements that cannot have children.
// These are automatically self-closed if they appear as start tags.
var voidElements = map[string]bool{
	"area":    true,
	"base":    true,
	"br":      true,
	"col":     true,
	"embed":   true,
	"hr":      true,
	"img":     true,
	"input":   true,
	"link":    true,
	"meta":    true,
	"param":   true,
	"source":  true,
	"track":   true,
	"wbr":     true,
}

// implicitCloseMap defines which opening tags implicitly close a currently
// open tag of the given type. For example, a <p> inside another <p> closes
// the outer <p>.
var implicitCloseMap = map[string]map[string]bool{
	"p": {
		"p": true, "div": true, "h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true, "ul": true, "ol": true,
		"dl": true, "table": true, "blockquote": true, "hr": true,
		"pre": true, "form": true, "fieldset": true, "address": true,
		"section": true, "article": true, "aside": true, "header": true,
		"footer": true, "nav": true, "figure": true, "figcaption": true,
		"main": true, "details": true, "summary": true,
	},
	"li": {
		"li": true,
	},
	"dt": {
		"dt": true, "dd": true,
	},
	"dd": {
		"dt": true, "dd": true,
	},
	"th": {
		"th": true, "td": true,
	},
	"td": {
		"th": true, "td": true,
	},
	"tr": {
		"tr": true,
	},
	"thead": {
		"tbody": true, "tfoot": true,
	},
	"tbody": {
		"tbody": true, "tfoot": true,
	},
	"tfoot": {
		"tbody": true,
	},
	"option": {
		"option": true,
	},
	"optgroup": {
		"optgroup": true,
	},
}

// rawTextElements contain raw text (not parsed for HTML tags).
var rawTextElements = map[string]bool{
	"script":   true,
	"style":    true,
	"textarea": true,
	"title":    true,
}

// ParserOption configures the HTML parser.
type ParserOption func(*parser)

// WithMaxDepth sets the maximum nesting depth.
func WithMaxDepth(depth int) ParserOption {
	return func(p *parser) {
		p.maxDepth = depth
	}
}

// WithMaxInputSize sets the maximum input size in bytes.
func WithMaxInputSize(size int) ParserOption {
	return func(p *parser) {
		p.maxInputSize = size
	}
}

// WithMaxNodeCount sets the maximum number of DOM nodes.
func WithMaxNodeCount(count int) ParserOption {
	return func(p *parser) {
		p.maxNodeCount = count
	}
}

// Parse parses HTML from an io.Reader and returns the root DOM node.
func Parse(r io.Reader, opts ...ParserOption) (*Node, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("html: read error: %w", err)
	}
	return ParseBytes(data, opts...)
}

// ParseString parses an HTML string and returns the root DOM node.
func ParseString(s string, opts ...ParserOption) (*Node, error) {
	return ParseBytes([]byte(s), opts...)
}

// ParseBytes parses HTML from a byte slice and returns the root DOM node.
func ParseBytes(data []byte, opts ...ParserOption) (*Node, error) {
	p := &parser{
		maxDepth:     DefaultMaxDepth,
		maxInputSize: DefaultMaxInputSize,
		maxNodeCount: DefaultMaxNodeCount,
	}
	for _, opt := range opts {
		opt(p)
	}

	if len(data) > p.maxInputSize {
		return nil, fmt.Errorf("html: input size %d exceeds maximum %d", len(data), p.maxInputSize)
	}

	p.tok = newTokenizer(data)
	return p.parse()
}

// parser builds a DOM tree from HTML tokens.
type parser struct {
	tok          *tokenizer
	maxDepth     int
	maxInputSize int
	maxNodeCount int
	nodeCount    int
	openStack    []*Node // stack of currently open elements
}

// parse runs the parser and returns the root node.
func (p *parser) parse() (*Node, error) {
	root := &Node{Type: DocumentNode}
	p.nodeCount++

	// Start with root as the only open element
	p.openStack = []*Node{root}

	for {
		tok := p.tok.next()
		switch tok.typ {
		case tokenEOF:
			return root, nil

		case tokenError:
			return root, fmt.Errorf("html: tokenizer error: %s", tok.data)

		case tokenText:
			if err := p.handleText(tok); err != nil {
				return root, err
			}

		case tokenStartTag:
			if err := p.handleStartTag(tok); err != nil {
				return root, err
			}

		case tokenSelfClose:
			if err := p.handleSelfCloseTag(tok); err != nil {
				return root, err
			}

		case tokenEndTag:
			p.handleEndTag(tok)

		case tokenComment:
			if err := p.handleComment(tok); err != nil {
				return root, err
			}

		case tokenDoctype:
			// DOCTYPE is noted but does not create a DOM node
		}
	}
}

// current returns the current parent node (top of the open stack).
func (p *parser) current() *Node {
	return p.openStack[len(p.openStack)-1]
}

// depth returns the current nesting depth.
func (p *parser) depth() int {
	return len(p.openStack) - 1 // exclude root
}

// allocNode creates a new node, checking the node count limit.
func (p *parser) allocNode() (*Node, error) {
	p.nodeCount++
	if p.nodeCount > p.maxNodeCount {
		return nil, fmt.Errorf("html: node count exceeds maximum %d", p.maxNodeCount)
	}
	return &Node{}, nil
}

// handleText appends a text node to the current parent.
func (p *parser) handleText(tok token) error {
	// Skip empty text nodes
	if tok.data == "" {
		return nil
	}
	node, err := p.allocNode()
	if err != nil {
		return err
	}
	node.Type = TextNode
	node.Content = tok.data
	p.current().AppendChild(node)
	return nil
}

// handleComment appends a comment node to the current parent.
func (p *parser) handleComment(tok token) error {
	node, err := p.allocNode()
	if err != nil {
		return err
	}
	node.Type = CommentNode
	node.Content = tok.data
	p.current().AppendChild(node)
	return nil
}

// handleStartTag processes an opening tag.
func (p *parser) handleStartTag(tok token) error {
	tag := tok.tag

	// Check for implicit closure of the current element
	p.implicitClose(tag)

	// Check depth limit
	if p.depth() >= p.maxDepth {
		return fmt.Errorf("html: nesting depth exceeds maximum %d at <%s>", p.maxDepth, tag)
	}

	node, err := p.allocNode()
	if err != nil {
		return err
	}
	node.Type = ElementNode
	node.Tag = tag
	node.Attrs = tok.attrs

	p.current().AppendChild(node)

	// Void elements are never pushed onto the stack
	if voidElements[tag] {
		return nil
	}

	// Raw text elements — consume all content until the matching end tag
	if rawTextElements[tag] {
		p.openStack = append(p.openStack, node)
		p.consumeRawText(tag)
		return nil
	}

	// Push onto the open stack
	p.openStack = append(p.openStack, node)
	return nil
}

// handleSelfCloseTag processes a self-closing tag (<br/>, <img .../>).
func (p *parser) handleSelfCloseTag(tok token) error {
	node, err := p.allocNode()
	if err != nil {
		return err
	}
	node.Type = ElementNode
	node.Tag = tok.tag
	node.Attrs = tok.attrs
	p.current().AppendChild(node)
	// Self-closing tags are never pushed onto the stack
	return nil
}

// handleEndTag processes a closing tag.
func (p *parser) handleEndTag(tok token) {
	tag := tok.tag

	// Find the matching open element in the stack (search from top)
	idx := -1
	for i := len(p.openStack) - 1; i >= 1; i-- {
		if p.openStack[i].Tag == tag {
			idx = i
			break
		}
	}

	if idx < 0 {
		// No matching open tag — ignore this end tag
		return
	}

	// Pop all elements from idx to top (cross-nesting recovery)
	p.openStack = p.openStack[:idx]
}

// implicitClose checks if the current open element should be implicitly
// closed by the arrival of a new opening tag.
func (p *parser) implicitClose(newTag string) {
	for {
		cur := p.current()
		if cur.Type != ElementNode {
			return
		}
		closers, ok := implicitCloseMap[cur.Tag]
		if !ok {
			return
		}
		if !closers[newTag] {
			return
		}
		// Close the current element
		p.openStack = p.openStack[:len(p.openStack)-1]
	}
}

// consumeRawText reads raw text content for elements like <script>, <style>.
// It reads until the matching end tag is found.
func (p *parser) consumeRawText(tag string) {
	// Build the end tag pattern to search for
	endTag := "</" + tag
	var sb strings.Builder
	input := p.tok.input
	pos := p.tok.pos
	size := p.tok.size

	for pos < size {
		// Look for potential end tag
		if input[pos] == '<' && pos+1 < size && input[pos+1] == '/' {
			// Check if this is our end tag
			remaining := strings.ToLower(string(input[pos:]))
			if strings.HasPrefix(remaining, endTag) {
				checkPos := pos + len(endTag)
				// Must be followed by '>' or whitespace then '>'
				for checkPos < size && isASCIISpace(input[checkPos]) {
					checkPos++
				}
				if checkPos < size && input[checkPos] == '>' {
					// Found the end tag
					content := sb.String()
					if content != "" {
						textNode := &Node{
							Type:    TextNode,
							Content: content,
						}
						p.current().AppendChild(textNode)
						p.nodeCount++
					}
					p.tok.pos = checkPos + 1
					// Pop the raw text element
					p.openStack = p.openStack[:len(p.openStack)-1]
					return
				}
			}
		}
		sb.WriteByte(input[pos])
		pos++
	}

	// Unterminated — consume everything as text
	content := sb.String()
	if content != "" {
		textNode := &Node{
			Type:    TextNode,
			Content: content,
		}
		p.current().AppendChild(textNode)
		p.nodeCount++
	}
	p.tok.pos = size
	p.openStack = p.openStack[:len(p.openStack)-1]
}
