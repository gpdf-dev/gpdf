package document

import (
	"fmt"
	"math"
	"testing"
)

func TestSizeLandscape(t *testing.T) {
	tests := []struct {
		name string
		in   Size
		want Size
	}{
		{"portrait A4 swaps", A4, Size{Width: A4.Height, Height: A4.Width}},
		{"portrait Letter swaps", Letter, Size{Width: Letter.Height, Height: Letter.Width}},
		{"already landscape unchanged", Size{Width: 800, Height: 600}, Size{Width: 800, Height: 600}},
		{"square unchanged", Size{Width: 500, Height: 500}, Size{Width: 500, Height: 500}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.Landscape()
			if got != tc.want {
				t.Errorf("Landscape() = %+v, want %+v", got, tc.want)
			}
			if again := got.Landscape(); again != got {
				t.Errorf("Landscape() is not idempotent: %+v -> %+v", got, again)
			}
		})
	}
}

// TestISOPageSizes checks every predefined ISO size against its nominal
// millimetre dimensions.
func TestISOPageSizes(t *testing.T) {
	mm := map[string][2]float64{
		"A0": {841, 1189}, "A1": {594, 841}, "A2": {420, 594}, "A3": {297, 420},
		"A4": {210, 297}, "A5": {148, 210}, "A6": {105, 148}, "A7": {74, 105},
		"A8": {52, 74}, "A9": {37, 52}, "A10": {26, 37},
		"B0": {1000, 1414}, "B1": {707, 1000}, "B2": {500, 707}, "B3": {353, 500},
		"B4": {250, 353}, "B5": {176, 250}, "B6": {125, 176}, "B7": {88, 125},
		"B8": {62, 88}, "B9": {44, 62}, "B10": {31, 44},
		"C0": {917, 1297}, "C1": {648, 917}, "C2": {458, 648}, "C3": {324, 458},
		"C4": {229, 324}, "C5": {162, 229}, "C6": {114, 162}, "C7": {81, 114},
		"C8": {57, 81}, "C9": {40, 57}, "C10": {28, 40},
		"DL": {110, 220},
	}
	sizes := map[string]Size{
		"A0": A0, "A1": A1, "A2": A2, "A3": A3, "A4": A4, "A5": A5,
		"A6": A6, "A7": A7, "A8": A8, "A9": A9, "A10": A10,
		"B0": B0, "B1": B1, "B2": B2, "B3": B3, "B4": B4, "B5": B5,
		"B6": B6, "B7": B7, "B8": B8, "B9": B9, "B10": B10,
		"C0": C0, "C1": C1, "C2": C2, "C3": C3, "C4": C4, "C5": C5,
		"C6": C6, "C7": C7, "C8": C8, "C9": C9, "C10": C10,
		"DL": DL,
	}
	if len(sizes) != len(mm) {
		t.Fatalf("size table mismatch: %d sizes, %d expectations", len(sizes), len(mm))
	}
	for name, dims := range mm {
		t.Run(name, func(t *testing.T) {
			got := sizes[name]
			want := Size{Width: dims[0] * 72 / 25.4, Height: dims[1] * 72 / 25.4}
			if math.Abs(got.Width-want.Width) > 0.01 || math.Abs(got.Height-want.Height) > 0.01 {
				t.Errorf("%s = %+v, want %s", name, got, fmt.Sprintf("%.2f x %.2f", want.Width, want.Height))
			}
			if got.Width >= got.Height {
				t.Errorf("%s is not portrait: %+v", name, got)
			}
		})
	}
}
