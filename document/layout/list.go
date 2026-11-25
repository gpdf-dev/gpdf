package layout

import (
	"fmt"

	"github.com/gpdf-dev/gpdf/document"
)

// defaultMarkerIndent is the default width in points reserved for list
// markers (bullet or number) when MarkerIndent is not explicitly set.
const defaultMarkerIndent = 20

// layoutList lays out a List node by placing a marker and indented content
// for each item sequentially.
func (bl *BlockLayout) layoutList(lst *document.List, constraints Constraints) Result {
	indent := lst.MarkerIndent
	if indent <= 0 {
		indent = defaultMarkerIndent
	}

	var placed []PlacedNode
	cursorY := 0.0

	for i := range lst.Items {
		item := &lst.Items[i]

		markerText := listMarkerText(lst.ListType, i)
		markerStyle := lst.ListStyle
		if item.ItemStyle.FontSize > 0 {
			markerStyle = item.ItemStyle
		}

		// Layout the marker text in the indent area.
		fl := &FlowLayout{}
		markerConstraints := Constraints{
			AvailableWidth:  indent,
			AvailableHeight: constraints.AvailableHeight - cursorY,
			FontResolver:    constraints.FontResolver,
		}
		markerResult := fl.LayoutText(markerText, markerStyle, markerConstraints)

		// Layout item content to the right of the indent.
		contentBox := &document.Box{Content: item.Content}
		contentConstraints := Constraints{
			AvailableWidth:  constraints.AvailableWidth - indent,
			AvailableHeight: constraints.AvailableHeight - cursorY,
			FontResolver:    constraints.FontResolver,
			belowContent:    constraints.belowContent || cursorY > 0,
		}
		contentResult := bl.Layout(contentBox, contentConstraints)

		// Row height is the taller of marker and content.
		rowHeight := markerResult.Bounds.Height
		if contentResult.Bounds.Height > rowHeight {
			rowHeight = contentResult.Bounds.Height
		}

		// Place marker.
		placed = append(placed, PlacedNode{
			Node: &document.Text{
				Content:   markerText,
				TextStyle: markerStyle,
			},
			Position: document.Point{X: 0, Y: cursorY},
			Size:     document.Size{Width: indent, Height: rowHeight},
			Children: markerResult.Children,
		})

		// Place content.
		placed = append(placed, PlacedNode{
			Node:     contentBox,
			Position: document.Point{X: indent, Y: cursorY},
			Size:     document.Size{Width: constraints.AvailableWidth - indent, Height: rowHeight},
			Children: contentResult.Children,
		})

		cursorY += rowHeight

		// Handle overflow: remaining items go to next page.
		if contentResult.Overflow != nil {
			remaining := make([]document.ListItem, len(lst.Items)-i-1)
			copy(remaining, lst.Items[i+1:])
			// The overflowed content becomes the first item on the next page.
			overflowItem := document.ListItem{
				Content:   []document.DocumentNode{contentResult.Overflow},
				ItemStyle: item.ItemStyle,
			}
			overflowItems := append([]document.ListItem{overflowItem}, remaining...)
			overflow := &document.List{
				Items:        overflowItems,
				ListType:     lst.ListType,
				ListStyle:    lst.ListStyle,
				BreakPolicy:  lst.BreakPolicy,
				MarkerIndent: lst.MarkerIndent,
			}
			return Result{
				Bounds: document.Rectangle{
					Width:  constraints.AvailableWidth,
					Height: cursorY,
				},
				Children: placed,
				Overflow: overflow,
			}
		}
	}

	return Result{
		Bounds: document.Rectangle{
			Width:  constraints.AvailableWidth,
			Height: cursorY,
		},
		Children: placed,
	}
}

// listMarkerText returns the marker string for the given list type and
// zero-based item index.
func listMarkerText(lt document.ListType, index int) string {
	switch lt {
	case document.Ordered:
		return fmt.Sprintf("%d.", index+1)
	default:
		return "\u2022" // bullet •
	}
}
