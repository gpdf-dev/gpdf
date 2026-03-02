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

// CompoundSelector represents a compound selector like "div.main#app".
type CompoundSelector struct {
	Tag        string   // element type (lowercase), "" means any
	ID         string   // #id value, "" means no ID constraint
	Classes    []string // .class values
	Universal  bool     // true if '*' was explicitly specified
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

		default:
			return cs, matched
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

	return true
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
