package pdf

import (
	"fmt"
	"strconv"
)

// compressedObjRef locates an object stored inside a compressed object stream,
// as recorded by a type-2 cross-reference entry (ISO 32000-1 §7.5.8.3).
type compressedObjRef struct {
	streamNum int // object number of the containing /ObjStm
	index     int // index of the object within that stream
}

// objStmEntry describes one object inside a decoded object stream.
// start and end are absolute offsets into the decoded body.
type objStmEntry struct {
	num   int
	start int
	end   int
}

// objectStream is a decoded /ObjStm: the decompressed body together with the
// object number / offset pairs parsed from its header (ISO 32000-1 §7.5.7).
type objectStream struct {
	content []byte
	entries []objStmEntry
}

// getCompressedObject resolves an object stored inside an object stream.
func (r *Reader) getCompressedObject(objNum int, loc compressedObjRef) (Object, error) {
	stm, err := r.loadObjectStream(loc.streamNum)
	if err != nil {
		return nil, fmt.Errorf("pdf: object %d: %w", objNum, err)
	}

	entry, ok := stm.entryFor(objNum, loc.index)
	if !ok {
		return nil, fmt.Errorf("pdf: object %d not found in object stream %d", objNum, loc.streamNum)
	}

	// Parse the object from its own slice of the body. Bounding the parser to
	// the entry keeps it from reading into the next object — objects inside an
	// object stream are self-delimiting and may not be streams themselves.
	p := newParser(stm.content[entry.start:entry.end])
	obj, err := p.parseObject()
	if err != nil {
		return nil, fmt.Errorf("pdf: object %d in object stream %d: %w", objNum, loc.streamNum, err)
	}
	return obj, nil
}

// entryFor returns the entry for objNum. The index from the cross-reference
// entry is used when it agrees with the stream header; otherwise the object
// number is looked up directly, since some producers write stale indices.
func (stm *objectStream) entryFor(objNum, index int) (objStmEntry, bool) {
	if index >= 0 && index < len(stm.entries) && stm.entries[index].num == objNum {
		return stm.entries[index], true
	}
	for _, e := range stm.entries {
		if e.num == objNum {
			return e, true
		}
	}
	return objStmEntry{}, false
}

// loadObjectStream fetches, decodes and parses the header of the /ObjStm with
// the given object number. Decoded streams are cached so that sibling objects
// do not re-inflate the same body.
func (r *Reader) loadObjectStream(streamNum int) (*objectStream, error) {
	if stm, ok := r.objStms[streamNum]; ok {
		return stm, nil
	}
	if r.loadingStm == nil {
		r.loadingStm = make(map[int]bool)
	}
	if r.loadingStm[streamNum] {
		return nil, fmt.Errorf("object stream %d is self-referential", streamNum)
	}
	r.loadingStm[streamNum] = true
	defer delete(r.loadingStm, streamNum)

	obj, err := r.GetObject(streamNum)
	if err != nil {
		return nil, fmt.Errorf("object stream %d: %w", streamNum, err)
	}
	s, ok := obj.(Stream)
	if !ok {
		return nil, fmt.Errorf("object %d is not a stream", streamNum)
	}
	if typ, ok := s.Dict[Name("Type")].(Name); ok && typ != Name("ObjStm") {
		return nil, fmt.Errorf("object %d is /%s, not /ObjStm", streamNum, typ)
	}

	n, err := r.intFromDict(s.Dict, "N")
	if err != nil {
		return nil, fmt.Errorf("object stream %d: %w", streamNum, err)
	}
	first, err := r.intFromDict(s.Dict, "First")
	if err != nil {
		return nil, fmt.Errorf("object stream %d: %w", streamNum, err)
	}

	content, err := r.decodeStreamContent(s)
	if err != nil {
		return nil, fmt.Errorf("decode object stream %d: %w", streamNum, err)
	}
	if first < 0 || first > len(content) {
		return nil, fmt.Errorf("object stream %d: /First %d out of range (body is %d bytes)", streamNum, first, len(content))
	}

	entries, err := parseObjStmHeader(content[:first], n, first, len(content))
	if err != nil {
		return nil, fmt.Errorf("object stream %d: %w", streamNum, err)
	}

	stm := &objectStream{content: content, entries: entries}
	if r.objStms == nil {
		r.objStms = make(map[int]*objectStream)
	}
	r.objStms[streamNum] = stm
	return stm, nil
}

// parseObjStmHeader reads the N pairs of "objectNumber relativeOffset" that
// precede /First and converts them to absolute [start, end) ranges in the body.
func parseObjStmHeader(header []byte, n, first, bodyLen int) ([]objStmEntry, error) {
	if n < 0 {
		return nil, fmt.Errorf("/N is negative (%d)", n)
	}

	entries := make([]objStmEntry, 0, n)
	p := newParser(header)
	for i := 0; i < n; i++ {
		p.skipWhitespaceAndComments()
		numStr, isReal, err := p.scanNumber()
		if err != nil {
			return nil, fmt.Errorf("header pair %d: object number: %w", i, err)
		}
		if isReal {
			return nil, fmt.Errorf("header pair %d: object number %q is not an integer", i, numStr)
		}
		p.skipWhitespaceAndComments()
		offStr, isReal, err := p.scanNumber()
		if err != nil {
			return nil, fmt.Errorf("header pair %d: offset: %w", i, err)
		}
		if isReal {
			return nil, fmt.Errorf("header pair %d: offset %q is not an integer", i, offStr)
		}

		num, _ := strconv.Atoi(numStr)
		off, _ := strconv.Atoi(offStr)
		start := first + off
		if off < 0 || start > bodyLen {
			return nil, fmt.Errorf("header pair %d: offset %d out of range", i, off)
		}
		entries = append(entries, objStmEntry{num: num, start: start, end: bodyLen})
	}

	// Offsets are required to be ascending, so each object ends where the next
	// one begins. Leave an entry running to the end of the body if a producer
	// emitted them out of order.
	for i := 0; i+1 < len(entries); i++ {
		if entries[i+1].start >= entries[i].start {
			entries[i].end = entries[i+1].start
		}
	}
	return entries, nil
}

// intFromDict reads an integer entry from a dict, resolving it if it is an
// indirect reference.
func (r *Reader) intFromDict(d Dict, key string) (int, error) {
	obj, ok := d[Name(key)]
	if !ok {
		return 0, fmt.Errorf("missing /%s", key)
	}
	resolved, err := r.Resolve(obj)
	if err != nil {
		return 0, fmt.Errorf("resolve /%s: %w", key, err)
	}
	v, ok := resolved.(Integer)
	if !ok {
		return 0, fmt.Errorf("/%s is not an integer", key)
	}
	return int(v), nil
}
