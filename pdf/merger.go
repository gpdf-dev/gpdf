package pdf

import (
	"bytes"
	"fmt"
)

// MergeSource represents one input PDF in a merge operation.
type MergeSource struct {
	Data     []byte // raw PDF bytes
	FromPage int    // 0-based inclusive start page
	ToPage   int    // 0-based inclusive end page; -1 means last page
}

// MergeConfig configures the merge operation.
type MergeConfig struct {
	Info DocumentInfo
}

// MergePDFs combines pages from multiple PDF sources into a single PDF.
func MergePDFs(sources []MergeSource, cfg MergeConfig) ([]byte, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("pdf: merge requires at least one source")
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetInfo(cfg.Info)

	for i, src := range sources {
		reader, err := NewReader(src.Data)
		if err != nil {
			return nil, fmt.Errorf("pdf: source %d: %w", i, err)
		}

		pageCount, err := reader.PageCount()
		if err != nil {
			return nil, fmt.Errorf("pdf: source %d: %w", i, err)
		}

		from := src.FromPage
		to := src.ToPage
		if to < 0 {
			to = pageCount - 1
		}

		if from < 0 || from >= pageCount {
			return nil, fmt.Errorf("pdf: source %d: from page %d out of range [0, %d)", i, from, pageCount)
		}
		if to < from || to >= pageCount {
			return nil, fmt.Errorf("pdf: source %d: to page %d out of range [%d, %d)", i, to, from, pageCount)
		}

		mapper := newObjectMapper(reader)
		for p := from; p <= to; p++ {
			if err := mapper.copyPage(p, w); err != nil {
				return nil, fmt.Errorf("pdf: source %d, page %d: %w", i, p, err)
			}
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// objectMapper tracks object number remapping from one source PDF to the output.
type objectMapper struct {
	reader  *Reader
	mapping map[int]int // source obj number -> output obj number
}

func newObjectMapper(r *Reader) *objectMapper {
	return &objectMapper{
		reader:  r,
		mapping: make(map[int]int, r.MaxObjectNumber()),
	}
}

// copyPage copies the i-th page from the source PDF into the writer.
func (m *objectMapper) copyPage(pageIndex int, w *Writer) error {
	pageDict, err := m.reader.PageDict(pageIndex)
	if err != nil {
		return err
	}

	// Deep copy the page dict, remapping all object references.
	copied, err := m.copyObject(pageDict, w)
	if err != nil {
		return err
	}

	dict := copied.(Dict)

	// Replace /Parent with the output page tree.
	delete(dict, Name("Parent"))
	dict[Name("Type")] = Name("Page")
	dict[Name("Parent")] = w.PageTreeRef()

	// Allocate and write the page object.
	pageRef := w.AllocObject()
	if err := w.WriteObject(pageRef, dict); err != nil {
		return err
	}
	w.AddRawPage(pageRef)
	return nil
}

// copyObject recursively copies a PDF object, remapping all indirect references
// to new object numbers in the output writer.
func (m *objectMapper) copyObject(obj Object, w *Writer) (Object, error) {
	switch v := obj.(type) {
	case ObjectRef:
		return m.copyRef(v, w)
	case Dict:
		return m.copyDict(v, w)
	case Array:
		return m.copyArray(v, w)
	case Stream:
		return m.copyStream(v, w)
	default:
		// Integer, Real, Boolean, Name, LiteralString, HexString, Null, Rectangle
		return obj, nil
	}
}

func (m *objectMapper) copyRef(ref ObjectRef, w *Writer) (ObjectRef, error) {
	// Already mapped?
	if newNum, ok := m.mapping[ref.Number]; ok {
		return ObjectRef{Number: newNum}, nil
	}

	// Allocate new object number before recursing (handles cycles).
	newRef := w.AllocObject()
	m.mapping[ref.Number] = newRef.Number

	// Resolve the source object.
	resolved, err := m.reader.GetObject(ref.Number)
	if err != nil {
		return ObjectRef{}, fmt.Errorf("resolve object %d: %w", ref.Number, err)
	}

	// Deep copy the resolved object.
	copied, err := m.copyObject(resolved, w)
	if err != nil {
		return ObjectRef{}, err
	}

	// Write to output.
	if err := w.WriteObject(newRef, copied); err != nil {
		return ObjectRef{}, err
	}
	return newRef, nil
}

func (m *objectMapper) copyDict(d Dict, w *Writer) (Dict, error) {
	newDict := make(Dict, len(d))
	for key, val := range d {
		copied, err := m.copyObject(val, w)
		if err != nil {
			return nil, err
		}
		newDict[key] = copied
	}
	return newDict, nil
}

func (m *objectMapper) copyArray(a Array, w *Writer) (Array, error) {
	newArr := make(Array, len(a))
	for i, item := range a {
		copied, err := m.copyObject(item, w)
		if err != nil {
			return nil, err
		}
		newArr[i] = copied
	}
	return newArr, nil
}

func (m *objectMapper) copyStream(s Stream, w *Writer) (Stream, error) {
	newDict, err := m.copyDict(s.Dict, w)
	if err != nil {
		return Stream{}, err
	}
	// Copy content bytes verbatim (no decompression/recompression).
	content := make([]byte, len(s.Content))
	copy(content, s.Content)
	return Stream{Dict: newDict, Content: content}, nil
}
