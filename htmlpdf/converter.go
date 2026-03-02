package htmlpdf

import (
	"strings"

	"github.com/gpdf-dev/gpdf/css"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/html"
)

// nodeAdapter adapts html.Node to the css.Matchable interface.
type nodeAdapter struct {
	node *html.Node
}

func (a *nodeAdapter) NodeTag() string { return a.node.Tag }
func (a *nodeAdapter) NodeID() string  { return a.node.ID() }
func (a *nodeAdapter) NodeClasses() []string {
	classAttr, ok := a.node.Attr("class")
	if !ok {
		return nil
	}
	return strings.Fields(classAttr)
}
func (a *nodeAdapter) NodeAttr(name string) (string, bool) { return a.node.Attr(name) }
func (a *nodeAdapter) NodeParent() css.Matchable {
	if a.node.Parent == nil {
		return nil
	}
	return &nodeAdapter{node: a.node.Parent}
}

// UA stylesheet (built-in browser defaults for PDF rendering).
const uaCSS = `
html, body          { display: block; }
h1                  { font-size: 24pt; font-weight: bold; margin-top: 8pt; margin-bottom: 8pt; }
h2                  { font-size: 18pt; font-weight: bold; margin-top: 6pt; margin-bottom: 6pt; }
h3                  { font-size: 14pt; font-weight: bold; margin-top: 5pt; margin-bottom: 5pt; }
h4                  { font-weight: bold; margin-top: 5pt; margin-bottom: 5pt; }
h5                  { font-size: 10pt; font-weight: bold; margin-top: 4pt; margin-bottom: 4pt; }
h6                  { font-size: 8pt; font-weight: bold; margin-top: 4pt; margin-bottom: 4pt; }
p                   { display: block; margin-top: 12pt; margin-bottom: 12pt; }
div                 { display: block; }
strong, b           { font-weight: bold; }
em, i               { font-style: italic; }
u                   { text-decoration: underline; }
s, del, strike      { text-decoration: line-through; }
hr                  { margin-top: 6pt; margin-bottom: 6pt; }
ul, ol              { display: block; margin-top: 12pt; margin-bottom: 12pt; padding-left: 40pt; }
li                  { display: list-item; }
blockquote          { display: block; margin-top: 12pt; margin-bottom: 12pt; margin-left: 40pt; }
pre                 { display: block; margin-top: 12pt; margin-bottom: 12pt; }
code                { font-family: monospace; }
table               { display: table; border-collapse: separate; border-spacing: 2pt; }
thead               { display: table-header-group; }
tbody               { display: table-row-group; }
tfoot               { display: table-footer-group; }
tr                  { display: table-row; }
th                  { display: table-cell; font-weight: bold; text-align: center; padding: 1pt 2pt; }
td                  { display: table-cell; padding: 1pt 2pt; }
caption             { display: block; text-align: center; }
`

// converter handles the HTML→DocumentNode conversion.
type converter struct {
	config       *Config
	uaStylesheet *css.Stylesheet
	stylesheets  []*css.Stylesheet
	styleCache   map[*html.Node]css.ComputedStyles
}

// newConverter creates a new converter with parsed stylesheets.
func newConverter(config *Config, htmlRoot *html.Node) *converter {
	ua, _ := css.ParseStylesheet(uaCSS)

	// Collect <style> elements from the document
	var stylesheets []*css.Stylesheet
	for _, styleNode := range htmlRoot.StyleElements() {
		cssText := styleNode.TextContent()
		if ss, err := css.ParseStylesheet(cssText); err == nil {
			stylesheets = append(stylesheets, ss)
		}
	}

	// Add user-provided additional CSS
	if config.ExtraCSS != "" {
		if ss, err := css.ParseStylesheet(config.ExtraCSS); err == nil {
			stylesheets = append(stylesheets, ss)
		}
	}

	return &converter{
		config:       config,
		uaStylesheet: ua,
		stylesheets:  stylesheets,
		styleCache:   make(map[*html.Node]css.ComputedStyles),
	}
}

// convert transforms the HTML DOM tree into a document.Document.
func (c *converter) convert(root *html.Node) *document.Document {
	// Find <body> or use root
	body := findBody(root)
	if body == nil {
		body = root
	}

	// Build the default style
	defaultStyle := c.config.defaultStyle()

	// Compute styles for body
	bodyComputed := c.computeStyles(body, nil)

	// Convert body's children to DocumentNode tree
	children := c.convertChildren(body, bodyComputed)

	// Wrap in a page
	page := &document.Page{
		Size:    c.config.PageSize,
		Margins: c.config.Margins,
		Content: children,
		PageStyle: document.Style{
			FontFamily: defaultStyle.FontFamily,
			FontSize:   defaultStyle.FontSize,
			LineHeight: defaultStyle.FontSize * 1.2,
			Color:      defaultStyle.Color,
		},
	}

	return &document.Document{
		Pages:        []*document.Page{page},
		DefaultStyle: defaultStyle,
	}
}

// convertChildren converts child nodes into DocumentNode slices,
// handling inline/block mixing and anonymous block boxes.
func (c *converter) convertChildren(parent *html.Node, parentComputed css.ComputedStyles) []document.DocumentNode {
	var result []document.DocumentNode
	var inlineBuffer []*html.Node

	flushInline := func() {
		if len(inlineBuffer) == 0 {
			return
		}
		// Collect inline content as RichText fragments
		var fragments []document.RichTextFragment
		for _, inlineNode := range inlineBuffer {
			computed := c.computeStyles(inlineNode, parentComputed)
			frags := collectInlineFragments(inlineNode, computed, func(n *html.Node) css.ComputedStyles {
				return c.computeStyles(n, computed)
			})
			fragments = append(fragments, frags...)
		}
		if len(fragments) > 0 {
			blockStyle := applyStyle(parentComputed)
			rt := wrapInlineAsRichText(fragments, blockStyle)
			if rt != nil {
				result = append(result, rt)
			}
		}
		inlineBuffer = nil
	}

	for _, child := range parent.Children {
		// Skip comments
		if child.Type == html.CommentNode {
			continue
		}

		computed := c.computeStyles(child, parentComputed)

		// Skip display: none
		if child.Type == html.ElementNode && isDisplayNone(computed) {
			continue
		}

		if isInlineElement(child, computed) {
			inlineBuffer = append(inlineBuffer, child)
		} else {
			// Block element encountered — flush any pending inline content
			flushInline()
			// Recursively convert the block element
			blockNode := c.convertElement(child, computed)
			if blockNode != nil {
				result = append(result, blockNode)
			}
		}
	}

	// Flush remaining inline content
	flushInline()

	return result
}

// convertElement converts a single block-level HTML element to a DocumentNode.
func (c *converter) convertElement(node *html.Node, computed css.ComputedStyles) document.DocumentNode {
	if node.Type == html.TextNode {
		// Standalone text (shouldn't happen for block, but handle gracefully)
		style := applyStyle(computed)
		return &document.Text{
			Content:   node.Content,
			TextStyle: style,
		}
	}

	if node.Type != html.ElementNode {
		return nil
	}

	// Special handling for certain elements
	switch node.Tag {
	case "br", "wbr":
		return nil // handled inline
	case "img":
		return mapImage(node, applyStyle(computed))
	case "hr":
		return mapHR(computed)
	case "table":
		return c.convertTable(node, computed)
	}

	// Recurse into children
	children := c.convertChildren(node, computed)

	return mapBlockElement(node, computed, children)
}

// computeStyles computes CSS cascade for a node.
func (c *converter) computeStyles(node *html.Node, parentComputed css.ComputedStyles) css.ComputedStyles {
	if cached, ok := c.styleCache[node]; ok {
		return cached
	}

	var inlineDecls []css.Declaration
	if node.Type == html.ElementNode {
		if styleAttr, ok := node.Attr("style"); ok {
			inlineDecls = css.ParseDeclarations(styleAttr)
		}
	}

	matchable := &nodeAdapter{node: node}
	computed := css.Cascade(matchable, c.stylesheets, inlineDecls, parentComputed, c.uaStylesheet)

	// Resolve computed values
	parentFontSize := 12.0
	containerWidth := c.config.PageSize.Width - c.config.Margins.Left.Resolve(0, 0) - c.config.Margins.Right.Resolve(0, 0)
	if parentComputed != nil {
		parentFontSize = css.ParsePt(parentComputed["font-size"], 12)
	}
	computed = css.ResolveComputed(computed, parentFontSize, containerWidth)

	c.styleCache[node] = computed
	return computed
}

// convertTable converts a <table> HTML element into a document.Table.
func (c *converter) convertTable(node *html.Node, computed css.ComputedStyles) document.DocumentNode {
	tbl := &document.Table{}

	// Apply table-level styles
	boxStyle := applyBoxStyle(computed)
	collapse := strings.EqualFold(strings.TrimSpace(computed["border-collapse"]), "collapse")
	tbl.TableStyle = document.TableStyle{
		BoxStyle:       boxStyle,
		BorderCollapse: collapse,
	}

	// Walk <table> children to find thead/tbody/tfoot/tr/caption
	var captionNode *html.Node
	var directRows []*html.Node // <tr> directly under <table> (no explicit section)

	for _, child := range node.Children {
		if child.Type != html.ElementNode {
			continue
		}
		childComputed := c.computeStyles(child, computed)
		if isDisplayNone(childComputed) {
			continue
		}

		switch child.Tag {
		case "caption":
			captionNode = child
		case "thead":
			tbl.Header = append(tbl.Header, c.convertTableSection(child, computed)...)
		case "tfoot":
			tbl.Footer = append(tbl.Footer, c.convertTableSection(child, computed)...)
		case "tbody":
			tbl.Body = append(tbl.Body, c.convertTableSection(child, computed)...)
		case "tr":
			directRows = append(directRows, child)
		case "colgroup", "col":
			// TODO: column width hints (Phase 4-B basic — ignore for now)
		}
	}

	// Direct <tr> children (no explicit <tbody>) go into Body
	for _, tr := range directRows {
		trComputed := c.computeStyles(tr, computed)
		row := c.convertTableRow(tr, trComputed)
		tbl.Body = append(tbl.Body, row)
	}

	// Determine column count from the widest row
	numCols := 0
	allRows := append(append(tbl.Header, tbl.Body...), tbl.Footer...)
	for _, row := range allRows {
		cols := 0
		for _, cell := range row.Cells {
			span := cell.ColSpan
			if span < 1 {
				span = 1
			}
			cols += span
		}
		if cols > numCols {
			numCols = cols
		}
	}

	// Set auto columns if none defined
	if len(tbl.Columns) == 0 && numCols > 0 {
		tbl.Columns = make([]document.TableColumn, numCols)
		for i := range tbl.Columns {
			tbl.Columns[i] = document.TableColumn{Width: document.Auto}
		}
	}

	// Apply border-collapse: remove duplicate borders between adjacent cells.
	if collapse {
		collapseCellBorders(tbl)
	}

	// If there's a caption, wrap table in a Box with caption + table
	if captionNode != nil {
		captionComputed := c.computeStyles(captionNode, computed)
		captionChildren := c.convertChildren(captionNode, captionComputed)
		captionStyle := applyBoxStyle(captionComputed)

		captionBox := &document.Box{
			Content:  captionChildren,
			BoxStyle: captionStyle,
		}

		side := strings.ToLower(strings.TrimSpace(computed["caption-side"]))
		var content []document.DocumentNode
		if side == "bottom" {
			content = []document.DocumentNode{tbl, captionBox}
		} else {
			content = []document.DocumentNode{captionBox, tbl}
		}
		return &document.Box{
			Content:  content,
			BoxStyle: document.BoxStyle{Direction: document.DirectionVertical},
		}
	}

	return tbl
}

// convertTableSection converts a <thead>/<tbody>/<tfoot> into []TableRow.
func (c *converter) convertTableSection(section *html.Node, tableComputed css.ComputedStyles) []document.TableRow {
	var rows []document.TableRow
	for _, child := range section.Children {
		if child.Type != html.ElementNode || child.Tag != "tr" {
			continue
		}
		trComputed := c.computeStyles(child, tableComputed)
		rows = append(rows, c.convertTableRow(child, trComputed))
	}
	return rows
}

// convertTableRow converts a <tr> element into a document.TableRow.
func (c *converter) convertTableRow(tr *html.Node, trComputed css.ComputedStyles) document.TableRow {
	var cells []document.TableCell
	for _, child := range tr.Children {
		if child.Type != html.ElementNode {
			continue
		}
		if child.Tag != "td" && child.Tag != "th" {
			continue
		}
		cellComputed := c.computeStyles(child, trComputed)
		cells = append(cells, c.convertTableCell(child, cellComputed))
	}
	return document.TableRow{Cells: cells}
}

// convertTableCell converts a <td> or <th> element into a document.TableCell.
func (c *converter) convertTableCell(cell *html.Node, cellComputed css.ComputedStyles) document.TableCell {
	// Convert cell content
	children := c.convertChildren(cell, cellComputed)

	// Parse colspan/rowspan attributes
	colspan := parseIntAttr(cell, "colspan", 1)
	rowspan := parseIntAttr(cell, "rowspan", 1)

	// Build cell style
	style := applyStyle(cellComputed)
	style.VerticalAlign = parseVerticalAlign(cellComputed["vertical-align"])

	return document.TableCell{
		Content:   children,
		ColSpan:   colspan,
		RowSpan:   rowspan,
		CellStyle: style,
	}
}

// collapseCellBorders removes duplicate borders between adjacent cells when
// border-collapse is enabled. For non-first rows, top borders are removed;
// for non-first columns, left borders are removed.
func collapseCellBorders(tbl *document.Table) {
	rowIdx := 0
	for i := range tbl.Header {
		collapseBordersInRow(tbl.Header[i].Cells, rowIdx)
		rowIdx++
	}
	for i := range tbl.Body {
		collapseBordersInRow(tbl.Body[i].Cells, rowIdx)
		rowIdx++
	}
	for i := range tbl.Footer {
		collapseBordersInRow(tbl.Footer[i].Cells, rowIdx)
		rowIdx++
	}
}

// collapseBordersInRow removes duplicate borders within a single row.
func collapseBordersInRow(cells []document.TableCell, rowIdx int) {
	colIdx := 0
	for i := range cells {
		if rowIdx > 0 {
			cells[i].CellStyle.Border.Top = document.BorderSide{}
		}
		if colIdx > 0 {
			cells[i].CellStyle.Border.Left = document.BorderSide{}
		}
		span := cells[i].ColSpan
		if span < 1 {
			span = 1
		}
		colIdx += span
	}
}

// parseIntAttr parses an integer HTML attribute, returning fallback if not present or invalid.
func parseIntAttr(node *html.Node, name string, fallback int) int {
	val, ok := node.Attr(name)
	if !ok {
		return fallback
	}
	n := 0
	for _, ch := range val {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		} else {
			break
		}
	}
	if n < 1 {
		return fallback
	}
	return n
}

// findBody finds the <body> element in the DOM tree.
func findBody(root *html.Node) *html.Node {
	var body *html.Node
	root.Walk(func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Tag == "body" {
			body = n
			return false
		}
		return true
	})
	return body
}
