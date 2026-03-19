package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"testing"
)

func TestParseStringEscape(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		errMsg string
	}{
		{
			name:  "newline escape",
			input: `(Hello\nWorld)`,
			want:  "Hello\nWorld",
		},
		{
			name:  "carriage return escape",
			input: `(Hello\rWorld)`,
			want:  "Hello\rWorld",
		},
		{
			name:  "tab escape",
			input: `(Hello\tWorld)`,
			want:  "Hello\tWorld",
		},
		{
			name:  "backspace escape",
			input: `(Hello\bWorld)`,
			want:  "Hello\bWorld",
		},
		{
			name:  "form feed escape",
			input: `(Hello\fWorld)`,
			want:  "Hello\fWorld",
		},
		{
			name:  "escaped backslash",
			input: `(Hello\\World)`,
			want:  "Hello\\World",
		},
		{
			name:  "escaped open paren",
			input: `(Hello\(World)`,
			want:  "Hello(World",
		},
		{
			name:  "escaped close paren",
			input: `(Hello\)World)`,
			want:  "Hello)World",
		},
		{
			name:  "line continuation CR",
			input: "(Hello\\\rWorld)",
			want:  "HelloWorld",
		},
		{
			name:  "line continuation LF",
			input: "(Hello\\\nWorld)",
			want:  "HelloWorld",
		},
		{
			name:  "line continuation CRLF",
			input: "(Hello\\\r\nWorld)",
			want:  "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser([]byte(tt.input))
			obj, err := p.parseObject()
			if tt.errMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseObject: %v", err)
			}
			s, ok := obj.(LiteralString)
			if !ok {
				t.Fatalf("got %T, want LiteralString", obj)
			}
			if string(s) != tt.want {
				t.Errorf("got %q, want %q", s, tt.want)
			}
		})
	}
}

func TestParseOctalOrLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "octal 3 digits",
			input: "(\\101)", // \101 = 'A' (65)
			want:  "A",
		},
		{
			name:  "octal 2 digits",
			input: "(\\11)", // \11 = tab (9)
			want:  "\t",
		},
		{
			name:  "octal 1 digit",
			input: "(\\0)", // \0 = null byte
			want:  "\x00",
		},
		{
			name:  "unknown escape treated as literal",
			input: "(\\q)", // \q → q
			want:  "q",
		},
		{
			name:  "octal 110 = H",
			input: "(\\110)",
			want:  "H",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser([]byte(tt.input))
			obj, err := p.parseObject()
			if err != nil {
				t.Fatalf("parseObject: %v", err)
			}
			s, ok := obj.(LiteralString)
			if !ok {
				t.Fatalf("got %T, want LiteralString", obj)
			}
			if string(s) != tt.want {
				t.Errorf("got %q, want %q", s, tt.want)
			}
		})
	}
}

func TestParseStream(t *testing.T) {
	// Build a dict + stream inline.
	content := "BT /F1 12 Tf (Hello) Tj ET"
	input := []byte(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))

	p := newParser(input)
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject: %v", err)
	}
	s, ok := obj.(Stream)
	if !ok {
		t.Fatalf("got %T, want Stream", obj)
	}
	// The content should match.
	got := string(s.Content)
	if got != content {
		t.Errorf("stream content = %q, want %q", got, content)
	}
}

func TestParseStreamScanEndstream(t *testing.T) {
	// Stream without /Length (or Length=0) → scanner-based fallback.
	content := "q 1 0 0 1 0 0 cm Q"
	input := []byte("<< >>\nstream\n" + content + "\nendstream")

	p := newParser(input)
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject: %v", err)
	}
	s, ok := obj.(Stream)
	if !ok {
		t.Fatalf("got %T, want Stream", obj)
	}
	if string(s.Content) != content {
		t.Errorf("stream content = %q, want %q", s.Content, content)
	}
}

func TestParseStreamCRLF(t *testing.T) {
	// Stream with CRLF after "stream" keyword.
	content := "hello"
	input := []byte("<< /Length 5 >>\r\nstream\r\nhello\r\nendstream")

	p := newParser(input)
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject: %v", err)
	}
	s, ok := obj.(Stream)
	if !ok {
		t.Fatalf("got %T, want Stream", obj)
	}
	if string(s.Content) != content {
		t.Errorf("stream content = %q, want %q", s.Content, content)
	}
}

func TestDecompressFlate(t *testing.T) {
	original := []byte("Hello, this is a test of flate decompression in PDF streams!")

	// Compress with zlib.
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, _ = w.Write(original)
	_ = w.Close()

	// Decompress.
	result, err := decompressFlate(buf.Bytes())
	if err != nil {
		t.Fatalf("decompressFlate: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("decompressed = %q, want %q", result, original)
	}
}

func TestDecompressFlateInvalidData(t *testing.T) {
	_, err := decompressFlate([]byte("not valid zlib data"))
	if err == nil {
		t.Error("expected error for invalid zlib data")
	}
}

func TestParserSkipComments(t *testing.T) {
	// A comment line should be skipped.
	input := "% this is a comment\n42 "
	p := newParser([]byte(input))
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject: %v", err)
	}
	if v, ok := obj.(Integer); !ok || v != 42 {
		t.Errorf("got %v (%T), want Integer(42)", obj, obj)
	}
}

func TestParserPeekAtEnd(t *testing.T) {
	p := newParser([]byte{})
	if p.peek() != 0 {
		t.Errorf("peek at end = %d, want 0", p.peek())
	}
}

func TestParserUnexpectedCharacter(t *testing.T) {
	p := newParser([]byte("@"))
	_, err := p.parseObject()
	if err == nil {
		t.Error("expected error for unexpected character")
	}
}

func TestParserUnexpectedEndOfData(t *testing.T) {
	p := newParser([]byte(""))
	_, err := p.parseObject()
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestUnhexAllRanges(t *testing.T) {
	tests := []struct {
		input byte
		want  int
	}{
		{'0', 0}, {'9', 9},
		{'a', 10}, {'f', 15},
		{'A', 10}, {'F', 15},
		{'g', -1}, {'G', -1}, {'z', -1}, {' ', -1},
	}
	for _, tt := range tests {
		got := unhex(tt.input)
		if got != tt.want {
			t.Errorf("unhex(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestHexStringOddDigits(t *testing.T) {
	// Odd number of hex digits: trailing 0 appended.
	// <ABC> → <AB C0> → bytes 0xAB, 0xC0
	p := newParser([]byte("<ABC>"))
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject: %v", err)
	}
	hs, ok := obj.(HexString)
	if !ok {
		t.Fatalf("got %T, want HexString", obj)
	}
	if len(hs) != 2 || hs[0] != 0xAB || hs[1] != 0xC0 {
		t.Errorf("got %x, want abc0", []byte(hs))
	}
}

func TestHexStringWithWhitespace(t *testing.T) {
	p := newParser([]byte("< 48 65 6C 6C 6F >"))
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject: %v", err)
	}
	hs, ok := obj.(HexString)
	if !ok {
		t.Fatalf("got %T, want HexString", obj)
	}
	if string(hs) != "Hello" {
		t.Errorf("got %q, want %q", hs, "Hello")
	}
}
