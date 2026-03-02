package css

import (
	"sort"
	"strings"
)

// CascadeOrigin represents the origin of a CSS declaration.
type CascadeOrigin int

const (
	OriginUA     CascadeOrigin = iota // User-Agent stylesheet
	OriginAuthor                      // Author stylesheet (<style> tags)
	OriginInline                      // Inline style attribute
)

// CascadeEntry represents a single declaration with its cascade metadata.
type CascadeEntry struct {
	Declaration Declaration
	Specificity Specificity
	Origin      CascadeOrigin
	Order       int // source order (higher = later)
}

// ComputedStyles maps property names to their resolved values for a single element.
type ComputedStyles map[string]string

// Cascade computes the final styles for a node by applying the cascade algorithm.
// It collects all matching rules, sorts by cascade order, and resolves to a property map.
//
// Parameters:
//   - node: the DOM node to compute styles for
//   - stylesheets: author stylesheets (from <style> tags), in source order
//   - inlineStyle: parsed inline style declarations (from style attribute)
//   - parentStyles: computed styles of the parent node (for inheritance)
//   - uaStylesheet: User-Agent stylesheet
func Cascade(
	node Matchable,
	stylesheets []*Stylesheet,
	inlineStyle []Declaration,
	parentStyles ComputedStyles,
	uaStylesheet *Stylesheet,
) ComputedStyles {
	var entries []CascadeEntry
	order := 0

	// 1. UA stylesheet (lowest priority)
	if uaStylesheet != nil {
		for _, rule := range uaStylesheet.Rules {
			selectors := ParseSelectorGroup(rule.Selectors)
			for _, sel := range selectors {
				if Match(sel, node) {
					for _, decl := range ExpandShorthands(rule.Declarations) {
						entries = append(entries, CascadeEntry{
							Declaration: decl,
							Specificity: sel.Spec,
							Origin:      OriginUA,
							Order:       order,
						})
						order++
					}
				}
			}
		}
	}

	// 2. Author stylesheets
	for _, ss := range stylesheets {
		for _, rule := range ss.Rules {
			selectors := ParseSelectorGroup(rule.Selectors)
			for _, sel := range selectors {
				if Match(sel, node) {
					for _, decl := range ExpandShorthands(rule.Declarations) {
						entries = append(entries, CascadeEntry{
							Declaration: decl,
							Specificity: sel.Spec,
							Origin:      OriginAuthor,
							Order:       order,
						})
						order++
					}
				}
			}
		}
	}

	// 3. Inline styles (highest specificity)
	for _, decl := range ExpandShorthands(inlineStyle) {
		entries = append(entries, CascadeEntry{
			Declaration: decl,
			Specificity: InlineSpecificity(),
			Origin:      OriginInline,
			Order:       order,
		})
		order++
	}

	// Sort by cascade order:
	// 1. !important declarations win (regardless of origin)
	// 2. Origin: inline > author > UA
	// 3. Specificity
	// 4. Source order (later wins)
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]

		// !important trumps everything
		if a.Declaration.Important != b.Declaration.Important {
			return !a.Declaration.Important // important goes later (wins)
		}

		// Origin
		if a.Origin != b.Origin {
			return a.Origin < b.Origin // higher origin goes later
		}

		// Specificity
		if a.Specificity != b.Specificity {
			return a.Specificity.Less(b.Specificity)
		}

		// Source order
		return a.Order < b.Order
	})

	// Build property map (last value wins after sorting)
	props := make(ComputedStyles)
	for _, entry := range entries {
		props[entry.Declaration.Property] = strings.TrimSpace(entry.Declaration.Value)
	}

	// Apply inheritance and initial values
	resolveInheritance(props, parentStyles)

	return props
}

// resolveInheritance handles inherit, initial, unset keywords and
// default inheritance for properties not explicitly set.
func resolveInheritance(props ComputedStyles, parentStyles ComputedStyles) {
	// First, resolve explicit keywords
	for prop, val := range props {
		lower := strings.ToLower(strings.TrimSpace(val))
		switch lower {
		case "inherit":
			if parentStyles != nil {
				if pv, ok := parentStyles[prop]; ok {
					props[prop] = pv
				} else {
					props[prop] = InitialValue(prop)
				}
			} else {
				props[prop] = InitialValue(prop)
			}
		case "initial":
			props[prop] = InitialValue(prop)
		case "unset":
			if IsInherited(prop) {
				// Inherited property: behave like inherit
				if parentStyles != nil {
					if pv, ok := parentStyles[prop]; ok {
						props[prop] = pv
					} else {
						props[prop] = InitialValue(prop)
					}
				} else {
					props[prop] = InitialValue(prop)
				}
			} else {
				// Non-inherited: behave like initial
				props[prop] = InitialValue(prop)
			}
		}
	}

	// Apply default inheritance for properties not in the map
	if parentStyles != nil {
		for prop, def := range PropertyTable {
			if _, exists := props[prop]; !exists && def.Inherited {
				if pv, ok := parentStyles[prop]; ok {
					props[prop] = pv
				}
			}
		}
	}
}

// CollectMatchingRules returns all declarations from stylesheets that match the given node.
// This is a lower-level API for when you need the raw entries before cascade resolution.
func CollectMatchingRules(node Matchable, stylesheets []*Stylesheet) []CascadeEntry {
	var entries []CascadeEntry
	order := 0
	for _, ss := range stylesheets {
		for _, rule := range ss.Rules {
			selectors := ParseSelectorGroup(rule.Selectors)
			for _, sel := range selectors {
				if Match(sel, node) {
					for _, decl := range rule.Declarations {
						entries = append(entries, CascadeEntry{
							Declaration: decl,
							Specificity: sel.Spec,
							Origin:      OriginAuthor,
							Order:       order,
						})
						order++
					}
				}
			}
		}
	}
	return entries
}
