package template

import (
	"bytes"
	"encoding/xml"
	"math"
	"strconv"
	"strings"

	"github.com/gpdf-dev/gpdf/barcode"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/qrcode"
)

// gridColumns is the total number of columns in the grid system.
const gridColumns = 12

// PageBuilder constructs the content of a single page using a row-based
// grid system. In addition to flow-based rows, it supports absolute
// positioning for placing content at fixed coordinates.
type PageBuilder struct {
	doc       *Document
	rows      []rowEntry
	absolutes []absoluteEntry
}

type rowEntry struct {
	height document.Value
	auto   bool
	fn     func(r *RowBuilder)
}

type absoluteEntry struct {
	x      document.Value
	y      document.Value
	width  document.Value
	height document.Value
	origin document.PositionOrigin
	fn     func(c *ColBuilder)
}

// Row adds a row with a specified height.
func (p *PageBuilder) Row(height document.Value, fn func(r *RowBuilder)) {
	p.rows = append(p.rows, rowEntry{height: height, fn: fn})
}

// AutoRow adds a row whose height is determined automatically by its content.
func (p *PageBuilder) AutoRow(fn func(r *RowBuilder)) {
	p.rows = append(p.rows, rowEntry{auto: true, fn: fn})
}

// Absolute places content at fixed XY coordinates, removed from the
// normal document flow. Coordinates are relative to the content area
// (inside page margins) by default. Use [AbsoluteOriginPage] to
// position relative to the page corner instead.
//
//	page.Absolute(document.Mm(100), document.Mm(200), func(c *ColBuilder) {
//	    c.Image(stampData, template.FitWidth(document.Mm(30)))
//	})
func (p *PageBuilder) Absolute(x, y document.Value, fn func(c *ColBuilder), opts ...AbsoluteOption) {
	cfg := absoluteConfig{
		origin: document.OriginContentArea,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	p.absolutes = append(p.absolutes, absoluteEntry{
		x:      x,
		y:      y,
		width:  cfg.width,
		height: cfg.height,
		origin: cfg.origin,
		fn:     fn,
	})
}

// buildNodes converts the page builder state into document nodes.
func (p *PageBuilder) buildNodes() []document.DocumentNode {
	var nodes []document.DocumentNode
	for _, row := range p.rows {
		rb := &RowBuilder{doc: p.doc}
		if row.fn != nil {
			row.fn(rb)
		}
		node := rb.build(row.height, row.auto)
		nodes = append(nodes, node)
	}

	// Absolute-positioned nodes are appended after flow nodes.
	for _, abs := range p.absolutes {
		cb := &ColBuilder{doc: p.doc}
		if abs.fn != nil {
			abs.fn(cb)
		}
		box := &document.Box{
			Content: cb.buildNodes(),
			BoxStyle: document.BoxStyle{
				Width:  abs.width,
				Height: abs.height,
				Position: document.Position{
					Mode:   document.PositionAbsolute,
					X:      abs.x,
					Y:      abs.y,
					Origin: abs.origin,
				},
			},
		}
		nodes = append(nodes, box)
	}

	return nodes
}

// RowBuilder constructs columns within a row.
type RowBuilder struct {
	doc  *Document
	cols []colEntry
}

type colEntry struct {
	span int
	fn   func(c *ColBuilder)
}

// Col adds a column with the given grid span (out of 12).
func (r *RowBuilder) Col(span int, fn func(c *ColBuilder)) {
	if span < 1 {
		span = 1
	}
	if span > gridColumns {
		span = gridColumns
	}
	r.cols = append(r.cols, colEntry{span: span, fn: fn})
}

// build converts the row into a Box document node. Each column becomes
// a child Box with a width proportional to its grid span.
func (r *RowBuilder) build(height document.Value, auto bool) document.DocumentNode {
	box := &document.Box{
		BoxStyle: document.BoxStyle{
			Direction: document.DirectionHorizontal,
		},
	}

	if !auto {
		box.BoxStyle.Height = height
	}

	for _, col := range r.cols {
		cb := &ColBuilder{doc: r.doc}
		if col.fn != nil {
			col.fn(cb)
		}

		colBox := &document.Box{
			Content: cb.buildNodes(),
			BoxStyle: document.BoxStyle{
				Width: document.Pct(float64(col.span) / float64(gridColumns) * 100),
			},
		}
		box.Content = append(box.Content, colBox)
	}

	return box
}

// ColBuilder constructs content within a column.
type ColBuilder struct {
	doc   *Document
	nodes []document.DocumentNode
}

// Text adds a text element to the column.
func (c *ColBuilder) Text(text string, opts ...TextOption) {
	style := c.defaultStyle()
	for _, opt := range opts {
		opt(&style)
	}
	c.nodes = append(c.nodes, &document.Text{
		Content:   text,
		TextStyle: style,
	})
}

// Image adds an image element to the column.
func (c *ColBuilder) Image(src []byte, opts ...ImageOption) {
	imgCfg := imageConfig{
		fitMode: document.FitContain,
	}
	for _, opt := range opts {
		opt(&imgCfg)
	}

	format := detectImageFormat(src)
	w, h := extractImageDimensions(src, format)

	imgNode := &document.Image{
		Source: document.ImageSource{
			Data:   src,
			Format: format,
			Width:  w,
			Height: h,
		},
		FitMode: imgCfg.fitMode,
	}

	if imgCfg.align != document.AlignLeft {
		imgNode.ImgStyle.TextAlign = imgCfg.align
	}

	if imgCfg.width.Amount > 0 {
		imgNode.DisplayWidth = imgCfg.width
	}
	if imgCfg.height.Amount > 0 {
		imgNode.DisplayHeight = imgCfg.height
	}

	c.nodes = append(c.nodes, imgNode)
}

// Table adds a table with header and body rows.
func (c *ColBuilder) Table(header []string, rows [][]string, opts ...TableOption) {
	tbl := &document.Table{}

	tblCfg := tableConfig{}
	for _, opt := range opts {
		opt(&tblCfg)
	}

	// Build header row.
	if len(header) > 0 {
		headerRow := document.TableRow{}
		for _, h := range header {
			cellStyle := c.defaultStyle()
			cellStyle.FontWeight = document.WeightBold
			if tblCfg.headerBgColor != nil {
				cellStyle.Background = tblCfg.headerBgColor
			}
			if tblCfg.headerTextColor != nil {
				cellStyle.Color = *tblCfg.headerTextColor
			}
			headerRow.Cells = append(headerRow.Cells, document.TableCell{
				Content: []document.DocumentNode{
					&document.Text{Content: h, TextStyle: cellStyle},
				},
				ColSpan: 1,
				RowSpan: 1,
			})
		}
		tbl.Header = []document.TableRow{headerRow}
	}

	// Build body rows.
	for i, row := range rows {
		bodyRow := document.TableRow{}
		for _, cell := range row {
			cellStyle := c.defaultStyle()
			if tblCfg.stripeColor != nil && i%2 == 1 {
				cellStyle.Background = tblCfg.stripeColor
			}
			if tblCfg.hasCellVAlign {
				cellStyle.VerticalAlign = tblCfg.cellVAlign
			}
			bodyRow.Cells = append(bodyRow.Cells, document.TableCell{
				Content: []document.DocumentNode{
					&document.Text{Content: cell, TextStyle: cellStyle},
				},
				ColSpan: 1,
				RowSpan: 1,
			})
		}
		tbl.Body = append(tbl.Body, bodyRow)
	}

	// Set column widths.
	if len(tblCfg.columnWidths) > 0 {
		for _, w := range tblCfg.columnWidths {
			tbl.Columns = append(tbl.Columns, document.TableColumn{
				Width: document.Pct(w),
			})
		}
	}

	c.nodes = append(c.nodes, tbl)
}

// Line adds a horizontal line (rule) to the column.
func (c *ColBuilder) Line(opts ...LineOption) {
	cfg := lineConfig{
		color:     pdf.Gray(0.8),
		thickness: document.Pt(1),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	c.nodes = append(c.nodes, &document.Box{
		BoxStyle: document.BoxStyle{
			Height:     cfg.thickness,
			Background: &cfg.color,
		},
	})
}

// List adds an unordered (bulleted) list to the column.
func (c *ColBuilder) List(items []string, opts ...ListOption) {
	c.addList(document.Unordered, items, opts)
}

// OrderedList adds an ordered (numbered) list to the column.
func (c *ColBuilder) OrderedList(items []string, opts ...ListOption) {
	c.addList(document.Ordered, items, opts)
}

// addList creates a List node with the given type and items.
func (c *ColBuilder) addList(lt document.ListType, items []string, opts []ListOption) {
	cfg := listConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	style := c.defaultStyle()
	lst := &document.List{
		ListType:  lt,
		ListStyle: style,
	}
	if cfg.indent > 0 {
		lst.MarkerIndent = cfg.indent
	}
	for _, item := range items {
		lst.Items = append(lst.Items, document.ListItem{
			Content: []document.DocumentNode{
				&document.Text{Content: item, TextStyle: style},
			},
		})
	}
	c.nodes = append(c.nodes, lst)
}

// PageNumber adds a text element containing the page number placeholder.
// The placeholder is replaced with the actual page number after pagination.
func (c *ColBuilder) PageNumber(opts ...TextOption) {
	style := c.defaultStyle()
	for _, opt := range opts {
		opt(&style)
	}
	c.nodes = append(c.nodes, &document.Text{
		Content:   document.PageNumberPlaceholder,
		TextStyle: style,
	})
}

// TotalPages adds a text element containing the total pages placeholder.
// The placeholder is replaced with the total page count after pagination.
func (c *ColBuilder) TotalPages(opts ...TextOption) {
	style := c.defaultStyle()
	for _, opt := range opts {
		opt(&style)
	}
	c.nodes = append(c.nodes, &document.Text{
		Content:   document.TotalPagesPlaceholder,
		TextStyle: style,
	})
}

// RichText adds a rich text element with mixed inline styles.
// The callback receives a RichTextBuilder for adding styled spans.
// Options apply to the paragraph-level (block) style.
func (c *ColBuilder) RichText(fn func(rt *RichTextBuilder), opts ...TextOption) {
	blockStyle := c.defaultStyle()
	for _, opt := range opts {
		opt(&blockStyle)
	}
	rtb := &RichTextBuilder{defaultStyle: blockStyle}
	fn(rtb)
	c.nodes = append(c.nodes, &document.RichText{
		Fragments:  rtb.fragments,
		BlockStyle: blockStyle,
	})
}

// QRCode adds a QR code image to the column.
// The data is encoded as a QR code and rendered as a PNG image.
func (c *ColBuilder) QRCode(data string, opts ...QRCodeOption) {
	cfg := qrCodeConfig{
		ecLevel: qrcode.LevelM,
		scale:   10,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	qr, err := qrcode.Encode(data, cfg.ecLevel)
	if err != nil {
		return // silently skip on error, consistent with other builder methods
	}

	pngData, err := qr.PNG(cfg.scale)
	if err != nil {
		return
	}

	w, h := extractImageDimensions(pngData, document.ImagePNG)

	imgNode := &document.Image{
		Source: document.ImageSource{
			Data:   pngData,
			Format: document.ImagePNG,
			Width:  w,
			Height: h,
		},
		FitMode: document.FitContain,
	}

	if cfg.size.Amount > 0 {
		imgNode.DisplayWidth = cfg.size
		imgNode.DisplayHeight = cfg.size
	}

	c.nodes = append(c.nodes, imgNode)
}

// Barcode adds a barcode image to the column.
// The data is encoded as a barcode and rendered as a PNG image.
func (c *ColBuilder) Barcode(data string, opts ...BarcodeOption) {
	cfg := barcodeConfig{
		format: barcode.Code128,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	bc, err := barcode.Encode(data, cfg.format)
	if err != nil {
		return
	}

	pngData, err := bc.PNG(2, 100)
	if err != nil {
		return
	}

	w, h := extractImageDimensions(pngData, document.ImagePNG)

	imgNode := &document.Image{
		Source: document.ImageSource{
			Data:   pngData,
			Format: document.ImagePNG,
			Width:  w,
			Height: h,
		},
		FitMode: document.FitContain,
	}

	if cfg.width.Amount > 0 {
		imgNode.DisplayWidth = cfg.width
	}
	if cfg.height.Amount > 0 {
		imgNode.DisplayHeight = cfg.height
	}

	c.nodes = append(c.nodes, imgNode)
}

// Spacer adds vertical space.
func (c *ColBuilder) Spacer(height document.Value) {
	c.nodes = append(c.nodes, &document.Box{
		BoxStyle: document.BoxStyle{
			Height: height,
		},
	})
}

// buildNodes returns the accumulated document nodes.
func (c *ColBuilder) buildNodes() []document.DocumentNode {
	return c.nodes
}

// defaultStyle returns the default style from the document configuration.
func (c *ColBuilder) defaultStyle() document.Style {
	s := document.DefaultStyle()
	if c.doc != nil {
		if c.doc.config.DefaultFont != "" {
			s.FontFamily = c.doc.config.DefaultFont
		}
		if c.doc.config.FontSize > 0 {
			s.FontSize = c.doc.config.FontSize
		}
	}
	return s
}

// detectImageFormat guesses the image format from the file header bytes.
func detectImageFormat(data []byte) document.ImageFormat {
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return document.ImageJPEG
	}
	if isSVGData(data) {
		return document.ImageSVG
	}
	return document.ImagePNG
}

// isSVGData returns true if data appears to be an SVG document.
func isSVGData(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	if bytes.HasPrefix(trimmed, []byte("<?xml")) {
		// Look for <svg within first 2KB.
		search := trimmed
		if len(search) > 2048 {
			search = search[:2048]
		}
		lower := bytes.ToLower(search)
		return bytes.Contains(lower, []byte("<svg"))
	}
	lower := bytes.ToLower(trimmed)
	return bytes.HasPrefix(lower, []byte("<svg"))
}

// extractImageDimensions returns the pixel width and height from image binary data.
func extractImageDimensions(data []byte, format document.ImageFormat) (int, int) {
	switch format {
	case document.ImagePNG:
		return extractPNGDimensions(data)
	case document.ImageJPEG:
		return extractJPEGDimensions(data)
	case document.ImageSVG:
		return extractSVGDimensions(data)
	}
	return 0, 0
}

// extractSVGDimensions parses the SVG root element to get the viewport dimensions
// (from viewBox, or from width/height attributes). Returns dimensions in user units
// rounded to the nearest integer.
func extractSVGDimensions(data []byte) (int, int) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if strings.ToLower(se.Name.Local) != "svg" {
			break
		}
		// Build attribute map (lowercase keys).
		attrs := make(map[string]string, len(se.Attr))
		for _, a := range se.Attr {
			attrs[strings.ToLower(a.Name.Local)] = a.Value
		}
		// Try viewBox first.
		if vb, ok2 := attrs["viewbox"]; ok2 {
			parts := strings.Fields(strings.ReplaceAll(vb, ",", " "))
			if len(parts) == 4 {
				w, errW := strconv.ParseFloat(parts[2], 64)
				h, errH := strconv.ParseFloat(parts[3], 64)
				if errW == nil && errH == nil && w > 0 && h > 0 {
					return int(math.Round(w)), int(math.Round(h))
				}
			}
		}
		// Fall back to width/height attributes.
		w := parseSVGLengthAttr(attrs["width"])
		h := parseSVGLengthAttr(attrs["height"])
		if w > 0 && h > 0 {
			return int(math.Round(w)), int(math.Round(h))
		}
		break
	}
	return 100, 100 // fallback
}

// parseSVGLengthAttr converts an SVG length string (e.g. "100px", "50mm") to px units.
func parseSVGLengthAttr(s string) float64 {
	s = strings.TrimSpace(s)
	units := []struct {
		suffix string
		factor float64
	}{
		{"px", 1},
		{"pt", 4.0 / 3},
		{"mm", 96.0 / 25.4},
		{"cm", 96.0 / 2.54},
		{"in", 96},
		{"rem", 16},
		{"em", 16},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			v, err := strconv.ParseFloat(strings.TrimSuffix(s, u.suffix), 64)
			if err == nil {
				return v * u.factor
			}
		}
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// extractPNGDimensions reads width and height from the PNG IHDR chunk.
// PNG layout: 8-byte signature, then IHDR chunk with width at offset 16
// and height at offset 20, both as 4-byte big-endian integers.
func extractPNGDimensions(data []byte) (int, int) {
	if len(data) < 24 {
		return 0, 0
	}
	w := int(data[16])<<24 | int(data[17])<<16 | int(data[18])<<8 | int(data[19])
	h := int(data[20])<<24 | int(data[21])<<16 | int(data[22])<<8 | int(data[23])
	return w, h
}

// extractJPEGDimensions reads width and height from JPEG SOF markers.
func extractJPEGDimensions(data []byte) (int, int) {
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0, 0
	}
	i := 2
	for i < len(data)-1 {
		if data[i] != 0xFF {
			i++
			continue
		}
		marker := data[i+1]
		// Skip padding bytes.
		if marker == 0x00 || marker == 0xFF {
			i++
			continue
		}
		// SOF0, SOF1, SOF2 markers contain image dimensions.
		if marker >= 0xC0 && marker <= 0xC2 {
			if i+9 < len(data) {
				h := int(data[i+5])<<8 | int(data[i+6])
				w := int(data[i+7])<<8 | int(data[i+8])
				return w, h
			}
		}
		// Skip to next marker segment.
		if i+3 < len(data) {
			segLen := int(data[i+2])<<8 | int(data[i+3])
			i += 2 + segLen
		} else {
			break
		}
	}
	return 0, 0
}
