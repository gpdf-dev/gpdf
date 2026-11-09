package svg

import (
	"math"
	"strconv"
	"strings"
)

// unitScale converts CSS absolute units to points. One user unit (px) is
// one point, matching gpdf's 72 DPI image convention.
var unitScale = map[string]float64{
	"":   1,
	"px": 1,
	"pt": 1,
	"pc": 12,
	"in": 72,
	"cm": 72 / 2.54,
	"mm": 72 / 25.4,
	"q":  72 / 101.6,
	"em": 16, // approximate: default font size
	"ex": 8,
}

// parseLength parses a length with an optional unit. Percentages resolve
// against ref.
func parseLength(s string, ref float64) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if strings.HasSuffix(s, "%") {
		v, err := strconv.ParseFloat(strings.TrimSpace(s[:len(s)-1]), 64)
		if err != nil {
			return 0, false
		}
		return v / 100 * ref, true
	}
	i := len(s)
	for i > 0 && (s[i-1] >= 'a' && s[i-1] <= 'z' || s[i-1] >= 'A' && s[i-1] <= 'Z') {
		i--
	}
	scale, ok := unitScale[strings.ToLower(s[i:])]
	if !ok {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s[:i]), 64)
	if err != nil {
		return 0, false
	}
	return v * scale, true
}

// lengthAttr reads a length attribute, returning def when missing or invalid.
func lengthAttr(n *node, name string, ref, def float64) float64 {
	if v, ok := parseLength(n.attrs[name], ref); ok {
		return v
	}
	return def
}

// numScanner reads numbers from SVG number lists and path data, where
// separators may be whitespace, commas, or omitted entirely ("1-2.5.5").
type numScanner struct {
	s   string
	pos int
}

func (sc *numScanner) skipSep() {
	for sc.pos < len(sc.s) {
		switch sc.s[sc.pos] {
		case ' ', '\t', '\n', '\r', '\f', ',':
			sc.pos++
		default:
			return
		}
	}
}

// number reads the next number. It reports false (without consuming
// input) when no number starts at the current position.
func (sc *numScanner) number() (float64, bool) {
	sc.skipSep()
	start := sc.pos
	i := start
	if i < len(sc.s) && (sc.s[i] == '+' || sc.s[i] == '-') {
		i++
	}
	i, intDigits := sc.digits(i)
	fracDigits := 0
	if i < len(sc.s) && sc.s[i] == '.' {
		i, fracDigits = sc.digits(i + 1)
	}
	if intDigits+fracDigits == 0 {
		return 0, false
	}
	i = sc.exponent(i)
	v, err := strconv.ParseFloat(sc.s[start:i], 64)
	if err != nil {
		return 0, false
	}
	sc.pos = i
	return v, true
}

// digits skips decimal digits starting at i, returning the new index and
// the number of digits read.
func (sc *numScanner) digits(i int) (int, int) {
	start := i
	for i < len(sc.s) && sc.s[i] >= '0' && sc.s[i] <= '9' {
		i++
	}
	return i, i - start
}

// exponent consumes an exponent ("e-3") at i when one is present and
// complete, returning the new index.
func (sc *numScanner) exponent(i int) int {
	if i >= len(sc.s) || (sc.s[i] != 'e' && sc.s[i] != 'E') {
		return i
	}
	j := i + 1
	if j < len(sc.s) && (sc.s[j] == '+' || sc.s[j] == '-') {
		j++
	}
	k, n := sc.digits(j)
	if n == 0 {
		return i
	}
	return k
}

// flag reads an arc flag, which is a single '0' or '1' that need not be
// separated from the following number.
func (sc *numScanner) flag() (bool, bool) {
	sc.skipSep()
	if sc.pos < len(sc.s) && (sc.s[sc.pos] == '0' || sc.s[sc.pos] == '1') {
		v := sc.s[sc.pos] == '1'
		sc.pos++
		return v, true
	}
	return false, false
}

func parseNumberList(s string) []float64 {
	sc := numScanner{s: s}
	var out []float64
	for {
		v, ok := sc.number()
		if !ok {
			return out
		}
		out = append(out, v)
	}
}

// fmtNum formats a number for a PDF content stream: fixed-point (PDF has
// no exponent syntax) with up to prec decimals and no trailing zeros.
func fmtNum(v float64, prec int) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', prec, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "-0" || s == "" {
		return "0"
	}
	return s
}
