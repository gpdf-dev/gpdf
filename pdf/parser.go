package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strconv"
)

// parser reads PDF tokens and objects from a byte slice.
// It maintains a current position and supports random-access seeking.
type parser struct {
	data []byte
	pos  int
}

// newParser creates a new parser over the given data.
func newParser(data []byte) *parser {
	return &parser{data: data}
}

// atEnd returns true if the parser has reached the end of data.
func (p *parser) atEnd() bool {
	return p.pos >= len(p.data)
}

// peek returns the byte at the current position without advancing.
func (p *parser) peek() byte {
	if p.pos >= len(p.data) {
		return 0
	}
	return p.data[p.pos]
}

// skipWhitespaceAndComments advances past whitespace and PDF comments (% ... EOL).
func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.data) {
		ch := p.data[p.pos]
		if isWhitespace(ch) {
			p.pos++
			continue
		}
		if ch == '%' {
			// Skip until end of line.
			for p.pos < len(p.data) && p.data[p.pos] != '\n' && p.data[p.pos] != '\r' {
				p.pos++
			}
			continue
		}
		break
	}
}

// parseObject parses the next PDF object at the current position.
// It handles all PDF types: bool, null, integer, real, name, string,
// hex string, array, dict, and indirect references.
func (p *parser) parseObject() (Object, error) {
	p.skipWhitespaceAndComments()
	if p.atEnd() {
		return nil, fmt.Errorf("pdf: unexpected end of data")
	}

	ch := p.peek()

	switch {
	case ch == '/':
		return p.parseName()
	case ch == '(':
		return p.parseLiteralString()
	case ch == '<':
		if p.pos+1 < len(p.data) && p.data[p.pos+1] == '<' {
			return p.parseDict()
		}
		return p.parseHexString()
	case ch == '[':
		return p.parseArray()
	case ch == 't' || ch == 'f':
		return p.parseBoolOrKeyword()
	case ch == 'n':
		return p.parseNull()
	case ch == '-' || ch == '+' || ch == '.' || isDigit(ch):
		return p.parseNumberOrRef()
	default:
		return nil, fmt.Errorf("pdf: unexpected character %q at offset %d", ch, p.pos)
	}
}

// parseName parses a PDF name object starting with '/'.
func (p *parser) parseName() (Name, error) {
	if p.data[p.pos] != '/' {
		return "", fmt.Errorf("pdf: expected '/' at offset %d", p.pos)
	}
	p.pos++ // skip '/'

	var b []byte
	for p.pos < len(p.data) {
		ch := p.data[p.pos]
		if isWhitespace(ch) || isDelimiter(ch) {
			break
		}
		if ch == '#' && p.pos+2 < len(p.data) {
			hi := unhex(p.data[p.pos+1])
			lo := unhex(p.data[p.pos+2])
			if hi >= 0 && lo >= 0 {
				b = append(b, byte(hi<<4|lo))
				p.pos += 3
				continue
			}
		}
		b = append(b, ch)
		p.pos++
	}
	return Name(b), nil
}

// parseLiteralString parses a PDF literal string enclosed in parentheses.
func (p *parser) parseLiteralString() (LiteralString, error) {
	if p.data[p.pos] != '(' {
		return "", fmt.Errorf("pdf: expected '(' at offset %d", p.pos)
	}
	p.pos++

	var b []byte
	depth := 1
	for p.pos < len(p.data) && depth > 0 {
		ch := p.data[p.pos]
		switch ch {
		case '(':
			depth++
			b = append(b, ch)
			p.pos++
		case ')':
			depth--
			if depth > 0 {
				b = append(b, ch)
			}
			p.pos++
		case '\\':
			escaped, skip, err := p.parseStringEscape()
			if err != nil {
				return "", err
			}
			if skip {
				continue
			}
			b = append(b, escaped)
		default:
			b = append(b, ch)
			p.pos++
		}
	}
	return LiteralString(b), nil
}

// parseStringEscape handles a backslash escape sequence in a literal string.
// It returns the decoded byte, whether to skip appending (line continuation), and any error.
func (p *parser) parseStringEscape() (byte, bool, error) {
	p.pos++ // skip backslash
	if p.pos >= len(p.data) {
		return 0, false, fmt.Errorf("pdf: unexpected end in string escape")
	}
	esc := p.data[p.pos]
	switch esc {
	case 'n':
		p.pos++
		return '\n', false, nil
	case 'r':
		p.pos++
		return '\r', false, nil
	case 't':
		p.pos++
		return '\t', false, nil
	case 'b':
		p.pos++
		return '\b', false, nil
	case 'f':
		p.pos++
		return '\f', false, nil
	case '(', ')', '\\':
		p.pos++
		return esc, false, nil
	case '\r':
		// Line continuation: skip \r and optional \n.
		p.pos++
		if p.pos < len(p.data) && p.data[p.pos] == '\n' {
			p.pos++
		}
		return 0, true, nil
	case '\n':
		// Line continuation: skip \n.
		p.pos++
		return 0, true, nil
	default:
		return p.parseOctalOrLiteral(esc)
	}
}

// parseOctalOrLiteral parses an octal escape or returns the byte as-is.
func (p *parser) parseOctalOrLiteral(esc byte) (byte, bool, error) {
	if esc >= '0' && esc <= '7' {
		oct := int(esc - '0')
		for i := 0; i < 2 && p.pos+1 < len(p.data); i++ {
			next := p.data[p.pos+1]
			if next < '0' || next > '7' {
				break
			}
			p.pos++
			oct = oct*8 + int(next-'0')
		}
		p.pos++
		return byte(oct), false, nil
	}
	p.pos++
	return esc, false, nil
}

// parseHexString parses a PDF hex string enclosed in angle brackets.
func (p *parser) parseHexString() (HexString, error) {
	if p.data[p.pos] != '<' {
		return "", fmt.Errorf("pdf: expected '<' at offset %d", p.pos)
	}
	p.pos++

	var hexChars []byte
	for p.pos < len(p.data) {
		ch := p.data[p.pos]
		if ch == '>' {
			p.pos++
			break
		}
		if isWhitespace(ch) {
			p.pos++
			continue
		}
		hexChars = append(hexChars, ch)
		p.pos++
	}

	// If odd number of hex digits, append a trailing 0.
	if len(hexChars)%2 != 0 {
		hexChars = append(hexChars, '0')
	}

	decoded := make([]byte, len(hexChars)/2)
	for i := 0; i < len(hexChars); i += 2 {
		hi := unhex(hexChars[i])
		lo := unhex(hexChars[i+1])
		if hi < 0 || lo < 0 {
			return "", fmt.Errorf("pdf: invalid hex digit in hex string")
		}
		decoded[i/2] = byte(hi<<4 | lo)
	}
	return HexString(decoded), nil
}

// parseArray parses a PDF array [ ... ].
func (p *parser) parseArray() (Array, error) {
	if p.data[p.pos] != '[' {
		return nil, fmt.Errorf("pdf: expected '[' at offset %d", p.pos)
	}
	p.pos++

	var arr Array
	for {
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			return nil, fmt.Errorf("pdf: unexpected end of array")
		}
		if p.data[p.pos] == ']' {
			p.pos++
			return arr, nil
		}
		obj, err := p.parseObject()
		if err != nil {
			return nil, err
		}
		arr = append(arr, obj)
	}
}

// parseDict parses a PDF dictionary << ... >>.
// If followed by "stream", it parses the stream content as well.
func (p *parser) parseDict() (Object, error) {
	if p.pos+1 >= len(p.data) || p.data[p.pos] != '<' || p.data[p.pos+1] != '<' {
		return nil, fmt.Errorf("pdf: expected '<<' at offset %d", p.pos)
	}
	p.pos += 2

	d := make(Dict)
	for {
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			return nil, fmt.Errorf("pdf: unexpected end of dict")
		}
		if p.data[p.pos] == '>' && p.pos+1 < len(p.data) && p.data[p.pos+1] == '>' {
			p.pos += 2
			break
		}
		key, err := p.parseName()
		if err != nil {
			return nil, fmt.Errorf("pdf: dict key: %w", err)
		}
		val, err := p.parseObject()
		if err != nil {
			return nil, fmt.Errorf("pdf: dict value for /%s: %w", key, err)
		}
		d[key] = val
	}

	// Check if followed by "stream".
	saved := p.pos
	p.skipWhitespaceAndComments()
	if p.pos+6 <= len(p.data) && string(p.data[p.pos:p.pos+6]) == "stream" {
		return p.parseStream(d)
	}
	p.pos = saved
	return d, nil
}

// parseStream parses a PDF stream, given its dictionary.
// The parser must be positioned at the "stream" keyword.
func (p *parser) parseStream(d Dict) (Stream, error) {
	// Skip "stream" keyword.
	p.pos += 6
	// Skip single EOL after "stream" (CR, LF, or CRLF).
	if p.pos < len(p.data) && p.data[p.pos] == '\r' {
		p.pos++
	}
	if p.pos < len(p.data) && p.data[p.pos] == '\n' {
		p.pos++
	}

	// Determine stream length.
	length := 0
	if lenObj, ok := d[Name("Length")]; ok {
		switch v := lenObj.(type) {
		case Integer:
			length = int(v)
		}
		// If Length is an indirect ref, we fall through to endstream scanning.
	}

	var content []byte
	if length > 0 && p.pos+length <= len(p.data) {
		content = make([]byte, length)
		copy(content, p.data[p.pos:p.pos+length])
		p.pos += length
	} else {
		// Scan for "endstream" marker.
		end := bytes.Index(p.data[p.pos:], []byte("endstream"))
		if end < 0 {
			return Stream{}, fmt.Errorf("pdf: endstream not found")
		}
		content = make([]byte, end)
		copy(content, p.data[p.pos:p.pos+end])
		p.pos += end
		// Trim trailing EOL before endstream.
		content = bytes.TrimRight(content, "\r\n")
	}

	// Skip "endstream".
	p.skipWhitespaceAndComments()
	if p.pos+9 <= len(p.data) && string(p.data[p.pos:p.pos+9]) == "endstream" {
		p.pos += 9
	}

	return Stream{Dict: d, Content: content}, nil
}

// parseNumberOrRef parses a number (integer or real) and possibly an indirect reference (N G R).
func (p *parser) parseNumberOrRef() (Object, error) {
	start := p.pos
	num, isReal, err := p.scanNumber()
	if err != nil {
		return nil, err
	}

	if isReal {
		f, _ := strconv.ParseFloat(num, 64)
		return Real(f), nil
	}

	// Try to parse as indirect reference: int int "R".
	saved := p.pos
	p.skipWhitespaceAndComments()
	if !p.atEnd() && isDigit(p.peek()) {
		gen, genIsReal, err := p.scanNumber()
		if err == nil && !genIsReal {
			p.skipWhitespaceAndComments()
			if !p.atEnd() && p.peek() == 'R' {
				// Check that R is not part of a longer keyword.
				if p.pos+1 >= len(p.data) || isWhitespace(p.data[p.pos+1]) || isDelimiter(p.data[p.pos+1]) {
					p.pos++ // skip 'R'
					n, _ := strconv.Atoi(num)
					g, _ := strconv.Atoi(gen)
					return ObjectRef{Number: n, Generation: g}, nil
				}
			}
		}
	}

	// Not a reference, restore position and return integer.
	p.pos = saved
	_ = start
	n, _ := strconv.ParseInt(num, 10, 64)
	return Integer(n), nil
}

// scanNumber scans a numeric token and returns it as a string,
// along with a flag indicating whether it contains a decimal point.
func (p *parser) scanNumber() (string, bool, error) {
	start := p.pos
	isReal := false

	if p.pos < len(p.data) && (p.data[p.pos] == '-' || p.data[p.pos] == '+') {
		p.pos++
	}
	if p.pos >= len(p.data) || (!isDigit(p.data[p.pos]) && p.data[p.pos] != '.') {
		return "", false, fmt.Errorf("pdf: expected number at offset %d", start)
	}
	for p.pos < len(p.data) && isDigit(p.data[p.pos]) {
		p.pos++
	}
	if p.pos < len(p.data) && p.data[p.pos] == '.' {
		isReal = true
		p.pos++
		for p.pos < len(p.data) && isDigit(p.data[p.pos]) {
			p.pos++
		}
	}
	return string(p.data[start:p.pos]), isReal, nil
}

// parseBoolOrKeyword parses "true" or "false".
func (p *parser) parseBoolOrKeyword() (Object, error) {
	if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "true" {
		next := p.pos + 4
		if next >= len(p.data) || isWhitespace(p.data[next]) || isDelimiter(p.data[next]) {
			p.pos = next
			return Boolean(true), nil
		}
	}
	if p.pos+5 <= len(p.data) && string(p.data[p.pos:p.pos+5]) == "false" {
		next := p.pos + 5
		if next >= len(p.data) || isWhitespace(p.data[next]) || isDelimiter(p.data[next]) {
			p.pos = next
			return Boolean(false), nil
		}
	}
	return nil, fmt.Errorf("pdf: unexpected keyword at offset %d", p.pos)
}

// parseNull parses the "null" keyword.
func (p *parser) parseNull() (Object, error) {
	if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "null" {
		next := p.pos + 4
		if next >= len(p.data) || isWhitespace(p.data[next]) || isDelimiter(p.data[next]) {
			p.pos = next
			return Null{}, nil
		}
	}
	return nil, fmt.Errorf("pdf: expected 'null' at offset %d", p.pos)
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f' || ch == 0
}

func isDelimiter(ch byte) bool {
	return ch == '(' || ch == ')' || ch == '<' || ch == '>' ||
		ch == '[' || ch == ']' || ch == '{' || ch == '}' ||
		ch == '/' || ch == '%'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func unhex(ch byte) int {
	switch {
	case ch >= '0' && ch <= '9':
		return int(ch - '0')
	case ch >= 'a' && ch <= 'f':
		return int(ch-'a') + 10
	case ch >= 'A' && ch <= 'F':
		return int(ch-'A') + 10
	default:
		return -1
	}
}

// decompressFlate decompresses zlib/deflate data.
func decompressFlate(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("pdf: zlib open: %w", err)
	}
	defer func() { _ = r.Close() }()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pdf: zlib decompress: %w", err)
	}
	return out, nil
}
