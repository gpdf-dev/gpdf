package document

import "github.com/gpdf-dev/gpdf/pdf"

const (
	// PageNumberPlaceholder is a sentinel string embedded in text nodes to
	// mark where the current page number should appear. It is replaced after
	// pagination with the actual page number.
	PageNumberPlaceholder = "\x00GPDF_PAGE\x00"
	// TotalPagesPlaceholder is a sentinel string replaced after pagination
	// with the total number of pages in the document.
	TotalPagesPlaceholder = "\x00GPDF_TOTAL\x00"
)

// Page represents a single page within a document. It defines the physical
// page size, margins, and content nodes laid out on that page.
type Page struct {
	// Size specifies the page dimensions (e.g., A4, Letter).
	Size Size
	// Margins defines the space between the page edge and the content area.
	Margins Edges
	// Content holds the document nodes rendered on this page.
	Content []DocumentNode
	// PageStyle provides default style properties for content on this page.
	PageStyle Style
	// Background is painted beneath all other content on this page and on
	// any continuation pages created when its content overflows. Nil means
	// no background.
	Background *PageBackground
}

// PageBackground describes a full-page background drawn before any other
// page content: an optional solid color, then an optional image on top.
type PageBackground struct {
	// Color fills the whole page when non-nil.
	Color *pdf.Color
	// Image is drawn over Color when non-nil. It is centered on the page
	// and scaled according to FitMode.
	Image *ImageSource
	// FitMode controls how Image is scaled relative to the page:
	// FitContain (the zero value) fits the whole image inside the page,
	// FitCover fills the page and crops the overflow, FitStretch fills the
	// page exactly, and FitOriginal uses the intrinsic size (1px = 1pt).
	FitMode ImageFitMode
	// Opacity is the constant alpha applied to the background (color and
	// image). The zero value means fully opaque.
	Opacity float64
}

// NodeType returns NodePage.
func (p *Page) NodeType() NodeType { return NodePage }

// Children returns the page's content nodes.
func (p *Page) Children() []DocumentNode { return p.Content }

// Style returns the page's default style.
func (p *Page) Style() Style { return p.PageStyle }

// Document is the root node of the document tree. It holds an ordered list
// of pages, document-level metadata, and a default style inherited by all
// descendant nodes that do not override specific properties.
type Document struct {
	// Pages holds the pages in document order.
	Pages []*Page
	// Metadata holds document-level information such as title and author.
	Metadata DocumentMetadata
	// DefaultStyle provides default values for style properties that are
	// not explicitly set on descendant nodes.
	DefaultStyle Style
}

// DocumentMetadata carries descriptive information about the document,
// corresponding to the PDF Info dictionary fields.
type DocumentMetadata struct {
	Title    string
	Author   string
	Subject  string
	Creator  string
	Producer string
}

// NodeType returns NodeDocument.
func (d *Document) NodeType() NodeType { return NodeDocument }

// Children returns each page as a DocumentNode.
func (d *Document) Children() []DocumentNode {
	nodes := make([]DocumentNode, len(d.Pages))
	for i, p := range d.Pages {
		nodes[i] = p
	}
	return nodes
}

// Style returns the document's default style.
func (d *Document) Style() Style { return d.DefaultStyle }
