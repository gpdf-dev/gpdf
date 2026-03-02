package css

// Specificity represents CSS specificity as a (A, B, C) tuple.
//   A = number of ID selectors
//   B = number of class selectors, attribute selectors, pseudo-classes
//   C = number of type selectors, pseudo-elements
type Specificity [3]int

// Less reports whether s has lower specificity than other.
func (s Specificity) Less(other Specificity) bool {
	if s[0] != other[0] {
		return s[0] < other[0]
	}
	if s[1] != other[1] {
		return s[1] < other[1]
	}
	return s[2] < other[2]
}

// calcSelectorSpecificity computes the specificity of a parsed selector.
func calcSelectorSpecificity(parts []CompoundSelector) Specificity {
	var spec Specificity
	for _, p := range parts {
		if p.ID != "" {
			spec[0]++
		}
		spec[1] += len(p.Classes)
		if p.Tag != "" {
			spec[2]++
		}
	}
	return spec
}

// InlineSpecificity returns the specificity for inline styles (1,0,0,0).
// We represent this by convention as (1000, 0, 0) to always win over selectors.
func InlineSpecificity() Specificity {
	return Specificity{1000, 0, 0}
}
