// Package css provides a zero-dependency CSS tokenizer, parser,
// selector matcher, and cascade engine for use with the gpdf
// PDF generation library.
//
// The implementation follows the CSS Syntax Module Level 3 specification
// for tokenization and parsing, with scope limited to properties
// relevant for PDF generation.
package css

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents the type of a CSS token.
type TokenType int

const (
	TokenEOF         TokenType = iota
	TokenIdent                 // e.g. "color", "div"
	TokenHash                  // e.g. "#fff", "#main"
	TokenString                // e.g. "hello", 'world'
	TokenNumber                // e.g. "42", "3.14"
	TokenDimension             // e.g. "12px", "1.5em"
	TokenPercentage            // e.g. "50%"
	TokenFunction              // e.g. "rgb("
	TokenDelim                 // single character delimiter
	TokenWhitespace            // space, tab, newline
	TokenAtKeyword             // e.g. "@media", "@page"
	TokenColon                 // ':'
	TokenSemicolon             // ';'
	TokenComma                 // ','
	TokenOpenBrace             // '{'
	TokenCloseBrace            // '}'
	TokenOpenParen             // '('
	TokenCloseParen            // ')'
	TokenOpenBracket           // '['
	TokenCloseBracket          // ']'
	TokenCDO                   // '<!--'
	TokenCDC                   // '-->'
)

// Token represents a single CSS token.
type Token struct {
	Type  TokenType
	Value string  // raw string value
	Num   float64 // numeric value for Number/Dimension/Percentage
	Unit  string  // unit for Dimension (e.g. "px", "em")
}

// IsIdent reports whether the token is an ident with the given (case-insensitive) name.
func (t Token) IsIdent(name string) bool {
	return t.Type == TokenIdent && strings.EqualFold(t.Value, name)
}

// Tokenizer performs lexical analysis of CSS input.
type Tokenizer struct {
	input []byte
	pos   int
	size  int
}

// NewTokenizer creates a new CSS tokenizer.
func NewTokenizer(input []byte) *Tokenizer {
	return &Tokenizer{input: input, pos: 0, size: len(input)}
}

// Next returns the next token from the input.
func (t *Tokenizer) Next() Token {
	if t.pos >= t.size {
		return Token{Type: TokenEOF}
	}

	r, sz := utf8.DecodeRune(t.input[t.pos:])

	// Whitespace
	if isWhitespace(r) {
		return t.consumeWhitespace()
	}

	// String
	if r == '"' || r == '\'' {
		return t.consumeString(byte(r))
	}

	// Hash
	if r == '#' {
		t.pos += sz
		if t.pos < t.size && isNameChar(t.peekRune()) {
			name := t.consumeName()
			return Token{Type: TokenHash, Value: name}
		}
		return Token{Type: TokenDelim, Value: "#"}
	}

	// Number / Dimension / Percentage
	if isDigit(r) || (r == '.' && t.pos+sz < t.size && isDigit(peekRuneAt(t.input, t.pos+sz))) {
		return t.consumeNumeric()
	}

	// Sign followed by digit or dot-digit
	if (r == '+' || r == '-') && t.pos+sz < t.size {
		next := peekRuneAt(t.input, t.pos+sz)
		if isDigit(next) || (next == '.' && t.pos+sz+1 < t.size && isDigit(peekRuneAt(t.input, t.pos+sz+1))) {
			return t.consumeNumeric()
		}
	}

	// At-keyword
	if r == '@' {
		t.pos += sz
		if t.pos < t.size && isNameStart(t.peekRune()) {
			name := t.consumeName()
			return Token{Type: TokenAtKeyword, Value: name}
		}
		return Token{Type: TokenDelim, Value: "@"}
	}

	// Ident or Function (including escape-started idents like \41)
	if isNameStart(r) || r == '\\' || (r == '-' && t.pos+sz < t.size && (isNameStart(peekRuneAt(t.input, t.pos+sz)) || peekRuneAt(t.input, t.pos+sz) == '-' || peekRuneAt(t.input, t.pos+sz) == '\\')) {
		return t.consumeIdentLike()
	}

	// Comment: /* ... */
	if r == '/' && t.pos+sz < t.size && t.input[t.pos+sz] == '*' {
		t.consumeComment()
		return t.Next() // skip comments, get next real token
	}

	// Single character tokens
	t.pos += sz
	switch r {
	case ':':
		return Token{Type: TokenColon, Value: ":"}
	case ';':
		return Token{Type: TokenSemicolon, Value: ";"}
	case ',':
		return Token{Type: TokenComma, Value: ","}
	case '{':
		return Token{Type: TokenOpenBrace, Value: "{"}
	case '}':
		return Token{Type: TokenCloseBrace, Value: "}"}
	case '(':
		return Token{Type: TokenOpenParen, Value: "("}
	case ')':
		return Token{Type: TokenCloseParen, Value: ")"}
	case '[':
		return Token{Type: TokenOpenBracket, Value: "["}
	case ']':
		return Token{Type: TokenCloseBracket, Value: "]"}
	default:
		return Token{Type: TokenDelim, Value: string(r)}
	}
}

// Peek returns the next token without consuming it.
func (t *Tokenizer) Peek() Token {
	saved := t.pos
	tok := t.Next()
	t.pos = saved
	return tok
}

// consumeWhitespace reads a run of whitespace.
func (t *Tokenizer) consumeWhitespace() Token {
	start := t.pos
	for t.pos < t.size {
		r, sz := utf8.DecodeRune(t.input[t.pos:])
		if !isWhitespace(r) {
			break
		}
		t.pos += sz
	}
	return Token{Type: TokenWhitespace, Value: string(t.input[start:t.pos])}
}

// consumeString reads a quoted string.
func (t *Tokenizer) consumeString(quote byte) Token {
	t.pos++ // skip opening quote
	var sb strings.Builder
	for t.pos < t.size {
		ch := t.input[t.pos]
		if ch == quote {
			t.pos++ // skip closing quote
			return Token{Type: TokenString, Value: sb.String()}
		}
		if ch == '\\' {
			t.pos++
			if t.pos < t.size {
				esc := t.input[t.pos]
				if esc == '\n' {
					// Escaped newline: line continuation
					t.pos++
					continue
				}
				// Hex escape: \ABCDEF (1-6 hex digits)
				if isHexDigit(rune(esc)) {
					sb.WriteRune(t.consumeEscape())
					continue
				}
				sb.WriteByte(esc)
				t.pos++
			}
			continue
		}
		if ch == '\n' {
			// Unescaped newline in string: bad string (per spec),
			// but we tolerate it by ending the string here
			return Token{Type: TokenString, Value: sb.String()}
		}
		r, sz := utf8.DecodeRune(t.input[t.pos:])
		sb.WriteRune(r)
		t.pos += sz
	}
	// Unterminated string
	return Token{Type: TokenString, Value: sb.String()}
}

// consumeNumeric reads a number, dimension, or percentage.
func (t *Tokenizer) consumeNumeric() Token {
	numStr, numVal := t.consumeNumber()

	if t.pos < t.size && isNameStart(t.peekRune()) {
		unit := t.consumeName()
		return Token{Type: TokenDimension, Value: numStr, Num: numVal, Unit: strings.ToLower(unit)}
	}
	if t.pos < t.size && t.input[t.pos] == '%' {
		t.pos++
		return Token{Type: TokenPercentage, Value: numStr, Num: numVal}
	}
	return Token{Type: TokenNumber, Value: numStr, Num: numVal}
}

// consumeNumber reads a CSS number and returns its string representation and float64 value.
func (t *Tokenizer) consumeNumber() (string, float64) {
	start := t.pos

	// Optional sign
	if t.pos < t.size && (t.input[t.pos] == '+' || t.input[t.pos] == '-') {
		t.pos++
	}

	// Integer part
	for t.pos < t.size && isDigit(rune(t.input[t.pos])) {
		t.pos++
	}

	// Fractional part
	if t.pos < t.size && t.input[t.pos] == '.' {
		t.pos++
		for t.pos < t.size && isDigit(rune(t.input[t.pos])) {
			t.pos++
		}
	}

	// Exponent
	if t.pos < t.size && (t.input[t.pos] == 'e' || t.input[t.pos] == 'E') {
		saved := t.pos
		t.pos++
		if t.pos < t.size && (t.input[t.pos] == '+' || t.input[t.pos] == '-') {
			t.pos++
		}
		if t.pos < t.size && isDigit(rune(t.input[t.pos])) {
			for t.pos < t.size && isDigit(rune(t.input[t.pos])) {
				t.pos++
			}
		} else {
			// Not a valid exponent — roll back
			t.pos = saved
		}
	}

	raw := string(t.input[start:t.pos])
	val := parseFloat(raw)
	return raw, val
}

// consumeName reads a CSS name (ident-like sequence).
func (t *Tokenizer) consumeName() string {
	var sb strings.Builder
	for t.pos < t.size {
		r, sz := utf8.DecodeRune(t.input[t.pos:])
		if r == '\\' && t.pos+sz < t.size {
			t.pos += sz
			sb.WriteRune(t.consumeEscape())
			continue
		}
		if !isNameChar(r) {
			break
		}
		sb.WriteRune(r)
		t.pos += sz
	}
	return sb.String()
}

// consumeIdentLike reads an ident or function token.
func (t *Tokenizer) consumeIdentLike() Token {
	name := t.consumeName()
	if t.pos < t.size && t.input[t.pos] == '(' {
		t.pos++ // consume '('
		return Token{Type: TokenFunction, Value: strings.ToLower(name)}
	}
	return Token{Type: TokenIdent, Value: name}
}

// consumeComment skips a CSS comment /* ... */.
func (t *Tokenizer) consumeComment() {
	t.pos += 2 // skip '/*'
	for t.pos+1 < t.size {
		if t.input[t.pos] == '*' && t.input[t.pos+1] == '/' {
			t.pos += 2
			return
		}
		t.pos++
	}
	t.pos = t.size // unterminated comment
}

// consumeEscape reads a CSS escape sequence (after the backslash).
// Handles 1-6 hex digits optionally followed by whitespace.
func (t *Tokenizer) consumeEscape() rune {
	if t.pos >= t.size {
		return unicode.ReplacementChar
	}

	r, sz := utf8.DecodeRune(t.input[t.pos:])
	if !isHexDigit(r) {
		// Non-hex escape: return the character itself
		t.pos += sz
		return r
	}

	// Hex escape: up to 6 hex digits
	var val rune
	count := 0
	for t.pos < t.size && count < 6 {
		r, sz = utf8.DecodeRune(t.input[t.pos:])
		if !isHexDigit(r) {
			break
		}
		val = val*16 + hexVal(r)
		t.pos += sz
		count++
	}

	// Consume optional single trailing whitespace
	if t.pos < t.size {
		r, sz = utf8.DecodeRune(t.input[t.pos:])
		if isWhitespace(r) {
			t.pos += sz
		}
	}

	if val == 0 || !utf8.ValidRune(val) {
		return unicode.ReplacementChar
	}
	return val
}

// peekRune returns the next rune without consuming it.
func (t *Tokenizer) peekRune() rune {
	r, _ := utf8.DecodeRune(t.input[t.pos:])
	return r
}

// ─── Character classification ─────────────────────────────────────

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isNameStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r > 0x7F
}

func isNameChar(r rune) bool {
	return isNameStart(r) || isDigit(r) || r == '-'
}

func hexVal(r rune) rune {
	switch {
	case r >= '0' && r <= '9':
		return r - '0'
	case r >= 'a' && r <= 'f':
		return r - 'a' + 10
	case r >= 'A' && r <= 'F':
		return r - 'A' + 10
	}
	return 0
}

// peekRuneAt decodes a rune at the given byte position.
func peekRuneAt(data []byte, pos int) rune {
	if pos >= len(data) {
		return 0
	}
	r, _ := utf8.DecodeRune(data[pos:])
	return r
}

// parseFloat parses a float64 from a string using simple manual parsing
// to avoid importing strconv in the hot path. Falls back to 0 on error.
func parseFloat(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	neg := false
	i := 0
	if s[0] == '-' {
		neg = true
		i++
	} else if s[0] == '+' {
		i++
	}

	var intPart float64
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		intPart = intPart*10 + float64(s[i]-'0')
		i++
	}

	var fracPart float64
	if i < len(s) && s[i] == '.' {
		i++
		divisor := 10.0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			fracPart += float64(s[i]-'0') / divisor
			divisor *= 10
			i++
		}
	}

	result := intPart + fracPart

	// Exponent
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		expNeg := false
		if i < len(s) && s[i] == '-' {
			expNeg = true
			i++
		} else if i < len(s) && s[i] == '+' {
			i++
		}
		var exp float64
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			exp = exp*10 + float64(s[i]-'0')
			i++
		}
		mul := 1.0
		for range int(exp) {
			if expNeg {
				mul /= 10
			} else {
				mul *= 10
			}
		}
		result *= mul
	}

	if neg {
		result = -result
	}
	return result
}
