package htmlpdf

import (
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
)

// CSSLayoutEngine implements layout.Engine for HTML+CSS content.
// Phase 4-A: delegates entirely to the existing BlockLayout engine.
type CSSLayoutEngine struct {
	blockLayout *layout.BlockLayout
}

// NewCSSLayoutEngine creates a new CSSLayoutEngine.
func NewCSSLayoutEngine() *CSSLayoutEngine {
	return &CSSLayoutEngine{
		blockLayout: &layout.BlockLayout{},
	}
}

// Layout performs layout calculation for the given node tree.
func (e *CSSLayoutEngine) Layout(node document.DocumentNode, constraints layout.Constraints) layout.Result {
	return e.blockLayout.Layout(node, constraints)
}
