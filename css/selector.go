package css

import (
	"strings"
)

// Matchable represents a DOM node that can be matched against CSS selectors.
// This interface allows the css package to remain independent of the html package.
type Matchable interface {
	NodeTag() string
	NodeID() string
	NodeClasses() []string
	NodeAttr(name string) (string, bool)
	NodeParent() Matchable
}

// Selector represents a parsed CSS selector (a single selector, not a group).
// Parts are stored in document order (left to right as written in CSS).
// Parts[len-1] is the subject (the element that the selector matches).
type Selector struct {
	Parts []CompoundSelector
	Spec  Specificity
}

// AttrMatcher specifies how an attribute selector matches.
type AttrMatcher int

const (
	AttrExists   AttrMatcher = iota // [attr]
	AttrEquals                      // [attr="val"]
	AttrIncludes                    // [attr~="val"] (whitespace-separated word)
	AttrDashMatch                   // [attr|="val"] (exact or prefix followed by '-')
	AttrPrefix                      // [attr^="val"]
	AttrSuffix                      // [attr$="val"]
	AttrSubstring                   // [attr*="val"]
)

// AttrSelector represents a single attribute selector like [href="..."].
type AttrSelector struct {
	Name    string      // attribute name (lowercase)
	Matcher AttrMatcher // match type
	Value   string      // value to match against (empty for AttrExists)
}

// CompoundSelector represents a compound selector like "div.main#app".
type CompoundSelector struct {
	Tag        string         // element type (lowercase), "" means any
	ID         string         // #id value, "" means no ID constraint
	Classes    []string       // .class values
	Attrs      []AttrSelector // attribute selectors
	Universal  bool           // true if '*' was explicitly specified
	Combinator Combinator
}

// Combinator represents the relationship between adjacent compound selectors.
type Combinator int

const (
	CombinatorNone       Combinator = iota // no combinator (for the leftmost part)
	CombinatorDescendant                   // ' ' (whitespace)
	CombinatorChild                        // '>'
)

// ParseSelectorGroup parses a comma-separated group of selectors.
// e.g., "h1, h2, h3" → 3 selectors.
func ParseSelectorGroup(raw string) []Selector {
	parts := splitSelectorGroup(raw)
	var selectors []Selector
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		sel, ok := ParseSelector(part)
		if ok {
			selectors = append(selectors, sel)
		}
	}
	return selectors
}

// ParseSelector parses a single CSS selector string.
func ParseSelector(raw string) (Selector, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Selector{}, false
	}

	tok := NewTokenizer([]byte(raw))
	var parts []CompoundSelector
	combinator := CombinatorNone

	for {
		// Skip leading whitespace (but note it as potential descendant combinator)
		hadWhitespace := false
		for tok.Peek().Type == TokenWhitespace {
			tok.Next()
			hadWhitespace = true
		}

		if tok.Peek().Type == TokenEOF {
			break
		}

		// Check for explicit combinator
		peek := tok.Peek()
		if peek.Type == TokenDelim && peek.Value == ">" {
			tok.Next()
			combinator = CombinatorChild
			// Skip whitespace after combinator
			for tok.Peek().Type == TokenWhitespace {
				tok.Next()
			}
		} else if hadWhitespace && len(parts) > 0 {
			combinator = CombinatorDescendant
		}

		// Parse compound selector
		compound, ok := parseCompound(tok)
		if !ok {
			break
		}
		compound.Combinator = combinator
		parts = append(parts, compound)
		combinator = CombinatorNone
	}

	if len(parts) == 0 {
		return Selector{}, false
	}

	sel := Selector{Parts: parts}
	sel.Spec = calcSelectorSpecificity(parts)
	return sel, true
}

// parseCompound parses a single compound selector (e.g., "div.main#app").
func parseCompound(tok *Tokenizer) (CompoundSelector, bool) {
	var cs CompoundSelector
	matched := false

	for {
		peek := tok.Peek()
		switch {
		case peek.Type == TokenIdent:
			// Type selector: div, p, h1, etc.
			t := tok.Next()
			cs.Tag = strings.ToLower(t.Value)
			matched = true

		case peek.Type == TokenDelim && peek.Value == ".":
			// Class selector
			tok.Next() // consume '.'
			next := tok.Peek()
			if next.Type == TokenIdent {
				t := tok.Next()
				cs.Classes = append(cs.Classes, t.Value)
				matched = true
			}

		case peek.Type == TokenHash:
			// ID selector
			t := tok.Next()
			cs.ID = t.Value
			matched = true

		case peek.Type == TokenDelim && peek.Value == "*":
			// Universal selector
			tok.Next()
			cs.Universal = true
			matched = true

		case peek.Type == TokenOpenBracket:
			// Attribute selector [attr], [attr="val"], etc.
			tok.Next() // consume '['
			attr, ok := parseAttrSelector(tok)
			if ok {
				cs.Attrs = append(cs.Attrs, attr)
				matched = true
			}

		default:
			return cs, matched
		}
	}
}

// parseAttrSelector parses the contents inside [...] and consumes the closing ']'.
func parseAttrSelector(tok *Tokenizer) (AttrSelector, bool) {
	// Skip whitespace
	for tok.Peek().Type == TokenWhitespace {
		tok.Next()
	}

	// Expect attribute name (ident)
	if tok.Peek().Type != TokenIdent {
		// Skip to closing bracket
		skipToCloseBracket(tok)
		return AttrSelector{}, false
	}
	name := strings.ToLower(tok.Next().Value)

	// Skip whitespace
	for tok.Peek().Type == TokenWhitespace {
		tok.Next()
	}

	// Check for closing bracket (existence selector) or matcher
	if tok.Peek().Type == TokenCloseBracket {
		tok.Next() // consume ']'
		return AttrSelector{Name: name, Matcher: AttrExists}, true
	}

	// Determine matcher type
	var matcher AttrMatcher
	peek := tok.Peek()

	switch {
	case peek.Type == TokenDelim && peek.Value == "=":
		tok.Next()
		matcher = AttrEquals
	case peek.Type == TokenDelim && peek.Value == "~":
		tok.Next()
		if tok.Peek().Type == TokenDelim && tok.Peek().Value == "=" {
			tok.Next()
			matcher = AttrIncludes
		} else {
			skipToCloseBracket(tok)
			return AttrSelector{}, false
		}
	case peek.Type == TokenDelim && peek.Value == "|":
		tok.Next()
		if tok.Peek().Type == TokenDelim && tok.Peek().Value == "=" {
			tok.Next()
			matcher = AttrDashMatch
		} else {
			skipToCloseBracket(tok)
			return AttrSelector{}, false
		}
	case peek.Type == TokenDelim && peek.Value == "^":
		tok.Next()
		if tok.Peek().Type == TokenDelim && tok.Peek().Value == "=" {
			tok.Next()
			matcher = AttrPrefix
		} else {
			skipToCloseBracket(tok)
			return AttrSelector{}, false
		}
	case peek.Type == TokenDelim && peek.Value == "$":
		tok.Next()
		if tok.Peek().Type == TokenDelim && tok.Peek().Value == "=" {
			tok.Next()
			matcher = AttrSuffix
		} else {
			skipToCloseBracket(tok)
			return AttrSelector{}, false
		}
	case peek.Type == TokenDelim && peek.Value == "*":
		tok.Next()
		if tok.Peek().Type == TokenDelim && tok.Peek().Value == "=" {
			tok.Next()
			matcher = AttrSubstring
		} else {
			skipToCloseBracket(tok)
			return AttrSelector{}, false
		}
	default:
		skipToCloseBracket(tok)
		return AttrSelector{}, false
	}

	// Skip whitespace
	for tok.Peek().Type == TokenWhitespace {
		tok.Next()
	}

	// Parse value (ident or string)
	var value string
	switch tok.Peek().Type {
	case TokenString:
		value = tok.Next().Value
	case TokenIdent:
		value = tok.Next().Value
	default:
		skipToCloseBracket(tok)
		return AttrSelector{}, false
	}

	// Skip whitespace and expect closing bracket
	for tok.Peek().Type == TokenWhitespace {
		tok.Next()
	}
	if tok.Peek().Type == TokenCloseBracket {
		tok.Next()
	}

	return AttrSelector{Name: name, Matcher: matcher, Value: value}, true
}

// skipToCloseBracket advances the tokenizer past the next ']'.
func skipToCloseBracket(tok *Tokenizer) {
	for {
		t := tok.Next()
		if t.Type == TokenCloseBracket || t.Type == TokenEOF {
			return
		}
	}
}

// Match reports whether the given selector matches the node.
func Match(sel Selector, node Matchable) bool {
	if len(sel.Parts) == 0 {
		return false
	}

	// Match rightmost part (subject) against the node
	subject := sel.Parts[len(sel.Parts)-1]
	if !matchCompound(subject, node) {
		return false
	}

	// Match remaining parts right-to-left against ancestors
	current := node
	for i := len(sel.Parts) - 2; i >= 0; i-- {
		part := sel.Parts[i+1] // get combinator from the next-rightward part
		compound := sel.Parts[i]

		switch part.Combinator {
		case CombinatorNone:
			// Leftmost part matched already
			continue

		case CombinatorChild:
			// Must match the direct parent
			parent := current.NodeParent()
			if parent == nil || !matchCompound(compound, parent) {
				return false
			}
			current = parent

		case CombinatorDescendant:
			// Must match some ancestor
			found := false
			ancestor := current.NodeParent()
			for ancestor != nil {
				if matchCompound(compound, ancestor) {
					found = true
					current = ancestor
					break
				}
				ancestor = ancestor.NodeParent()
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// matchCompound checks if a compound selector matches a single node.
func matchCompound(cs CompoundSelector, node Matchable) bool {
	// Tag check
	if cs.Tag != "" && cs.Tag != strings.ToLower(node.NodeTag()) {
		return false
	}

	// ID check
	if cs.ID != "" && cs.ID != node.NodeID() {
		return false
	}

	// Class check
	if len(cs.Classes) > 0 {
		nodeClasses := node.NodeClasses()
		classSet := make(map[string]bool, len(nodeClasses))
		for _, c := range nodeClasses {
			classSet[c] = true
		}
		for _, required := range cs.Classes {
			if !classSet[required] {
				return false
			}
		}
	}

	// Attribute selector check
	for _, attr := range cs.Attrs {
		val, exists := node.NodeAttr(attr.Name)
		if !matchAttr(attr, val, exists) {
			return false
		}
	}

	return true
}

// matchAttr checks whether an attribute value satisfies an attribute selector.
func matchAttr(sel AttrSelector, val string, exists bool) bool {
	switch sel.Matcher {
	case AttrExists:
		return exists
	case AttrEquals:
		return exists && val == sel.Value
	case AttrIncludes:
		if !exists {
			return false
		}
		for _, word := range strings.Fields(val) {
			if word == sel.Value {
				return true
			}
		}
		return false
	case AttrDashMatch:
		return exists && (val == sel.Value || strings.HasPrefix(val, sel.Value+"-"))
	case AttrPrefix:
		return exists && strings.HasPrefix(val, sel.Value)
	case AttrSuffix:
		return exists && strings.HasSuffix(val, sel.Value)
	case AttrSubstring:
		return exists && strings.Contains(val, sel.Value)
	}
	return false
}

// splitSelectorGroup splits "h1, h2, h3" into ["h1", " h2", " h3"].
// It respects parentheses (for :not() etc. in future).
func splitSelectorGroup(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}
