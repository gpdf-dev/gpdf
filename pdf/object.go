package pdf

import (
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Object is the common interface for all PDF objects that can serialize
// themselves into PDF binary format.
type Object interface {
	WriteTo(w io.Writer) (int64, error)
}

// ---------------------------------------------------------------------------
// ObjectRef - indirect reference (e.g., "1 0 R")
// ---------------------------------------------------------------------------

// ObjectRef represents a PDF indirect object reference such as "1 0 R".
type ObjectRef struct {
	Number     int
	Generation int
}

// WriteTo writes the indirect reference in the form "N G R".
func (r ObjectRef) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "%d %d R", r.Number, r.Generation)
	return int64(n), err
}

// ---------------------------------------------------------------------------
// Name - PDF name object (e.g., /Type)
// ---------------------------------------------------------------------------

// Name represents a PDF name object such as /Type or /Font.
type Name string

// WriteTo writes the name with a leading slash, e.g. "/Type".
// Characters outside the printable ASCII range and the '#' character
// are hex-encoded as per the PDF specification.
func (n Name) WriteTo(w io.Writer) (int64, error) {
	var b strings.Builder
	b.WriteByte('/')
	for i := 0; i < len(string(n)); i++ {
		ch := n[i]
		// Encode characters that are not regular printable ASCII,
		// or that are PDF delimiters / whitespace, or '#'.
		if ch < '!' || ch > '~' || ch == '#' || ch == '/' ||
			ch == '(' || ch == ')' || ch == '<' || ch == '>' ||
			ch == '[' || ch == ']' || ch == '{' || ch == '}' || ch == '%' {
			fmt.Fprintf(&b, "#%02X", ch)
		} else {
			b.WriteByte(ch)
		}
	}
	written, err := io.WriteString(w, b.String())
	return int64(written), err
}

// ---------------------------------------------------------------------------
// String types
// ---------------------------------------------------------------------------

// LiteralString represents a PDF literal string enclosed in parentheses,
// e.g. (Hello World). Special characters are escaped.
type LiteralString string

// WriteTo writes the literal string with proper escaping of (, ), and \.
func (s LiteralString) WriteTo(w io.Writer) (int64, error) {
	var b strings.Builder
	b.WriteByte('(')
	for _, ch := range []byte(s) {
		switch ch {
		case '(':
			b.WriteString(`\(`)
		case ')':
			b.WriteString(`\)`)
		case '\\':
			b.WriteString(`\\`)
		case '\r':
			b.WriteString(`\r`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteByte(ch)
		}
	}
	b.WriteByte(')')
	written, err := io.WriteString(w, b.String())
	return int64(written), err
}

// HexString represents a PDF hexadecimal string enclosed in angle brackets,
// e.g. <48656C6C6F>.
type HexString string

// WriteTo writes the hex-encoded string in the form <hex>.
func (s HexString) WriteTo(w io.Writer) (int64, error) {
	encoded := hex.EncodeToString([]byte(s))
	n, err := fmt.Fprintf(w, "<%s>", strings.ToUpper(encoded))
	return int64(n), err
}

// ---------------------------------------------------------------------------
// Number types
// ---------------------------------------------------------------------------

// Integer represents a PDF integer object.
type Integer int

// WriteTo writes the integer as a decimal number.
func (i Integer) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, strconv.Itoa(int(i)))
	return int64(n), err
}

// Real represents a PDF real (floating-point) number.
type Real float64

// WriteTo writes the real number with enough precision to represent it
// accurately. Trailing zeros after the decimal point are trimmed.
func (r Real) WriteTo(w io.Writer) (int64, error) {
	s := strconv.FormatFloat(float64(r), 'f', -1, 64)
	n, err := io.WriteString(w, s)
	return int64(n), err
}

// ---------------------------------------------------------------------------
// Boolean
// ---------------------------------------------------------------------------

// Boolean represents a PDF boolean value (true or false).
type Boolean bool

// WriteTo writes "true" or "false".
func (b Boolean) WriteTo(w io.Writer) (int64, error) {
	var s string
	if b {
		s = "true"
	} else {
		s = "false"
	}
	n, err := io.WriteString(w, s)
	return int64(n), err
}

// ---------------------------------------------------------------------------
// Null
// ---------------------------------------------------------------------------

// Null represents the PDF null object.
type Null struct{}

// WriteTo writes the keyword "null".
func (n Null) WriteTo(w io.Writer) (int64, error) {
	written, err := io.WriteString(w, "null")
	return int64(written), err
}

// ---------------------------------------------------------------------------
// Dict - PDF dictionary
// ---------------------------------------------------------------------------

// Dict represents a PDF dictionary object mapping Name keys to Object values.
type Dict map[Name]Object

// WriteTo writes the dictionary in the form << /Key value ... >>.
// Keys are written in sorted order for deterministic output.
func (d Dict) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}

	if _, err := io.WriteString(cw, "<<"); err != nil {
		return cw.n, err
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := io.WriteString(cw, " "); err != nil {
			return cw.n, err
		}
		name := Name(k)
		if _, err := name.WriteTo(cw); err != nil {
			return cw.n, err
		}
		if _, err := io.WriteString(cw, " "); err != nil {
			return cw.n, err
		}
		if _, err := d[name].WriteTo(cw); err != nil {
			return cw.n, err
		}
	}

	if _, err := io.WriteString(cw, " >>"); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// ---------------------------------------------------------------------------
// Array - PDF array
// ---------------------------------------------------------------------------

// Array represents a PDF array object containing a list of Objects.
type Array []Object

// WriteTo writes the array in the form [ item1 item2 ... ].
func (a Array) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}

	if _, err := io.WriteString(cw, "["); err != nil {
		return cw.n, err
	}
	for i, obj := range a {
		if i > 0 {
			if _, err := io.WriteString(cw, " "); err != nil {
				return cw.n, err
			}
		}
		if _, err := obj.WriteTo(cw); err != nil {
			return cw.n, err
		}
	}
	if _, err := io.WriteString(cw, "]"); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// ---------------------------------------------------------------------------
// Stream - PDF stream with dict + content
// ---------------------------------------------------------------------------

// Stream represents a PDF stream object consisting of a dictionary
// and binary content data.
type Stream struct {
	Dict    Dict
	Content []byte
}

// WriteTo writes the stream in the form:
//
//	<< dict >> stream\n...content...\nendstream
func (s Stream) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}

	// Ensure /Length is set in the dictionary.
	d := make(Dict, len(s.Dict)+1)
	for k, v := range s.Dict {
		d[k] = v
	}
	d[Name("Length")] = Integer(len(s.Content))

	if _, err := d.WriteTo(cw); err != nil {
		return cw.n, err
	}
	if _, err := io.WriteString(cw, "\nstream\n"); err != nil {
		return cw.n, err
	}
	if _, err := cw.Write(s.Content); err != nil {
		return cw.n, err
	}
	if _, err := io.WriteString(cw, "\nendstream"); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// ---------------------------------------------------------------------------
// Rectangle - PDF rectangle [llx lly urx ury]
// ---------------------------------------------------------------------------

// Rectangle represents a PDF rectangle defined by its lower-left and
// upper-right corners: [LLX LLY URX URY].
type Rectangle struct {
	LLX, LLY, URX, URY float64
}

// WriteTo writes the rectangle as a PDF array [llx lly urx ury].
func (r Rectangle) WriteTo(w io.Writer) (int64, error) {
	arr := Array{Real(r.LLX), Real(r.LLY), Real(r.URX), Real(r.URY)}
	return arr.WriteTo(w)
}

// ---------------------------------------------------------------------------
// ResourceDict
// ---------------------------------------------------------------------------

// ResourceDict represents a PDF resource dictionary containing
// font, XObject, and other resource references.
type ResourceDict struct {
	Font       Dict
	XObject    Dict
	ExtGState  Dict
	ColorSpace Dict
	Pattern    Dict
}

// ToDict converts the ResourceDict to a PDF Dict for serialization.
func (rd ResourceDict) ToDict() Dict {
	d := make(Dict)
	if len(rd.Font) > 0 {
		d[Name("Font")] = rd.Font
	}
	if len(rd.XObject) > 0 {
		d[Name("XObject")] = rd.XObject
	}
	if len(rd.ExtGState) > 0 {
		d[Name("ExtGState")] = rd.ExtGState
	}
	if len(rd.ColorSpace) > 0 {
		d[Name("ColorSpace")] = rd.ColorSpace
	}
	if len(rd.Pattern) > 0 {
		d[Name("Pattern")] = rd.Pattern
	}
	return d
}

// ---------------------------------------------------------------------------
// PageObject
// ---------------------------------------------------------------------------

// PageObject represents a PDF page with its media box, resources, and
// content stream references.
type PageObject struct {
	MediaBox  Rectangle
	Resources ResourceDict
	Contents  []ObjectRef
}

// ---------------------------------------------------------------------------
// DocumentInfo
// ---------------------------------------------------------------------------

// DocumentInfo holds optional PDF document metadata such as title, author,
// and producer.
type DocumentInfo struct {
	Title    string
	Author   string
	Subject  string
	Creator  string
	Producer string
}

// ToDict converts the DocumentInfo to a PDF Dict for serialization.
// Only non-empty fields are included.
func (di DocumentInfo) ToDict() Dict {
	d := make(Dict)
	if di.Title != "" {
		d[Name("Title")] = LiteralString(di.Title)
	}
	if di.Author != "" {
		d[Name("Author")] = LiteralString(di.Author)
	}
	if di.Subject != "" {
		d[Name("Subject")] = LiteralString(di.Subject)
	}
	if di.Creator != "" {
		d[Name("Creator")] = LiteralString(di.Creator)
	}
	if di.Producer != "" {
		d[Name("Producer")] = LiteralString(di.Producer)
	}
	return d
}

// ---------------------------------------------------------------------------
// countingWriter - internal helper
// ---------------------------------------------------------------------------

// countingWriter wraps an io.Writer and counts the total bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}
