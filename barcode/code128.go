package barcode

import "fmt"

// Code 128 special code values.
const (
	code128SwitchC = 99
	code128SwitchB = 100
	code128SwitchA = 101
	code128FNC1    = 102
	code128StartA  = 103
	code128StartB  = 104
	code128StartC  = 105
	code128Stop    = 106
)

// code128Patterns contains the bar/space width patterns for all 107 Code 128
// symbols (indices 0-105) plus the stop pattern (index 106). Each symbol has
// 6 alternating bar/space widths summing to 11 modules. The stop pattern has
// 7 elements summing to 13 modules.
var code128Patterns = [107][7]int{
	{2, 1, 2, 2, 2, 2, 0}, // 0
	{2, 2, 2, 1, 2, 2, 0}, // 1
	{2, 2, 2, 2, 2, 1, 0}, // 2
	{1, 2, 1, 2, 2, 3, 0}, // 3
	{1, 2, 1, 3, 2, 2, 0}, // 4
	{1, 3, 1, 2, 2, 2, 0}, // 5
	{1, 2, 2, 2, 1, 3, 0}, // 6
	{1, 2, 2, 3, 1, 2, 0}, // 7
	{1, 3, 2, 2, 1, 2, 0}, // 8
	{2, 2, 1, 2, 1, 3, 0}, // 9
	{2, 2, 1, 3, 1, 2, 0}, // 10
	{2, 3, 1, 2, 1, 2, 0}, // 11
	{1, 1, 2, 2, 3, 2, 0}, // 12
	{1, 2, 2, 1, 3, 2, 0}, // 13
	{1, 2, 2, 2, 3, 1, 0}, // 14
	{1, 1, 3, 2, 2, 2, 0}, // 15
	{1, 2, 3, 1, 2, 2, 0}, // 16
	{1, 2, 3, 2, 2, 1, 0}, // 17
	{2, 2, 3, 2, 1, 1, 0}, // 18
	{2, 2, 1, 1, 3, 2, 0}, // 19
	{2, 2, 1, 2, 3, 1, 0}, // 20
	{2, 1, 3, 2, 1, 2, 0}, // 21
	{2, 2, 3, 1, 1, 2, 0}, // 22
	{3, 1, 2, 1, 3, 1, 0}, // 23
	{3, 1, 1, 2, 2, 2, 0}, // 24
	{3, 2, 1, 1, 2, 2, 0}, // 25
	{3, 2, 1, 2, 2, 1, 0}, // 26
	{3, 1, 2, 2, 1, 2, 0}, // 27
	{3, 2, 2, 1, 1, 2, 0}, // 28
	{3, 2, 2, 2, 1, 1, 0}, // 29
	{2, 1, 2, 1, 2, 3, 0}, // 30
	{2, 1, 2, 3, 2, 1, 0}, // 31
	{2, 3, 2, 1, 2, 1, 0}, // 32
	{1, 1, 1, 3, 2, 3, 0}, // 33
	{1, 3, 1, 1, 2, 3, 0}, // 34
	{1, 3, 1, 3, 2, 1, 0}, // 35
	{1, 1, 2, 3, 1, 3, 0}, // 36
	{1, 3, 2, 1, 1, 3, 0}, // 37
	{1, 3, 2, 3, 1, 1, 0}, // 38
	{2, 1, 1, 3, 1, 3, 0}, // 39
	{2, 3, 1, 1, 1, 3, 0}, // 40
	{2, 3, 1, 3, 1, 1, 0}, // 41
	{1, 1, 2, 1, 3, 3, 0}, // 42
	{1, 1, 2, 3, 3, 1, 0}, // 43
	{1, 3, 2, 1, 3, 1, 0}, // 44
	{1, 1, 3, 1, 2, 3, 0}, // 45
	{1, 1, 3, 3, 2, 1, 0}, // 46
	{1, 3, 3, 1, 2, 1, 0}, // 47
	{3, 1, 3, 1, 2, 1, 0}, // 48
	{2, 1, 1, 3, 3, 1, 0}, // 49
	{2, 3, 1, 1, 3, 1, 0}, // 50
	{2, 1, 3, 1, 1, 3, 0}, // 51
	{2, 1, 3, 3, 1, 1, 0}, // 52
	{2, 1, 3, 1, 3, 1, 0}, // 53
	{3, 1, 1, 1, 2, 3, 0}, // 54
	{3, 1, 1, 3, 2, 1, 0}, // 55
	{3, 3, 1, 1, 2, 1, 0}, // 56
	{3, 1, 2, 1, 1, 3, 0}, // 57
	{3, 1, 2, 3, 1, 1, 0}, // 58
	{3, 3, 2, 1, 1, 1, 0}, // 59
	{3, 1, 4, 1, 1, 1, 0}, // 60
	{2, 2, 1, 4, 1, 1, 0}, // 61
	{4, 3, 1, 1, 1, 1, 0}, // 62
	{1, 1, 1, 2, 2, 4, 0}, // 63
	{1, 1, 1, 4, 2, 2, 0}, // 64
	{1, 2, 1, 1, 2, 4, 0}, // 65
	{1, 2, 1, 4, 2, 1, 0}, // 66
	{1, 4, 1, 1, 2, 2, 0}, // 67
	{1, 4, 1, 2, 2, 1, 0}, // 68
	{1, 1, 2, 2, 1, 4, 0}, // 69
	{1, 1, 2, 4, 1, 2, 0}, // 70
	{1, 2, 2, 1, 1, 4, 0}, // 71
	{1, 2, 2, 4, 1, 1, 0}, // 72
	{1, 4, 2, 1, 1, 2, 0}, // 73
	{1, 4, 2, 2, 1, 1, 0}, // 74
	{2, 4, 1, 2, 1, 1, 0}, // 75
	{2, 2, 1, 1, 1, 4, 0}, // 76
	{4, 1, 3, 1, 1, 1, 0}, // 77
	{2, 4, 1, 1, 1, 2, 0}, // 78
	{1, 3, 4, 1, 1, 1, 0}, // 79
	{1, 1, 1, 2, 4, 2, 0}, // 80
	{1, 2, 1, 1, 4, 2, 0}, // 81
	{1, 2, 1, 2, 4, 1, 0}, // 82
	{1, 1, 4, 2, 1, 2, 0}, // 83
	{1, 2, 4, 1, 1, 2, 0}, // 84
	{1, 2, 4, 2, 1, 1, 0}, // 85
	{4, 1, 1, 2, 1, 2, 0}, // 86
	{4, 2, 1, 1, 1, 2, 0}, // 87
	{4, 2, 1, 2, 1, 1, 0}, // 88
	{2, 1, 2, 1, 4, 1, 0}, // 89
	{2, 1, 4, 1, 2, 1, 0}, // 90
	{4, 1, 2, 1, 2, 1, 0}, // 91
	{1, 1, 1, 1, 4, 3, 0}, // 92
	{1, 1, 1, 3, 4, 1, 0}, // 93
	{1, 3, 1, 1, 4, 1, 0}, // 94
	{1, 1, 4, 1, 1, 3, 0}, // 95
	{1, 1, 4, 3, 1, 1, 0}, // 96
	{4, 1, 1, 1, 1, 3, 0}, // 97
	{4, 1, 1, 3, 1, 1, 0}, // 98
	{1, 1, 3, 1, 4, 1, 0}, // 99  (CodeC switch)
	{1, 1, 4, 1, 3, 1, 0}, // 100 (CodeB switch)
	{3, 1, 1, 1, 4, 1, 0}, // 101 (CodeA switch)
	{4, 1, 1, 1, 3, 1, 0}, // 102 (FNC1)
	{2, 1, 1, 4, 1, 2, 0}, // 103 (StartA)
	{2, 1, 1, 2, 1, 4, 0}, // 104 (StartB)
	{2, 1, 1, 2, 3, 2, 0}, // 105 (StartC)
	{2, 3, 3, 1, 1, 1, 2}, // 106 (Stop - 7 elements, 13 units)
}

// encodeCode128 encodes data as a Code 128 barcode and returns the symbol
// values (including start code, data symbols, checksum, and stop code).
func encodeCode128(data string) ([]int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("barcode: empty data")
	}

	// Validate that all characters are encodable (ASCII 0-127).
	for i := 0; i < len(data); i++ {
		if data[i] > 127 {
			return nil, fmt.Errorf("barcode: character at position %d (0x%02x) is not encodable in Code 128", i, data[i])
		}
	}

	symbols := make([]int, 0, len(data)+4)
	pos := 0

	// Determine start code.
	startCode, currentSet := chooseStartCode(data)
	symbols = append(symbols, startCode)

	for pos < len(data) {
		switch currentSet {
		case 'C':
			pos, symbols, currentSet = encodeSetC(data, pos, symbols)
		case 'B':
			pos, symbols, currentSet = encodeSetB(data, pos, symbols)
		case 'A':
			pos, symbols, currentSet = encodeSetA(data, pos, symbols)
		}
	}

	// Calculate checksum: (startCode + sum(position * value)) mod 103
	checksum := symbols[0] // Start code value
	for i := 1; i < len(symbols); i++ {
		checksum += i * symbols[i]
	}
	checksum %= 103

	symbols = append(symbols, checksum)
	symbols = append(symbols, code128Stop)

	return symbols, nil
}

// encodeSetC processes Code Set C encoding at the current position.
func encodeSetC(data string, pos int, symbols []int) (int, []int, byte) {
	digitRun := countDigits(data, pos)
	if digitRun < 2 {
		// Less than 2 digits remaining in Code C, switch to B.
		symbols = append(symbols, code128SwitchB)
		return pos, symbols, 'B'
	}
	// Encode pairs of digits.
	pairs := digitRun / 2
	for i := 0; i < pairs; i++ {
		d1 := int(data[pos] - '0')
		d2 := int(data[pos+1] - '0')
		symbols = append(symbols, d1*10+d2)
		pos += 2
	}
	// If we have remaining non-digit chars or odd digit, switch to B.
	if pos < len(data) {
		symbols = append(symbols, code128SwitchB)
		return pos, symbols, 'B'
	}
	return pos, symbols, 'C'
}

// encodeSetB processes Code Set B encoding at the current position.
func encodeSetB(data string, pos int, symbols []int) (int, []int, byte) {
	// Check if we should switch to Code C for a run of digits.
	if shouldSwitchToC(data, pos) {
		symbols = append(symbols, code128SwitchC)
		return pos, symbols, 'C'
	}

	ch := data[pos]
	if ch < 32 {
		// Control character: switch to Code A for this character.
		symbols = append(symbols, code128SwitchA)
		symbols = append(symbols, int(ch))
		pos++
		// Switch back to B.
		symbols = append(symbols, code128SwitchB)
	} else {
		// Normal Code B character.
		symbols = append(symbols, int(ch)-32)
		pos++
	}
	return pos, symbols, 'B'
}

// encodeSetA processes Code Set A encoding at the current position.
func encodeSetA(data string, pos int, symbols []int) (int, []int, byte) {
	// Check if we should switch to Code C for a run of digits.
	if shouldSwitchToC(data, pos) {
		symbols = append(symbols, code128SwitchC)
		return pos, symbols, 'C'
	}

	ch := data[pos]
	if ch >= 32 && ch <= 95 {
		// Printable ASCII in Code A range.
		symbols = append(symbols, int(ch)-32)
		pos++
	} else if ch < 32 {
		// Control characters.
		symbols = append(symbols, int(ch)+64)
		pos++
	} else {
		// Character > 95, switch to Code B.
		symbols = append(symbols, code128SwitchB)
		return pos, symbols, 'B'
	}
	return pos, symbols, 'A'
}

// shouldSwitchToC returns true if the current position has a digit run
// of 4+ that warrants switching to Code C.
func shouldSwitchToC(data string, pos int) bool {
	digitRun := countDigits(data, pos)
	if digitRun < 4 {
		return false
	}
	return digitRun%2 == 0 || pos+digitRun == len(data)
}

// chooseStartCode determines the optimal starting code set and returns the
// start symbol value and the code set character ('A', 'B', or 'C').
func chooseStartCode(data string) (int, byte) {
	// Start with Code C if first 4+ characters are digits.
	if countDigits(data, 0) >= 4 {
		return code128StartC, 'C'
	}
	// Use Code A if first character is a control character.
	if len(data) > 0 && data[0] < 32 {
		return code128StartA, 'A'
	}
	// Default to Code B.
	return code128StartB, 'B'
}

// countDigits returns the number of consecutive ASCII digit characters
// starting at position pos in data.
func countDigits(data string, pos int) int {
	count := 0
	for i := pos; i < len(data); i++ {
		if data[i] >= '0' && data[i] <= '9' {
			count++
		} else {
			break
		}
	}
	return count
}

// code128ToPattern expands the symbol values into a boolean bar pattern.
// true represents a bar (black), false represents a space (white).
func code128ToPattern(symbols []int) []bool {
	// Calculate total modules.
	total := 0
	for _, sym := range symbols {
		pat := code128Patterns[sym]
		if sym == code128Stop {
			for _, w := range pat {
				total += w
			}
		} else {
			for i := 0; i < 6; i++ {
				total += pat[i]
			}
		}
	}

	pattern := make([]bool, 0, total)
	for _, sym := range symbols {
		pat := code128Patterns[sym]
		elems := 6
		if sym == code128Stop {
			elems = 7
		}
		for i := 0; i < elems; i++ {
			bar := (i % 2) == 0 // even indices are bars, odd are spaces
			for j := 0; j < pat[i]; j++ {
				pattern = append(pattern, bar)
			}
		}
	}

	return pattern
}
