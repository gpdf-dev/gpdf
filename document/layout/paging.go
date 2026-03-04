package layout

import (
	"fmt"
	"strings"

	"github.com/gpdf-dev/gpdf/document"
)

// Paginator splits a document into individually laid-out pages. It uses a
// BlockLayout engine to place content within each page's content area and
// carries overflow content to subsequent pages automatically.
type Paginator struct {
	pageSize     document.Size
	margins      document.Edges
	fontResolver FontResolver
	headerNodes  []document.DocumentNode
	footerNodes  []document.DocumentNode
}

// NewPaginator creates a Paginator with the specified page size, margins,
// and font resolver.
func NewPaginator(pageSize document.Size, margins document.Edges, fontResolver FontResolver) *Paginator {
	return &Paginator{
		pageSize:     pageSize,
		margins:      margins,
		fontResolver: fontResolver,
	}
}

// SetHeaderFooter registers document nodes to be placed as header and
// footer on every page. These are laid out once to measure their height,
// which is subtracted from the available body height on each page.
func (p *Paginator) SetHeaderFooter(header, footer []document.DocumentNode) {
	p.headerNodes = header
	p.footerNodes = footer
}

// PageLayout contains the final placed nodes for a single page along with
// the page dimensions.
type PageLayout struct {
	// Size is the page's physical dimensions.
	Size document.Size
	// Children holds the positioned nodes on this page.
	Children []PlacedNode
}

// Paginate processes the document's pages and their content, producing a
// list of PageLayout values. Each input page is laid out within its content
// area (page size minus margins). Content that overflows is automatically
// continued on a new page with the same size and margins.
func (p *Paginator) Paginate(doc *document.Document) []PageLayout {
	if doc == nil {
		return nil
	}

	var pages []PageLayout
	block := NewBlockLayout()

	for _, page := range doc.Pages {
		pageSize := page.Size
		if pageSize.Width <= 0 || pageSize.Height <= 0 {
			pageSize = p.pageSize
		}

		margins := page.Margins
		if margins == (document.Edges{}) {
			margins = p.margins
		}

		resolvedMargins := margins.Resolve(pageSize.Width, pageSize.Height, 12)

		contentWidth := pageSize.Width - resolvedMargins.Horizontal()
		contentHeight := pageSize.Height - resolvedMargins.Vertical()

		// Measure header and footer to determine body available height.
		headerHeight, headerPlaced := p.measureSection(p.headerNodes, contentWidth, block)
		footerHeight, footerPlaced := p.measureSection(p.footerNodes, contentWidth, block)

		bodyHeight := contentHeight - headerHeight - footerHeight
		if bodyHeight < 0 {
			bodyHeight = 0
		}

		// Create a virtual container node for the page content.
		container := &document.Box{
			Content: page.Content,
		}

		var remaining document.DocumentNode = container

		for remaining != nil {
			constraints := Constraints{
				AvailableWidth:  contentWidth,
				AvailableHeight: bodyHeight,
				FontResolver:    p.fontResolver,
			}

			result := block.Layout(remaining, constraints)

			pl := p.composePageLayout(
				pageSize, resolvedMargins, contentHeight,
				headerPlaced, headerHeight,
				footerPlaced, footerHeight,
				result.Children,
			)
			pages = append(pages, pl)

			remaining = result.Overflow
		}
	}

	return pages
}

// measureSection lays out a set of document nodes and returns the total
// height and the placed nodes. The placed nodes are not offset by margins;
// the caller is responsible for positioning.
func (p *Paginator) measureSection(nodes []document.DocumentNode, contentWidth float64, block *BlockLayout) (float64, []PlacedNode) {
	if len(nodes) == 0 {
		return 0, nil
	}
	container := &document.Box{Content: nodes}
	constraints := Constraints{
		AvailableWidth:  contentWidth,
		AvailableHeight: 1e9, // effectively unlimited
		FontResolver:    p.fontResolver,
	}
	result := block.Layout(container, constraints)
	return result.Bounds.Height, result.Children
}

// composePageLayout assembles a PageLayout by positioning header at the top
// of the content area, body below the header, and footer at the bottom.
func (p *Paginator) composePageLayout(
	pageSize document.Size,
	margins document.ResolvedEdges,
	contentHeight float64,
	headerPlaced []PlacedNode, headerHeight float64,
	footerPlaced []PlacedNode, footerHeight float64,
	bodyPlaced []PlacedNode,
) PageLayout {
	var children []PlacedNode

	mx := margins.Left
	my := margins.Top

	// Header at top of content area.
	if len(headerPlaced) > 0 {
		children = append(children, offsetNodes(headerPlaced, mx, my)...)
	}

	// Body below header.
	bodyOffsetY := my + headerHeight
	if len(bodyPlaced) > 0 {
		children = append(children, offsetNodes(bodyPlaced, mx, bodyOffsetY)...)
	}

	// Footer at bottom of content area.
	if len(footerPlaced) > 0 {
		footerY := my + contentHeight - footerHeight
		children = append(children, offsetNodes(footerPlaced, mx, footerY)...)
	}

	// Adjust absolute-positioned nodes with OriginPage so their
	// coordinates are relative to the page corner, not the content area.
	adjustAbsoluteOrigins(children, mx, my)

	return PageLayout{
		Size:     pageSize,
		Children: children,
	}
}

// adjustAbsoluteOrigins walks the top-level placed nodes and adjusts
// coordinates for absolute-positioned nodes that use OriginPage. These
// nodes were laid out relative to the content area, so we subtract the
// margin offset to make them relative to the page corner.
func adjustAbsoluteOrigins(nodes []PlacedNode, marginX, marginY float64) {
	for i := range nodes {
		box, ok := nodes[i].Node.(*document.Box)
		if !ok {
			continue
		}
		if box.BoxStyle.Position.Mode != document.PositionAbsolute {
			continue
		}
		if box.BoxStyle.Position.Origin == document.OriginPage {
			nodes[i].Position.X -= marginX
			nodes[i].Position.Y -= marginY
		}
	}
}

// ResolvePageNumbers walks all pages and replaces page-number placeholder
// strings in text nodes with the actual page number and total page count.
func ResolvePageNumbers(pages []PageLayout) {
	total := len(pages)
	for i := range pages {
		resolvePageNumbersInNodes(pages[i].Children, i+1, total)
	}
}

// resolvePageNumbersInNodes recursively replaces placeholder strings in
// text nodes.
func resolvePageNumbersInNodes(nodes []PlacedNode, pageNum, totalPages int) {
	for i := range nodes {
		if textNode, ok := nodes[i].Node.(*document.Text); ok {
			if strings.Contains(textNode.Content, document.PageNumberPlaceholder) ||
				strings.Contains(textNode.Content, document.TotalPagesPlaceholder) {
				textNode.Content = strings.ReplaceAll(textNode.Content, document.PageNumberPlaceholder, fmt.Sprintf("%d", pageNum))
				textNode.Content = strings.ReplaceAll(textNode.Content, document.TotalPagesPlaceholder, fmt.Sprintf("%d", totalPages))
			}
		}
		if len(nodes[i].Children) > 0 {
			resolvePageNumbersInNodes(nodes[i].Children, pageNum, totalPages)
		}
	}
}

// offsetNodes shifts all placed node positions by the given dx and dy.
func offsetNodes(nodes []PlacedNode, dx, dy float64) []PlacedNode {
	if len(nodes) == 0 {
		return nodes
	}
	result := make([]PlacedNode, len(nodes))
	for i, n := range nodes {
		result[i] = PlacedNode{
			Node: n.Node,
			Position: document.Point{
				X: n.Position.X + dx,
				Y: n.Position.Y + dy,
			},
			Size:     n.Size,
			Children: offsetNodes(n.Children, 0, 0), // children are relative to parent
		}
	}
	return result
}
