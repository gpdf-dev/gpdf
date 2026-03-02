package layout

import "github.com/gpdf-dev/gpdf/document"

// tableContext holds resolved geometry for table layout.
type tableContext struct {
	contentX     float64
	contentY     float64
	contentWidth float64
	colWidths    []float64
	colOffsets   []float64
	numCols      int
	margin       document.ResolvedEdges
	padding      document.ResolvedEdges
	borderWidths document.ResolvedEdges
}

// layoutTable arranges a Table node's rows and cells into a grid. Each row
// is laid out horizontally, with cell widths derived from the table's column
// definitions. When no column widths are specified, columns share the
// available width equally. Header rows are repeated on overflow pages.
func (bl *BlockLayout) layoutTable(tbl *document.Table, constraints Constraints) Result {
	tc := bl.resolveTableContext(tbl, constraints)
	if tc.numCols == 0 {
		return Result{Bounds: document.Rectangle{Width: constraints.AvailableWidth}}
	}

	contentAvailH := constraints.AvailableHeight -
		tc.margin.Vertical() - tc.padding.Vertical() - tc.borderWidths.Vertical()

	var placed []PlacedNode
	cursorY := 0.0

	// Layout header rows (always placed first).
	headerPlaced, headerHeight := bl.layoutTableSection(tbl.Header, tc, constraints)
	offsetPlacedNodes(headerPlaced, tc.contentX, tc.contentY)
	placed = append(placed, headerPlaced...)
	cursorY += headerHeight

	// Layout body rows one at a time, checking for overflow.
	for i, row := range tbl.Body {
		rowPlaced, rowHeight := bl.layoutTableRow(row, tc.colWidths, tc.colOffsets, tc.numCols, constraints)

		if cursorY+rowHeight > contentAvailH && len(placed) > 0 {
			// This row doesn't fit. Build an overflow table with header + remaining body + footer.
			overflow := &document.Table{
				Columns:    tbl.Columns,
				Header:     tbl.Header,
				Body:       tbl.Body[i:],
				Footer:     tbl.Footer,
				TableStyle: tbl.TableStyle,
			}
			return bl.tableOverflowResult(tbl, constraints, tc, placed, cursorY, overflow)
		}

		offsetPlacedNodes(rowPlaced, tc.contentX, tc.contentY+cursorY)
		placed = append(placed, rowPlaced...)
		cursorY += rowHeight
	}

	// Layout footer rows.
	footerPlaced, footerHeight := bl.layoutTableSection(tbl.Footer, tc, constraints)
	offsetPlacedNodes(footerPlaced, tc.contentX, tc.contentY+cursorY)
	placed = append(placed, footerPlaced...)
	cursorY += footerHeight

	totalHeight := tc.margin.Top + tc.borderWidths.Top + tc.padding.Top +
		cursorY +
		tc.padding.Bottom + tc.borderWidths.Bottom + tc.margin.Bottom
	return Result{
		Bounds:   document.Rectangle{Width: constraints.AvailableWidth, Height: totalHeight},
		Children: placed,
	}
}

// resolveTableContext computes spacing, column widths, and offsets for a table.
func (bl *BlockLayout) resolveTableContext(tbl *document.Table, constraints Constraints) tableContext {
	style := tbl.Style()
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	margin := style.Margin.Resolve(constraints.AvailableWidth, constraints.AvailableHeight, fontSize)
	padding := style.Padding.Resolve(constraints.AvailableWidth, constraints.AvailableHeight, fontSize)
	bw := resolveBorderWidths(style.Border, constraints.AvailableWidth, fontSize)

	outerWidth := constraints.AvailableWidth - margin.Horizontal()
	cw := outerWidth - padding.Horizontal() - bw.Horizontal()
	if cw < 0 {
		cw = 0
	}

	numCols := tableColumnCount(tbl)

	// Use content-based proportional widths when all columns are auto.
	var colWidths []float64
	if allColumnsAuto(tbl.Columns) {
		colWidths = estimateAutoColumnWidths(tbl, numCols, cw)
	} else {
		colWidths = resolveTableColumnWidths(tbl.Columns, numCols, cw, fontSize)
	}
	colOffsets := make([]float64, numCols)
	for i := 1; i < numCols; i++ {
		colOffsets[i] = colOffsets[i-1] + colWidths[i-1]
	}

	return tableContext{
		contentX:     margin.Left + bw.Left + padding.Left,
		contentY:     margin.Top + bw.Top + padding.Top,
		contentWidth: cw,
		colWidths:    colWidths,
		colOffsets:   colOffsets,
		numCols:      numCols,
		margin:       margin,
		padding:      padding,
		borderWidths: bw,
	}
}

// layoutTableSection lays out a slice of table rows, returning all placed
// nodes and the total section height.
func (bl *BlockLayout) layoutTableSection(rows []document.TableRow, tc tableContext, constraints Constraints) ([]PlacedNode, float64) {
	var placed []PlacedNode
	var totalHeight float64
	for _, row := range rows {
		rp, rh := bl.layoutTableRow(row, tc.colWidths, tc.colOffsets, tc.numCols, constraints)
		offsetPlacedNodes(rp, 0, totalHeight)
		placed = append(placed, rp...)
		totalHeight += rh
	}
	return placed, totalHeight
}

// tableOverflowResult builds a Result that overflows the given table.
func (bl *BlockLayout) tableOverflowResult(tbl *document.Table, constraints Constraints, tc tableContext, placed []PlacedNode, cursorY float64, overflow *document.Table) Result {
	totalHeight := tc.margin.Top + tc.borderWidths.Top + tc.padding.Top +
		cursorY +
		tc.padding.Bottom + tc.borderWidths.Bottom + tc.margin.Bottom
	return Result{
		Bounds:   document.Rectangle{Width: constraints.AvailableWidth, Height: totalHeight},
		Children: placed,
		Overflow: overflow,
	}
}

// offsetPlacedNodes shifts all placed nodes by dx and dy.
func offsetPlacedNodes(nodes []PlacedNode, dx, dy float64) {
	for i := range nodes {
		nodes[i].Position.X += dx
		nodes[i].Position.Y += dy
	}
}

// layoutTableRow lays out a single table row, returning the placed cells and
// the row height (tallest cell).
func (bl *BlockLayout) layoutTableRow(row document.TableRow, colWidths []float64, colOffsets []float64, numCols int, constraints Constraints) ([]PlacedNode, float64) {
	var placed []PlacedNode
	maxHeight := 0.0

	colIdx := 0
	for i := range row.Cells {
		if colIdx >= numCols {
			break
		}

		cell := &row.Cells[i]
		colSpan := cell.ColSpan
		if colSpan < 1 {
			colSpan = 1
		}
		if colIdx+colSpan > numCols {
			colSpan = numCols - colIdx
		}

		// Calculate cell width by summing spanned columns.
		cellWidth := 0.0
		for j := colIdx; j < colIdx+colSpan; j++ {
			cellWidth += colWidths[j]
		}

		cellX := colOffsets[colIdx]

		// Create a virtual container for the cell's content.
		cellBox := &document.Box{
			Content: cell.Content,
			BoxStyle: document.BoxStyle{
				Padding:    cell.CellStyle.Padding,
				Border:     cell.CellStyle.Border,
				Background: cell.CellStyle.Background,
			},
		}

		cellConstraints := Constraints{
			AvailableWidth:  cellWidth,
			AvailableHeight: constraints.AvailableHeight,
			FontResolver:    constraints.FontResolver,
		}

		cellResult := bl.Layout(cellBox, cellConstraints)

		placed = append(placed, PlacedNode{
			Node: &document.CellNode{Cell: cell},
			Position: document.Point{
				X: cellX,
				Y: 0,
			},
			Size:     document.Size{Width: cellWidth, Height: cellResult.Bounds.Height},
			Children: cellResult.Children,
		})

		if cellResult.Bounds.Height > maxHeight {
			maxHeight = cellResult.Bounds.Height
		}

		colIdx += colSpan
	}

	// Stretch all cells to match the tallest cell in the row and apply
	// vertical alignment.
	for i := range placed {
		contentHeight := placed[i].Size.Height
		placed[i].Size.Height = maxHeight
		if placed[i].Node != nil {
			applyVerticalAlign(placed[i].Children, placed[i].Node.Style().VerticalAlign, contentHeight, maxHeight)
		}
	}

	return placed, maxHeight
}

// tableColumnCount determines the number of columns from the table definition.
func tableColumnCount(tbl *document.Table) int {
	if len(tbl.Columns) > 0 {
		return len(tbl.Columns)
	}
	if len(tbl.Header) > 0 && len(tbl.Header[0].Cells) > 0 {
		return countRowColumns(tbl.Header[0])
	}
	if len(tbl.Body) > 0 && len(tbl.Body[0].Cells) > 0 {
		return countRowColumns(tbl.Body[0])
	}
	return 0
}

// countRowColumns counts the effective number of columns in a row,
// accounting for column spans.
func countRowColumns(row document.TableRow) int {
	n := 0
	for _, cell := range row.Cells {
		span := cell.ColSpan
		if span < 1 {
			span = 1
		}
		n += span
	}
	return n
}

// applyVerticalAlign offsets the Y position of children within a cell based
// on the given alignment. contentHeight is the natural height of the cell
// content, cellHeight is the stretched row height.
func applyVerticalAlign(children []PlacedNode, align document.VerticalAlign, contentHeight, cellHeight float64) {
	if len(children) == 0 || contentHeight >= cellHeight {
		return
	}

	var dy float64
	switch align {
	case document.VAlignMiddle:
		dy = (cellHeight - contentHeight) / 2
	case document.VAlignBottom:
		dy = cellHeight - contentHeight
	default:
		return
	}

	for i := range children {
		children[i].Position.Y += dy
	}
}

// resolveTableColumnWidths computes the width of each column in points.
// Columns with explicit widths are resolved first; remaining space is
// distributed equally among auto-width columns.
func resolveTableColumnWidths(cols []document.TableColumn, numCols int, contentWidth, fontSize float64) []float64 {
	widths := make([]float64, numCols)

	if len(cols) == 0 {
		// No column definitions: equal widths.
		w := contentWidth / float64(numCols)
		for i := range widths {
			widths[i] = w
		}
		return widths
	}

	usedWidth := 0.0
	autoCount := 0

	for i := 0; i < numCols; i++ {
		if i < len(cols) && cols[i].Width.Unit != document.UnitAuto && cols[i].Width.Amount > 0 {
			widths[i] = cols[i].Width.Resolve(contentWidth, fontSize)
			usedWidth += widths[i]
		} else {
			autoCount++
		}
	}

	if autoCount > 0 {
		remaining := contentWidth - usedWidth
		if remaining < 0 {
			remaining = 0
		}
		autoWidth := remaining / float64(autoCount)
		for i := range widths {
			if widths[i] == 0 {
				widths[i] = autoWidth
			}
		}
	}

	return widths
}

// allColumnsAuto reports whether every column in cols uses automatic width.
// Returns false for empty slices (no explicit column definitions).
func allColumnsAuto(cols []document.TableColumn) bool {
	if len(cols) == 0 {
		return false
	}
	for _, c := range cols {
		if c.Width.Unit != document.UnitAuto {
			return false
		}
	}
	return true
}

// estimateAutoColumnWidths distributes the available width proportionally
// based on the maximum text content length in each column. This produces
// more natural-looking tables than equal distribution.
func estimateAutoColumnWidths(tbl *document.Table, numCols int, contentWidth float64) []float64 {
	maxLen := make([]float64, numCols)

	allRows := make([]document.TableRow, 0, len(tbl.Header)+len(tbl.Body)+len(tbl.Footer))
	allRows = append(allRows, tbl.Header...)
	allRows = append(allRows, tbl.Body...)
	allRows = append(allRows, tbl.Footer...)

	for _, row := range allRows {
		colIdx := 0
		for _, cell := range row.Cells {
			if colIdx >= numCols {
				break
			}
			span := cell.ColSpan
			if span < 1 {
				span = 1
			}
			// Only use non-spanning cells for width estimation.
			if span == 1 {
				textLen := float64(textContentLength(cell.Content))
				if textLen > maxLen[colIdx] {
					maxLen[colIdx] = textLen
				}
			}
			colIdx += span
		}
	}

	totalLen := 0.0
	for i, l := range maxLen {
		if l < 1 {
			maxLen[i] = 1
			l = 1
		}
		totalLen += l
	}

	if totalLen == 0 {
		w := contentWidth / float64(numCols)
		widths := make([]float64, numCols)
		for i := range widths {
			widths[i] = w
		}
		return widths
	}

	widths := make([]float64, numCols)
	for i, l := range maxLen {
		widths[i] = contentWidth * l / totalLen
	}
	return widths
}

// textContentLength counts the total number of characters in a tree of
// document nodes, used as a rough proxy for content width.
func textContentLength(nodes []document.DocumentNode) int {
	total := 0
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if text, ok := node.(*document.Text); ok {
			total += len(text.Content)
		}
		total += textContentLength(node.Children())
	}
	return total
}
