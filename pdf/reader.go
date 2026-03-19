package pdf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Reader reads and parses an existing PDF document from a byte slice.
// It provides random access to objects via the cross-reference table
// and can traverse the page tree to enumerate pages.
type Reader struct {
	data    []byte
	xref    map[int]int64  // object number -> byte offset
	trailer Dict           // the trailer dictionary
	root    ObjectRef      // catalog reference
	pages   []PageInfo     // flattened page list (populated lazily)
	cache   map[int]Object // parsed object cache
}

// PageInfo describes a page in the existing PDF.
type PageInfo struct {
	Ref      ObjectRef // indirect reference to the page object
	MediaBox Rectangle // page dimensions
	Index    int       // 0-based page index
}

// NewReader creates a Reader from raw PDF data.
// It parses the xref table and trailer to enable object lookups.
func NewReader(data []byte) (*Reader, error) {
	r := &Reader{
		data:  data,
		xref:  make(map[int]int64),
		cache: make(map[int]Object),
	}

	if err := r.parseXRefAndTrailer(); err != nil {
		return nil, fmt.Errorf("pdf: %w", err)
	}

	// Extract /Root reference.
	rootObj, ok := r.trailer[Name("Root")]
	if !ok {
		return nil, fmt.Errorf("pdf: trailer missing /Root")
	}
	rootRef, ok := rootObj.(ObjectRef)
	if !ok {
		return nil, fmt.Errorf("pdf: /Root is not an indirect reference")
	}
	r.root = rootRef

	return r, nil
}

// PageCount returns the number of pages in the document.
func (r *Reader) PageCount() (int, error) {
	if err := r.loadPages(); err != nil {
		return 0, err
	}
	return len(r.pages), nil
}

// Page returns information about the i-th page (0-based).
func (r *Reader) Page(i int) (PageInfo, error) {
	if err := r.loadPages(); err != nil {
		return PageInfo{}, err
	}
	if i < 0 || i >= len(r.pages) {
		return PageInfo{}, fmt.Errorf("pdf: page index %d out of range [0, %d)", i, len(r.pages))
	}
	return r.pages[i], nil
}

// GetObject reads and parses the indirect object with the given number.
// Results are cached for repeated lookups.
func (r *Reader) GetObject(objNum int) (Object, error) {
	if obj, ok := r.cache[objNum]; ok {
		return obj, nil
	}

	offset, ok := r.xref[objNum]
	if !ok {
		return nil, fmt.Errorf("pdf: object %d not found in xref", objNum)
	}

	obj, err := r.parseIndirectObjectAt(offset)
	if err != nil {
		return nil, err
	}
	r.cache[objNum] = obj
	return obj, nil
}

// Resolve dereferences an Object: if it is an ObjectRef, fetch the actual object.
// Non-reference objects are returned as-is.
func (r *Reader) Resolve(obj Object) (Object, error) {
	ref, ok := obj.(ObjectRef)
	if !ok {
		return obj, nil
	}
	return r.GetObject(ref.Number)
}

// ResolveDict resolves an object and asserts it is a Dict.
func (r *Reader) ResolveDict(obj Object) (Dict, error) {
	resolved, err := r.Resolve(obj)
	if err != nil {
		return nil, err
	}
	switch v := resolved.(type) {
	case Dict:
		return v, nil
	case Stream:
		return v.Dict, nil
	default:
		return nil, fmt.Errorf("pdf: expected dict, got %T", resolved)
	}
}

// Trailer returns the trailer dictionary.
func (r *Reader) Trailer() Dict {
	return r.trailer
}

// RootRef returns the catalog object reference.
func (r *Reader) RootRef() ObjectRef {
	return r.root
}

// Data returns the raw PDF data.
func (r *Reader) Data() []byte {
	return r.data
}

// MaxObjectNumber returns the highest object number found in the xref table.
func (r *Reader) MaxObjectNumber() int {
	max := 0
	for n := range r.xref {
		if n > max {
			max = n
		}
	}
	return max
}

// PageDict returns the raw dictionary for the i-th page.
func (r *Reader) PageDict(i int) (Dict, error) {
	info, err := r.Page(i)
	if err != nil {
		return nil, err
	}
	return r.ResolveDict(info.Ref)
}

// ---------------------------------------------------------------------------
// Internal: xref + trailer parsing
// ---------------------------------------------------------------------------

// parseXRefAndTrailer locates and parses the xref table and trailer.
// It follows /Prev links to handle incremental updates.
func (r *Reader) parseXRefAndTrailer() error {
	// Find startxref.
	startxrefOffset, err := r.findStartXRef()
	if err != nil {
		return err
	}

	// Parse xref tables following /Prev chain (newest first).
	visited := make(map[int64]bool)
	offset := startxrefOffset
	var firstTrailer Dict

	for !visited[offset] {
		visited[offset] = true

		p := newParser(r.data)
		p.pos = int(offset)
		p.skipWhitespaceAndComments()

		// Check if this is a cross-reference stream (PDF 1.5+) or traditional xref table.
		if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "xref" {
			trailer, err := r.parseXRefTable(p)
			if err != nil {
				return err
			}
			if firstTrailer == nil {
				firstTrailer = trailer
			}

			// Follow /Prev if present.
			if prev, ok := trailer[Name("Prev")]; ok {
				if prevInt, ok := prev.(Integer); ok {
					offset = int64(prevInt)
					continue
				}
			}
			break
		}

		// Cross-reference stream.
		trailer, err := r.parseXRefStream(p)
		if err != nil {
			return fmt.Errorf("xref stream: %w", err)
		}
		if firstTrailer == nil {
			firstTrailer = trailer
		}

		if prev, ok := trailer[Name("Prev")]; ok {
			if prevInt, ok := prev.(Integer); ok {
				offset = int64(prevInt)
				continue
			}
		}
		break
	}

	if firstTrailer == nil {
		return fmt.Errorf("no trailer found")
	}
	r.trailer = firstTrailer
	return nil
}

// findStartXRef searches backward from the end of file for "startxref"
// and returns the byte offset value that follows it.
func (r *Reader) findStartXRef() (int64, error) {
	// Search in the last 1024 bytes.
	searchLen := 1024
	start := len(r.data) - searchLen
	if start < 0 {
		start = 0
	}
	tail := r.data[start:]

	idx := bytes.LastIndex(tail, []byte("startxref"))
	if idx < 0 {
		return 0, fmt.Errorf("startxref not found")
	}

	// Parse the offset number after "startxref".
	p := newParser(tail)
	p.pos = idx + len("startxref")
	p.skipWhitespaceAndComments()

	numStr, _, err := p.scanNumber()
	if err != nil {
		return 0, fmt.Errorf("could not read startxref offset: %w", err)
	}
	offset, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid startxref offset %q: %w", numStr, err)
	}
	return offset, nil
}

// parseXRefTable parses a traditional xref table and the following trailer dict.
func (r *Reader) parseXRefTable(p *parser) (Dict, error) {
	// Skip "xref".
	p.pos += 4
	p.skipWhitespaceAndComments()

	// Parse subsections: "startObj count"
	for {
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			return nil, fmt.Errorf("unexpected end in xref table")
		}
		// Stop at "trailer".
		if p.pos+7 <= len(p.data) && string(p.data[p.pos:p.pos+7]) == "trailer" {
			break
		}

		startStr, _, err := p.scanNumber()
		if err != nil {
			return nil, fmt.Errorf("xref subsection start: %w", err)
		}
		startObj, _ := strconv.Atoi(startStr)

		p.skipWhitespaceAndComments()
		countStr, _, err := p.scanNumber()
		if err != nil {
			return nil, fmt.Errorf("xref subsection count: %w", err)
		}
		count, _ := strconv.Atoi(countStr)

		for i := 0; i < count; i++ {
			p.skipWhitespaceAndComments()
			offsetStr, _, err := p.scanNumber()
			if err != nil {
				return nil, fmt.Errorf("xref entry offset: %w", err)
			}
			p.skipWhitespaceAndComments()
			genStr, _, err := p.scanNumber()
			if err != nil {
				return nil, fmt.Errorf("xref entry generation: %w", err)
			}
			_ = genStr

			p.skipWhitespaceAndComments()
			if p.atEnd() {
				return nil, fmt.Errorf("unexpected end in xref entry")
			}
			marker := p.data[p.pos]
			p.pos++

			objNum := startObj + i
			if marker == 'n' {
				// Only store if not already set (newest wins).
				if _, exists := r.xref[objNum]; !exists {
					off, _ := strconv.ParseInt(offsetStr, 10, 64)
					r.xref[objNum] = off
				}
			}
		}
	}

	// Parse trailer dict.
	p.pos += 7 // skip "trailer"
	p.skipWhitespaceAndComments()

	obj, err := p.parseObject()
	if err != nil {
		return nil, fmt.Errorf("trailer dict: %w", err)
	}
	trailer, ok := obj.(Dict)
	if !ok {
		return nil, fmt.Errorf("trailer is not a dict")
	}
	return trailer, nil
}

// parseXRefStream parses a cross-reference stream object (PDF 1.5+).
func (r *Reader) parseXRefStream(p *parser) (Dict, error) {
	// The xref stream is an indirect object: "N G obj ... stream ... endstream endobj"
	obj, err := r.parseIndirectObjectAt(int64(p.pos))
	if err != nil {
		return nil, fmt.Errorf("parse xref stream object: %w", err)
	}

	stream, ok := obj.(Stream)
	if !ok {
		return nil, fmt.Errorf("xref stream is not a stream object")
	}

	// Decompress stream content.
	content, err := r.decodeStreamContent(stream)
	if err != nil {
		return nil, fmt.Errorf("decode xref stream: %w", err)
	}

	w, err := parseXRefFieldWidths(stream.Dict)
	if err != nil {
		return nil, err
	}
	entrySize := w[0] + w[1] + w[2]
	if entrySize == 0 {
		return stream.Dict, nil
	}

	indices := parseXRefIndices(stream.Dict)
	r.parseXRefEntries(content, indices, w, entrySize)

	return stream.Dict, nil
}

// parseXRefFieldWidths extracts and validates the /W array from an xref stream dict.
func parseXRefFieldWidths(d Dict) ([3]int, error) {
	wObj, ok := d[Name("W")]
	if !ok {
		return [3]int{}, fmt.Errorf("xref stream missing /W")
	}
	wArr, ok := wObj.(Array)
	if !ok || len(wArr) != 3 {
		return [3]int{}, fmt.Errorf("xref stream /W must be array of 3")
	}
	var w [3]int
	for i, a := range wArr {
		if v, ok := a.(Integer); ok {
			w[i] = int(v)
		}
	}
	return w, nil
}

// parseXRefIndices extracts the /Index array from an xref stream dict.
// Returns [0, Size] as default if /Index is absent.
func parseXRefIndices(d Dict) []int {
	sizeObj := d[Name("Size")]
	size := 0
	if v, ok := sizeObj.(Integer); ok {
		size = int(v)
	}

	indices := []int{0, size}
	if idxObj, ok := d[Name("Index")]; ok {
		if idxArr, ok := idxObj.(Array); ok {
			indices = make([]int, len(idxArr))
			for i, a := range idxArr {
				if v, ok := a.(Integer); ok {
					indices[i] = int(v)
				}
			}
		}
	}
	return indices
}

// parseXRefEntries reads xref entries from stream content and populates the xref table.
func (r *Reader) parseXRefEntries(content []byte, indices []int, w [3]int, entrySize int) {
	pos := 0
	for i := 0; i < len(indices)-1; i += 2 {
		startObj := indices[i]
		count := indices[i+1]
		for j := 0; j < count; j++ {
			if pos+entrySize > len(content) {
				break
			}
			fields := readXRefFields(content[pos:pos+entrySize], w)
			pos += entrySize

			objNum := startObj + j
			if fields[0] == 1 {
				if _, exists := r.xref[objNum]; !exists {
					r.xref[objNum] = fields[1]
				}
			}
		}
	}
}

// readXRefFields reads the three fields from a single xref stream entry.
func readXRefFields(data []byte, w [3]int) [3]int64 {
	var fields [3]int64
	offset := 0
	for f := 0; f < 3; f++ {
		for k := 0; k < w[f]; k++ {
			fields[f] = fields[f]<<8 | int64(data[offset])
			offset++
		}
		if f == 0 && w[0] == 0 {
			fields[0] = 1
		}
	}
	return fields
}

// parseIndirectObjectAt parses "N G obj ... endobj" at the given offset.
func (r *Reader) parseIndirectObjectAt(offset int64) (Object, error) {
	p := newParser(r.data)
	p.pos = int(offset)
	p.skipWhitespaceAndComments()

	// Parse "N G obj".
	_, _, err := p.scanNumber() // object number
	if err != nil {
		return nil, fmt.Errorf("indirect object number at %d: %w", offset, err)
	}
	p.skipWhitespaceAndComments()
	_, _, err = p.scanNumber() // generation number
	if err != nil {
		return nil, fmt.Errorf("indirect object generation at %d: %w", offset, err)
	}
	p.skipWhitespaceAndComments()

	// Expect "obj".
	if p.pos+3 > len(p.data) || string(p.data[p.pos:p.pos+3]) != "obj" {
		return nil, fmt.Errorf("expected 'obj' at offset %d", p.pos)
	}
	p.pos += 3

	// Parse the object value.
	obj, err := p.parseObject()
	if err != nil {
		return nil, err
	}

	// If the object is a dict that was followed by stream, it's already a Stream.
	// Otherwise skip to "endobj".
	p.skipWhitespaceAndComments()
	if p.pos+6 <= len(p.data) && string(p.data[p.pos:p.pos+6]) == "endobj" {
		p.pos += 6
	}

	// For streams, resolve /Length if it was an indirect reference.
	if s, ok := obj.(Stream); ok {
		if lenRef, ok := s.Dict[Name("Length")].(ObjectRef); ok {
			lenObj, err := r.GetObject(lenRef.Number)
			if err == nil {
				s.Dict[Name("Length")] = lenObj
			}
		}
	}

	return obj, nil
}

// decodeStreamContent decompresses a stream's content based on /Filter.
func (r *Reader) decodeStreamContent(s Stream) ([]byte, error) {
	filterObj, hasFilter := s.Dict[Name("Filter")]
	if !hasFilter {
		return s.Content, nil
	}

	filters := []string{}
	switch f := filterObj.(type) {
	case Name:
		filters = []string{string(f)}
	case Array:
		for _, item := range f {
			if n, ok := item.(Name); ok {
				filters = append(filters, string(n))
			}
		}
	}

	data := s.Content
	for _, filter := range filters {
		switch filter {
		case "FlateDecode":
			decoded, err := decompressFlate(data)
			if err != nil {
				return nil, err
			}
			data = decoded
		default:
			return nil, fmt.Errorf("pdf: unsupported filter %q", filter)
		}
	}
	return data, nil
}

// ---------------------------------------------------------------------------
// Internal: page tree traversal
// ---------------------------------------------------------------------------

// loadPages traverses the page tree and populates the flat page list.
func (r *Reader) loadPages() error {
	if r.pages != nil {
		return nil
	}

	catalog, err := r.ResolveDict(r.root)
	if err != nil {
		return fmt.Errorf("resolve catalog: %w", err)
	}
	pagesRef, ok := catalog[Name("Pages")]
	if !ok {
		return fmt.Errorf("pdf: catalog missing /Pages")
	}

	r.pages = []PageInfo{}
	return r.walkPageTree(pagesRef, nil)
}

// walkPageTree recursively walks the page tree, collecting leaf Page nodes.
// inheritedMediaBox propagates /MediaBox from parent Pages nodes.
func (r *Reader) walkPageTree(node Object, inheritedMediaBox *Rectangle) error {
	d, err := r.ResolveDict(node)
	if err != nil {
		return err
	}

	// Get MediaBox if present (inheritable).
	mediaBox := inheritedMediaBox
	if mbObj, ok := d[Name("MediaBox")]; ok {
		mb, err := r.parseRectangle(mbObj)
		if err == nil {
			mediaBox = &mb
		}
	}

	typeObj := d[Name("Type")]
	typeName, _ := typeObj.(Name)

	switch string(typeName) {
	case "Pages":
		kids, ok := d[Name("Kids")]
		if !ok {
			return nil
		}
		kidsResolved, err := r.Resolve(kids)
		if err != nil {
			return err
		}
		kidsArr, ok := kidsResolved.(Array)
		if !ok {
			return fmt.Errorf("pdf: /Kids is not an array")
		}
		for _, kid := range kidsArr {
			if err := r.walkPageTree(kid, mediaBox); err != nil {
				return err
			}
		}
	case "Page":
		ref, _ := node.(ObjectRef)
		mb := Rectangle{URX: 612, URY: 792} // default US Letter
		if mediaBox != nil {
			mb = *mediaBox
		}
		r.pages = append(r.pages, PageInfo{
			Ref:      ref,
			MediaBox: mb,
			Index:    len(r.pages),
		})
	default:
		// Try to determine by presence of /Kids (Pages) vs /Contents (Page).
		if _, hasKids := d[Name("Kids")]; hasKids {
			return r.walkPageTree(node, mediaBox)
		}
		ref, _ := node.(ObjectRef)
		mb := Rectangle{URX: 612, URY: 792}
		if mediaBox != nil {
			mb = *mediaBox
		}
		r.pages = append(r.pages, PageInfo{
			Ref:      ref,
			MediaBox: mb,
			Index:    len(r.pages),
		})
	}
	return nil
}

// parseRectangle converts a PDF array to a Rectangle.
func (r *Reader) parseRectangle(obj Object) (Rectangle, error) {
	resolved, err := r.Resolve(obj)
	if err != nil {
		return Rectangle{}, err
	}
	arr, ok := resolved.(Array)
	if !ok || len(arr) != 4 {
		return Rectangle{}, fmt.Errorf("pdf: rectangle must be array of 4")
	}
	vals := [4]float64{}
	for i, item := range arr {
		switch v := item.(type) {
		case Integer:
			vals[i] = float64(v)
		case Real:
			vals[i] = float64(v)
		default:
			return Rectangle{}, fmt.Errorf("pdf: rectangle element %d is not a number", i)
		}
	}
	return Rectangle{LLX: vals[0], LLY: vals[1], URX: vals[2], URY: vals[3]}, nil
}

// ---------------------------------------------------------------------------
// String representation helpers (for debugging)
// ---------------------------------------------------------------------------

// String returns a summary of the reader state.
func (r *Reader) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Reader: %d bytes, %d objects in xref", len(r.data), len(r.xref))
	if r.pages != nil {
		fmt.Fprintf(&b, ", %d pages", len(r.pages))
	}
	return b.String()
}
