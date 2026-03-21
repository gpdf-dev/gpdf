package pdf

import (
	"bytes"
	"fmt"
	"io"
)

// Modifier applies incremental updates to an existing PDF.
// It appends new/modified objects after the original data, then writes
// a new xref table and trailer referencing the original via /Prev.
type Modifier struct {
	reader     *Reader
	newObjects map[int]Object // objNum -> new/modified object
	nextObjNum int
}

// NewModifier creates a Modifier for the given Reader.
func NewModifier(r *Reader) *Modifier {
	return &Modifier{
		reader:     r,
		newObjects: make(map[int]Object),
		nextObjNum: r.MaxObjectNumber() + 1,
	}
}

// AllocObject allocates a new object number for use in the incremental update.
func (m *Modifier) AllocObject() ObjectRef {
	ref := ObjectRef{Number: m.nextObjNum, Generation: 0}
	m.nextObjNum++
	return ref
}

// SetObject registers a new or replacement object for the given ref.
// If the ref already exists in the original PDF, the new object will
// override it in the incremental update.
func (m *Modifier) SetObject(ref ObjectRef, obj Object) {
	m.newObjects[ref.Number] = obj
}

// Reader returns the underlying reader.
func (m *Modifier) Reader() *Reader {
	return m.reader
}

// OverlayPage adds content on top of an existing page.
// The overlayContent is raw PDF content stream data (e.g., text operations).
// It wraps the original page contents in q/Q and appends the overlay.
func (m *Modifier) OverlayPage(pageIndex int, overlayContent []byte, resources *Dict) error {
	info, err := m.reader.Page(pageIndex)
	if err != nil {
		return err
	}

	pageDict, err := m.reader.ResolveDict(info.Ref)
	if err != nil {
		return err
	}

	// Create a stream that wraps original content in q/Q then draws overlay.
	// We use the /Contents array approach:
	//   [q_stream, original_refs..., Q_stream, overlay_stream]

	// Allocate streams.
	qRef := m.AllocObject()
	bigQRef := m.AllocObject()
	overlayRef := m.AllocObject()

	m.SetObject(qRef, Stream{
		Dict:    Dict{},
		Content: []byte("q\n"),
	})
	m.SetObject(bigQRef, Stream{
		Dict:    Dict{},
		Content: []byte("\nQ\n"),
	})
	m.SetObject(overlayRef, Stream{
		Dict:    Dict{},
		Content: overlayContent,
	})

	// Build new /Contents array.
	var contentRefs Array
	contentRefs = append(contentRefs, qRef)

	// Add original content references.
	if origContents, ok := pageDict[Name("Contents")]; ok {
		switch v := origContents.(type) {
		case ObjectRef:
			contentRefs = append(contentRefs, v)
		case Array:
			contentRefs = append(contentRefs, v...)
		}
	}
	contentRefs = append(contentRefs, bigQRef, overlayRef)

	// Build updated page dict.
	newPageDict := make(Dict, len(pageDict)+2)
	for k, v := range pageDict {
		newPageDict[k] = v
	}
	newPageDict[Name("Contents")] = contentRefs

	// Merge overlay resources into page resources.
	if resources != nil {
		existingRes, _ := m.reader.ResolveDict(pageDict[Name("Resources")])
		if existingRes == nil {
			existingRes = make(Dict)
		}
		merged := mergeResources(existingRes, *resources)
		newPageDict[Name("Resources")] = merged
	}

	m.SetObject(info.Ref, newPageDict)
	return nil
}

// Write outputs the complete modified PDF (original data + incremental update)
// to the given writer.
func (m *Modifier) Write(w io.Writer) error {
	cw := &countWriter{w: w}

	// 1. Write the original PDF data as-is.
	if _, err := cw.Write(m.reader.Data()); err != nil {
		return err
	}

	if len(m.newObjects) == 0 {
		return nil
	}

	// 2. Write new/modified objects.
	xref, err := m.writeNewObjects(cw)
	if err != nil {
		return err
	}

	// 3. Write new xref table (only modified entries).
	xrefOffset := cw.count
	if err := m.writeIncrementalXRef(cw, xref); err != nil {
		return err
	}

	// 4. Write trailer.
	return m.writeIncrementalTrailer(cw, xrefOffset)
}

func (m *Modifier) writeNewObjects(cw *countWriter) (*XRefTable, error) {
	xref := NewXRefTable()
	for objNum, obj := range m.newObjects {
		offset := cw.count
		xref.Add(objNum, offset, 0)

		if _, err := fmt.Fprintf(cw, "%d 0 obj\n", objNum); err != nil {
			return nil, err
		}
		if _, err := obj.WriteTo(cw); err != nil {
			return nil, err
		}
		if _, err := io.WriteString(cw, "\nendobj\n"); err != nil {
			return nil, err
		}
	}
	return xref, nil
}

func (m *Modifier) writeIncrementalTrailer(cw *countWriter, xrefOffset int64) error {
	prevXRef, err := m.reader.findStartXRef()
	if err != nil {
		return err
	}

	size := m.nextObjNum
	if origSize, ok := m.reader.trailer[Name("Size")]; ok {
		if v, ok := origSize.(Integer); ok && int(v) > size {
			size = int(v)
		}
	}

	trailerDict := Dict{
		Name("Size"): Integer(size),
		Name("Root"): m.reader.RootRef(),
		Name("Prev"): Integer(prevXRef),
	}
	if info, ok := m.reader.trailer[Name("Info")]; ok {
		trailerDict[Name("Info")] = info
	}

	if _, err := io.WriteString(cw, "trailer\n"); err != nil {
		return err
	}
	if _, err := trailerDict.WriteTo(cw); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(cw, "\nstartxref\n%d\n%%%%EOF\n", xrefOffset); err != nil {
		return err
	}
	return nil
}

// Bytes is a convenience method that writes to a buffer and returns the result.
func (m *Modifier) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := m.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeIncrementalXRef writes xref subsections for only the modified objects.
func (m *Modifier) writeIncrementalXRef(w io.Writer, xref *XRefTable) error {
	// Collect and sort object numbers.
	type entry struct {
		num    int
		offset int64
	}
	var entries []entry
	for objNum := range m.newObjects {
		for i, e := range xref.entries {
			if i == objNum && e.InUse {
				entries = append(entries, entry{num: objNum, offset: e.Offset})
				break
			}
		}
	}

	// Sort by object number.
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].num < entries[j-1].num; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}

	if _, err := io.WriteString(w, "xref\n"); err != nil {
		return err
	}

	// Group into contiguous subsections.
	i := 0
	for i < len(entries) {
		start := entries[i].num
		j := i
		for j < len(entries) && entries[j].num == start+(j-i) {
			j++
		}

		count := j - i
		if _, err := fmt.Fprintf(w, "%d %d\n", start, count); err != nil {
			return err
		}
		for k := i; k < j; k++ {
			line := fmt.Sprintf("%010d %05d n \r\n", entries[k].offset, 0)
			if _, err := io.WriteString(w, line); err != nil {
				return err
			}
		}
		i = j
	}

	return nil
}

// mergeResources merges overlay resources into existing page resources.
func mergeResources(existing, overlay Dict) Dict {
	result := make(Dict, len(existing))
	for k, v := range existing {
		result[k] = v
	}

	for k, v := range overlay {
		if existingSubDict, ok := result[k].(Dict); ok {
			if overlaySubDict, ok := v.(Dict); ok {
				merged := make(Dict, len(existingSubDict)+len(overlaySubDict))
				for sk, sv := range existingSubDict {
					merged[sk] = sv
				}
				for sk, sv := range overlaySubDict {
					merged[sk] = sv
				}
				result[k] = merged
				continue
			}
		}
		result[k] = v
	}
	return result
}
