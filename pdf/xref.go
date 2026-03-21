package pdf

import (
	"fmt"
	"io"
)

// XRefEntry represents a single entry in the PDF cross-reference table.
type XRefEntry struct {
	Offset     int64
	Generation int
	InUse      bool
}

// XRefTable manages object byte offsets for the PDF cross-reference table.
// Entry 0 is always the head of the free-object linked list.
type XRefTable struct {
	entries []XRefEntry
}

// NewXRefTable creates a new XRefTable with the mandatory free entry at index 0.
func NewXRefTable() *XRefTable {
	return &XRefTable{
		entries: []XRefEntry{
			{Offset: 0, Generation: 65535, InUse: false}, // entry 0: free list head
		},
	}
}

// Add registers an object's byte offset in the cross-reference table.
// If objNum exceeds the current table size, intermediate entries are added
// as free entries.
func (t *XRefTable) Add(objNum int, offset int64, generation int) {
	// Grow the slice if necessary.
	for len(t.entries) <= objNum {
		t.entries = append(t.entries, XRefEntry{Offset: 0, Generation: 0, InUse: false})
	}
	t.entries[objNum] = XRefEntry{
		Offset:     offset,
		Generation: generation,
		InUse:      true,
	}
}

// Size returns the total number of entries in the cross-reference table,
// including the mandatory free entry at index 0.
func (t *XRefTable) Size() int {
	return len(t.entries)
}

// WriteTo writes the cross-reference table in standard PDF format:
//
//	xref
//	0 N
//	0000000000 65535 f \r\n
//	0000000009 00000 n \r\n
//	...
//
// Each entry line is exactly 20 bytes (including the trailing \r\n).
func (t *XRefTable) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}

	if _, err := fmt.Fprintf(cw, "xref\n0 %d\n", len(t.entries)); err != nil {
		return cw.n, err
	}

	for _, e := range t.entries {
		marker := byte('n')
		if !e.InUse {
			marker = 'f'
		}
		// Each line is exactly 20 bytes: 10-digit offset + space + 5-digit gen + space + marker + EOL(2).
		// EOL is CR+LF per PDF spec (ISO 32000-2 §7.5.4).
		line := fmt.Sprintf("%010d %05d %c\r\n", e.Offset, e.Generation, marker)
		if _, err := io.WriteString(cw, line); err != nil {
			return cw.n, err
		}
	}

	return cw.n, nil
}
