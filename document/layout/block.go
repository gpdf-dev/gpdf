package layout

import "github.com/gpdf-dev/gpdf/document"

// BlockLayout arranges children vertically from top to bottom, similar
// to CSS block-level layout. Each child occupies the full available width
// and is placed below the previous child. When the available height is
// exhausted, remaining content is returned as overflow for placement on
// the next page.
type BlockLayout struct{}

// NewBlockLayout creates a new BlockLayout engine.
func NewBlockLayout() *BlockLayout {
	return &BlockLayout{}
}

// blockContext holds intermediate state used during vertical child placement.
type blockContext struct {
	node          document.DocumentNode
	constraints   Constraints
	contentWidth  float64
	contentHeight float64
	contentX      float64
	contentY      float64
	margin        document.ResolvedEdges
	padding       document.ResolvedEdges
	borderWidths  document.ResolvedEdges
}

// wrapHeight returns the total height including margin, border, and padding
// around the given content height.
func (bc *blockContext) wrapHeight(contentH float64) float64 {
	return bc.margin.Top + bc.borderWidths.Top + bc.padding.Top +
		contentH +
		bc.padding.Bottom + bc.borderWidths.Bottom + bc.margin.Bottom
}

// overflowResult builds a Result that overflows the given remaining nodes.
func (bc *blockContext) overflowResult(placed []PlacedNode, cursorY float64, remaining []document.DocumentNode) Result {
	return Result{
		Bounds: document.Rectangle{
			Width:  bc.constraints.AvailableWidth,
			Height: bc.wrapHeight(cursorY),
		},
		Children: placed,
		Overflow: createOverflowNode(bc.node, remaining),
	}
}

// Layout places the given node's children vertically within the
// constraints. It resolves margins, padding, and border widths from
// the node's style, then positions each child sequentially.
func (bl *BlockLayout) Layout(node document.DocumentNode, constraints Constraints) Result {
	// Horizontal layout for row-like containers.
	if box, ok := node.(*document.Box); ok && box.BoxStyle.Direction == document.DirectionHorizontal {
		return bl.layoutHorizontal(node, constraints)
	}

	bc := bl.newBlockContext(node, constraints)

	var placed []PlacedNode
	var absoluteNodes []PlacedNode
	cursorY := 0.0
	children := node.Children()

	for i, child := range children {
		if child == nil {
			continue
		}

		// Absolute-positioned nodes are removed from normal flow.
		if box, ok := child.(*document.Box); ok && box.BoxStyle.Position.Mode == document.PositionAbsolute {
			absNode := bl.layoutAbsolute(box, &bc)
			absoluteNodes = append(absoluteNodes, absNode)
			continue
		}

		result, done := bl.layoutVerticalChild(&bc, child, children, i, placed, cursorY)
		if done {
			// Append absolute nodes to the overflow result so they
			// still appear on this page.
			result.Children = append(result.Children, absoluteNodes...)
			return result
		}

		placed = append(placed, PlacedNode{
			Node: child,
			Position: document.Point{
				X: bc.contentX,
				Y: bc.contentY + cursorY,
			},
			Size:     document.Size{Width: result.Bounds.Width, Height: result.Bounds.Height},
			Children: result.Children,
		})
		cursorY += result.Bounds.Height

		if result.Overflow != nil {
			remaining := make([]document.DocumentNode, 0, 1+len(children[i+1:]))
			remaining = append(remaining, result.Overflow)
			remaining = append(remaining, children[i+1:]...)
			r := bc.overflowResult(placed, cursorY, remaining)
			r.Children = append(r.Children, absoluteNodes...)
			return r
		}

		if after, ok := bl.checkBreakAfter(child, children, i, &bc, placed, cursorY); ok {
			after.Children = append(after.Children, absoluteNodes...)
			return after
		}
	}

	result := bl.finishVerticalLayout(&bc, placed, cursorY)
	// Absolute nodes are rendered after flow nodes (drawn on top).
	result.Children = append(result.Children, absoluteNodes...)
	return result
}

// newBlockContext resolves spacing and computes the content area for
// vertical layout.
func (bl *BlockLayout) newBlockContext(node document.DocumentNode, constraints Constraints) blockContext {
	style := node.Style()
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	margin := style.Margin.Resolve(constraints.AvailableWidth, constraints.AvailableHeight, fontSize)
	padding := style.Padding.Resolve(constraints.AvailableWidth, constraints.AvailableHeight, fontSize)
	bw := resolveBorderWidths(style.Border, constraints.AvailableWidth, fontSize)

	outerWidth := constraints.AvailableWidth - margin.Horizontal()
	cw := outerWidth - padding.Horizontal() - bw.Horizontal()
	ch := constraints.AvailableHeight - margin.Vertical() - padding.Vertical() - bw.Vertical()
	if cw < 0 {
		cw = 0
	}
	if ch < 0 {
		ch = 0
	}

	return blockContext{
		node:          node,
		constraints:   constraints,
		contentWidth:  cw,
		contentHeight: ch,
		contentX:      margin.Left + bw.Left + padding.Left,
		contentY:      margin.Top + bw.Top + padding.Top,
		margin:        margin,
		padding:       padding,
		borderWidths:  bw,
	}
}

// layoutVerticalChild handles break-policy checks and layout for a single
// child during vertical placement. If done is true the returned Result is
// the final result for the Layout call. Otherwise the Result carries the
// child's layout output for the caller to place.
func (bl *BlockLayout) layoutVerticalChild(bc *blockContext, child document.DocumentNode, children []document.DocumentNode, i int, placed []PlacedNode, cursorY float64) (Result, bool) {
	bp := extractBreakPolicy(child)

	// BreakBefore: force overflow before this child (unless first).
	if i > 0 && (bp.BreakBefore == document.BreakAlways || bp.BreakBefore == document.BreakPage) {
		return bc.overflowResult(placed, cursorY, children[i:]), true
	}

	childConstraints := Constraints{
		AvailableWidth:  bc.contentWidth,
		AvailableHeight: bc.contentHeight - cursorY,
		FontResolver:    bc.constraints.FontResolver,
	}

	if childConstraints.AvailableHeight <= 0 {
		r := Result{
			Bounds:   document.Rectangle{Width: bc.constraints.AvailableWidth, Height: bc.constraints.AvailableHeight},
			Children: placed,
			Overflow: createOverflowNode(bc.node, children[i:]),
		}
		return r, true
	}

	childResult := bl.layoutChild(child, childConstraints)

	// BreakInside=BreakAvoid: if the child overflowed, move the
	// entire child to overflow instead of splitting it.
	if bp.BreakInside == document.BreakAvoid && childResult.Overflow != nil {
		return bc.overflowResult(placed, cursorY, children[i:]), true
	}

	return childResult, false
}

// checkBreakAfter returns an overflow result when the child's BreakAfter
// policy forces a page break and there are more children remaining.
func (bl *BlockLayout) checkBreakAfter(child document.DocumentNode, children []document.DocumentNode, i int, bc *blockContext, placed []PlacedNode, cursorY float64) (Result, bool) {
	bp := extractBreakPolicy(child)
	if (bp.BreakAfter == document.BreakAlways || bp.BreakAfter == document.BreakPage) && i+1 < len(children) {
		return bc.overflowResult(placed, cursorY, children[i+1:]), true
	}
	return Result{}, false
}

// finishVerticalLayout produces the final Result when all children fit.
func (bl *BlockLayout) finishVerticalLayout(bc *blockContext, placed []PlacedNode, cursorY float64) Result {
	style := bc.node.Style()
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}
	contentH := cursorY
	if fixedH := resolveFixedHeight(bc.node, bc.constraints.AvailableWidth, fontSize, bc.margin, bc.padding, bc.borderWidths); fixedH >= 0 && fixedH > contentH {
		contentH = fixedH
	}
	return Result{
		Bounds: document.Rectangle{
			Width:  bc.constraints.AvailableWidth,
			Height: bc.wrapHeight(contentH),
		},
		Children: placed,
	}
}

// layoutAbsolute places a node at its specified absolute coordinates.
// The node is laid out independently from the normal flow, using the
// full content area as available space minus the offset.
func (bl *BlockLayout) layoutAbsolute(box *document.Box, bc *blockContext) PlacedNode {
	const defaultFontSize = 12.0

	pos := box.BoxStyle.Position
	x := pos.X.Resolve(bc.contentWidth, defaultFontSize)
	y := pos.Y.Resolve(bc.contentHeight, defaultFontSize)

	// Determine available space for child layout.
	availW := bc.contentWidth - x
	if box.BoxStyle.Width.Unit != document.UnitAuto && box.BoxStyle.Width.Amount > 0 {
		availW = box.BoxStyle.Width.Resolve(bc.contentWidth, defaultFontSize)
	}
	availH := bc.contentHeight - y
	if box.BoxStyle.Height.Unit != document.UnitAuto && box.BoxStyle.Height.Amount > 0 {
		availH = box.BoxStyle.Height.Resolve(bc.contentHeight, defaultFontSize)
	}
	if availW < 0 {
		availW = 0
	}
	if availH < 0 {
		availH = 0
	}

	// Layout the content inside a plain Box (without the Position to
	// avoid infinite recursion).
	inner := &document.Box{
		Content: box.Content,
		BoxStyle: document.BoxStyle{
			Width:      box.BoxStyle.Width,
			Height:     box.BoxStyle.Height,
			Padding:    box.BoxStyle.Padding,
			Border:     box.BoxStyle.Border,
			Background: box.BoxStyle.Background,
			Direction:  box.BoxStyle.Direction,
		},
	}

	childConstraints := Constraints{
		AvailableWidth:  availW,
		AvailableHeight: availH,
		FontResolver:    bc.constraints.FontResolver,
	}

	result := bl.Layout(inner, childConstraints)

	return PlacedNode{
		Node: box,
		Position: document.Point{
			X: bc.contentX + x,
			Y: bc.contentY + y,
		},
		Size:     document.Size{Width: result.Bounds.Width, Height: result.Bounds.Height},
		Children: result.Children,
	}
}

// extractBreakPolicy returns the BreakPolicy for a node. Box and RichText
// nodes carry a BreakPolicy; all other node types return the zero value (BreakAuto).
func extractBreakPolicy(node document.DocumentNode) document.BreakPolicy {
	if box, ok := node.(*document.Box); ok {
		return box.BreakPolicy
	}
	if rt, ok := node.(*document.RichText); ok {
		return rt.BreakPolicy
	}
	return document.BreakPolicy{}
}

// layoutHorizontal arranges children left-to-right. Each child's width
// is determined by its BoxStyle.Width (typically a percentage of the
// parent). Children without an explicit width share the remaining space
// equally. The row height equals the tallest child.
func (bl *BlockLayout) layoutHorizontal(node document.DocumentNode, constraints Constraints) Result {
	style := node.Style()
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	margin := style.Margin.Resolve(constraints.AvailableWidth, constraints.AvailableHeight, fontSize)
	padding := style.Padding.Resolve(constraints.AvailableWidth, constraints.AvailableHeight, fontSize)
	borderWidths := resolveBorderWidths(style.Border, constraints.AvailableWidth, fontSize)

	outerWidth := constraints.AvailableWidth - margin.Horizontal()
	contentWidth := outerWidth - padding.Horizontal() - borderWidths.Horizontal()
	contentHeight := constraints.AvailableHeight - margin.Vertical() - padding.Vertical() - borderWidths.Vertical()

	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	contentX := margin.Left + borderWidths.Left + padding.Left
	contentY := margin.Top + borderWidths.Top + padding.Top

	fixedContentHeight := resolveFixedHeight(node, constraints.AvailableWidth, fontSize, margin, padding, borderWidths)

	children := node.Children()
	childWidths := resolveChildWidths(children, contentWidth, fontSize)

	childAvailHeight := contentHeight
	if fixedContentHeight >= 0 && fixedContentHeight < childAvailHeight {
		childAvailHeight = fixedContentHeight
	}

	var placed []PlacedNode
	cursorX := 0.0
	maxHeight := 0.0

	for i, child := range children {
		if child == nil {
			continue
		}

		childConstraints := Constraints{
			AvailableWidth:  childWidths[i],
			AvailableHeight: childAvailHeight,
			FontResolver:    constraints.FontResolver,
		}

		childResult := bl.layoutChild(child, childConstraints)

		placed = append(placed, PlacedNode{
			Node: child,
			Position: document.Point{
				X: contentX + cursorX,
				Y: contentY,
			},
			Size:     document.Size{Width: childWidths[i], Height: childResult.Bounds.Height},
			Children: childResult.Children,
		})

		cursorX += childWidths[i]
		if childResult.Bounds.Height > maxHeight {
			maxHeight = childResult.Bounds.Height
		}
	}

	// Use the fixed height if specified, otherwise use content-derived height.
	finalContentHeight := maxHeight
	if fixedContentHeight >= 0 && fixedContentHeight > finalContentHeight {
		finalContentHeight = fixedContentHeight
	}

	// Stretch all children (columns) and their direct children (e.g. text
	// nodes) to match the final row height so that backgrounds fill the cell.
	stretchPlacedNodes(placed, finalContentHeight)

	totalHeight := margin.Top + borderWidths.Top + padding.Top + finalContentHeight + padding.Bottom + borderWidths.Bottom + margin.Bottom
	return Result{
		Bounds: document.Rectangle{
			X:      0,
			Y:      0,
			Width:  constraints.AvailableWidth,
			Height: totalHeight,
		},
		Children: placed,
	}
}

// resolveChildWidths determines the width of each child node for
// horizontal layout. Children with an explicit BoxStyle.Width are
// resolved against parentWidth. The remaining space is divided equally
// among children without an explicit width.
func resolveChildWidths(children []document.DocumentNode, parentWidth, fontSize float64) []float64 {
	widths := make([]float64, len(children))
	usedWidth := 0.0
	autoCount := 0

	for i, child := range children {
		if child == nil {
			continue
		}
		if box, ok := child.(*document.Box); ok && box.BoxStyle.Width.Unit != document.UnitAuto && box.BoxStyle.Width.Amount > 0 {
			widths[i] = box.BoxStyle.Width.Resolve(parentWidth, fontSize)
			usedWidth += widths[i]
		} else {
			autoCount++
		}
	}

	if autoCount > 0 {
		remaining := parentWidth - usedWidth
		if remaining < 0 {
			remaining = 0
		}
		autoWidth := remaining / float64(autoCount)
		for i := range widths {
			if widths[i] == 0 && children[i] != nil {
				widths[i] = autoWidth
			}
		}
	}

	return widths
}

// layoutChild dispatches layout for a single child node.
func (bl *BlockLayout) layoutChild(child document.DocumentNode, constraints Constraints) Result {
	switch child.NodeType() {
	case document.NodeText:
		fl := &FlowLayout{}
		textNode, ok := child.(*document.Text)
		if !ok {
			return Result{}
		}
		return fl.LayoutText(textNode.Content, child.Style(), constraints)
	case document.NodeTable:
		tbl, ok := child.(*document.Table)
		if !ok {
			return bl.Layout(child, constraints)
		}
		return bl.layoutTable(tbl, constraints)
	case document.NodeList:
		lst, ok := child.(*document.List)
		if !ok {
			return bl.Layout(child, constraints)
		}
		return bl.layoutList(lst, constraints)
	case document.NodeImage:
		return bl.layoutImage(child, constraints)
	case document.NodeRichText:
		fl := &FlowLayout{}
		rtNode, ok := child.(*document.RichText)
		if !ok {
			return Result{}
		}
		return fl.LayoutRichText(rtNode, constraints)
	default:
		// Recurse using block layout for container nodes.
		return bl.Layout(child, constraints)
	}
}

// layoutImage computes the display dimensions for an image node,
// respecting explicit size constraints, fit mode, and aspect ratio.
func (bl *BlockLayout) layoutImage(child document.DocumentNode, constraints Constraints) Result {
	img, ok := child.(*document.Image)
	if !ok {
		return Result{}
	}

	// Intrinsic dimensions in points (1 pixel = 1 point at 72 DPI).
	intrinsicW := float64(img.Source.Width)
	intrinsicH := float64(img.Source.Height)
	if intrinsicW <= 0 || intrinsicH <= 0 {
		return Result{}
	}

	aspectRatio := intrinsicW / intrinsicH

	var displayW, displayH float64

	switch img.FitMode {
	case document.FitOriginal:
		displayW = intrinsicW
		displayH = intrinsicH
	case document.FitStretch:
		if w, h, ok := resolveExplicitDimensions(img, constraints, aspectRatio); ok {
			displayW, displayH = w, h
		} else {
			displayW = constraints.AvailableWidth
			displayH = displayW / aspectRatio
		}
	case document.FitCover:
		displayW, displayH = computeCoverSize(img, constraints, intrinsicW, intrinsicH)
	default: // FitContain
		if w, h, ok := resolveExplicitDimensions(img, constraints, aspectRatio); ok {
			displayW, displayH = w, h
		} else {
			displayW = intrinsicW
			displayH = intrinsicH
			if displayW > constraints.AvailableWidth {
				displayW = constraints.AvailableWidth
				displayH = displayW / aspectRatio
			}
		}
	}

	displayW, displayH = clampImageSize(img.FitMode, displayW, displayH, aspectRatio, constraints)

	return Result{
		Bounds: document.Rectangle{
			Width:  displayW,
			Height: displayH,
		},
	}
}

// resolveExplicitDimensions resolves display dimensions from explicit
// width/height settings. Returns (0, 0, false) when no explicit dimensions are set.
func resolveExplicitDimensions(img *document.Image, constraints Constraints, aspectRatio float64) (float64, float64, bool) {
	const defaultFontSize = 12.0
	hasW := img.DisplayWidth.Unit != document.UnitAuto && img.DisplayWidth.Amount > 0
	hasH := img.DisplayHeight.Unit != document.UnitAuto && img.DisplayHeight.Amount > 0

	switch {
	case hasW && hasH:
		return img.DisplayWidth.Resolve(constraints.AvailableWidth, defaultFontSize),
			img.DisplayHeight.Resolve(constraints.AvailableHeight, defaultFontSize), true
	case hasW:
		w := img.DisplayWidth.Resolve(constraints.AvailableWidth, defaultFontSize)
		return w, w / aspectRatio, true
	case hasH:
		h := img.DisplayHeight.Resolve(constraints.AvailableHeight, defaultFontSize)
		return h * aspectRatio, h, true
	default:
		return 0, 0, false
	}
}

// computeCoverSize calculates dimensions for FitCover mode, scaling to
// cover the bounds completely.
func computeCoverSize(img *document.Image, constraints Constraints, intrinsicW, intrinsicH float64) (float64, float64) {
	const defaultFontSize = 12.0
	boundsW := constraints.AvailableWidth
	boundsH := constraints.AvailableHeight
	if img.DisplayWidth.Unit != document.UnitAuto && img.DisplayWidth.Amount > 0 {
		boundsW = img.DisplayWidth.Resolve(constraints.AvailableWidth, defaultFontSize)
	}
	if img.DisplayHeight.Unit != document.UnitAuto && img.DisplayHeight.Amount > 0 {
		boundsH = img.DisplayHeight.Resolve(constraints.AvailableHeight, defaultFontSize)
	}
	scaleW := boundsW / intrinsicW
	scaleH := boundsH / intrinsicH
	scale := scaleW
	if scaleH > scaleW {
		scale = scaleH
	}
	return intrinsicW * scale, intrinsicH * scale
}

// clampImageSize constrains display dimensions to available space.
func clampImageSize(fitMode document.ImageFitMode, w, h, aspectRatio float64, constraints Constraints) (float64, float64) {
	if fitMode != document.FitStretch {
		if w > constraints.AvailableWidth {
			w = constraints.AvailableWidth
			h = w / aspectRatio
		}
		if h > constraints.AvailableHeight {
			h = constraints.AvailableHeight
			w = h * aspectRatio
		}
	} else {
		if w > constraints.AvailableWidth {
			w = constraints.AvailableWidth
		}
		if h > constraints.AvailableHeight {
			h = constraints.AvailableHeight
		}
	}
	return w, h
}

// resolveBorderWidths extracts the four border widths as resolved edges.
func resolveBorderWidths(border document.BorderEdges, parentWidth, fontSize float64) document.ResolvedEdges {
	var top, right, bottom, left float64
	if border.Top.Style != document.BorderNone {
		top = border.Top.Width.Resolve(parentWidth, fontSize)
	}
	if border.Right.Style != document.BorderNone {
		right = border.Right.Width.Resolve(parentWidth, fontSize)
	}
	if border.Bottom.Style != document.BorderNone {
		bottom = border.Bottom.Width.Resolve(parentWidth, fontSize)
	}
	if border.Left.Style != document.BorderNone {
		left = border.Left.Width.Resolve(parentWidth, fontSize)
	}
	return document.ResolvedEdges{Top: top, Right: right, Bottom: bottom, Left: left}
}

// resolveFixedHeight returns the fixed content height for a Box node with
// an explicit Height value. Returns -1 when no fixed height is set.
func resolveFixedHeight(node document.DocumentNode, availWidth, fontSize float64, margin, padding, borderWidths document.ResolvedEdges) float64 {
	box, ok := node.(*document.Box)
	if !ok || box.BoxStyle.Height.Unit == document.UnitAuto || box.BoxStyle.Height.Amount <= 0 {
		return -1
	}
	resolved := box.BoxStyle.Height.Resolve(availWidth, fontSize)
	if resolved <= 0 {
		return -1
	}
	h := resolved - padding.Vertical() - borderWidths.Vertical() - margin.Vertical()
	if h < 0 {
		h = 0
	}
	return h
}

// stretchPlacedNodes sets the height of each placed column node to
// targetHeight, ensuring column backgrounds fill the full row cell.
// It also stretches the last child within each column so that
// background colors extend to the bottom of the row.
func stretchPlacedNodes(nodes []PlacedNode, targetHeight float64) {
	for i := range nodes {
		nodes[i].Size.Height = targetHeight
		stretchLastChild(nodes[i].Children, targetHeight)
	}
}

// stretchLastChild recursively stretches the last child in a chain so
// that its height fills the remaining space. This ensures backgrounds
// on leaf content nodes (e.g. Text) extend to the bottom of their
// containing row cell. Image nodes are never stretched so that their
// aspect ratio is preserved.
func stretchLastChild(children []PlacedNode, parentHeight float64) {
	if len(children) == 0 {
		return
	}
	last := &children[len(children)-1]
	// Do not stretch image nodes — they have intrinsic aspect ratios.
	if last.Node != nil && last.Node.NodeType() == document.NodeImage {
		return
	}
	remaining := parentHeight - last.Position.Y
	if remaining > last.Size.Height {
		last.Size.Height = remaining
	}
	stretchLastChild(last.Children, last.Size.Height)
}

// createOverflowNode wraps remaining children in a lightweight container
// that preserves the parent's style for continued layout on the next page.
func createOverflowNode(parent document.DocumentNode, remaining []document.DocumentNode) document.DocumentNode {
	if len(remaining) == 0 {
		return nil
	}
	return &document.Box{
		Content: remaining,
		BoxStyle: document.BoxStyle{
			Margin:  parent.Style().Margin,
			Padding: parent.Style().Padding,
			Border:  parent.Style().Border,
		},
	}
}
