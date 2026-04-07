package pdf

import (
	"fmt"
)

// FlattenForms flattens AcroForm fields into page content streams,
// making form data part of the static page content and removing
// all interactive form elements.
//
// For each widget annotation that has an appearance stream (/AP),
// the appearance is rendered onto the page at the annotation's /Rect
// position. The annotation is then removed from the page's /Annots
// array, and the /AcroForm entry is removed from the document catalog.
//
// Returns nil if the document has no AcroForm.
func (m *Modifier) FlattenForms() error {
	r := m.reader

	catalog, err := r.ResolveDict(r.RootRef())
	if err != nil {
		return fmt.Errorf("pdf: flatten: resolve catalog: %w", err)
	}

	// Check for AcroForm.
	acroFormObj, ok := catalog[Name("AcroForm")]
	if !ok {
		return nil // no forms
	}
	_, err = r.ResolveDict(acroFormObj)
	if err != nil {
		return nil // cannot resolve, treat as no forms
	}

	// Process each page.
	pageCount, err := r.PageCount()
	if err != nil {
		return fmt.Errorf("pdf: flatten: %w", err)
	}

	for i := range pageCount {
		if err := m.flattenPageAnnotations(i); err != nil {
			return fmt.Errorf("pdf: flatten page %d: %w", i, err)
		}
	}

	// Remove /AcroForm from catalog.
	newCatalog := make(Dict, len(catalog))
	for k, v := range catalog {
		if k != Name("AcroForm") {
			newCatalog[k] = v
		}
	}
	m.SetObject(r.RootRef(), newCatalog)

	return nil
}

// flattenPageAnnotations processes widget annotations on a single page.
func (m *Modifier) flattenPageAnnotations(pageIndex int) error {
	r := m.reader

	info, err := r.Page(pageIndex)
	if err != nil {
		return err
	}

	pageDict, err := r.ResolveDict(info.Ref)
	if err != nil {
		return err
	}

	annotsObj, ok := pageDict[Name("Annots")]
	if !ok {
		return nil // no annotations on this page
	}

	annotsResolved, err := r.Resolve(annotsObj)
	if err != nil {
		return err
	}
	annotsArr, ok := annotsResolved.(Array)
	if !ok {
		return nil
	}

	var overlayContent []byte
	overlayResources := Dict{
		Name("XObject"): Dict{},
	}
	xobjDict := overlayResources[Name("XObject")].(Dict)
	var remainingAnnots Array
	xobjIndex := 0

	for _, annotObj := range annotsArr {
		ops, kept := m.flattenAnnotation(annotObj, xobjDict, &xobjIndex)
		if kept {
			remainingAnnots = append(remainingAnnots, annotObj)
		}
		overlayContent = append(overlayContent, ops...)
	}

	if len(overlayContent) == 0 && len(remainingAnnots) == len(annotsArr) {
		return nil // nothing changed
	}

	// Update page dict.
	newPageDict := make(Dict, len(pageDict))
	for k, v := range pageDict {
		newPageDict[k] = v
	}

	// Update or remove /Annots.
	if len(remainingAnnots) == 0 {
		delete(newPageDict, Name("Annots"))
	} else {
		newPageDict[Name("Annots")] = remainingAnnots
	}

	// Overlay the flattened content onto the page.
	if len(overlayContent) > 0 {
		m.overlayFlattenedContent(info, newPageDict, overlayContent, &overlayResources)
	} else {
		m.SetObject(info.Ref, newPageDict)
	}

	return nil
}

// flattenAnnotation processes a single annotation for flattening.
// It returns the content stream operators (nil if none) and whether the annotation should be kept.
func (m *Modifier) flattenAnnotation(annotObj Object, xobjDict Dict, xobjIndex *int) (ops []byte, keep bool) {
	annotDict, err := m.reader.ResolveDict(annotObj)
	if err != nil {
		return nil, true
	}

	// Non-widget annotations are preserved.
	if !m.isWidgetAnnotation(annotDict) {
		return nil, true
	}

	// Get appearance stream.
	apStream, err := m.resolveAppearanceStream(annotDict)
	if err != nil || apStream == nil {
		// No appearance stream — just remove the annotation.
		return nil, false
	}

	// Get annotation rect.
	rect, err := m.resolveAnnotRect(annotDict)
	if err != nil {
		return nil, true
	}

	// Register the appearance stream as an XObject.
	xobjName := Name(fmt.Sprintf("_Flat%d", *xobjIndex))
	*xobjIndex++

	xobjRef := m.AllocObject()
	formXObj := m.buildFormXObject(apStream, rect)
	m.SetObject(xobjRef, formXObj)
	xobjDict[xobjName] = xobjRef

	// Generate content stream operators to draw the XObject.
	rectW := rect.URX - rect.LLX
	rectH := rect.URY - rect.LLY
	if rectW <= 0 || rectH <= 0 {
		return nil, false
	}

	// Determine scale from BBox to Rect.
	sx, sy := 1.0, 1.0
	bbox := m.resolveFormBBox(formXObj)
	bboxW := bbox.URX - bbox.LLX
	bboxH := bbox.URY - bbox.LLY
	if bboxW > 0 && bboxH > 0 {
		sx = rectW / bboxW
		sy = rectH / bboxH
	}

	// Translate to rect origin, offset by BBox origin, scale to fit.
	content := fmt.Sprintf("q %.4f 0 0 %.4f %.4f %.4f cm /%s Do Q\n",
		sx, sy,
		rect.LLX-bbox.LLX*sx, rect.LLY-bbox.LLY*sy,
		string(xobjName))
	return []byte(content), false
}

// isWidgetAnnotation checks if an annotation dict is a widget (form field).
func (m *Modifier) isWidgetAnnotation(d Dict) bool {
	subtypeObj, ok := d[Name("Subtype")]
	if ok {
		if subtype, ok := subtypeObj.(Name); ok {
			return string(subtype) == "Widget"
		}
	}
	// Some form fields don't have /Subtype but have /FT (field type).
	_, hasFT := d[Name("FT")]
	return hasFT
}

// resolveAppearanceStream gets the normal appearance stream for an annotation.
func (m *Modifier) resolveAppearanceStream(annotDict Dict) (*Stream, error) {
	r := m.reader

	apObj, ok := annotDict[Name("AP")]
	if !ok {
		return nil, nil
	}

	apDict, err := r.ResolveDict(apObj)
	if err != nil {
		return nil, err
	}

	// Get /N (normal appearance). It can be a stream directly or a dict of states.
	nObj, ok := apDict[Name("N")]
	if !ok {
		return nil, nil
	}

	resolved, err := r.Resolve(nObj)
	if err != nil {
		return nil, err
	}

	switch v := resolved.(type) {
	case Stream:
		return &v, nil
	case Dict:
		// It's a dict of appearance states. Use /AS to select the right one.
		asObj, ok := annotDict[Name("AS")]
		if !ok {
			// Try "Yes" as common default for checkboxes.
			if yesObj, ok := v[Name("Yes")]; ok {
				resolved2, err := r.Resolve(yesObj)
				if err != nil {
					return nil, err
				}
				if s, ok := resolved2.(Stream); ok {
					return &s, nil
				}
			}
			return nil, nil
		}
		asName, ok := asObj.(Name)
		if !ok {
			return nil, nil
		}
		stateObj, ok := v[asName]
		if !ok {
			return nil, nil
		}
		resolved2, err := r.Resolve(stateObj)
		if err != nil {
			return nil, err
		}
		if s, ok := resolved2.(Stream); ok {
			return &s, nil
		}
		return nil, nil
	default:
		return nil, nil
	}
}

// resolveAnnotRect extracts the /Rect from an annotation dict.
func (m *Modifier) resolveAnnotRect(annotDict Dict) (Rectangle, error) {
	rectObj, ok := annotDict[Name("Rect")]
	if !ok {
		return Rectangle{}, fmt.Errorf("annotation missing /Rect")
	}
	return m.reader.parseRectangle(rectObj)
}

// buildFormXObject creates a Form XObject stream from an appearance stream.
// If the appearance stream already has /Subtype /Form and a /BBox, it is
// used directly. Otherwise, a proper Form XObject wrapper is created.
func (m *Modifier) buildFormXObject(ap *Stream, rect Rectangle) Stream {
	content := ap.Content

	// Decompress content if needed.
	decoded, err := m.reader.decodeStreamContent(*ap)
	if err == nil {
		content = decoded
	}

	// Build the Form XObject dict.
	formDict := Dict{
		Name("Type"):    Name("XObject"),
		Name("Subtype"): Name("Form"),
		Name("BBox"):    Array{Real(0), Real(0), Real(rect.URX - rect.LLX), Real(rect.URY - rect.LLY)},
	}

	// Copy /Matrix if present.
	if matrix, ok := ap.Dict[Name("Matrix")]; ok {
		formDict[Name("Matrix")] = matrix
	}

	// Use the original BBox if present.
	if bbox, ok := ap.Dict[Name("BBox")]; ok {
		formDict[Name("BBox")] = bbox
	}

	// Copy /Resources from the appearance stream.
	if res, ok := ap.Dict[Name("Resources")]; ok {
		formDict[Name("Resources")] = res
	}

	return Stream{
		Dict:    formDict,
		Content: content,
	}
}

// resolveFormBBox extracts the BBox rectangle from a Form XObject stream.
func (m *Modifier) resolveFormBBox(s Stream) Rectangle {
	bboxObj, ok := s.Dict[Name("BBox")]
	if !ok {
		return Rectangle{}
	}
	r, err := m.reader.parseRectangle(bboxObj)
	if err != nil {
		return Rectangle{}
	}
	return r
}

// overlayFlattenedContent merges flattened content onto a page,
// similar to OverlayPage but operating on an already-modified page dict.
func (m *Modifier) overlayFlattenedContent(info PageInfo, pageDict Dict, content []byte, resources *Dict) {
	// Allocate streams.
	qRef := m.AllocObject()
	bigQRef := m.AllocObject()
	contentRef := m.AllocObject()

	m.SetObject(qRef, Stream{
		Dict:    Dict{},
		Content: []byte("q\n"),
	})
	m.SetObject(bigQRef, Stream{
		Dict:    Dict{},
		Content: []byte("\nQ\n"),
	})
	m.SetObject(contentRef, Stream{
		Dict:    Dict{},
		Content: content,
	})

	// Build new /Contents array.
	var contentRefs Array
	contentRefs = append(contentRefs, qRef)

	if origContents, ok := pageDict[Name("Contents")]; ok {
		switch v := origContents.(type) {
		case ObjectRef:
			contentRefs = append(contentRefs, v)
		case Array:
			contentRefs = append(contentRefs, v...)
		}
	}
	contentRefs = append(contentRefs, bigQRef, contentRef)
	pageDict[Name("Contents")] = contentRefs

	// Merge resources.
	if resources != nil {
		existingRes, _ := m.reader.ResolveDict(pageDict[Name("Resources")])
		if existingRes == nil {
			existingRes = make(Dict)
		}
		merged := mergeResources(existingRes, *resources)
		pageDict[Name("Resources")] = merged
	}

	m.SetObject(info.Ref, pageDict)
}
