package layout

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
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
	// Background is painted beneath Children, or nil for none. It is
	// copied from the source [document.Page], including onto pages
	// created when that page's content overflows.
	Background *document.PageBackground
}

// ErrContentDoesNotFit is returned by [Paginator.PaginateWithError] when
// content cannot be placed even on an empty page, for example because the
// page size is zero or the margins, header and footer leave less room than
// a single line of text needs. Without this check pagination would emit
// empty pages forever. Use [errors.Is] to test for it; the wrapped error
// message carries the page and content-area dimensions.
var ErrContentDoesNotFit = errors.New("layout: content does not fit on an empty page")

// Paginate processes the document's pages and their content, producing a
// list of PageLayout values. Each input page is laid out within its content
// area (page size minus margins). Content that overflows is automatically
// continued on a new page with the same size and margins.
//
// If some content can never be placed (see [ErrContentDoesNotFit]),
// Paginate stops and returns the pages produced so far. Use
// [Paginator.PaginateWithError] to detect that case.
func (p *Paginator) Paginate(doc *document.Document) []PageLayout {
	pages, _ := p.PaginateWithError(doc)
	return pages
}

// PaginateWithError is like [Paginator.Paginate] but reports content that
// cannot be placed even on an empty page. In that case it returns the
// pages produced so far together with an error wrapping
// [ErrContentDoesNotFit].
func (p *Paginator) PaginateWithError(doc *document.Document) ([]PageLayout, error) {
	if doc == nil {
		return nil, nil
	}

	var pages []PageLayout
	block := NewBlockLayout()

	for _, page := range doc.Pages {
		var err error
		pages, err = p.paginatePage(page, block, pages)
		if err != nil {
			return pages, err
		}
	}

	return pages, nil
}

// paginatePage lays out a single source page, appending one PageLayout per
// physical page (overflow included) to pages.
func (p *Paginator) paginatePage(page *document.Page, block *BlockLayout, pages []PageLayout) ([]PageLayout, error) {
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

	constraints := Constraints{
		AvailableWidth:  contentWidth,
		AvailableHeight: bodyHeight,
		PageWidth:       pageSize.Width,
		PageHeight:      pageSize.Height,
		FontResolver:    p.fontResolver,
	}

	// Create a virtual container node for the page content.
	var remaining document.DocumentNode = &document.Box{
		Content: page.Content,
	}

	emptyPages := 0
	for remaining != nil {
		result := block.Layout(remaining, constraints)

		if result.Overflow != nil {
			// Layout is deterministic, so an overflow identical to the
			// input means the next page would come out exactly the same:
			// nothing on it can ever be placed. As a backstop against
			// overflow that changes without shrinking, also stop after a
			// long run of pages that place nothing visible.
			if hasVisibleContent(result.Children) {
				emptyPages = 0
			} else {
				emptyPages++
			}
			if emptyPages > maxConsecutiveEmptyPages || reflect.DeepEqual(result.Overflow, remaining) {
				return pages, fmt.Errorf("%w: body area is %.2f x %.2f pt on a %.2f x %.2f pt page; "+
					"enlarge the page or reduce margins, header, footer or element size",
					ErrContentDoesNotFit,
					contentWidth, bodyHeight, pageSize.Width, pageSize.Height)
			}
		}

		pl := p.composePageLayout(
			pageSize, resolvedMargins, contentHeight,
			headerPlaced, headerHeight,
			footerPlaced, footerHeight,
			result.Children,
		)
		pl.Background = page.Background
		pages = append(pages, pl)

		remaining = result.Overflow
	}

	return pages, nil
}

// maxConsecutiveEmptyPages bounds how many pages in a row may overflow
// without placing any visible content before pagination gives up. Page
// breaks around zero-height content can legitimately produce a few blank
// pages; a long run means the layout is not making progress.
const maxConsecutiveEmptyPages = 64

// hasVisibleContent reports whether any leaf among the placed nodes
// occupies vertical space, i.e. whether the page shows something.
func hasVisibleContent(nodes []PlacedNode) bool {
	for i := range nodes {
		if len(nodes[i].Children) == 0 {
			if nodes[i].Size.Height > 0 {
				return true
			}
			continue
		}
		if hasVisibleContent(nodes[i].Children) {
			return true
		}
	}
	return false
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
//
// Lines keep the position computed during layout. Use
// [ResolvePageNumbersWithFonts] to also re-measure the resolved lines so
// that centered and right-aligned page numbers stay aligned.
func ResolvePageNumbers(pages []PageLayout) {
	ResolvePageNumbersWithFonts(pages, nil)
}

// ResolvePageNumbersWithFonts is like [ResolvePageNumbers] but, when fr
// is non-nil, re-measures every resolved text line and shifts centered
// and right-aligned lines so they stay aligned with the actual digits.
// During layout each placeholder is measured as a single digit; this
// corrects the difference once the real number is known.
func ResolvePageNumbersWithFonts(pages []PageLayout, fr FontResolver) {
	total := len(pages)
	for i := range pages {
		resolvePageNumbersInNodes(pages[i].Children, i+1, total, fr)
	}
}

// resolvePageNumbersInNodes recursively replaces placeholder strings in
// text nodes.
//
// Header/footer nodes are laid out once and the same *document.Text is
// shared by every physical page, so placeholders are resolved on a
// per-page copy (copy-on-write) instead of mutating the shared node.
// This also leaves the caller's Document untouched so it can be
// rendered again.
func resolvePageNumbersInNodes(nodes []PlacedNode, pageNum, totalPages int, fr FontResolver) {
	for i := range nodes {
		if textNode, ok := nodes[i].Node.(*document.Text); ok {
			if strings.Contains(textNode.Content, document.PageNumberPlaceholder) ||
				strings.Contains(textNode.Content, document.TotalPagesPlaceholder) {
				clone := *textNode
				clone.Content = strings.ReplaceAll(clone.Content, document.PageNumberPlaceholder, strconv.Itoa(pageNum))
				clone.Content = strings.ReplaceAll(clone.Content, document.TotalPagesPlaceholder, strconv.Itoa(totalPages))
				nodes[i].Node = &clone
				if fr != nil && len(nodes[i].Children) == 0 {
					realignLine(&nodes[i], &clone, fr)
				}
			}
		}
		if len(nodes[i].Children) > 0 {
			resolvePageNumbersInNodes(nodes[i].Children, pageNum, totalPages, fr)
		}
	}
}

// realignLine re-measures a resolved text line and keeps it anchored to
// its alignment: the center for centered text, the right edge for
// right-aligned text, and the left edge otherwise. Justified lines are
// left as laid out.
func realignLine(pn *PlacedNode, t *document.Text, fr FontResolver) {
	style := t.TextStyle
	if style.TextAlign == document.AlignJustify {
		return
	}
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}
	newWidth := measureLineWidth(t.Content, style, fontSize, Constraints{FontResolver: fr})
	switch style.TextAlign {
	case document.AlignCenter:
		pn.Position.X += (pn.Size.Width - newWidth) / 2
	case document.AlignRight:
		pn.Position.X += pn.Size.Width - newWidth
	}
	pn.Size.Width = newWidth
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
