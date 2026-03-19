package pdf

import (
	"fmt"
	"io"
)

// pdfHeader is the PDF version header written at the start of every file.
const pdfHeader = "%PDF-1.7\n"

// objectEntry records an allocated object.
type objectEntry struct {
	ref ObjectRef
}

// Writer assembles a complete PDF document and writes it to an io.Writer.
// Objects are written sequentially; the cross-reference table and trailer
// are produced when Close is called.
type Writer struct {
	w          *countWriter
	xref       *XRefTable
	objects    []*objectEntry
	pages      []ObjectRef
	fonts      map[string]ObjectRef // font name -> object ref
	images     map[string]ObjectRef // image name -> object ref
	info       DocumentInfo
	catalog    ObjectRef
	pageTree   ObjectRef
	nextObjNum int
	compress   bool
	closed     bool

	// Extension hooks for gpdf-pro features (PDF/A, encryption, signatures).
	catalogExtra  Dict                                // extra entries merged into catalog dict
	trailerExtra  Dict                                // extra entries merged into trailer dict
	onWriteObject func(ref ObjectRef, obj Object) Object // object transformation hook
	beforeClose   []func(pw *Writer) error            // callbacks run before Close finalizes
}

// countWriter wraps an io.Writer and tracks the total number of bytes written.
// This is used to record byte offsets for the cross-reference table.
type countWriter struct {
	w     io.Writer
	count int64
}

func (cw *countWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.count += int64(n)
	return n, err
}

// NewWriter creates a new PDF Writer that writes to w.
// The PDF header is written immediately.
func NewWriter(w io.Writer) *Writer {
	cw := &countWriter{w: w}

	pw := &Writer{
		w:          cw,
		xref:       NewXRefTable(),
		fonts:      make(map[string]ObjectRef),
		images:     make(map[string]ObjectRef),
		nextObjNum: 1,
		compress:   true,
	}

	// Write the PDF header immediately.
	// We ignore errors here; they will surface on subsequent writes.
	_, _ = io.WriteString(cw, pdfHeader)

	// Pre-allocate catalog and page tree objects so they have stable refs.
	pw.catalog = pw.AllocObject()
	pw.pageTree = pw.AllocObject()

	return pw
}

// AllocObject allocates a new object number and returns its ObjectRef.
// The object is not written to the output until WriteObject is called.
func (pw *Writer) AllocObject() ObjectRef {
	ref := ObjectRef{Number: pw.nextObjNum, Generation: 0}
	pw.nextObjNum++
	pw.objects = append(pw.objects, &objectEntry{ref: ref})
	return ref
}

// WriteObject writes a PDF indirect object to the output stream in the form:
//
//	N G obj
//	<object>
//	endobj
//
// It records the byte offset in the cross-reference table.
func (pw *Writer) WriteObject(ref ObjectRef, obj Object) error {
	// Apply object transformation hook if set (e.g., encryption).
	if pw.onWriteObject != nil {
		obj = pw.onWriteObject(ref, obj)
	}

	offset := pw.w.count
	pw.xref.Add(ref.Number, offset, ref.Generation)

	if _, err := fmt.Fprintf(pw.w, "%d %d obj\n", ref.Number, ref.Generation); err != nil {
		return err
	}
	if _, err := obj.WriteTo(pw.w); err != nil {
		return err
	}
	if _, err := io.WriteString(pw.w, "\nendobj\n"); err != nil {
		return err
	}
	return nil
}

// AddPage adds a page to the document. The page's content streams should
// already have been written via WriteObject; their refs are in page.Contents.
func (pw *Writer) AddPage(page PageObject) error {
	pageRef := pw.AllocObject()

	// Build the page dictionary.
	pageDict := Dict{
		Name("Type"):     Name("Page"),
		Name("Parent"):   pw.pageTree,
		Name("MediaBox"): page.MediaBox,
	}

	// Resources.
	resDict := page.Resources.ToDict()
	if len(resDict) > 0 {
		pageDict[Name("Resources")] = resDict
	}

	// Contents: single ref or array.
	switch len(page.Contents) {
	case 0:
		// No content stream.
	case 1:
		pageDict[Name("Contents")] = page.Contents[0]
	default:
		arr := make(Array, len(page.Contents))
		for i, ref := range page.Contents {
			arr[i] = ref
		}
		pageDict[Name("Contents")] = arr
	}

	if err := pw.WriteObject(pageRef, pageDict); err != nil {
		return err
	}

	pw.pages = append(pw.pages, pageRef)
	return nil
}

// ReserveFontRef reserves a font resource name and object reference for the
// given font name without writing any PDF objects. This allows Type0/CIDFont
// structures to be written later via OnBeforeClose hooks.
func (pw *Writer) ReserveFontRef(name string) (string, ObjectRef) {
	if ref, ok := pw.fonts[name]; ok {
		idx := 1
		for k := range pw.fonts {
			if k == name {
				break
			}
			idx++
		}
		return fmt.Sprintf("F%d", idx), ref
	}

	fontRef := pw.AllocObject()
	resName := fmt.Sprintf("F%d", len(pw.fonts)+1)
	pw.fonts[name] = fontRef
	return resName, fontRef
}

// RegisterFont registers a font with the given name and font data.
// It writes the font as a PDF font object and returns the PDF resource
// name (e.g., "F1") and the object reference.
func (pw *Writer) RegisterFont(name string, fontData []byte) (string, ObjectRef, error) {
	if ref, ok := pw.fonts[name]; ok {
		// Find the resource name index.
		idx := 1
		for k := range pw.fonts {
			if k == name {
				break
			}
			idx++
		}
		return fmt.Sprintf("F%d", idx), ref, nil
	}

	fontRef := pw.AllocObject()
	resName := fmt.Sprintf("F%d", len(pw.fonts)+1)

	// Write font descriptor and stream if font data is provided.
	if len(fontData) > 0 {
		fontFileRef := pw.AllocObject()

		streamDict := Dict{
			Name("Length1"): Integer(len(fontData)),
		}

		content := fontData
		if pw.compress {
			compressed, err := CompressFlate(fontData)
			if err != nil {
				return "", ObjectRef{}, fmt.Errorf("pdf: failed to compress font data: %w", err)
			}
			streamDict[Name("Filter")] = Name("FlateDecode")
			content = compressed
		}

		fontStream := Stream{
			Dict:    streamDict,
			Content: content,
		}
		if err := pw.WriteObject(fontFileRef, fontStream); err != nil {
			return "", ObjectRef{}, err
		}

		// Write a basic TrueType font dictionary.
		fontDict := Dict{
			Name("Type"):     Name("Font"),
			Name("Subtype"):  Name("TrueType"),
			Name("BaseFont"): Name(name),
			Name("FontDescriptor"): func() ObjectRef {
				descRef := pw.AllocObject()
				descDict := Dict{
					Name("Type"):      Name("FontDescriptor"),
					Name("FontName"):  Name(name),
					Name("FontFile2"): fontFileRef,
				}
				// Write the descriptor (errors will be caught downstream).
				_ = pw.WriteObject(descRef, descDict)
				return descRef
			}(),
		}
		if err := pw.WriteObject(fontRef, fontDict); err != nil {
			return "", ObjectRef{}, err
		}
	} else {
		// Standard 14 font (no embedding needed).
		fontDict := Dict{
			Name("Type"):     Name("Font"),
			Name("Subtype"):  Name("Type1"),
			Name("BaseFont"): Name(name),
		}
		if err := pw.WriteObject(fontRef, fontDict); err != nil {
			return "", ObjectRef{}, err
		}
	}

	pw.fonts[name] = fontRef
	return resName, fontRef, nil
}

// RegisterImage registers an image and returns its PDF resource name
// (e.g., "Im1") and the object reference. The filter parameter selects
// the PDF stream filter: "DCTDecode" for JPEG data (stored as-is) or
// empty for raw pixel data (compressed with FlateDecode).
// If smaskData is non-nil, an SMask (soft mask) image object is created
// for the alpha channel and referenced from the main image.
func (pw *Writer) RegisterImage(name string, data []byte, width, height int, colorSpace, filter string, smaskData []byte) (string, ObjectRef, error) {
	if ref, ok := pw.images[name]; ok {
		idx := 1
		for k := range pw.images {
			if k == name {
				break
			}
			idx++
		}
		return fmt.Sprintf("Im%d", idx), ref, nil
	}

	// Write SMask object first if alpha data is provided.
	var smaskRef ObjectRef
	if len(smaskData) > 0 {
		smaskRef = pw.AllocObject()
		smaskDict := Dict{
			Name("Type"):             Name("XObject"),
			Name("Subtype"):          Name("Image"),
			Name("Width"):            Integer(width),
			Name("Height"):           Integer(height),
			Name("ColorSpace"):       Name("DeviceGray"),
			Name("BitsPerComponent"): Integer(8),
		}
		smaskContent := smaskData
		if pw.compress {
			compressed, err := CompressFlate(smaskData)
			if err != nil {
				return "", ObjectRef{}, fmt.Errorf("pdf: failed to compress smask data: %w", err)
			}
			smaskDict[Name("Filter")] = Name("FlateDecode")
			smaskContent = compressed
		}
		smaskStream := Stream{
			Dict:    smaskDict,
			Content: smaskContent,
		}
		if err := pw.WriteObject(smaskRef, smaskStream); err != nil {
			return "", ObjectRef{}, err
		}
	}

	imgRef := pw.AllocObject()
	resName := fmt.Sprintf("Im%d", len(pw.images)+1)

	imgDict := Dict{
		Name("Type"):             Name("XObject"),
		Name("Subtype"):          Name("Image"),
		Name("Width"):            Integer(width),
		Name("Height"):           Integer(height),
		Name("ColorSpace"):       Name(colorSpace),
		Name("BitsPerComponent"): Integer(8),
	}

	if smaskRef.Number > 0 {
		imgDict[Name("SMask")] = smaskRef
	}

	content := data
	switch filter {
	case "DCTDecode":
		imgDict[Name("Filter")] = Name("DCTDecode")
	default:
		if pw.compress {
			compressed, err := CompressFlate(data)
			if err != nil {
				return "", ObjectRef{}, fmt.Errorf("pdf: failed to compress image data: %w", err)
			}
			imgDict[Name("Filter")] = Name("FlateDecode")
			content = compressed
		}
	}

	imgStream := Stream{
		Dict:    imgDict,
		Content: content,
	}
	if err := pw.WriteObject(imgRef, imgStream); err != nil {
		return "", ObjectRef{}, err
	}

	pw.images[name] = imgRef
	return resName, imgRef, nil
}

// SetCompression enables or disables flate compression for streams.
func (pw *Writer) SetCompression(enabled bool) {
	pw.compress = enabled
}

// AddCatalogEntry adds an entry to the catalog dictionary.
// This is used by extensions (e.g., PDF/A OutputIntents, signatures AcroForm).
func (pw *Writer) AddCatalogEntry(key Name, value Object) {
	if pw.catalogExtra == nil {
		pw.catalogExtra = make(Dict)
	}
	pw.catalogExtra[key] = value
}

// AddTrailerEntry adds an entry to the trailer dictionary.
// This is used by extensions (e.g., encryption Encrypt dict, ID array).
func (pw *Writer) AddTrailerEntry(key Name, value Object) {
	if pw.trailerExtra == nil {
		pw.trailerExtra = make(Dict)
	}
	pw.trailerExtra[key] = value
}

// SetObjectHook registers a function that transforms each object before
// it is written. This is used by encryption to encrypt strings and streams.
func (pw *Writer) SetObjectHook(fn func(ref ObjectRef, obj Object) Object) {
	pw.onWriteObject = fn
}

// OnBeforeClose registers a callback that runs before Close finalizes the PDF.
// Multiple callbacks are executed in registration order.
// This is used to write ICC profiles, XMP metadata, or encryption dictionaries.
func (pw *Writer) OnBeforeClose(fn func(pw *Writer) error) {
	pw.beforeClose = append(pw.beforeClose, fn)
}

// BytesWritten returns the total number of bytes written so far.
// This is useful for calculating ByteRange offsets for digital signatures.
func (pw *Writer) BytesWritten() int64 {
	return pw.w.count
}

// RawWrite writes raw bytes directly to the output stream.
// This is used for signature placeholders where exact byte control is needed.
func (pw *Writer) RawWrite(data []byte) (int, error) {
	return pw.w.Write(data)
}

// Close finishes writing the PDF document. It writes the page tree,
// catalog, cross-reference table, trailer, and %%EOF marker.
// Close must be called exactly once.
func (pw *Writer) Close() error {
	if pw.closed {
		return fmt.Errorf("pdf: writer already closed")
	}
	pw.closed = true

	// 0. Run beforeClose hooks (e.g., write ICC profiles, XMP metadata, encrypt dicts).
	for _, fn := range pw.beforeClose {
		if err := fn(pw); err != nil {
			return err
		}
	}

	// 1. Write the page tree object.
	kids := make(Array, len(pw.pages))
	for i, ref := range pw.pages {
		kids[i] = ref
	}
	pageTreeDict := Dict{
		Name("Type"):  Name("Pages"),
		Name("Kids"):  kids,
		Name("Count"): Integer(len(pw.pages)),
	}
	if err := pw.WriteObject(pw.pageTree, pageTreeDict); err != nil {
		return err
	}

	// 2. Write info dictionary if any metadata is set.
	var infoRef ObjectRef
	infoDict := pw.info.ToDict()
	if len(infoDict) > 0 {
		infoRef = pw.AllocObject()
		if err := pw.WriteObject(infoRef, infoDict); err != nil {
			return err
		}
	}

	// 3. Write the catalog object, merging any extra entries.
	catalogDict := Dict{
		Name("Type"):  Name("Catalog"),
		Name("Pages"): pw.pageTree,
	}
	for k, v := range pw.catalogExtra {
		catalogDict[k] = v
	}
	if err := pw.WriteObject(pw.catalog, catalogDict); err != nil {
		return err
	}

	// 4. Write the cross-reference table.
	xrefOffset := pw.w.count
	if _, err := pw.xref.WriteTo(pw.w); err != nil {
		return err
	}

	// 5. Write the trailer, merging any extra entries.
	trailerDict := Dict{
		Name("Size"): Integer(pw.xref.Size()),
		Name("Root"): pw.catalog,
	}
	if infoRef.Number > 0 {
		trailerDict[Name("Info")] = infoRef
	}
	for k, v := range pw.trailerExtra {
		trailerDict[k] = v
	}

	if _, err := io.WriteString(pw.w, "trailer\n"); err != nil {
		return err
	}
	if _, err := trailerDict.WriteTo(pw.w); err != nil {
		return err
	}

	// 6. Write startxref.
	if _, err := fmt.Fprintf(pw.w, "\nstartxref\n%d\n", xrefOffset); err != nil {
		return err
	}

	// 7. Write %%EOF.
	if _, err := io.WriteString(pw.w, "%%EOF\n"); err != nil {
		return err
	}

	return nil
}

// SetInfo sets the document metadata.
func (pw *Writer) SetInfo(info DocumentInfo) {
	pw.info = info
}
