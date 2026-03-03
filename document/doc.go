// Package document provides the intermediate document model that bridges the
// high-level template API and the low-level PDF writer. It is Layer 2 of the
// gpdf architecture.
//
// # Document Tree
//
// A document is represented as a tree of nodes implementing the
// [DocumentNode] interface. Each node carries a [NodeType], optional children,
// and a [Style]:
//
//   - [Document] — root node containing pages and metadata
//   - [Page] — a single page with size, margins, and content
//   - [Box] — a generic container with CSS-like box model (horizontal/vertical)
//   - [Text] — a leaf node containing styled text content
//   - [Image] — a leaf node containing an embedded image (JPEG/PNG)
//   - [Table] — tabular data with header, body, and footer sections
//   - [List] — ordered or unordered list items
//   - [RichText] — inline formatting context with multiple styled fragments
//
// # Styling
//
// The [Style] type holds the complete set of visual properties following CSS
// box model conventions: font family/size/weight, color, text alignment,
// margin, padding, and border. Styles are inherited from parent to child via
// [InheritStyle].
//
// # Units
//
// Dimensions are expressed as [Value] with an associated [Unit]. The library
// supports pt, mm, cm, in, em, and percentage units. Constructor functions
// provide a convenient way to create values:
//
//	v := document.Mm(15)       // 15 millimeters
//	v := document.Pt(12)       // 12 points
//	v := document.Pct(50)      // 50 percent
//
// Standard page sizes are provided as [Size] values: [A4], [A3], [Letter],
// [Legal].
//
// # Page Breaks
//
// [BreakPolicy] controls page-break behavior (before, after, or inside a
// node). The layout engine in the layout sub-package uses these hints when
// paginating content.
package document
