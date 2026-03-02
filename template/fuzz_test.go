package template

import (
	"testing"
)

// FuzzFromJSON tests that arbitrary JSON input does not panic
// during schema parsing and document building.
func FuzzFromJSON(f *testing.F) {
	// Seed corpus with valid schemas.
	f.Add([]byte(`{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":12,"text":"Hello"}]}}]}`))
	f.Add([]byte(`{"page":{"size":"Letter","margins":"15mm"},"body":[]}`))
	f.Add([]byte(`{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":6,"text":"Left"},{"span":6,"text":"Right"}]}}]}`))
	f.Add([]byte(`{"page":{"size":"A4"},"metadata":{"title":"Test","author":"A"},"body":[{"row":{"cols":[{"span":12,"elements":[{"type":"text","content":"Hello","style":{"size":24,"bold":true}}]}]}}]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":12,"table":{"header":["A","B"],"rows":[["1","2"]]}}]}}]}`))
	f.Add([]byte(`{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":12,"list":{"items":["a","b","c"]}}]}}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		doc, err := FromJSON(data, nil)
		if err != nil {
			return // parse errors are expected for random input
		}
		// If parsing succeeded, generation should not panic.
		_, _ = doc.Generate()
	})
}

// FuzzParseValue tests that arbitrary dimension strings do not panic.
func FuzzParseValue(f *testing.F) {
	f.Add("15mm")
	f.Add("12pt")
	f.Add("auto")
	f.Add("2.5cm")
	f.Add("1in")
	f.Add("1.5em")
	f.Add("50%")
	f.Add("")
	f.Add("abc")
	f.Add("99999999999mm")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = parseValue(s)
	})
}

// FuzzParseColor tests that arbitrary color strings do not panic.
func FuzzParseColor(f *testing.F) {
	f.Add("#FF0000")
	f.Add("#000000")
	f.Add("red")
	f.Add("black")
	f.Add("")
	f.Add("#ZZZZZZ")
	f.Add("unknown")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = parseColor(s)
	})
}
