package css

// PropertyDef defines a CSS property's metadata.
type PropertyDef struct {
	Initial   string // initial (default) value
	Inherited bool   // whether the property inherits
}

// PropertyTable maps CSS property names to their definitions.
// Phase 4-A properties only.
var PropertyTable = map[string]PropertyDef{
	// Display
	"display": {Initial: "inline", Inherited: false},

	// Box model
	"margin-top":    {Initial: "0", Inherited: false},
	"margin-right":  {Initial: "0", Inherited: false},
	"margin-bottom": {Initial: "0", Inherited: false},
	"margin-left":   {Initial: "0", Inherited: false},

	"padding-top":    {Initial: "0", Inherited: false},
	"padding-right":  {Initial: "0", Inherited: false},
	"padding-bottom": {Initial: "0", Inherited: false},
	"padding-left":   {Initial: "0", Inherited: false},

	"border-top-width":    {Initial: "medium", Inherited: false},
	"border-right-width":  {Initial: "medium", Inherited: false},
	"border-bottom-width": {Initial: "medium", Inherited: false},
	"border-left-width":   {Initial: "medium", Inherited: false},

	"border-top-style":    {Initial: "none", Inherited: false},
	"border-right-style":  {Initial: "none", Inherited: false},
	"border-bottom-style": {Initial: "none", Inherited: false},
	"border-left-style":   {Initial: "none", Inherited: false},

	"border-top-color":    {Initial: "currentcolor", Inherited: false},
	"border-right-color":  {Initial: "currentcolor", Inherited: false},
	"border-bottom-color": {Initial: "currentcolor", Inherited: false},
	"border-left-color":   {Initial: "currentcolor", Inherited: false},

	// Sizing
	"width":      {Initial: "auto", Inherited: false},
	"height":     {Initial: "auto", Inherited: false},
	"min-width":  {Initial: "0", Inherited: false},
	"max-width":  {Initial: "none", Inherited: false},
	"min-height": {Initial: "0", Inherited: false},
	"max-height": {Initial: "none", Inherited: false},

	// Typography (inherited)
	"font-family":    {Initial: "serif", Inherited: true},
	"font-size":      {Initial: "12pt", Inherited: true},
	"font-weight":    {Initial: "normal", Inherited: true},
	"font-style":     {Initial: "normal", Inherited: true},
	"color":          {Initial: "#000000", Inherited: true},
	"text-align":     {Initial: "left", Inherited: true},
	"text-decoration": {Initial: "none", Inherited: false},
	"text-indent":    {Initial: "0", Inherited: true},
	"line-height":    {Initial: "1.2", Inherited: true},
	"letter-spacing": {Initial: "normal", Inherited: true},
	"word-spacing":   {Initial: "normal", Inherited: true},
	"vertical-align": {Initial: "baseline", Inherited: false},

	// Background
	"background-color": {Initial: "transparent", Inherited: false},

	// List
	"list-style-type": {Initial: "disc", Inherited: true},

	// Table (Phase 4-B)
	"border-collapse": {Initial: "separate", Inherited: true},
	"border-spacing":  {Initial: "0", Inherited: true},
	"caption-side":    {Initial: "top", Inherited: true},
	"table-layout":    {Initial: "auto", Inherited: false},
}

// IsInherited reports whether the named property inherits by default.
func IsInherited(property string) bool {
	def, ok := PropertyTable[property]
	if !ok {
		return false
	}
	return def.Inherited
}

// InitialValue returns the initial value for the named property.
func InitialValue(property string) string {
	def, ok := PropertyTable[property]
	if !ok {
		return ""
	}
	return def.Initial
}
