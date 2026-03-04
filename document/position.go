package document

// PositionMode specifies how a node is positioned within its container.
type PositionMode int

const (
	// PositionStatic is the default flow-based positioning.
	PositionStatic PositionMode = iota
	// PositionAbsolute places the node at fixed coordinates,
	// removed from the normal document flow.
	PositionAbsolute
)

// PositionOrigin specifies the reference point for absolute coordinates.
type PositionOrigin int

const (
	// OriginContentArea measures from the top-left of the content area
	// (inside page margins). This is the default.
	OriginContentArea PositionOrigin = iota
	// OriginPage measures from the top-left of the physical page
	// (0, 0 = top-left corner, ignoring margins).
	OriginPage
)

// Position defines the placement of a node when using absolute positioning.
type Position struct {
	// Mode selects the positioning strategy.
	Mode PositionMode
	// X is the horizontal offset from the origin.
	X Value
	// Y is the vertical offset from the origin.
	Y Value
	// Origin selects the reference point for X/Y coordinates.
	Origin PositionOrigin
}
