package html

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenType represents the type of an HTML token.
type tokenType int

const (
	tokenError      tokenType = iota
	tokenEOF                  // end of input
	tokenStartTag             // <div>, <p class="x">
	tokenEndTag               // </div>
	tokenSelfClose            // <br/>, <img .../>
	tokenText                 // text content
	tokenComment              // <!-- comment -->
	tokenDoctype              // <!DOCTYPE html>
)

// token represents a single HTML token.
type token struct {
	typ   tokenType
	tag   string            // lowercase tag name (for start/end/self-close)
	attrs map[string]string // attributes (for start/self-close)
	data  string            // text content or comment body or error message
}

// tokenizer performs lexical analysis of HTML input.
type tokenizer struct {
	input []byte
	pos   int
	size  int // cached len(input)
}

// newTokenizer creates a new tokenizer for the given input.
func newTokenizer(input []byte) *tokenizer {
	return &tokenizer{
		input: input,
		pos:   0,
		size:  len(input),
	}
}

// next returns the next token from the input.
func (t *tokenizer) next() token {
	if t.pos >= t.size {
		return token{typ: tokenEOF}
	}

	if t.input[t.pos] == '<' {
		return t.readTag()
	}
	return t.readText()
}

// readText reads text content until the next '<' or end of input.
func (t *tokenizer) readText() token {
	start := t.pos
	for t.pos < t.size && t.input[t.pos] != '<' {
		t.pos++
	}
	raw := string(t.input[start:t.pos])
	decoded := decodeEntities(raw)
	return token{typ: tokenText, data: decoded}
}

// readTag reads a tag starting with '<'.
func (t *tokenizer) readTag() token {
	// Skip '<'
	t.pos++
	if t.pos >= t.size {
		return token{typ: tokenText, data: "<"}
	}

	ch := t.input[t.pos]

	// Comment: <!-- ... -->
	if ch == '!' {
		return t.readBangTag()
	}

	// End tag: </...>
	if ch == '/' {
		t.pos++
		return t.readEndTag()
	}

	// Check for valid tag start character
	if ch == '?' {
		// Processing instruction — skip until ?>
		return t.skipUntil("?>")
	}

	r, _ := utf8.DecodeRune(t.input[t.pos:])
	if !isTagNameStart(r) {
		// Not a valid tag — treat '<' as text
		return token{typ: tokenText, data: "<"}
	}

	return t.readStartTag()
}

// readBangTag handles tags starting with <!
func (t *tokenizer) readBangTag() token {
	// t.pos is at '!'
	t.pos++

	// Check for comment
	if t.pos+1 < t.size && t.input[t.pos] == '-' && t.input[t.pos+1] == '-' {
		t.pos += 2 // skip '--'
		return t.readComment()
	}

	// Check for DOCTYPE
	remaining := string(t.input[t.pos:])
	if len(remaining) >= 7 && strings.EqualFold(remaining[:7], "DOCTYPE") {
		return t.readDoctype()
	}

	// Check for CDATA
	if strings.HasPrefix(remaining, "[CDATA[") {
		t.pos += 7
		return t.readCDATA()
	}

	// Unknown <! construct — skip until >
	return t.skipUntil(">")
}

// readComment reads an HTML comment (<!-- ... -->).
func (t *tokenizer) readComment() token {
	start := t.pos
	for t.pos < t.size {
		if t.pos+2 < t.size && t.input[t.pos] == '-' && t.input[t.pos+1] == '-' && t.input[t.pos+2] == '>' {
			data := string(t.input[start:t.pos])
			t.pos += 3 // skip '-->'
			return token{typ: tokenComment, data: data}
		}
		t.pos++
	}
	// Unterminated comment — consume rest as comment
	data := string(t.input[start:t.pos])
	return token{typ: tokenComment, data: data}
}

// readDoctype reads a <!DOCTYPE ...> declaration.
func (t *tokenizer) readDoctype() token {
	start := t.pos
	for t.pos < t.size && t.input[t.pos] != '>' {
		t.pos++
	}
	data := string(t.input[start:t.pos])
	if t.pos < t.size {
		t.pos++ // skip '>'
	}
	return token{typ: tokenDoctype, data: data}
}

// readCDATA reads a CDATA section (<![CDATA[ ... ]]>).
func (t *tokenizer) readCDATA() token {
	start := t.pos
	for t.pos < t.size {
		if t.pos+2 < t.size && t.input[t.pos] == ']' && t.input[t.pos+1] == ']' && t.input[t.pos+2] == '>' {
			data := string(t.input[start:t.pos])
			t.pos += 3 // skip ']]>'
			return token{typ: tokenText, data: data}
		}
		t.pos++
	}
	// Unterminated — return what we have
	data := string(t.input[start:t.pos])
	return token{typ: tokenText, data: data}
}

// readStartTag reads an opening tag (with attributes).
func (t *tokenizer) readStartTag() token {
	tag := t.readTagName()
	attrs := t.readAttributes()

	selfClose := false
	t.skipWhitespace()
	if t.pos < t.size && t.input[t.pos] == '/' {
		selfClose = true
		t.pos++
	}
	if t.pos < t.size && t.input[t.pos] == '>' {
		t.pos++
	}

	typ := tokenStartTag
	if selfClose {
		typ = tokenSelfClose
	}
	return token{typ: typ, tag: tag, attrs: attrs}
}

// readEndTag reads a closing tag (</tag>).
func (t *tokenizer) readEndTag() token {
	tag := t.readTagName()
	// Skip any extra content and find '>'
	for t.pos < t.size && t.input[t.pos] != '>' {
		t.pos++
	}
	if t.pos < t.size {
		t.pos++ // skip '>'
	}
	return token{typ: tokenEndTag, tag: tag}
}

// readTagName reads and returns a lowercase tag name.
func (t *tokenizer) readTagName() string {
	var sb strings.Builder
	for t.pos < t.size {
		r, sz := utf8.DecodeRune(t.input[t.pos:])
		if !isTagNameChar(r) {
			break
		}
		sb.WriteRune(unicode.ToLower(r))
		t.pos += sz
	}
	return sb.String()
}

// readAttributes reads tag attributes and returns them as a map.
func (t *tokenizer) readAttributes() map[string]string {
	attrs := make(map[string]string)
	for {
		t.skipWhitespace()
		if t.pos >= t.size {
			break
		}
		ch := t.input[t.pos]
		if ch == '>' || ch == '/' {
			break
		}

		name := t.readAttrName()
		if name == "" {
			// Skip invalid character
			t.pos++
			continue
		}

		t.skipWhitespace()
		if t.pos < t.size && t.input[t.pos] == '=' {
			t.pos++ // skip '='
			t.skipWhitespace()
			attrs[name] = t.readAttrValue()
		} else {
			// Boolean attribute (e.g., <input disabled>)
			attrs[name] = ""
		}
	}
	return attrs
}

// readAttrName reads an attribute name.
func (t *tokenizer) readAttrName() string {
	var sb strings.Builder
	for t.pos < t.size {
		r, sz := utf8.DecodeRune(t.input[t.pos:])
		if r == '=' || r == '>' || r == '/' || unicode.IsSpace(r) {
			break
		}
		sb.WriteRune(unicode.ToLower(r))
		t.pos += sz
	}
	return sb.String()
}

// readAttrValue reads an attribute value (quoted or unquoted).
func (t *tokenizer) readAttrValue() string {
	if t.pos >= t.size {
		return ""
	}

	ch := t.input[t.pos]
	if ch == '"' || ch == '\'' {
		return t.readQuotedValue(ch)
	}
	return t.readUnquotedValue()
}

// readQuotedValue reads a quoted attribute value.
func (t *tokenizer) readQuotedValue(quote byte) string {
	t.pos++ // skip opening quote
	var sb strings.Builder
	for t.pos < t.size && t.input[t.pos] != quote {
		sb.WriteByte(t.input[t.pos])
		t.pos++
	}
	if t.pos < t.size {
		t.pos++ // skip closing quote
	}
	return decodeEntities(sb.String())
}

// readUnquotedValue reads an unquoted attribute value.
func (t *tokenizer) readUnquotedValue() string {
	var sb strings.Builder
	for t.pos < t.size {
		ch := t.input[t.pos]
		if ch == '>' || ch == '/' || ch == '\'' || ch == '"' || isASCIISpace(ch) {
			break
		}
		sb.WriteByte(ch)
		t.pos++
	}
	return decodeEntities(sb.String())
}

// skipWhitespace advances past any whitespace characters.
func (t *tokenizer) skipWhitespace() {
	for t.pos < t.size && isASCIISpace(t.input[t.pos]) {
		t.pos++
	}
}

// skipUntil skips characters until the given delimiter is found.
func (t *tokenizer) skipUntil(delim string) token {
	delimBytes := []byte(delim)
	for t.pos < t.size {
		if t.pos+len(delimBytes) <= t.size {
			match := true
			for i, b := range delimBytes {
				if t.input[t.pos+i] != b {
					match = false
					break
				}
			}
			if match {
				t.pos += len(delimBytes)
				return token{typ: tokenComment, data: ""}
			}
		}
		t.pos++
	}
	return token{typ: tokenComment, data: ""}
}

// decodeEntities decodes HTML character references in a string.
// Supports named entities (&amp;), decimal (&#123;), and hex (&#x7B;) references.
func decodeEntities(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != '&' {
			sb.WriteByte(s[i])
			i++
			continue
		}

		// Find the end of the entity
		end := i + 1
		for end < len(s) && end-i < 32 && s[end] != ';' && s[end] != '&' && !isASCIISpace(s[end]) {
			end++
		}

		if end >= len(s) || s[end] != ';' {
			// Not a valid entity — output '&' as-is
			sb.WriteByte('&')
			i++
			continue
		}

		entity := s[i+1 : end]
		if r, ok := resolveEntity(entity); ok {
			sb.WriteRune(r)
			i = end + 1
		} else {
			// Unknown entity — output as-is
			sb.WriteString(s[i : end+1])
			i = end + 1
		}
	}
	return sb.String()
}

// resolveEntity resolves a single entity reference (without & and ;).
func resolveEntity(entity string) (rune, bool) {
	if len(entity) == 0 {
		return 0, false
	}

	// Numeric reference
	if entity[0] == '#' {
		return resolveNumericEntity(entity[1:])
	}

	// Named reference
	return lookupEntity(entity)
}

// resolveNumericEntity resolves a numeric character reference.
func resolveNumericEntity(s string) (rune, bool) {
	if len(s) == 0 {
		return 0, false
	}

	var n uint64
	var err error

	if s[0] == 'x' || s[0] == 'X' {
		// Hex reference: &#x7B;
		if len(s) < 2 {
			return 0, false
		}
		n, err = strconv.ParseUint(s[1:], 16, 32)
	} else {
		// Decimal reference: &#123;
		n, err = strconv.ParseUint(s, 10, 32)
	}

	if err != nil {
		return 0, false
	}

	r := rune(n)
	// Reject invalid code points — return false so the entity is preserved as-is
	if r == 0 || !utf8.ValidRune(r) {
		return 0, false
	}
	return r, true
}

// isTagNameStart reports whether r can start an HTML tag name.
func isTagNameStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isTagNameChar reports whether r can appear in an HTML tag name.
func isTagNameChar(r rune) bool {
	return isTagNameStart(r) || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.'
}

// isASCIISpace reports whether b is an ASCII whitespace byte.
func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

// String returns a debug representation of the token.
func (tok token) String() string {
	switch tok.typ {
	case tokenStartTag:
		return fmt.Sprintf("<StartTag %s attrs=%v>", tok.tag, tok.attrs)
	case tokenEndTag:
		return fmt.Sprintf("<EndTag %s>", tok.tag)
	case tokenSelfClose:
		return fmt.Sprintf("<SelfClose %s attrs=%v>", tok.tag, tok.attrs)
	case tokenText:
		return fmt.Sprintf("<Text %q>", tok.data)
	case tokenComment:
		return fmt.Sprintf("<Comment %q>", tok.data)
	case tokenDoctype:
		return fmt.Sprintf("<Doctype %q>", tok.data)
	case tokenEOF:
		return "<EOF>"
	default:
		return fmt.Sprintf("<Error %q>", tok.data)
	}
}
