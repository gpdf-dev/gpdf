package css

import (
	"strings"
)

// Default limits for parser safety.
const (
	DefaultMaxCSSDepth     = 64
	DefaultMaxCSSInputSize = 50 * 1024 * 1024 // 50MB
)

// Stylesheet represents a parsed CSS stylesheet.
type Stylesheet struct {
	Rules []Rule
}

// Rule represents a CSS style rule (selector + declarations).
type Rule struct {
	Selectors    string        // raw selector text (e.g. "div.main > p")
	Declarations []Declaration // property: value pairs
}

// Declaration represents a single CSS property-value pair.
type Declaration struct {
	Property  string // lowercase property name (e.g. "font-size")
	Value     string // raw value text (e.g. "12px")
	Important bool   // has !important
}

// CSSParserOption configures the CSS parser.
type CSSParserOption func(*cssParser)

// WithMaxCSSDepth sets the maximum nesting depth for CSS parsing.
func WithMaxCSSDepth(depth int) CSSParserOption {
	return func(p *cssParser) {
		p.maxDepth = depth
	}
}

// WithMaxCSSInputSize sets the maximum input size for CSS parsing.
func WithMaxCSSInputSize(size int) CSSParserOption {
	return func(p *cssParser) {
		p.maxInputSize = size
	}
}

// ParseStylesheet parses a CSS stylesheet string into rules.
func ParseStylesheet(css string, opts ...CSSParserOption) (*Stylesheet, error) {
	p := &cssParser{
		maxDepth:     DefaultMaxCSSDepth,
		maxInputSize: DefaultMaxCSSInputSize,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p.parseStylesheet(css)
}

// ParseDeclarations parses a CSS declaration block (e.g. from a style attribute).
// Input should be the content between { and }, or the value of a style attribute.
func ParseDeclarations(css string) []Declaration {
	p := &cssParser{
		maxDepth:     DefaultMaxCSSDepth,
		maxInputSize: DefaultMaxCSSInputSize,
	}
	return p.parseDeclarationList(css)
}

// ParseValue parses a CSS value string into tokens.
func ParseValue(value string) []Token {
	tok := NewTokenizer([]byte(value))
	var tokens []Token
	for {
		t := tok.Next()
		if t.Type == TokenEOF {
			break
		}
		if t.Type == TokenWhitespace {
			continue
		}
		tokens = append(tokens, t)
	}
	return tokens
}

// ─── Internal parser ──────────────────────────────────────────────

type cssParser struct {
	maxDepth     int
	maxInputSize int
}

func (p *cssParser) parseStylesheet(css string) (*Stylesheet, error) {
	tok := NewTokenizer([]byte(css))
	ss := &Stylesheet{}

	for {
		p.skipWhitespaceAndComments(tok)
		next := tok.Peek()
		if next.Type == TokenEOF {
			break
		}

		// At-rule (skip for now in Phase 4-A, except we ignore them gracefully)
		if next.Type == TokenAtKeyword {
			p.skipAtRule(tok)
			continue
		}

		// Try to parse a style rule
		rule, ok := p.parseRule(tok)
		if ok {
			ss.Rules = append(ss.Rules, rule)
		}
	}

	return ss, nil
}

// parseRule parses a single CSS style rule: selector { declarations }
func (p *cssParser) parseRule(tok *Tokenizer) (Rule, bool) {
	// Read selector until '{'
	selector := p.readUntilBrace(tok)
	selector = strings.TrimSpace(selector)
	if selector == "" {
		// No selector found — skip to after the closing brace
		p.skipBlock(tok)
		return Rule{}, false
	}

	// Expect '{'
	next := tok.Next()
	if next.Type != TokenOpenBrace {
		return Rule{}, false
	}

	// Parse declarations until '}'
	decls := p.parseDeclarationsFromTokenizer(tok)

	return Rule{
		Selectors:    selector,
		Declarations: decls,
	}, true
}

// readUntilBrace reads tokens and builds a string until '{' is found.
func (p *cssParser) readUntilBrace(tok *Tokenizer) string {
	var sb strings.Builder
	for {
		next := tok.Peek()
		if next.Type == TokenEOF || next.Type == TokenOpenBrace {
			break
		}
		t := tok.Next()
		sb.WriteString(tokenToString(t))
	}
	return sb.String()
}

// parseDeclarationsFromTokenizer parses declarations inside a block { ... }.
func (p *cssParser) parseDeclarationsFromTokenizer(tok *Tokenizer) []Declaration {
	var decls []Declaration
	for {
		p.skipWhitespaceAndComments(tok)
		next := tok.Peek()
		if next.Type == TokenEOF || next.Type == TokenCloseBrace {
			tok.Next() // consume '}'
			break
		}

		decl, ok := p.parseDeclaration(tok)
		if ok {
			decls = append(decls, decl)
		}
	}
	return decls
}

// parseDeclaration parses a single declaration: property: value [!important] ;
func (p *cssParser) parseDeclaration(tok *Tokenizer) (Declaration, bool) {
	p.skipWhitespaceAndComments(tok)

	// Read property name
	propTok := tok.Next()
	if propTok.Type != TokenIdent {
		// Skip to next ';' or '}'
		p.skipToSemicolonOrBrace(tok)
		return Declaration{}, false
	}
	property := strings.ToLower(propTok.Value)

	p.skipWhitespaceAndComments(tok)

	// Expect ':'
	colon := tok.Next()
	if colon.Type != TokenColon {
		p.skipToSemicolonOrBrace(tok)
		return Declaration{}, false
	}

	// Read value until ';' or '}' or EOF
	value, important := p.readValue(tok)
	value = strings.TrimSpace(value)

	if value == "" {
		return Declaration{}, false
	}

	return Declaration{
		Property:  property,
		Value:     value,
		Important: important,
	}, true
}

// parseDeclarationList parses declarations from a style attribute string.
func (p *cssParser) parseDeclarationList(css string) []Declaration {
	tok := NewTokenizer([]byte(css))
	var decls []Declaration
	for {
		p.skipWhitespaceAndComments(tok)
		if tok.Peek().Type == TokenEOF {
			break
		}
		decl, ok := p.parseDeclaration(tok)
		if ok {
			decls = append(decls, decl)
		}
	}
	return decls
}

// readValue reads the value portion of a declaration, stopping at ';' or '}'.
// It also detects !important.
func (p *cssParser) readValue(tok *Tokenizer) (string, bool) {
	var sb strings.Builder
	important := false
	parenDepth := 0

	for {
		next := tok.Peek()
		if next.Type == TokenEOF {
			break
		}
		if next.Type == TokenCloseBrace {
			// Don't consume '}' — let the caller handle it
			break
		}
		if next.Type == TokenSemicolon && parenDepth == 0 {
			tok.Next() // consume ';'
			break
		}

		t := tok.Next()

		if t.Type == TokenOpenParen {
			parenDepth++
		} else if t.Type == TokenCloseParen {
			if parenDepth > 0 {
				parenDepth--
			}
		}

		// Check for !important
		if t.Type == TokenDelim && t.Value == "!" {
			p.skipWhitespaceAndComments(tok)
			peek := tok.Peek()
			if peek.Type == TokenIdent && strings.EqualFold(peek.Value, "important") {
				tok.Next() // consume "important"
				important = true
				continue
			}
		}

		sb.WriteString(tokenToString(t))
	}

	return sb.String(), important
}

// skipAtRule skips an at-rule (e.g. @media, @import).
func (p *cssParser) skipAtRule(tok *Tokenizer) {
	tok.Next() // consume @keyword
	depth := 0
	for {
		t := tok.Next()
		if t.Type == TokenEOF {
			return
		}
		if t.Type == TokenOpenBrace {
			depth++
		}
		if t.Type == TokenCloseBrace {
			if depth <= 1 {
				return
			}
			depth--
		}
		if t.Type == TokenSemicolon && depth == 0 {
			return
		}
	}
}

// skipBlock skips everything up to and including the matching '}'.
func (p *cssParser) skipBlock(tok *Tokenizer) {
	depth := 0
	for {
		t := tok.Next()
		if t.Type == TokenEOF {
			return
		}
		if t.Type == TokenOpenBrace {
			depth++
		}
		if t.Type == TokenCloseBrace {
			if depth <= 0 {
				return
			}
			depth--
		}
	}
}

// skipToSemicolonOrBrace skips tokens until ';' or '}' is found.
func (p *cssParser) skipToSemicolonOrBrace(tok *Tokenizer) {
	for {
		next := tok.Peek()
		if next.Type == TokenEOF || next.Type == TokenCloseBrace {
			return
		}
		t := tok.Next()
		if t.Type == TokenSemicolon {
			return
		}
	}
}

// skipWhitespaceAndComments skips whitespace tokens.
// (Comments are already skipped by the tokenizer.)
func (p *cssParser) skipWhitespaceAndComments(tok *Tokenizer) {
	for {
		next := tok.Peek()
		if next.Type != TokenWhitespace {
			return
		}
		tok.Next()
	}
}

// tokenToString converts a token back to its string representation.
func tokenToString(t Token) string {
	switch t.Type {
	case TokenString:
		return "\"" + t.Value + "\""
	case TokenHash:
		return "#" + t.Value
	case TokenFunction:
		return t.Value + "("
	case TokenDimension:
		return t.Value + t.Unit
	case TokenPercentage:
		return t.Value + "%"
	case TokenAtKeyword:
		return "@" + t.Value
	default:
		return t.Value
	}
}
